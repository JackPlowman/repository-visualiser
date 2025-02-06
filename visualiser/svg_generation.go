package visualiser

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-enry/go-enry/v2"
)

// ...existing code...

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
