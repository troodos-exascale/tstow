package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runCmdE runs a command and returns output and error, allowing tests to verify expected failures.
func runCmdE(binPath string, args ...string) (string, error) {
	cmd := exec.Command(binPath, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
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

func TestTstowE2E(t *testing.T) {
	// 1. Compile the binary ONCE for all tests
	buildDir, _ := os.MkdirTemp("", "tstow-build-*")
	defer os.RemoveAll(buildDir)
	binPath := filepath.Join(buildDir, "tstow")

	if out, err := exec.Command("go", "build", "-o", binPath, "main.go").CombinedOutput(); err != nil {
		t.Fatalf("Failed to compile tstow: %v\nOutput: %s", err, string(out))
	}

	// Helper to create isolated environments for each subtest
	setupEnv := func(t *testing.T) (string, string, []string) {
		homeDir, _ := os.MkdirTemp("", "tstow-home-*")
		repoDir, _ := os.MkdirTemp("", "tstow-repo-*")
		baseArgs := []string{"-i", homeDir, "-r", repoDir}

		// Jail the binary so os.UserHomeDir() points to our temp folder!
		origHome := os.Getenv("HOME")
		os.Setenv("HOME", homeDir)
		t.Cleanup(func() {
			os.Setenv("HOME", origHome)
		})

		return homeDir, repoDir, baseArgs
	}

	t.Run("Req: Add regular file moves it and links it", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		localFile := filepath.Join(homeDir, ".bashrc")
		os.WriteFile(localFile, []byte("echo hello"), 0644)

		args := append(baseArgs, "add", ".bashrc", "shell/.bashrc")
		out, err := runCmdE(binPath, args...)
		if err != nil {
			t.Fatalf("add failed: %v, output: %s", err, out)
		}

		absRepo := filepath.Join(repoDir, "shell", ".bashrc")
		if !fileExists(absRepo) {
			t.Errorf("File was not moved to repo")
		}
		if !isSymlinkTo(t, localFile, absRepo) {
			t.Errorf("Symlink not created at install location")
		}
	})

	t.Run("Req: Cannot add a symbolic link", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		targetFile := filepath.Join(homeDir, "real_file")
		symlinkFile := filepath.Join(homeDir, "symlink")
		os.WriteFile(targetFile, []byte("data"), 0644)
		os.Symlink(targetFile, symlinkFile)

		args := append(baseArgs, "add", "symlink", "shell/symlink")
		out, err := runCmdE(binPath, args...)
		if err == nil || !strings.Contains(out, "Cannot add a symbolic link") {
			t.Errorf("Expected failure when adding symlink. Output: %s", out)
		}
	})

	t.Run("Req: Uses relative paths ONLY in repo location", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		localFile := filepath.Join(homeDir, ".bashrc")
		os.WriteFile(localFile, []byte("data"), 0644)

		args := append(baseArgs, "add", ".bashrc", "/absolute/path/in/repo")
		out, err := runCmdE(binPath, args...)
		if err == nil || !strings.Contains(out, "must be relative") {
			t.Errorf("Expected failure for absolute repo path. Output: %s", out)
		}
	})

	t.Run("Req: Install entire placement list", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		os.MkdirAll(filepath.Join(repoDir, "shell"), 0755)
		os.WriteFile(filepath.Join(repoDir, "shell", ".bashrc"), []byte("data"), 0644)
		yamlContent := "mappings:\n  shell/.bashrc: .bashrc\n"
		os.WriteFile(filepath.Join(repoDir, "tstow.yaml"), []byte(yamlContent), 0644)

		runCmdE(binPath, append(baseArgs, "install")...)

		if !isSymlinkTo(t, filepath.Join(homeDir, ".bashrc"), filepath.Join(repoDir, "shell", ".bashrc")) {
			t.Errorf("Install failed to create symlink")
		}
	})

	t.Run("Req: Install recursive (folder subset)", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		os.MkdirAll(filepath.Join(repoDir, "shell"), 0755)
		os.MkdirAll(filepath.Join(repoDir, "emacs"), 0755)
		os.WriteFile(filepath.Join(repoDir, "shell", ".bashrc"), []byte("data"), 0644)
		os.WriteFile(filepath.Join(repoDir, "emacs", "init.el"), []byte("data"), 0644)

		yamlContent := "mappings:\n  shell/.bashrc: .bashrc\n  emacs/init.el: .emacs.d/init.el\n"
		os.WriteFile(filepath.Join(repoDir, "tstow.yaml"), []byte(yamlContent), 0644)

		// Install ONLY shell
		runCmdE(binPath, append(baseArgs, "install", "shell")...)

		if !isSymlinkTo(t, filepath.Join(homeDir, ".bashrc"), filepath.Join(repoDir, "shell", ".bashrc")) {
			t.Errorf("Failed to install folder subset")
		}
		if fileExists(filepath.Join(homeDir, ".emacs.d", "init.el")) {
			t.Errorf("Installed files outside of recursive target")
		}
	})

	t.Run("Req: Force fix wrong symbolic links", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		os.MkdirAll(filepath.Join(repoDir, "shell"), 0755)
		os.WriteFile(filepath.Join(repoDir, "shell", ".bashrc"), []byte("data"), 0644)
		yamlContent := "mappings:\n  shell/.bashrc: .bashrc\n"
		os.WriteFile(filepath.Join(repoDir, "tstow.yaml"), []byte(yamlContent), 0644)

		// Create WRONG symlink
		os.Symlink("/dev/null", filepath.Join(homeDir, ".bashrc"))

		// Install without force should SKIP
		runCmdE(binPath, append(baseArgs, "install")...)
		if isSymlinkTo(t, filepath.Join(homeDir, ".bashrc"), filepath.Join(repoDir, "shell", ".bashrc")) {
			t.Errorf("Install overwrote symlink without -f")
		}

		// Install WITH force should FIX
		runCmdE(binPath, append(baseArgs, "install", "-f")...)
		if !isSymlinkTo(t, filepath.Join(homeDir, ".bashrc"), filepath.Join(repoDir, "shell", ".bashrc")) {
			t.Errorf("Force install failed to fix wrong symlink")
		}
	})

	t.Run("Req: NEVER remove regular file or dir (Safety Boundary)", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		os.MkdirAll(filepath.Join(repoDir, "shell"), 0755)
		os.WriteFile(filepath.Join(repoDir, "shell", ".bashrc"), []byte("repo data"), 0644)
		yamlContent := "mappings:\n  shell/.bashrc: .bashrc\n"
		os.WriteFile(filepath.Join(repoDir, "tstow.yaml"), []byte(yamlContent), 0644)

		// Create a REAL FILE in the way (with DIFFERING content)
		os.WriteFile(filepath.Join(homeDir, ".bashrc"), []byte("REAL FILE"), 0644)

		// Even WITH force, it must refuse
		out, _ := runCmdE(binPath, append(baseArgs, "install", "-f")...)

		if !strings.Contains(out, "differing from repo") {
			t.Errorf("Failed to enforce safety boundary. Output: %s", out)
		}

		info, _ := os.Lstat(filepath.Join(homeDir, ".bashrc"))
		if info.Mode()&os.ModeSymlink != 0 {
			t.Errorf("CRITICAL FAILURE: tstow overwrote a real file with a symlink!")
		}
	})

	t.Run("Req: Safe Replace identical files", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		os.MkdirAll(filepath.Join(repoDir, "shell"), 0755)
		os.WriteFile(filepath.Join(repoDir, "shell", ".bashrc"), []byte("identical data"), 0644)
		yamlContent := "mappings:\n  shell/.bashrc: ~/.bashrc\n"
		os.WriteFile(filepath.Join(repoDir, "tstow.yaml"), []byte(yamlContent), 0644)

		// Create an IDENTICAL REAL FILE in the way
		os.WriteFile(filepath.Join(homeDir, ".bashrc"), []byte("identical data"), 0644)

		out, _ := runCmdE(binPath, append(baseArgs, "install")...)

		if !strings.Contains(out, "Safely replacing with symlink") {
			t.Errorf("Failed to replace identical file. Output: %s", out)
		}

		if !isSymlinkTo(t, filepath.Join(homeDir, ".bashrc"), filepath.Join(repoDir, "shell", ".bashrc")) {
			t.Errorf("Identical file was not replaced by a symlink")
		}
	})

	t.Run("Req: Skip functionality", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		os.MkdirAll(filepath.Join(repoDir, "shell"), 0755)
		os.WriteFile(filepath.Join(repoDir, "shell", ".bashrc"), []byte("data"), 0644)
		yamlContent := "mappings:\n  shell/.bashrc: .bashrc\n"
		os.WriteFile(filepath.Join(repoDir, "tstow.yaml"), []byte(yamlContent), 0644)

		// Mark as skipped
		runCmdE(binPath, append(baseArgs, "skip", "shell/.bashrc")...)

		// Run install
		runCmdE(binPath, append(baseArgs, "install")...)

		if fileExists(filepath.Join(homeDir, ".bashrc")) {
			t.Errorf("tstow install did not respect the skiplist")
		}
	})

	t.Run("Req: Delete removes location pair but NOT repo file", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		localFile := filepath.Join(homeDir, ".bashrc")
		os.WriteFile(localFile, []byte("data"), 0644)
		runCmdE(binPath, append(baseArgs, "add", ".bashrc", "shell/.bashrc")...)

		// Delete
		runCmdE(binPath, append(baseArgs, "delete", "shell/.bashrc")...)

		if isSymlinkTo(t, localFile, filepath.Join(repoDir, "shell", ".bashrc")) {
			t.Errorf("Delete did not remove symlink")
		}
		if !fileExists(filepath.Join(repoDir, "shell", ".bashrc")) {
			t.Errorf("Delete wrongfully removed the repo file")
		}
		yamlData, _ := os.ReadFile(filepath.Join(repoDir, "tstow.yaml"))
		if strings.Contains(string(yamlData), "shell/.bashrc") {
			t.Errorf("Delete did not remove mapping from yaml")
		}
	})

	t.Run("Req: Undo reverts physical file and removes mapping", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		localFile := filepath.Join(homeDir, ".bashrc")
		os.WriteFile(localFile, []byte("data"), 0644)
		runCmdE(binPath, append(baseArgs, "add", ".bashrc", "shell/.bashrc")...)

		// Run undo
		runCmdE(binPath, append(baseArgs, "undo", "shell/.bashrc")...)

		// Verify symlink is gone, real file is back
		info, err := os.Lstat(localFile)
		if err != nil || info.Mode()&os.ModeSymlink != 0 {
			t.Errorf("File was not restored as a regular file")
		}

		// Verify repo file is gone
		if fileExists(filepath.Join(repoDir, "shell", ".bashrc")) {
			t.Errorf("Repo file was not moved out")
		}

		// Verify YAML is cleaned up
		yamlData, _ := os.ReadFile(filepath.Join(repoDir, "tstow.yaml"))
		if strings.Contains(string(yamlData), "shell/.bashrc") {
			t.Errorf("YAML mapping was not removed")
		}
	})

	t.Run("Req: Portable ~ paths are enforced in yaml", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		localFile := filepath.Join(homeDir, ".bashrc")
		os.WriteFile(localFile, []byte("data"), 0644)

		// Pass absolute path to add
		runCmdE(binPath, append(baseArgs, "add", localFile, "shell/.bashrc")...)

		yamlData, _ := os.ReadFile(filepath.Join(repoDir, "tstow.yaml"))
		if !strings.Contains(string(yamlData), "~/.bashrc") {
			t.Errorf("YAML did not use portable path. Got: %s", string(yamlData))
		}
		if strings.Contains(string(yamlData), homeDir) {
			t.Errorf("YAML contains hardcoded absolute path!")
		}
	})

	t.Run("Req: Autopilot .tstowrc remembers repo", func(t *testing.T) {
		homeDir, repoDir, _ := setupEnv(t) // Not using baseArgs so we can omit -r
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		// 1. Run once WITH -r to teach it the location
		runCmdE(binPath, "-i", homeDir, "-r", repoDir, "show")

		rcFile := filepath.Join(homeDir, ".tstowrc")
		if !fileExists(rcFile) {
			t.Fatalf(".tstowrc was not created")
		}

		// 2. Set up a valid repo payload
		os.MkdirAll(filepath.Join(repoDir, "shell"), 0755)
		os.WriteFile(filepath.Join(repoDir, "shell", ".bashrc"), []byte("data"), 0644)
		yamlContent := "mappings:\n  shell/.bashrc: ~/.bashrc\n"
		os.WriteFile(filepath.Join(repoDir, "tstow.yaml"), []byte(yamlContent), 0644)

		// 3. Run install WITHOUT -r
		runCmdE(binPath, "-i", homeDir, "install")

		// 4. Verify it found the repo and linked the file
		if !isSymlinkTo(t, filepath.Join(homeDir, ".bashrc"), filepath.Join(repoDir, "shell", ".bashrc")) {
			t.Errorf("Autopilot failed to use remembered repo dir")
		}
	})
}
