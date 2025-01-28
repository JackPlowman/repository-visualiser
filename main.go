package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-enry/go-enry/v2"
)

// main is the entry point of the application.
func main() {
	languageCountArray := recursiveFileSearch(".")
	fmt.Println(languageCountArray)
	writeSummary(languageCountArray)
}

// writeSummary writes the language count array to the GitHub Action summary if available.
func writeSummary(languageCountArray LanguageCountArray) {
	actionSummaryPath := os.Getenv("GITHUB_STEP_SUMMARY")
	if actionSummaryPath != "" {
		file, err := os.OpenFile(actionSummaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer file.Close()
		for _, lc := range languageCountArray {
			fmt.Fprintf(file, "%s: %d\n", lc.Language, lc.Count)
		}
	}
}

type LanguageCount struct {
	Language string
	Count    int
}

type LanguageCountArray []LanguageCount

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
