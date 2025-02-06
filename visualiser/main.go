package visualiser

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// main is the entry point of the application.
func main() {
	languageCountArray := recursiveFileSearch(".")
	fmt.Println(languageCountArray)

	fileStats := getFileStats("/github/workspace")
	// Apply ignore list filtering.
	ignoreList, err := loadIgnoreList()
	if err != nil {
		fmt.Println("Error loading ignore list:", err)
	} else {
		fileStats = filterIgnoredFiles(fileStats, ignoreList)
	}
	svgOutput := generateSVG(fileStats)

	// Push the SVG to branch "repository-visualiser" in a commit-hash directory.
	svgURL, err := pushSVGToBranch(svgOutput)
	if err != nil {
		fmt.Println("Error pushing SVG:", err)
	}
	// Post PR comment with link to the SVG if running in a pull request.
	if err := commentOnPR(svgURL); err != nil {
		fmt.Println("Error commenting on PR:", err)
	}
	writeSummary(languageCountArray)
}

// writeDiagram writes the SVG output to a file named "diagram.svg".
func writeDiagram(svgOutput string) {
	// Write svg locally.
	err := os.WriteFile("diagram.svg", []byte(svgOutput), 0644)
	if err != nil {
		fmt.Println("Error writing SVG file:", err)
	}
}

// writeSummary writes the language count array to the GitHub Action summary if available.
func writeSummary(languageCountArray LanguageCountArray) {
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
		// Header row with empty left cell.
		sb.WriteString("|         | " + strings.Join(headers, " | ") + " |\n")
		// Separator row.
		sb.WriteString("|---------|" + strings.Repeat("---------|", len(headers)) + "\n")
		// Data row with left cell "Files".
		sb.WriteString("| Files   | " + strings.Join(counts, " | ") + " |\n")

		fmt.Fprintln(file, sb.String())
	}
}
