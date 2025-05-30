package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-enry/go-enry/v2"
	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
)

// main is the entry point of the application.
func main() {
	languageCountArray := recursiveFileSearch(".")
	fmt.Println(languageCountArray)

	fileStats := getFileStats("/github/workspace")
	if len(fileStats) == 0 {
		fmt.Println("No files found.")
		return
	}
	// Apply ignore list filtering.
	fileStats = filterIgnoredFiles(fileStats, defaultIgnoreList)
	svgOutput := generateSVG(fileStats)

	markdownContent := generateMarkdownContent(svgOutput)

	// Post PR comment with SVG image.
	if err := commentOnPR(markdownContent); err != nil {
		fmt.Println("Error commenting on PR:", err)
	}
	writeSummary(languageCountArray, markdownContent)
}

// writeDiagram writes the SVG output to a file named "diagram.svg".
func writeDiagram(svgOutput string) {
	// Write svg locally.
	err := os.WriteFile("/github/workspace/repository_visualisation.svg", []byte(svgOutput), 0644)
	if err != nil {
		fmt.Println("Error writing SVG file:", err)
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
		r := math.Max(float64(sum)/10+20, 10) // Ensure a minimum radius to avoid zero or negative values.
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
		angleStep := 360.0 / math.Max(float64(len(fd.files)), 1) // Avoid division by zero.
		for i, f := range fd.files {
			// ...existing code to calculate color, radius, and angle...
			color := getLanguageColor(f.Path)
			rad := float64(f.Lines)/10 + 5
			angle := float64(i) * angleStep
			fileX := centerX + math.Max(fd.radius-25, 5)*cosDeg(angle) // Ensure valid radius for placement.
			fileY := centerY + math.Max(fd.radius-25, 5)*sinDeg(angle)
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
	writeDiagram(output.String())
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

// generateMarkdownContent generates a Markdown string with the SVG image.
func generateMarkdownContent(svgContent string) string {
	encodedSVG := base64.StdEncoding.EncodeToString([]byte(svgContent))
	return fmt.Sprintf("## Repository Visualiser\n![SVG Image](data:image/svg+xml;base64,%s)", encodedSVG)
}

// Updated commentOnPR using go-github.
func commentOnPR(markdownContent string) error {
	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		return nil // Not running in a GitHub Actions event.
	}
	eventData, err := os.ReadFile(eventPath)
	if err != nil {
		return err
	}
	var event struct {
		PullRequest struct {
			Number int `json:"number"`
		} `json:"pull_request"`
	}
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	if event.PullRequest.Number == 0 {
		return nil
	}
	github_repository := os.Getenv("INPUT_GITHUB_REPOSITORY")
	if github_repository == "" {
		return errors.New("GITHUB_REPOSITORY not set")
	}
	github_repository_owner := os.Getenv("GITHUB_REPOSITORY_OWNER")
	if github_repository_owner == "" {
		return errors.New("GITHUB_REPOSITORY_OWNER not set")
	}
	token := os.Getenv("INPUT_GITHUB_TOKEN")
	if token == "" {
		return errors.New("GITHUB_TOKEN not set")
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	comment := &github.IssueComment{
		Body: github.String(markdownContent),
	}
	_, _, err = client.Issues.CreateComment(ctx, github_repository_owner, github_repository, event.PullRequest.Number, comment)
	return err
}

// writeSummary writes the language count array to the GitHub Action summary if available.
func writeSummary(languageCountArray LanguageCountArray, markdownContent string) {
	actionSummaryPath := os.Getenv("GITHUB_STEP_SUMMARY")
	if actionSummaryPath != "" {
		// Separate unknown language count.
		var unknownCount int
		var knownCounts LanguageCountArray
		for _, lc := range languageCountArray {
			if lc.Language == "Unknown" {
				unknownCount += lc.Count
			} else {
				knownCounts = append(knownCounts, lc)
			}
		}
		// Sort known counts by descending file count.
		sort.Slice(knownCounts, func(i, j int) bool {
			return knownCounts[i].Count > knownCounts[j].Count
		})

		// Build headers and counts.
		var headers, counts []string
		for _, lc := range knownCounts {
			headers = append(headers, lc.Language)
			counts = append(counts, fmt.Sprintf("%d", lc.Count))
		}
		// Append unknown column if found.
		if unknownCount > 0 {
			headers = append(headers, "Unknown")
			counts = append(counts, fmt.Sprintf("%d", unknownCount))
		}

		file, err := os.OpenFile(actionSummaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer file.Close()

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("%s\n", markdownContent))
		// Header row with empty left cell.
		sb.WriteString("|         | " + strings.Join(headers, " | ") + " |\n")
		// Separator row.
		sb.WriteString("|---------|" + strings.Repeat("---------|", len(headers)) + "\n")
		// Data row with left cell "Files".
		sb.WriteString("| Files   | " + strings.Join(counts, " | ") + " |\n")

		fmt.Fprintln(file, sb.String())
	}
}
