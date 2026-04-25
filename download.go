package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	widevine "github.com/iyear/gowidevine"
	"github.com/unki2aut/go-mpd"
)

const maxWorkers = 16

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

// downloadParts downloads all segments with resume support, progress bar and MB/s display.
func downloadParts(baseUrl, representationId *string, set *mpd.AdaptationSet, cacheDir string, label string) (string, error) {
	os.MkdirAll(cacheDir, 0755)

	// Init segment (cached)
	initFile := filepath.Join(cacheDir, "init.bin")
	var initData []byte
	if data, err := os.ReadFile(initFile); err == nil {
		initData = data
	} else {
		initUrl := buildUrl(*baseUrl, *representationId, *set.SegmentTemplate.Initialization, nil)
		var err error
		initData, err = downloadPart(initUrl)
		if err != nil {
			return "", fmt.Errorf("init segment: %w", err)
		}
		os.WriteFile(initFile, initData, 0644)
	}

	timeline := expandTimeline(set.SegmentTemplate.SegmentTimeline.S, 1)
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
			url := buildUrl(*baseUrl, *representationId, *set.SegmentTemplate.Media, &item)
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

func downloadEpisode(contentId string, videoQuality, audioQuality, subtitlesLang *string, info EpisodeInfo) {
	dirName, outputFile := buildOutputPath(info, *videoQuality, *audioLang)

	if _, err := os.Stat(dirName); err != nil {
		_ = os.MkdirAll(dirName, 0777)
	}

	if _, err := os.Stat(outputFile); err == nil {
		fmt.Printf("⏭  Already downloaded: S%02dE%02d, skipping...\n",
			info.EpisodeMetadata.SeasonNumber, info.EpisodeMetadata.EpisodeNumber)
		return
	}

	episode := getEpisode(contentId)
	fmt.Printf("⬇  S%02dE%02d - %s (%s)\n",
		info.EpisodeMetadata.SeasonNumber,
		info.EpisodeMetadata.EpisodeNumber,
		info.Title,
		info.EpisodeMetadata.SeriesTitle,
	)

	manifest := parseManifest(episode.ManifestURL)
	pssh := getPssh(manifest)
	if pssh == nil {
		panic("PSSH not found")
	}
	videoSet := manifest.Period[0].AdaptationSets[0]
	audioSet := manifest.Period[0].AdaptationSets[1]

	err := getLicense(*pssh, contentId, episode.Token)
	if err != nil {
		fmt.Printf("License error: %s\n", err)
		os.Exit(1)
	}

	cacheBase := fmt.Sprintf(".crdl_cache/%s", contentId)

	subtitles := episode.Subtitles[*subtitlesLang]
	var subsFile string
	if subtitles != nil {
		fmt.Printf("  Subtitles (%s)...\n", languageNames[*subtitlesLang])
		subsFile = downloadSubs(subtitles.URL)
	}

	baseUrl, representationId := getBaseUrl(videoSet, true, *videoQuality)
	if baseUrl == nil {
		fmt.Println("Failed to get video URL, check --video-quality")
		os.Exit(1)
	}
	videoFile, err := downloadParts(baseUrl, representationId, videoSet, cacheBase+"/video", "Vidéo")
	if err != nil {
		panic(err)
	}

	audioBaseUrl, audioRepresentationId := getBaseUrl(audioSet, false, *audioQuality)
	if audioBaseUrl == nil {
		fmt.Println("Failed to get audio URL, check --audio-quality")
		os.Exit(1)
	}
	audioFile, err := downloadParts(audioBaseUrl, audioRepresentationId, audioSet, cacheBase+"/audio", "Audio")
	if err != nil {
		panic(err)
	}

	if success := deleteStream(contentId, episode.Token); !success {
		fmt.Println("Warning: failed to delete stream token")
	}

	mergeEverything(videoFile, audioFile, subsFile, outputFile, subtitlesLang, info)
}

func downloadSeason(videoQuality, audioQuality, subtitlesLang *string, episodes []SeasonEpisode) {
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
		downloadEpisode(episode.ID, videoQuality, audioQuality, subtitlesLang, info)
	}
}
