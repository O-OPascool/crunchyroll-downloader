package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/unki2aut/go-mpd"
)

func parseManifest(url string) (*mpd.MPD, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("parseManifest: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:147.0) Gecko/20100101 Firefox/147.0")
	resp, err := DoRequest(req)
	if err != nil {
		return nil, fmt.Errorf("parseManifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("manifest request failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parseManifest: %w", err)
	}
	mpd := new(mpd.MPD)
	mpd.Decode(body)

	return mpd, nil
}

func getBaseUrl(set *mpd.AdaptationSet, isVideoSet bool, quality string) (*string, *string) {
	for _, representation := range set.Representations {
		if isVideoSet {
			toInt, _ := strconv.ParseInt(strings.ReplaceAll(quality, "p", ""), 10, 64)
			if *representation.Height == uint64(toInt) {
				return &representation.BaseURL[0].Value, representation.ID
			}
		} else {
			if representation.ID != nil && strings.Contains(*representation.ID, "audio/") {
				if strings.Contains(*representation.ID, quality) {
					return &representation.BaseURL[0].Value, representation.ID
				}
			} else if representation.Bandwidth != nil {
				num := strings.ReplaceAll(quality, "k", "")

				// Crunchyroll MPDs are weird on the "bandwidth" value, it can be 192002 (not just 192000) on certain manifests
				if num == "192" && *representation.Bandwidth >= 192000 {
					return &representation.BaseURL[0].Value, representation.ID
				} else if num == "128" && *representation.Bandwidth >= 128000 {
					return &representation.BaseURL[0].Value, representation.ID
				} else if num == "96" && *representation.Bandwidth >= 96000 {
					return &representation.BaseURL[0].Value, representation.ID
				}
			}
		}
	}
	// Fallback for audio: pick the highest bandwidth representation available
	if !isVideoSet && len(set.Representations) > 0 {
		var bestRep *mpd.Representation
		var bestBW uint64
		for _, rep := range set.Representations {
			if rep.Bandwidth != nil && *rep.Bandwidth > bestBW {
				bestBW = *rep.Bandwidth
				bestRep = &rep
			}
		}
		if bestRep != nil && len(bestRep.BaseURL) > 0 {
			fmt.Printf("  ⚠ Audio quality %q not found, using best available (%d bps)\n", quality, bestBW)
			return &bestRep.BaseURL[0].Value, bestRep.ID
		}
		// Last resort: just pick the first representation
		if len(set.Representations[0].BaseURL) > 0 {
			fmt.Printf("  ⚠ Audio quality %q not found, using first available representation\n", quality)
			return &set.Representations[0].BaseURL[0].Value, set.Representations[0].ID
		}
	}
	return nil, nil
}

// findAdaptationSet finds an AdaptationSet by content type ("video" or "audio").
// It checks MimeType, ContentType, and representation properties to handle different MPD layouts.
func findAdaptationSet(period *mpd.Period, contentType string) *mpd.AdaptationSet {
	for _, set := range period.AdaptationSets {
		// Check by MimeType (e.g. "video/mp4", "audio/mp4")
		if strings.HasPrefix(set.MimeType, contentType+"/") {
			return set
		}
		// Check by ContentType attribute
		if set.ContentType != nil && *set.ContentType == contentType {
			return set
		}
	}
	// Fallback: infer from representations
	for _, set := range period.AdaptationSets {
		for _, rep := range set.Representations {
			if contentType == "video" && rep.Height != nil {
				return set
			}
			if contentType == "audio" && rep.Height == nil && rep.Bandwidth != nil {
				return set
			}
			if contentType == "audio" && rep.ID != nil && strings.Contains(*rep.ID, "audio/") {
				return set
			}
		}
	}
	return nil
}

func expandTimeline(timeline []*mpd.SegmentTimelineS, startNumber int64) []int64 {
	var result []int64
	segNum := startNumber

	for _, s := range timeline {
		repeat := int64(0)
		if s.R != nil && *s.R > 0 {
			repeat = *s.R
		}

		total := repeat + 1 // DASH rule: total segments = r + 1

		for i := int64(0); i < total; i++ {
			result = append(result, segNum)
			segNum++
		}
	}

	return result
}
