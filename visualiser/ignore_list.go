package visualiser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ...existing code...

// Update loadIgnoreList to parse JSON with an "ignore" key.
func loadIgnoreList() ([]string, error) {
	data, err := os.ReadFile("ignore.json")
	if err != nil {
		return nil, err
	}
	var config struct {
		Ignore []string `json:"ignore"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return config.Ignore, nil
}

// filterIgnoredFiles filters out FileStat entries whose Path matches any pattern in the ignore list.
// If a pattern does not include wildcards and matches a folder name, all files under that folder are ignored.
func filterIgnoredFiles(stats []FileStat, ignoreList []string) []FileStat {
	var filtered []FileStat
	for _, stat := range stats {
		ignore := false
		for _, pattern := range ignoreList {
			// First, check if pattern has wildcards.
			if strings.ContainsAny(pattern, "*?[]") {
				if matched, _ := filepath.Match(pattern, stat.Path); matched {
					ignore = true
					break
				}
			} else {
				// No wildcard: if the file path equals the pattern or starts with pattern + separator, ignore it.
				if stat.Path == pattern || strings.HasPrefix(stat.Path, pattern+string(os.PathSeparator)) {
					ignore = true
					break
				}
			}
		}
		if !ignore {
			filtered = append(filtered, stat)
		}
	}
	return filtered
}
