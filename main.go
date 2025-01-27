package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-enry/go-enry/v2"
)

func main() {
	var languageCountArray LanguageCountArray
	languageCountArray = recursiveFileSearch(languageCountArray)
	fmt.Println(languageCountArray)
}

type LanguageCount struct {
	Language string
	Count    int
}

type LanguageCountArray []LanguageCount

func recursiveFileSearch(languageCountArray LanguageCountArray) LanguageCountArray {
	root := "."

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			languageCountArray = detectLanguage(languageCountArray, path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %q: %v\n", root, err)
	}

	return languageCountArray
}

func detectLanguage(languageCountArray LanguageCountArray, filePath string) LanguageCountArray {
	language, _ := enry.GetLanguageByExtension(filePath)
	fmt.Println(language)
	if language == "" {
		language = "Unknown"
	}
	for i := range languageCountArray {
		if languageCountArray[i].Language == language {
			languageCountArray[i].Count++
			return languageCountArray
		}
	}
	return append(languageCountArray, LanguageCount{Language: language, Count: 1})
}
