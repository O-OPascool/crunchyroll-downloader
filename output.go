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

// getLangTag returns VOSTFR, VF or Multi depending on available audio tracks
func getLangTag(info EpisodeInfo, audioLang string) string {
	if audioLang == "fr-FR" {
		return "VF"
	}
	for _, v := range info.EpisodeMetadata.Versions {
		if v != nil && v.AudioLocale == "fr-FR" {
			return "Multi"
		}
	}
	return "VOSTFR"
}

// buildOutputPath returns (directory, full output path) for an episode
// Format: Title.S01E01.CR.WEBDL.Multi.1080p.x265-tag.mkv
func buildOutputPath(info EpisodeInfo, videoQuality, audioLang string) (string, string) {
	dirName := sanitizeForFS(info.EpisodeMetadata.SeriesTitle)
	titleDots := strings.ReplaceAll(dirName, " ", ".")
	season := fmt.Sprintf("S%02d", info.EpisodeMetadata.SeasonNumber)
	episode := fmt.Sprintf("E%02d", info.EpisodeMetadata.EpisodeNumber)
	langTag := getLangTag(info, audioLang)

	filename := fmt.Sprintf("%s.%s%s.CR.WEBDL.%s.%s.x265-%s.mkv",
		titleDots, season, episode, langTag, videoQuality, *releaseTag)

	return dirName, fmt.Sprintf("%s/%s", dirName, filename)
}

// mergeEverything merges audio, video and subtitles in a single MKV container
func mergeEverything(videoFile, audioFile, subsFile, outputFile string, subtitlesLang *string, info EpisodeInfo) {
	args := []string{"-y", "-i", videoFile, "-i", audioFile}

	if subsFile != "" {
		args = append(args, "-i", subsFile)
	}

	args = append(args, "-c:v", "copy", "-c:a", "copy")

	if subsFile != "" {
		args = append(args,
			"-c:s", "ass",
			"-metadata:s:s:0", fmt.Sprintf("language=%s", *subtitlesLang),
			"-metadata:s:s:0", fmt.Sprintf("title=%s", languageNames[*subtitlesLang]),
			"-disposition:s:0", "default",
		)
	}

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

	_ = os.Remove(videoFile)
	_ = os.Remove(audioFile)
	if subsFile != "" {
		_ = os.Remove(subsFile)
	}

	fmt.Printf("\n✓ Saved: %s\n\n", outputFile)
}
