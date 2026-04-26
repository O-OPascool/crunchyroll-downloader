package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	widevine "github.com/iyear/gowidevine"
	"github.com/unki2aut/go-mpd"
)

const maxWorkers = 16

// audioTrack represents a downloaded audio track with its language
type audioTrack struct {
	file string
	lang string
}

// subTrack represents a downloaded subtitle track with its language
type subTrack struct {
	file string
	lang string
}

func buildUrl(base, representationId, file string, partNum *int64) string {
	if partNum != nil {
		file = strings.ReplaceAll(file, "$Number$", fmt.Sprintf("%05d", *partNum))
		file = strings.ReplaceAll(file, "$Number%05d$", fmt.Sprintf("%05d", *partNum))
	}
	return base + strings.ReplaceAll(file, "$RepresentationID$", representationId)
}

func downloadPart(url string) ([]byte, error) {
	maxRetries := 5
	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Origin", "https://static.crunchyroll.com")
		req.Header.Set("Referer", "https://static.crunchyroll.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:147.0) Gecko/20100101 Firefox/147.0")
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != 200 {
			if resp != nil {
				resp.Body.Close()
			}
			if attempt < maxRetries-1 {
				continue
			}
			return nil, fmt.Errorf("failed after %d retries", maxRetries)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			if attempt < maxRetries-1 {
				continue
			}
			return nil, fmt.Errorf("body read failed after %d retries: %w", maxRetries, err)
		}
		return body, nil
	}
	return nil, fmt.Errorf("failed after %d retries", maxRetries)
}

func getTempFilename(set *mpd.AdaptationSet) string {
	if set == nil {
		f, _ := os.CreateTemp("", "crdl-subs-*.ass")
		return f.Name()
	}
	for _, representation := range set.Representations {
		if representation.Height != nil {
			f, _ := os.CreateTemp("", "crdl-video-*.mp4")
			return f.Name()
		} else if representation.Bandwidth != nil {
			f, _ := os.CreateTemp("", "crdl-audio-*.mp3")
			return f.Name()
		}
	}
	return ""
}

type segmentJob struct {
	index int
	url   string
}

// downloadWhole handles SegmentBase-style MPDs where the content is a single file at the BaseURL.
func downloadWhole(baseUrl string, set *mpd.AdaptationSet, label string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, baseUrl, nil)
	if err != nil {
		return "", fmt.Errorf("downloadWhole: %w", err)
	}
	req.Header.Set("Origin", "https://static.crunchyroll.com")
	req.Header.Set("Referer", "https://static.crunchyroll.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:147.0) Gecko/20100101 Firefox/147.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("downloadWhole: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("downloadWhole: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("downloadWhole read: %w", err)
	}
	fmt.Printf("  %s    downloaded %.1f MB\n", label, float64(len(data))/1024/1024)

	filename := getTempFilename(set)
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err = widevine.DecryptMP4Auto(io.NopCloser(bytes.NewReader(data)), keys, file); err != nil {
		return "", fmt.Errorf("widevine.DecryptMP4Auto: %w", err)
	}

	return filename, nil
}

// downloadParts downloads all segments with resume support, progress bar and MB/s display.
func downloadParts(baseUrl, representationId *string, set *mpd.AdaptationSet, cacheDir string, label string) (string, error) {
	os.MkdirAll(cacheDir, 0755)

	// Resolve SegmentTemplate: check AdaptationSet level first, then Representation level
	segTemplate := set.SegmentTemplate
	if segTemplate == nil {
		for _, rep := range set.Representations {
			if rep.ID != nil && *rep.ID == *representationId && rep.SegmentTemplate != nil {
				segTemplate = rep.SegmentTemplate
				break
			}
		}
	}
	if segTemplate == nil {
		// SegmentBase / single-file MPD: download the whole file from the BaseURL
		return downloadWhole(*baseUrl, set, label)
	}

	// Init segment (cached)
	initFile := filepath.Join(cacheDir, "init.bin")
	var initData []byte
	if data, err := os.ReadFile(initFile); err == nil {
		initData = data
	} else {
		initUrl := buildUrl(*baseUrl, *representationId, *segTemplate.Initialization, nil)
		var err error
		initData, err = downloadPart(initUrl)
		if err != nil {
			return "", fmt.Errorf("init segment: %w", err)
		}
		os.WriteFile(initFile, initData, 0644)
	}

	timeline := expandTimeline(segTemplate.SegmentTimeline.S, 1)
	total := len(timeline)
	results := make([][]byte, total)

	// Check already cached segments
	pending := make([]segmentJob, 0, total)
	alreadyDone := int64(0)
	for i, item := range timeline {
		segFile := filepath.Join(cacheDir, fmt.Sprintf("seg_%05d.bin", i))
		if data, err := os.ReadFile(segFile); err == nil && len(data) > 0 {
			results[i] = data
			alreadyDone++
		} else {
			url := buildUrl(*baseUrl, *representationId, *segTemplate.Media, &item)
			pending = append(pending, segmentJob{index: i, url: url})
		}
	}

	pb := NewProgressBar(int64(total), label)
	pb.SetDone(alreadyDone)
	pb.render()

	if len(pending) > 0 {
		jobs := make(chan segmentJob, len(pending))
		var mu sync.Mutex
		var downloadErr error
		var wg sync.WaitGroup

		for w := 0; w < maxWorkers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for job := range jobs {
					data, err := downloadPart(job.url)
					if err != nil {
						mu.Lock()
						if downloadErr == nil {
							downloadErr = err
						}
						mu.Unlock()
						return
					}
					segFile := filepath.Join(cacheDir, fmt.Sprintf("seg_%05d.bin", job.index))
					os.WriteFile(segFile, data, 0644)

					mu.Lock()
					results[job.index] = data
					mu.Unlock()

					pb.Add(int64(len(data)))
				}
			}()
		}

		for _, job := range pending {
			jobs <- job
		}
		close(jobs)
		wg.Wait()

		if downloadErr != nil {
			return "", downloadErr
		}
	}

	pb.Done()

	// Concatenate init + all segments then decrypt
	var parts []byte
	parts = append(parts, initData...)
	for _, data := range results {
		parts = append(parts, data...)
	}

	filename := getTempFilename(set)
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err = widevine.DecryptMP4Auto(io.NopCloser(bytes.NewReader(parts)), keys, file); err != nil {
		return "", fmt.Errorf("widevine.DecryptMP4Auto: %w", err)
	}

	os.RemoveAll(cacheDir)
	return filename, nil
}

func downloadSubs(url string) string {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Origin", "https://static.crunchyroll.com")
	req.Header.Set("Referer", "https://static.crunchyroll.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:147.0) Gecko/20100101 Firefox/147.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	filename := getTempFilename(nil)
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	file.Write(body)
	file.Close()

	// Apply nice styling
	if err := styleSubtitles(filename); err != nil {
		fmt.Printf("  Warning: could not restyle subtitles: %v\n", err)
	}

	return filename
}

// resolveVersionGUID finds the correct episode GUID for a given audio locale
func resolveVersionGUID(info EpisodeInfo, audioLang string) (string, error) {
	if info.EpisodeMetadata.AudioLocale == audioLang {
		// Current version already matches
		return "", nil // empty means "use the current contentId"
	}
	idx := slices.IndexFunc(info.EpisodeMetadata.Versions, func(v *DubVersion) bool {
		return v.AudioLocale == audioLang
	})
	if idx == -1 {
		return "", fmt.Errorf("audio locale %q not available for this episode", audioLang)
	}
	return info.EpisodeMetadata.Versions[idx].GUID, nil
}

func downloadEpisode(contentId string, videoQuality, audioQuality *string, audioLangs, subsLangs []string, info EpisodeInfo) error {
	primaryAudioLang := audioLangs[0]

	dirName, outputFile := buildOutputPath(info, *videoQuality, audioLangs)

	if _, err := os.Stat(dirName); err != nil {
		_ = os.MkdirAll(dirName, 0777)
	}

	if _, err := os.Stat(outputFile); err == nil {
		fmt.Printf("⏭  Already downloaded: S%02dE%02d, skipping...\n",
			info.EpisodeMetadata.SeasonNumber, info.EpisodeMetadata.EpisodeNumber)
		return nil
	}

	// --- Resolve primary audio version ---
	primaryContentId := contentId
	guid, err := resolveVersionGUID(info, primaryAudioLang)
	if err != nil {
		return fmt.Errorf("primary audio: %w", err)
	}
	if guid != "" {
		primaryContentId = guid
	}

	episode := getEpisode(primaryContentId)
	fmt.Printf("⬇  S%02dE%02d - %s (%s)\n",
		info.EpisodeMetadata.SeasonNumber,
		info.EpisodeMetadata.EpisodeNumber,
		info.Title,
		info.EpisodeMetadata.SeriesTitle,
	)

	manifest, err := parseManifest(episode.ManifestURL)
	if err != nil {
		return fmt.Errorf("manifest: %w", err)
	}
	if len(manifest.Period) == 0 {
		return fmt.Errorf("manifest has no periods (empty or invalid MPD)")
	}
	pssh := getPssh(manifest)
	if pssh == nil {
		return fmt.Errorf("PSSH not found in manifest")
	}
	videoSet := findAdaptationSet(manifest.Period[0], "video")
	audioSet := findAdaptationSet(manifest.Period[0], "audio")
	if videoSet == nil {
		return fmt.Errorf("failed to find video adaptation set in manifest")
	}
	if audioSet == nil {
		fmt.Printf("  Failed to find audio adaptation set in manifest (%d adaptation sets available)\n", len(manifest.Period[0].AdaptationSets))
		for i, set := range manifest.Period[0].AdaptationSets {
			fmt.Printf("    [%d] MimeType=%q ContentType=%v Representations=%d\n", i, set.MimeType, set.ContentType, len(set.Representations))
		}
		return fmt.Errorf("no audio adaptation set found")
	}

	err = getLicense(*pssh, primaryContentId, episode.Token)
	if err != nil {
		return fmt.Errorf("license error: %w", err)
	}

	cacheBase := fmt.Sprintf(".crdl_cache/%s", primaryContentId)

	// --- Download subtitles (multi) ---
	var subTracks []subTrack
	if len(subsLangs) == 1 && strings.ToLower(subsLangs[0]) == "all" {
		// Download all available subtitles
		for lang, sub := range episode.Subtitles {
			if sub != nil {
				langName := languageNames[lang]
				if langName == "" {
					langName = lang
				}
				fmt.Printf("  Subtitles (%s)...\n", langName)
				subsFile := downloadSubs(sub.URL)
				subTracks = append(subTracks, subTrack{file: subsFile, lang: lang})
			}
		}
	} else {
		for _, lang := range subsLangs {
			sub := episode.Subtitles[lang]
			if sub != nil {
				langName := languageNames[lang]
				if langName == "" {
					langName = lang
				}
				fmt.Printf("  Subtitles (%s)...\n", langName)
				subsFile := downloadSubs(sub.URL)
				subTracks = append(subTracks, subTrack{file: subsFile, lang: lang})
			} else {
				fmt.Printf("  ⚠ Subtitles for %q not available, skipping...\n", lang)
			}
		}
	}

	// --- Download video (from primary manifest) ---
	baseUrl, representationId := getBaseUrl(videoSet, true, *videoQuality)
	if baseUrl == nil {
		return fmt.Errorf("failed to get video URL, check --video-quality")
	}
	videoFile, err := downloadParts(baseUrl, representationId, videoSet, cacheBase+"/video", "Vidéo")
	if err != nil {
		return fmt.Errorf("video download: %w", err)
	}

	// --- Download audio tracks (multi) ---
	var audioTracks []audioTrack

	// Primary audio track (from the already-parsed manifest)
	audioBaseUrl, audioRepresentationId := getBaseUrl(audioSet, false, *audioQuality)
	if audioBaseUrl == nil {
		return fmt.Errorf("failed to get audio URL for %s, check --audio-quality", primaryAudioLang)
	}
	langName := languageNames[primaryAudioLang]
	if langName == "" {
		langName = primaryAudioLang
	}
	audioFile, err := downloadParts(audioBaseUrl, audioRepresentationId, audioSet, cacheBase+"/audio_"+primaryAudioLang, fmt.Sprintf("Audio (%s)", langName))
	if err != nil {
		return fmt.Errorf("audio download (%s): %w", primaryAudioLang, err)
	}
	audioTracks = append(audioTracks, audioTrack{file: audioFile, lang: primaryAudioLang})

	// Additional audio tracks from other versions
	for _, lang := range audioLangs[1:] {
		guid, err := resolveVersionGUID(info, lang)
		if err != nil {
			fmt.Printf("  ⚠ Audio %q not available, skipping: %v\n", lang, err)
			continue
		}
		versionContentId := primaryContentId
		if guid != "" {
			versionContentId = guid
		}

		versionEpisode := getEpisode(versionContentId)
		versionManifest, err := parseManifest(versionEpisode.ManifestURL)
		if err != nil {
			fmt.Printf("  ⚠ Could not parse manifest for audio %q: %v\n", lang, err)
			continue
		}
		if len(versionManifest.Period) == 0 {
			fmt.Printf("  ⚠ Empty manifest for audio %q, skipping\n", lang)
			continue
		}

		versionPssh := getPssh(versionManifest)
		if versionPssh != nil {
			if err := getLicense(*versionPssh, versionContentId, versionEpisode.Token); err != nil {
				fmt.Printf("  ⚠ License error for audio %q: %v\n", lang, err)
				continue
			}
		}

		versionAudioSet := findAdaptationSet(versionManifest.Period[0], "audio")
		if versionAudioSet == nil {
			fmt.Printf("  ⚠ No audio adaptation set for %q, skipping\n", lang)
			continue
		}

		versionBaseUrl, versionRepId := getBaseUrl(versionAudioSet, false, *audioQuality)
		if versionBaseUrl == nil {
			fmt.Printf("  ⚠ Audio quality not found for %q, skipping\n", lang)
			continue
		}

		langName := languageNames[lang]
		if langName == "" {
			langName = lang
		}
		versionAudioFile, err := downloadParts(versionBaseUrl, versionRepId, versionAudioSet,
			fmt.Sprintf(".crdl_cache/%s/audio_%s", versionContentId, lang),
			fmt.Sprintf("Audio (%s)", langName))
		if err != nil {
			fmt.Printf("  ⚠ Audio download failed for %q: %v\n", lang, err)
			continue
		}

		if success := deleteStream(versionContentId, versionEpisode.Token); !success {
			fmt.Println("  Warning: failed to delete stream token for", lang)
		}

		audioTracks = append(audioTracks, audioTrack{file: versionAudioFile, lang: lang})
	}

	if success := deleteStream(primaryContentId, episode.Token); !success {
		fmt.Println("Warning: failed to delete stream token")
	}

	mergeEverything(videoFile, audioTracks, subTracks, outputFile, info)
	return nil
}

func downloadSeason(videoQuality, audioQuality *string, audioLangs, subsLangs []string, episodes []SeasonEpisode) {
	if len(episodes) == 0 {
		fmt.Println("⚠  No episodes found in this season, skipping...")
		return
	}
	fmt.Printf("📦 Season %d of %s — %d episodes\n\n",
		episodes[0].SeasonNumber, episodes[0].SeriesTitle, len(episodes))

	for _, episode := range episodes {
		info := EpisodeInfo{
			EpisodeMetadata: EpisodeMetadata{
				SeriesTitle:        episode.SeriesTitle,
				SeasonNumber:       episode.SeasonNumber,
				EpisodeNumber:      episode.EpisodeNumber,
				AudioLocale:        episode.AudioLocale,
				Versions:           episode.Versions,
				AvailabilityStarts: episode.AvailabilityStarts,
			},
			Title: episode.Title,
		}
		if err := downloadEpisode(episode.ID, videoQuality, audioQuality, audioLangs, subsLangs, info); err != nil {
			fmt.Printf("⚠  Error on S%02dE%02d: %s, skipping...\n\n",
				episode.SeasonNumber, episode.EpisodeNumber, err)
		}
	}
}
