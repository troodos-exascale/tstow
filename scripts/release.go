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

	// 3. Harvest artifacts
	var uploadFiles []string
	if matches, _ := filepath.Glob("out/*.*"); len(matches) > 0 {
		uploadFiles = matches
	}

	// 4. Create or Append to GitHub Release
	fmt.Printf("📦 Uploading to GitHub...\n")
	createArgs := append([]string{"release", "create", "v" + latest.Version, "--title", "v" + latest.Version, "--notes", latest.Notes}, uploadFiles...)
	cmd := exec.Command("gh", createArgs...)

	out, ghErr := cmd.CombinedOutput()
	outStr := string(out)

	if ghErr != nil {
		// If Mac already created it, gracefully fall back to uploading assets to the existing release
		if strings.Contains(outStr, "already exists") || strings.Contains(outStr, "Validation Failed") {
			fmt.Printf("⚠️ Release v%s already exists. Appending assets instead...\n", latest.Version)
			if len(uploadFiles) > 0 {
				uploadArgs := append([]string{"release", "upload", "v" + latest.Version, "--clobber"}, uploadFiles...)
				run("gh", uploadArgs...)
			}
		} else if strings.Contains(ghErr.Error(), "executable file not found") {
			fmt.Printf("\n❌ Fatal: 'gh' CLI is not installed on this machine.\nInstall it and run 'gh auth login'.\n")
			os.Exit(1)
		} else {
			fmt.Printf("\n❌ Fatal GitHub Error:\n%s\n%v\n", outStr, ghErr)
			os.Exit(1)
		}
	} else if outStr != "" {
		fmt.Print(outStr)
	}
	fmt.Printf("✅ GitHub Release v%s is synced!\n", latest.Version)

	// 5. Update Homebrew Tap
	fmt.Printf("🍺 Updating Homebrew Tap...\n")
	tempDir, err := os.MkdirTemp("", "homebrew-tap-*")
	if err != nil {
		fmt.Printf("Fatal: Could not create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

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

	os.WriteFile(filepath.Join(tempDir, "tstow.rb"), []byte(formula), 0644)

	run("git", "-C", tempDir, "add", "tstow.rb")
	run("git", "-C", tempDir, "commit", "-m", "tstow: bump to v"+latest.Version)
	run("git", "-C", tempDir, "push", "origin", "main")

	fmt.Printf("✅ Homebrew Tap updated to v%s!\n", latest.Version)
}

// run executes a command, streams output, and intelligently ignores safe errors
func run(cmdName string, args ...string) {
	cmd := exec.Command(cmdName, args...)
	out, err := cmd.CombinedOutput()
	outStr := string(out)

	if outStr != "" {
		fmt.Print(outStr)
	}

	if err != nil {
		if strings.Contains(outStr, "already exists") || strings.Contains(outStr, "nothing to commit") {
			return // Safely ignore these specific known states and keep going
		}
		fmt.Printf("\n❌ Command failed: %s %v\nError details: %v\n", cmdName, args, err)
		os.Exit(1)
	}
}
