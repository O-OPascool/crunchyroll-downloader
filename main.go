package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"flag"
)

var (
	token        = ""
	audioLangs   = flag.String("audio-langs", "ja-JP", "Comma-separated list of audio languages (e.g. \"ja-JP,fr-FR\")")
	subsLangs    = flag.String("subs-langs", "en-US", "Comma-separated list of subtitle languages, or \"all\" (e.g. \"en-US,fr-FR\")")
	videoQuality = flag.String("video-quality", "1080p", "Video quality")
	audioQuality = flag.String("audio-quality", "192k", "Audio quality")
	seasonNumber = flag.Int("season", 0, "Season number. Not used if an episode link is entered")
	etpRt        = flag.String("etp-rt", "", "The \"etp_rt\" cookie value of your account")
	releaseTag   = flag.String("tag", "Pascool", "Release tag appended to the filename")
)

// parseCommaSeparated splits a comma-separated flag value into a trimmed slice
func parseCommaSeparated(val string) []string {
	parts := strings.Split(val, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func processUrl(url string, aLangs, sLangs []string) {
	parts := strings.Split(url, "/")
	if len(parts) < 5 {
		fmt.Printf("Invalid URL format: %s\n", url)
		return
	}
	contentType := parts[3]
	contentId := parts[4]
	if len(contentId) != 9 && len(contentId) != 14 {
		fmt.Printf("Invalid URL format: %s\n", url)
		return
	}
	if contentType != "watch" && contentType != "series" {
		fmt.Printf("Invalid URL (must be /watch/ or /series/): %s\n", url)
		return
	}

	if contentType == "watch" {
		info, err := getEpisodeInfo(contentId)
		if err != nil {
			fmt.Printf("Error fetching episode info: %s\n", err)
			return
		}
		if err := downloadEpisode(contentId, videoQuality, audioQuality, aLangs, sLangs, info); err != nil {
			fmt.Printf("⚠  Error: %s\n", err)
		}
	} else {
		seasons := getSeasons(contentId)

		if *seasonNumber != 0 {
			var seasonId string
			for _, season := range seasons {
				if season.SeasonNumber == *seasonNumber {
					seasonId = season.ID
					break
				}
			}
			if seasonId == "" {
				fmt.Printf("This anime has no season %v!\n", *seasonNumber)
				return
			}
			episodes := getSeasonEpisodes(seasonId)
			downloadSeason(videoQuality, audioQuality, aLangs, sLangs, episodes)
		} else {
			fmt.Println("No season specified, downloading all seasons...")
			for _, season := range seasons {
				episodes := getSeasonEpisodes(season.ID)
				downloadSeason(videoQuality, audioQuality, aLangs, sLangs, episodes)
			}
		}
	}
}

func main() {
	url := flag.String("url", "", "URL of the episode/season to download")
	urlsFile := flag.String("urls", "", "Path to a text file with one URL per line")
	flag.Parse()

	if *url == "" && *urlsFile == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *etpRt == "" {
		fmt.Println("You must specify the \"-etp-rt\" option!")
		os.Exit(1)
	}

	aLangs := parseCommaSeparated(*audioLangs)
	sLangs := parseCommaSeparated(*subsLangs)

	token = GetAccessToken(*etpRt)

	if *urlsFile != "" {
		file, err := os.Open(*urlsFile)
		if err != nil {
			fmt.Printf("Failed to open URLs file: %s\n", err)
			os.Exit(1)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		var urls []string
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && strings.HasPrefix(line, "http") {
				urls = append(urls, line)
			}
		}

		fmt.Printf("Found %d URLs to download\n\n", len(urls))
		for i, u := range urls {
			fmt.Printf("=== [%d/%d] %s ===\n", i+1, len(urls), u)
			processUrl(u, aLangs, sLangs)
			fmt.Println()
		}
	} else {
		processUrl(*url, aLangs, sLangs)
	}
}
