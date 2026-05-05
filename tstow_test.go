package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

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
	buildDir, _ := os.MkdirTemp("", "tstow-build-*")
	defer os.RemoveAll(buildDir)
	binPath := filepath.Join(buildDir, "tstow")

	if out, err := exec.Command("go", "build", "-o", binPath, "main.go").CombinedOutput(); err != nil {
		t.Fatalf("Failed to compile tstow: %v\nOutput: %s", err, string(out))
	}

	setupEnv := func(t *testing.T) (string, string, []string) {
		homeDir, _ := os.MkdirTemp("", "tstow-home-*")
		repoDir, _ := os.MkdirTemp("", "tstow-repo-*")
		baseArgs := []string{"-p", homeDir, "-r", repoDir}

		origHome := os.Getenv("HOME")
		os.Setenv("HOME", homeDir)
		t.Cleanup(func() {
			os.Setenv("HOME", origHome)
		})

		return homeDir, repoDir, baseArgs
	}

	t.Run("Req: Ingest regular file copies it and links it", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		localFile := filepath.Join(homeDir, ".bashrc")
		os.WriteFile(localFile, []byte("echo hello"), 0644)

		args := append(baseArgs, "ingest", "shell/.bashrc", "~/.bashrc")
		out, err := runCmdE(binPath, args...)
		if err != nil {
			t.Fatalf("ingest failed: %v, output: %s", err, out)
		}

		absRepo := filepath.Join(repoDir, "shell", ".bashrc")
		if !fileExists(absRepo) {
			t.Errorf("File was not copied to repo")
		}
		if !isSymlinkTo(t, localFile, absRepo) {
			t.Errorf("Symlink not created at placement location")
		}
	})

	t.Run("Req: Ingest with ignore flag securely copies file but skips mapping and deletion", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		localFile := filepath.Join(homeDir, "btt_backup.json")
		os.WriteFile(localFile, []byte("data"), 0644)

		args := append(baseArgs, "ingest", "-i", "btt/btt_backup.json", "~/btt_backup.json")
		runCmdE(binPath, args...)

		absRepo := filepath.Join(repoDir, "btt", "btt_backup.json")
		if !fileExists(absRepo) {
			t.Errorf("Ignored file was not copied to repo")
		}

		if !fileExists(localFile) {
			t.Errorf("Local source file was wrongfully removed during ignored ingest")
		}

		yamlData, _ := os.ReadFile(filepath.Join(repoDir, "tstow.yaml"))

		// It SHOULD be in the file under 'skips', so a generic strings.Contains fails.
		// Check specifically that it didn't get added as a mapping key (with a colon).
		if strings.Contains(string(yamlData), "btt/btt_backup.json:") {
			t.Errorf("Ignored file was incorrectly added to yaml mappings")
		}

		// Ensure it actually made it to the skiplist
		if !strings.Contains(string(yamlData), "- btt/btt_backup.json") {
			t.Errorf("Ignored file was not added to the skiplist")
		}
	})

	t.Run("Req: Ingest can merge into an existing repo directory", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		localDir := filepath.Join(homeDir, "source_dir")
		os.MkdirAll(localDir, 0755)
		os.WriteFile(filepath.Join(localDir, "test.txt"), []byte("data"), 0644)

		os.MkdirAll(filepath.Join(repoDir, "repo_dir"), 0755)

		args := append(baseArgs, "ingest", "-i", "repo_dir", localDir)
		out, err := runCmdE(binPath, args...)
		if err != nil {
			t.Fatalf("Directory merge ingest failed: %v, output: %s", err, out)
		}

		if !fileExists(filepath.Join(repoDir, "repo_dir", "test.txt")) {
			t.Errorf("Failed to merge directory contents into existing repo directory")
		}
	})

	t.Run("Req: Uses relative paths ONLY in repo location", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		localFile := filepath.Join(homeDir, ".bashrc")
		os.WriteFile(localFile, []byte("data"), 0644)

		args := append(baseArgs, "ingest", "/absolute/path/in/repo", "~/.bashrc")
		out, err := runCmdE(binPath, args...)
		if err == nil || !strings.Contains(out, "must be relative") {
			t.Errorf("Expected failure for absolute repo path. Output: %s", out)
		}
	})

	t.Run("Req: Place entire placement list", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		os.MkdirAll(filepath.Join(repoDir, "shell"), 0755)
		os.WriteFile(filepath.Join(repoDir, "shell", ".bashrc"), []byte("data"), 0644)
		yamlContent := "mappings:\n  shell/.bashrc: .bashrc\n"
		os.WriteFile(filepath.Join(repoDir, "tstow.yaml"), []byte(yamlContent), 0644)

		runCmdE(binPath, append(baseArgs, "place")...)

		if !isSymlinkTo(t, filepath.Join(homeDir, ".bashrc"), filepath.Join(repoDir, "shell", ".bashrc")) {
			t.Errorf("Place failed to create symlink")
		}
	})

	t.Run("Req: Place recursive (folder subset)", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		os.MkdirAll(filepath.Join(repoDir, "shell"), 0755)
		os.MkdirAll(filepath.Join(repoDir, "emacs"), 0755)
		os.WriteFile(filepath.Join(repoDir, "shell", ".bashrc"), []byte("data"), 0644)
		os.WriteFile(filepath.Join(repoDir, "emacs", "init.el"), []byte("data"), 0644)

		yamlContent := "mappings:\n  shell/.bashrc: .bashrc\n  emacs/init.el: .emacs.d/init.el\n"
		os.WriteFile(filepath.Join(repoDir, "tstow.yaml"), []byte(yamlContent), 0644)

		runCmdE(binPath, append(baseArgs, "place", "shell")...)

		if !isSymlinkTo(t, filepath.Join(homeDir, ".bashrc"), filepath.Join(repoDir, "shell", ".bashrc")) {
			t.Errorf("Failed to place folder subset")
		}
		if fileExists(filepath.Join(homeDir, ".emacs.d", "init.el")) {
			t.Errorf("Placed files outside of recursive target")
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

		os.Symlink("/dev/null", filepath.Join(homeDir, ".bashrc"))

		runCmdE(binPath, append(baseArgs, "place")...)
		if isSymlinkTo(t, filepath.Join(homeDir, ".bashrc"), filepath.Join(repoDir, "shell", ".bashrc")) {
			t.Errorf("Place overwrote symlink without -f")
		}

		runCmdE(binPath, append(baseArgs, "place", "-f")...)
		if !isSymlinkTo(t, filepath.Join(homeDir, ".bashrc"), filepath.Join(repoDir, "shell", ".bashrc")) {
			t.Errorf("Force place failed to fix wrong symlink")
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

		os.WriteFile(filepath.Join(homeDir, ".bashrc"), []byte("REAL FILE"), 0644)

		out, _ := runCmdE(binPath, append(baseArgs, "place", "-f")...)

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

		os.WriteFile(filepath.Join(homeDir, ".bashrc"), []byte("identical data"), 0644)

		out, _ := runCmdE(binPath, append(baseArgs, "place")...)

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

		runCmdE(binPath, append(baseArgs, "skip", "shell/.bashrc")...)
		runCmdE(binPath, append(baseArgs, "place")...)

		if fileExists(filepath.Join(homeDir, ".bashrc")) {
			t.Errorf("tstow place did not respect the skiplist")
		}
	})

	t.Run("Req: Delete removes location pair but NOT repo file", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		localFile := filepath.Join(homeDir, ".bashrc")
		os.WriteFile(localFile, []byte("data"), 0644)
		runCmdE(binPath, append(baseArgs, "ingest", "shell/.bashrc", "~/.bashrc")...)

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

	t.Run("Req: Undo restores physical file and removes mapping", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		localFile := filepath.Join(homeDir, ".bashrc")
		os.WriteFile(localFile, []byte("data"), 0644)
		runCmdE(binPath, append(baseArgs, "ingest", "shell/.bashrc", "~/.bashrc")...)

		runCmdE(binPath, append(baseArgs, "undo", "shell/.bashrc")...)

		info, err := os.Lstat(localFile)
		if err != nil || info.Mode()&os.ModeSymlink != 0 {
			t.Errorf("File was not restored as a regular file")
		}

		if fileExists(filepath.Join(repoDir, "shell", ".bashrc")) {
			t.Errorf("Repo file was not removed during undo")
		}

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

		runCmdE(binPath, append(baseArgs, "ingest", "shell/.bashrc", "~/.bashrc")...)

		yamlData, _ := os.ReadFile(filepath.Join(repoDir, "tstow.yaml"))
		if !strings.Contains(string(yamlData), "~/.bashrc") {
			t.Errorf("YAML did not use portable path. Got: %s", string(yamlData))
		}
		if strings.Contains(string(yamlData), homeDir) {
			t.Errorf("YAML contains hardcoded absolute path!")
		}
	})

	t.Run("Req: Inspection uses .tstowrc, but Mutation requires explicit scope", func(t *testing.T) {
		homeDir, repoDir, _ := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		runCmdE(binPath, "-p", homeDir, "-r", repoDir, "show")

		rcFile := filepath.Join(homeDir, ".tstowrc")
		if !fileExists(rcFile) {
			t.Fatalf(".tstowrc was not created by explicit flags")
		}

		os.MkdirAll(filepath.Join(repoDir, "shell"), 0755)
		os.WriteFile(filepath.Join(repoDir, "shell", ".bashrc"), []byte("data"), 0644)
		yamlContent := "mappings:\n  shell/.bashrc: ~/.bashrc\n"
		os.WriteFile(filepath.Join(repoDir, "tstow.yaml"), []byte(yamlContent), 0644)

		runCmdE(binPath, "place")

		if isSymlinkTo(t, filepath.Join(homeDir, ".bashrc"), filepath.Join(repoDir, "shell", ".bashrc")) {
			t.Errorf("Mutation command dangerously used RC state!")
		}

		out, _ := runCmdE(binPath, "show")
		if !strings.Contains(out, "shell/.bashrc") {
			t.Errorf("Show command failed to use RC state for inspection")
		}
	})

	t.Run("Req: Ingest directory explicitly maps files, not the directory itself", func(t *testing.T) {
		homeDir, repoDir, baseArgs := setupEnv(t)
		defer os.RemoveAll(homeDir)
		defer os.RemoveAll(repoDir)

		localDir := filepath.Join(homeDir, ".config", "myapp")
		os.MkdirAll(localDir, 0755)
		os.WriteFile(filepath.Join(localDir, "config.yml"), []byte("data"), 0644)
		os.WriteFile(filepath.Join(localDir, "hooks.sh"), []byte("echo run"), 0755)

		args := append(baseArgs, "ingest", "config/myapp", "~/.config/myapp")
		runCmdE(binPath, args...)

		if !fileExists(filepath.Join(repoDir, "config", "myapp", "config.yml")) {
			t.Errorf("File inside directory was not copied to repo")
		}

		info, _ := os.Lstat(localDir)
		if info.Mode()&os.ModeSymlink != 0 {
			t.Errorf("tstow wrongfully symlinked the directory itself")
		}

		if !isSymlinkTo(t, filepath.Join(localDir, "config.yml"), filepath.Join(repoDir, "config", "myapp", "config.yml")) {
			t.Errorf("File inside directory was not replaced with a symlink")
		}

		yamlData, _ := os.ReadFile(filepath.Join(repoDir, "tstow.yaml"))
		if !strings.Contains(string(yamlData), "config/myapp/config.yml") {
			t.Errorf("YAML did not map individual files")
		}
	})
}
