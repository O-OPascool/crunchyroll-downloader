package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// sanitizeForFS removes characters illegal in filenames
func sanitizeForFS(s string) string {
	illegal := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	res := s
	for _, char := range illegal {
		res = strings.ReplaceAll(res, char, "")
	}
	return strings.TrimRight(res, " .")
}

// getLangTag returns VOSTFR, VF or Multi depending on audio tracks
func getLangTag(audioLangs []string) string {
	if len(audioLangs) > 1 {
		return "Multi"
	}
	if len(audioLangs) == 1 && audioLangs[0] == "fr-FR" {
		return "VF"
	}
	return "VOSTFR"
}

// buildOutputPath returns (directory, full output path) for an episode
// Format: Title.S01E01.CR.WEBDL.Multi.1080p.x265-tag.mkv
func buildOutputPath(info EpisodeInfo, videoQuality string, audioLangs []string) (string, string) {
	dirName := sanitizeForFS(info.EpisodeMetadata.SeriesTitle)
	titleDots := strings.ReplaceAll(dirName, " ", ".")
	season := fmt.Sprintf("S%02d", info.EpisodeMetadata.SeasonNumber)
	episode := fmt.Sprintf("E%02d", info.EpisodeMetadata.EpisodeNumber)
	langTag := getLangTag(audioLangs)

	filename := fmt.Sprintf("%s.%s%s.CR.WEBDL.%s.%s.x265-%s.mkv",
		titleDots, season, episode, langTag, videoQuality, *releaseTag)

	return dirName, fmt.Sprintf("%s/%s", dirName, filename)
}

// mergeEverything merges video, multiple audio tracks and multiple subtitle tracks in a single MKV container
func mergeEverything(videoFile string, audioTracks []audioTrack, subTracks []subTrack, outputFile string, info EpisodeInfo) {
	// Build ffmpeg input arguments
	args := []string{"-y", "-i", videoFile}

	// Add all audio inputs
	for _, a := range audioTracks {
		args = append(args, "-i", a.file)
	}
	// Add all subtitle inputs
	for _, s := range subTracks {
		args = append(args, "-i", s.file)
	}

	// Map video stream (input 0)
	args = append(args, "-map", "0:v")

	// Map all audio streams (inputs 1..len(audioTracks))
	for i := range audioTracks {
		args = append(args, "-map", fmt.Sprintf("%d:a", i+1))
	}

	// Map all subtitle streams
	subInputOffset := 1 + len(audioTracks)
	for i := range subTracks {
		args = append(args, "-map", fmt.Sprintf("%d:0", subInputOffset+i))
	}

	// Codec settings
	args = append(args, "-c:v", "copy", "-c:a", "copy")
	if len(subTracks) > 0 {
		args = append(args, "-c:s", "ass")
	}

	// Audio track metadata
	for i, a := range audioTracks {
		isoCode := getISOCode(a.lang)
		langName := languageNames[a.lang]
		if langName == "" {
			langName = a.lang
		}
		args = append(args,
			fmt.Sprintf("-metadata:s:a:%d", i), fmt.Sprintf("language=%s", isoCode),
			fmt.Sprintf("-metadata:s:a:%d", i), fmt.Sprintf("title=%s", langName),
		)
		if i == 0 {
			args = append(args, fmt.Sprintf("-disposition:a:%d", i), "default")
		} else {
			args = append(args, fmt.Sprintf("-disposition:a:%d", i), "0")
		}
	}

	// Subtitle track metadata
	for i, s := range subTracks {
		isoCode := getISOCode(s.lang)
		langName := languageNames[s.lang]
		if langName == "" {
			langName = s.lang
		}
		args = append(args,
			fmt.Sprintf("-metadata:s:s:%d", i), fmt.Sprintf("language=%s", isoCode),
			fmt.Sprintf("-metadata:s:s:%d", i), fmt.Sprintf("title=%s", langName),
		)
		if i == 0 {
			args = append(args, fmt.Sprintf("-disposition:s:%d", i), "default")
		} else {
			args = append(args, fmt.Sprintf("-disposition:s:%d", i), "0")
		}
	}

	// Global metadata
	args = append(args,
		"-metadata", fmt.Sprintf("title=S%02dE%02d - %s", info.EpisodeMetadata.SeasonNumber, info.EpisodeMetadata.EpisodeNumber, info.Title),
		"-metadata", fmt.Sprintf("show=%s", info.EpisodeMetadata.SeriesTitle),
		"-metadata", fmt.Sprintf("track=%d", info.EpisodeMetadata.EpisodeNumber),
		"-metadata", fmt.Sprintf("season_number=%d", info.EpisodeMetadata.SeasonNumber),
		outputFile,
	)

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(fmt.Errorf("ffmpeg merge failed: %w", err))
	}

	// Cleanup temp files
	_ = os.Remove(videoFile)
	for _, a := range audioTracks {
		_ = os.Remove(a.file)
	}
	for _, s := range subTracks {
		_ = os.Remove(s.file)
	}

	fmt.Printf("\n✓ Saved: %s\n\n", outputFile)
}
