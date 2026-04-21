package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Release struct {
	Version string `yaml:"version"`
	Notes   string `yaml:"notes"`
}

func main() {
	// 1. Parse the Changelog
	data, err := os.ReadFile("changelog.yaml")
	if err != nil {
		fmt.Printf("Fatal: Could not read changelog.yaml: %v\n", err)
		os.Exit(1)
	}

	var changelog []Release
	if err := yaml.Unmarshal(data, &changelog); err != nil || len(changelog) == 0 {
		fmt.Printf("Fatal: Invalid or empty changelog.yaml\n")
		os.Exit(1)
	}

	latest := changelog[0]
	fmt.Printf("🚀 Preparing Release v%s\n", latest.Version)

	// 2. Tag and Push
	run("git", "tag", "-a", "v"+latest.Version, "-m", "Release v"+latest.Version)
	run("git", "push", "origin", "v"+latest.Version)

	// 3. Create GitHub Release and attach artifacts
	args := []string{"release", "create", "v" + latest.Version, "--title", "v" + latest.Version, "--notes", latest.Notes}

	// Harvest packages and checksums from the out/ directory
	if matches, _ := filepath.Glob("out/*.*"); len(matches) > 0 {
		args = append(args, matches...)
	}

	fmt.Printf("📦 Uploading to GitHub...\n")
	run("gh", args...)
	fmt.Printf("✅ GitHub Release v%s is live!\n", latest.Version)

	// 4. Update Homebrew Tap
	fmt.Printf("🍺 Updating Homebrew Tap...\n")

	// Create a safe, temporary directory
	tempDir, err := os.MkdirTemp("", "homebrew-tap-*")
	if err != nil {
		fmt.Printf("Fatal: Could not create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir) // Ensure it gets cleaned up

	run("git", "clone", "git@github.com:troodos-exascale/homebrew-tap.git", tempDir)

	formula := fmt.Sprintf(`class Tstow < Formula
  desc "Explicit, idempotent deployment functor for dotfiles"
  homepage "https://github.com/troodos-exascale/tstow"
  url "https://github.com/troodos-exascale/tstow.git",
      tag:      "v%s",
      revision: "HEAD"
  license "Apache-2.0"
  head "https://github.com/troodos-exascale/tstow.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", "-o", bin/"tstow", "main.go"
    generate_completions_from_executable(bin/"tstow", "completion")
  end

  test do
    system "#{bin}/tstow", "--help"
  end
end`, latest.Version)

	formulaPath := filepath.Join(tempDir, "tstow.rb")
	os.WriteFile(formulaPath, []byte(formula), 0644)

	run("git", "-C", tempDir, "add", "tstow.rb")
	run("git", "-C", tempDir, "commit", "-m", "tstow: bump to v"+latest.Version)
	run("git", "-C", tempDir, "push", "origin", "main")

	fmt.Printf("✅ Homebrew Tap updated to v%s!\n", latest.Version)
}

// run executes a command, streams output, and intelligently ignores safe errors
func run(cmdName string, args ...string) {
	cmd := exec.Command(cmdName, args...)

	// Capture all stdout and stderr
	out, err := cmd.CombinedOutput()
	outputStr := string(out)

	// Print the output so you still see exactly what is happening
	if outputStr != "" {
		fmt.Print(outputStr)
	}

	if err != nil {
		// Now we are actually reading the raw Git/GH error message!
		if strings.Contains(outputStr, "already exists") || strings.Contains(outputStr, "nothing to commit") {
			return // Safely ignore these specific known states and keep going
		}
		os.Exit(1)
	}
}
