package visualiser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-enry/go-enry/v2"
)

// getFileStats recursively counts lines in each file.
func getFileStats(root string) []FileStat {
	var stats []FileStat
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			content, err := os.ReadFile(path)
			if err == nil {
				lineCount := len(strings.Split(string(content), "\n"))
				stats = append(stats, FileStat{Path: path, Lines: lineCount})
			}
		}
		return nil
	})
	return stats
}

// recursiveFileSearch recursively searches files in the provided root directory and returns language counts.
func recursiveFileSearch(root string) LanguageCountArray {
	languageCountMap := make(map[string]int)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			language := enry.GetLanguage(path, content)
			if language == "" {
				language = "Unknown"
			}
			languageCountMap[language]++
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error:", err)
	}

	return mapToLanguageCountArray(languageCountMap)
}

// mapToLanguageCountArray converts a map of language counts to a LanguageCountArray.
func mapToLanguageCountArray(languageCountMap map[string]int) LanguageCountArray {
	languageCountArray := make(LanguageCountArray, 0, len(languageCountMap))
	for language, count := range languageCountMap {
		languageCountArray = append(languageCountArray, LanguageCount{Language: language, Count: count})
	}
	return languageCountArray
}
