package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runCmd is a helper to execute the compiled gstow binary
func runCmd(t *testing.T, dir string, name string, args ...string) string {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command '%s %s' failed: %v\nOutput: %s", name, strings.Join(args, " "), err, string(out))
	}
	return string(out)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isSymlinkTo(t *testing.T, linkPath, expectedTarget string) bool {
	info, err := os.Lstat(linkPath)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		return false
	}
	target, err := os.Readlink(linkPath)
	if err != nil {
		return false
	}
	return target == expectedTarget
}

func TestGstowE2E(t *testing.T) {
	// 1. Setup isolated environment
	homeDir, err := os.MkdirTemp("", "gstow-home-*")
	if err != nil {
		t.Fatalf("Failed to create temp home dir: %v", err)
	}
	defer os.RemoveAll(homeDir)

	repoDir, err := os.MkdirTemp("", "gstow-repo-*")
	if err != nil {
		t.Fatalf("Failed to create temp repo dir: %v", err)
	}
	defer os.RemoveAll(repoDir)

	// 2. Compile the binary into the temp repo
	binPath := filepath.Join(repoDir, "gstow")
	cmd := exec.Command("go", "build", "-o", binPath, "main.go")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to compile gstow: %v\nOutput: %s", err, string(out))
	}

	// 3. Create a dummy config file in the fake home directory
	localConfigDir := filepath.Join(homeDir, ".emacs.d")
	os.MkdirAll(localConfigDir, 0755)
	localConfigFile := filepath.Join(localConfigDir, "init.el")
	os.WriteFile(localConfigFile, []byte("(setq dummy-var t)"), 0644)

	repoConfigPath := "emacs/init.el"
	localConfigRelative := "~/.emacs.d/init.el"
	absRepoConfigPath := filepath.Join(repoDir, repoConfigPath)

	t.Run("add", func(t *testing.T) {
		runCmd(t, repoDir, binPath, "-d", homeDir, "add", localConfigRelative, repoConfigPath)

		// Verify file moved to repo
		if !fileExists(absRepoConfigPath) {
			t.Errorf("File was not moved to repo: %s", absRepoConfigPath)
		}

		// Verify symlink created in home
		if !isSymlinkTo(t, localConfigFile, absRepoConfigPath) {
			t.Errorf("Symlink was not created correctly at %s", localConfigFile)
		}

		// Verify yaml updated
		yamlData, _ := os.ReadFile(filepath.Join(repoDir, "gs.yaml"))
		if !strings.Contains(string(yamlData), repoConfigPath) {
			t.Errorf("gs.yaml does not contain the mapping")
		}
	})

	t.Run("show", func(t *testing.T) {
		out := runCmd(t, repoDir, binPath, "-d", homeDir, "show")
		if !strings.Contains(out, repoConfigPath) {
			t.Errorf("show command did not output the expected mapping")
		}
	})

	t.Run("rmloc", func(t *testing.T) {
		runCmd(t, repoDir, binPath, "-d", homeDir, "rmloc", repoConfigPath)

		// Verify symlink removed
		if fileExists(localConfigFile) || isSymlinkTo(t, localConfigFile, absRepoConfigPath) {
			t.Errorf("Symlink was not removed from %s", localConfigFile)
		}

		// Verify repo file remains intact
		if !fileExists(absRepoConfigPath) {
			t.Errorf("Repo file was incorrectly deleted: %s", absRepoConfigPath)
		}

		// Verify yaml updated
		yamlData, _ := os.ReadFile(filepath.Join(repoDir, "gs.yaml"))
		if strings.Contains(string(yamlData), repoConfigPath) {
			t.Errorf("gs.yaml mapping was not removed")
		}
	})

	t.Run("install", func(t *testing.T) {
		// Manually re-add to yaml to test install functor
		yamlContent := repoConfigPath + ": " + localConfigRelative + "\n"
		os.WriteFile(filepath.Join(repoDir, "gs.yaml"), []byte(yamlContent), 0644)

		runCmd(t, repoDir, binPath, "-d", homeDir, "install")

		// Verify symlink restored
		if !isSymlinkTo(t, localConfigFile, absRepoConfigPath) {
			t.Errorf("install failed to recreate the symlink at %s", localConfigFile)
		}
	})

	t.Run("install_idempotency", func(t *testing.T) {
		// Running it again should exit cleanly 0 without errors
		out := runCmd(t, repoDir, binPath, "-d", homeDir, "install")
		if strings.Contains(out, "[ERROR]") {
			t.Errorf("Idempotent install produced an error: %s", out)
		}
	})
}
