package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-enry/go-enry/v2"
)

// main is the entry point of the application.
func main() {
	languageCountArray := recursiveFileSearch(".")
	fmt.Println(languageCountArray)
	writeSummary(languageCountArray)
	fileStats := getFileStats(".")
	svgOutput := generateSVG(fileStats)
	fmt.Println(svgOutput) // or write to a file
	err := os.WriteFile("diagram.svg", []byte(svgOutput), 0644)
	if err != nil {
		fmt.Println("Error writing SVG file:", err)
	}
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

type FileStat struct {
	Path  string
	Lines int
}

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

// generateSVG creates an SVG where bigger line counts yield larger circles.
func generateSVG(stats []FileStat) string {
	// Minimal example: each circle is placed in a simple grid
	// with radius proportional to the file's line count.
	svgHeader := `<svg xmlns="http://www.w3.org/2000/svg" width="800" height="600">`
	svgFooter := `</svg>`
	var output strings.Builder
	output.WriteString(svgHeader)
	x, y := 50.0, 50.0
	const step = 75.0

	for _, fs := range stats {
		color := getLanguageColor(fs.Path)
		rad := float64(fs.Lines) / 10
		if rad < 5 {
			rad = 5
		}
		output.WriteString(fmt.Sprintf(
			`<circle cx="%f" cy="%f" r="%f" fill="%s" />`,
			x, y, rad, color,
		))
		x += step
		if x > 700 {
			x = 50
			y += step
		}
	}
	output.WriteString(svgFooter)
	return output.String()
}

// getLanguageColor returns a basic color code per language.
func getLanguageColor(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return "none"
	}
	lang := enry.GetLanguage(path, content)
	switch lang {
	case "Go":
		return "lightblue"
	case "JavaScript":
		return "yellow"
	case "HTML":
		return "orange"
	// ...add more languages/colors as desired...
	default:
		return "grey"
	}
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
