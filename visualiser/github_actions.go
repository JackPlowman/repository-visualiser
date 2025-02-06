package visualiser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
)

// ...existing code...

// pushSVGToBranch creates or checks out branch "repository-visualiser", writes diagram.svg into a directory
// named with the current commit hash, then commits and pushes the change. If the branch already exists its
// history is preserved.
func pushSVGToBranch(svgContent string) (string, error) {
	commitHash := os.Getenv("GITHUB_SHA")
	if commitHash == "" {
		commitHash = "latest"
	}
	// Create a new directory for the repository.
	repoDir := "/tmp/repository-visualiser"
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Change to the repository directory.
	if err := os.Chdir(repoDir); err != nil {
		return "", fmt.Errorf("failed to change directory: %w", err)
	}

	// Clone the repository.
	repoURL := "https://github.com/JackPlowman/repository-visualiser"
	cmd := exec.Command("git", "clone", repoURL, repoDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to clone repository: %s", output)
	}

	// Create and checkout the branch.
	cmd = exec.Command("git", "checkout", "-B", "repository-visualiser")
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to create or checkout branch: %s", output)
	}

	// Create a directory for the commit hash.
	commitDir := filepath.Join(repoDir, commitHash)
	if err := os.MkdirAll(commitDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create commit directory: %w", err)
	}

	// Change to the repository directory.
	if err := os.Chdir(commitDir); err != nil {
		return "", fmt.Errorf("failed to change directory: %w", err)
	}

	err := os.WriteFile(filepath.Join(repoDir, commitHash, "diagram.svg"), []byte(svgContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write SVG file: %w", err)
	}

	cmd = exec.Command("git", "config", "--global", "user.name", "github-actions")
	cmd.Run()

	cmd = exec.Command("git", "config", "--global", "user.email", "github-actions@github.com")
	cmd.Run()

	// Add, commit, and push the changes.
	cmd = exec.Command("git", "add", "-f", "diagram.svg")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to add changes: %s", output)
	}

	cmd = exec.Command("git", "commit", "-m", "Add repository visualisation")
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to commit changes: %s", output)
	}

	// Use personal access token for authentication.
	token := os.Getenv("INPUT_GITHUB_TOKEN")
	if token == "" {
		return "", errors.New("GITHUB_TOKEN not set")
	}
	authURL := fmt.Sprintf("https://%s@github.com/JackPlowman/repository-visualiser.git", token)
	cmd = exec.Command("git", "push", "-u", authURL, "repository-visualiser")
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to push changes: %s", output)
	}

	return fmt.Sprintf("%s/%s/diagram.svg", repoURL, commitHash), nil
}

// Updated commentOnPR using go-github.
func commentOnPR(svgURL string) error {
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
	repoFull := os.Getenv("INPUT_GITHUB_REPOSITORY")
	if repoFull == "" {
		return errors.New("GITHUB_REPOSITORY not set")
	}
	parts := strings.Split(repoFull, "/")
	if len(parts) != 2 {
		return errors.New("invalid GITHUB_REPOSITORY format")
	}
	owner, repo := parts[0], parts[1]
	token := os.Getenv("INPUT_GITHUB_TOKEN")
	if token == "" {
		return errors.New("GITHUB_TOKEN not set")
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	comment := &github.IssueComment{
		Body: github.String(fmt.Sprintf("## Repository Visualiser\n![Diagram](%s)", svgURL)),
	}
	_, _, err = client.Issues.CreateComment(ctx, owner, repo, event.PullRequest.Number, comment)
	return err
}
