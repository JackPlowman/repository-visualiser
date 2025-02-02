package main

import (
	"fmt"
	"math"
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

// groupFileStatsByFolder groups file stats by their top-level folder.
func groupFileStatsByFolder(stats []FileStat) map[string][]FileStat {
	result := make(map[string][]FileStat)
	for _, fs := range stats {
		dir := filepath.Dir(fs.Path)
		result[dir] = append(result[dir], fs)
	}
	return result
}

// generateSVG creates an SVG with one circle per folder, then circles for each file inside it.
func generateSVG(stats []FileStat) string {
	folderMap := groupFileStatsByFolder(stats)
	// Collect folder data first.
	type folderData struct {
		folderPath string
		folderSum  int
		files      []FileStat
		radius     float64
	}
	var allFolders []folderData
	for folderPath, files := range folderMap {
		sum := 0
		for _, f := range files {
			sum += f.Lines
		}
		r := float64(sum)/10 + 20
		allFolders = append(allFolders, folderData{folderPath: folderPath, folderSum: sum, files: files, radius: r})
	}

	// Layout so folder circles don’t overlap.
	svgHeader := `<svg xmlns="http://www.w3.org/2000/svg" width="800" height="600">`
	svgFooter := `</svg>`
	var output strings.Builder
	output.WriteString(svgHeader)

	const maxWidth = 800.0
	const margin = 10.0
	x, y := 0.0, 100.0
	rowHeight := 0.0

	for _, fd := range allFolders {
		diameter := fd.radius * 2
		if x+diameter > maxWidth {
			// Move to the next row.
			x = 0
			y += rowHeight + margin
			rowHeight = 0
		}

		centerX := x + fd.radius
		centerY := y + fd.radius
		// Update row height if needed.
		if diameter > rowHeight {
			rowHeight = diameter
		}

		// Draw folder circle.
		output.WriteString(fmt.Sprintf(
			`<circle cx="%f" cy="%f" r="%f" fill="none" stroke="black" stroke-width="2" />`,
			centerX, centerY, fd.radius,
		))
		output.WriteString(fmt.Sprintf(
			`<text x="%f" y="%f" text-anchor="middle" alignment-baseline="baseline" font-size="12">%s</text>`,
			centerX, centerY-fd.radius-5, filepath.Base(fd.folderPath),
		))

		// Place file circles inside folder.
		angleStep := 360.0 / float64(len(fd.files)+1)
		for i, f := range fd.files {
			// ...existing code to calculate color, radius, and angle...
			color := getLanguageColor(f.Path)
			rad := float64(f.Lines)/10 + 5
			angle := float64(i) * angleStep
			fileX := centerX + (fd.radius-25)*cosDeg(angle)
			fileY := centerY + (fd.radius-25)*sinDeg(angle)
			// Draw file circle without title.
			output.WriteString(fmt.Sprintf(
				`<circle cx="%f" cy="%f" r="%f" fill="%s" />`,
				fileX, fileY, rad, color,
			))
			// Only add text if circle is large enough for legibility.
			if rad >= 15 {
				output.WriteString(fmt.Sprintf(
					`<text x="%f" y="%f" text-anchor="middle" alignment-baseline="middle" font-size="8">%s</text>`,
					fileX, fileY, filepath.Base(f.Path),
				))
			}
		}

		// Advance x to the right edge of this folder circle (touching edges).
		x += diameter
	}

	output.WriteString(svgFooter)
	return output.String()
}

// cosDeg and sinDeg are helper functions for degrees-based trig.
func cosDeg(deg float64) float64 { return math.Cos(deg * math.Pi / 180) }
func sinDeg(deg float64) float64 { return math.Sin(deg * math.Pi / 180) }

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
