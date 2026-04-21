package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

type Release struct {
	Version string `yaml:"version"`
	Notes   string `yaml:"notes"`
}

func main() {
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

	// 1. Tag and Push
	run("git", "tag", "-a", "v"+latest.Version, "-m", "Release v"+latest.Version)
	run("git", "push", "origin", "v"+latest.Version)

	// 2. Create GitHub Release
	args := []string{"release", "create", "v" + latest.Version, "--title", "v" + latest.Version, "--notes", latest.Notes}

	// If packages exist in the out/ directory, attach them
	if matches, _ := filepath.Glob("out/*.*"); len(matches) > 0 {
		args = append(args, matches...)
	}

	fmt.Printf("📦 Uploading to GitHub...\n")
	run("gh", args...)
	fmt.Printf("✅ Release v%s is live!\n", latest.Version)
}

func run(cmdName string, args ...string) {
	cmd := exec.Command(cmdName, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Ignore tag already exists errors, but fail on others
		if !strings.Contains(err.Error(), "already exists") {
			os.Exit(1)
		}
	}
}
fmt.Printf("🍺 Updating Homebrew Tap...\n")
	
	// 1. Clone the tap repository into a temporary directory
	run("git", "clone", "git@github.com:troodos-exascale/homebrew-tap.git", "/tmp/homebrew-tap")
	
	// 2. Dynamically generate the Ruby formula with the new version
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

	// 3. Write it, commit it, and push it
	os.WriteFile("/tmp/homebrew-tap/tstow.rb", []byte(formula), 0644)
	
	run("git", "-C", "/tmp/homebrew-tap", "add", "tstow.rb")
	run("git", "-C", "/tmp/homebrew-tap", "commit", "-m", "tstow: bump to v"+latest.Version)
	run("git", "-C", "/tmp/homebrew-tap", "push", "origin", "main")
	
	// 4. Cleanup
	os.RemoveAll("/tmp/homebrew-tap")
	fmt.Printf("✅ Homebrew Tap updated to v%s!\n", latest.Version)
