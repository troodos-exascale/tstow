package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const mappingFile = "tstow.yaml"
const rcFile = ".tstowrc"

type Config struct {
	Mappings map[string]string `yaml:"mappings"`
	Skips    []string          `yaml:"skips,omitempty"`
}

type RcConfig struct {
	RepoDir string `yaml:"repo_dir"`
}

var (
	repoDir       string
	installFolder string
)

func main() {
	var forceInstall bool

	var rootCmd = &cobra.Command{
		Use:   "tstow",
		Short: "Troodos Exascale dotfile manager",
		Long: `tstow is an explicit, idempotent deployment functor for dotfiles. 
It maps configuration files from a repository to an installation folder.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Resolve installation folder
			if installFolder == "" || installFolder == "~" {
				h, err := os.UserHomeDir()
				if err != nil {
					return err
				}
				installFolder = h
			} else {
				abs, err := filepath.Abs(installFolder)
				if err != nil {
					return err
				}
				installFolder = abs
			}

			// RC / Autopilot Logic for Repo Directory
			if !cmd.Flags().Changed("repo-dir") {
				if saved := loadRc(); saved != "" {
					repoDir = saved
				}
			}

			// Resolve absolute repo directory and memorize it
			absRepo, err := filepath.Abs(repoDir)
			if err != nil {
				return err
			}
			repoDir = absRepo
			saveRc(repoDir)

			return nil
		},
	}

	rootCmd.PersistentFlags().StringVarP(&repoDir, "repo-dir", "r", ".", "Target repository directory")
	rootCmd.PersistentFlags().StringVarP(&installFolder, "install-folder", "i", "~", "Target installation folder")

	var addCmd = &cobra.Command{
		Use:   "add <local_path> <repo_path>",
		Short: "Moves a regular file/dir to the repo, adds mapping, and symlinks it",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			handleAdd(loadConfig(), args[0], args[1])
		},
	}

	var installCmd = &cobra.Command{
		Use:   "install [repo_folder]",
		Short: "Enforces the mapping state. Optionally restrict to a specific repo folder.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			handleInstall(loadConfig(), forceInstall, args)
		},
	}
	installCmd.Flags().BoolVarP(&forceInstall, "force", "f", false, "Fix wrong symbolic links (NEVER deletes regular files)")

	var showCmd = &cobra.Command{
		Use:   "show",
		Short: "Lists installed configurations and conflicts",
		Run: func(cmd *cobra.Command, args []string) {
			handleShow(loadConfig())
		},
	}

	var skipCmd = &cobra.Command{
		Use:   "skip <repo_path>",
		Short: "Adds a repository file/folder to the skip list",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			handleSkip(loadConfig(), args[0])
		},
	}

	var deleteCmd = &cobra.Command{
		Use:   "delete <repo_path>",
		Short: "Removes a location pair or skip rule (leaves repo file intact)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			handleDelete(loadConfig(), args[0])
		},
	}

	var undoCmd = &cobra.Command{
		Use:   "undo <repo_path>",
		Short: "Reverts an add: moves the physical file back to the install dir and deletes the mapping",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			handleUndo(loadConfig(), args[0])
		},
	}

	var demoCmd = &cobra.Command{
		Use:    "demo",
		Short:  "Runs a visual demonstration of tstow's capabilities",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			runDemo()
		},
	}

	rootCmd.AddCommand(addCmd, installCmd, showCmd, skipCmd, deleteCmd, undoCmd, demoCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// -- Core Commands --

func handleAdd(cfg *Config, localPath, repoPath string) {
	if filepath.IsAbs(repoPath) {
		fatal("Repo path must be relative: %s", repoPath)
	}

	localExpanded := expandInstallPath(localPath)
	portableDest := makePortablePath(localPath) // Protect YAML from absolute shell expansion
	absRepoPath := filepath.Join(repoDir, repoPath)

	info, err := os.Lstat(localExpanded)
	if os.IsNotExist(err) {
		fatal("Local path does not exist: %s", localExpanded)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		fatal("Cannot add a symbolic link to the repository.")
	}

	if _, err := os.Stat(absRepoPath); err == nil {
		fatal("Repo path already exists: %s", absRepoPath)
	}

	if err := os.MkdirAll(filepath.Dir(absRepoPath), 0755); err != nil {
		fatal("Failed to create repo directory: %v", err)
	}

	if err := movePath(localExpanded, absRepoPath); err != nil {
		fatal("Failed to move path across filesystems: %v", err)
	}

	if cfg.Mappings == nil {
		cfg.Mappings = make(map[string]string)
	}
	cfg.Mappings[repoPath] = portableDest
	saveConfig(cfg)

	createSymlink(absRepoPath, localExpanded, false)
	fmt.Printf("Added: %s -> %s\n", portableDest, repoPath)
}

func handleInstall(cfg *Config, force bool, args []string) {
	fmt.Println("Reconciling state...")

	targetPrefix := ""
	if len(args) > 0 {
		targetPrefix = args[0]
	}

	for repoPath, destPath := range cfg.Mappings {
		// Recursive match: exact match OR is a subfile of the requested folder
		if targetPrefix != "" && repoPath != targetPrefix && !strings.HasPrefix(repoPath, targetPrefix+"/") {
			continue
		}

		if isSkipped(cfg, repoPath) {
			fmt.Printf("[SKIP] %s is in the skiplist.\n", repoPath)
			continue
		}

		absRepoPath := filepath.Join(repoDir, repoPath)
		if _, err := os.Stat(absRepoPath); os.IsNotExist(err) {
			fmt.Printf("[WARNING] Repo file missing: %s\n", absRepoPath)
			continue
		}

		expandedDest := expandInstallPath(destPath)
		createSymlink(absRepoPath, expandedDest, force)
	}
	fmt.Println("Install complete.")
}

func handleUndo(cfg *Config, repoPath string) {
	destPath, exists := cfg.Mappings[repoPath]
	if !exists {
		fatal("Mapping not found for: %s", repoPath)
	}

	absRepo := filepath.Join(repoDir, repoPath)
	expandedDest := expandInstallPath(destPath)

	// Sever symlink if it belongs to us
	info, err := os.Lstat(expandedDest)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		if target, _ := os.Readlink(expandedDest); target == absRepo {
			os.Remove(expandedDest)
		}
	}

	// Move the real file back from the repo
	if err := movePath(absRepo, expandedDest); err != nil {
		fatal("Failed to restore physical file to %s: %v", expandedDest, err)
	}

	// Clean up YAML
	delete(cfg.Mappings, repoPath)
	saveConfig(cfg)

	fmt.Printf("✅ Undo complete: %s moved back to %s\n", repoPath, destPath)
}

func handleShow(cfg *Config) {
	fmt.Printf("State for Install Folder: %s\n", installFolder)
	fmt.Printf("          Repository: %s\n\n", repoDir)

	if len(cfg.Mappings) == 0 && len(cfg.Skips) == 0 {
		fmt.Println("No mappings or skips found.")
		return
	}

	if len(cfg.Mappings) > 0 {
		fmt.Println("Mappings:")
		for repoPath, destPath := range cfg.Mappings {
			absRepoPath := filepath.Join(repoDir, repoPath)
			expandedDest := expandInstallPath(destPath)

			status := "✅ OK"
			if isSkipped(cfg, repoPath) {
				status = "⏭️  SKIPPED"
			} else if _, err := os.Stat(absRepoPath); os.IsNotExist(err) {
				status = "❌ MISSING IN REPO"
			} else {
				info, err := os.Lstat(expandedDest)
				if os.IsNotExist(err) {
					status = "⚠️  NOT INSTALLED"
				} else if info.Mode()&os.ModeSymlink != 0 {
					target, _ := os.Readlink(expandedDest)
					if target != absRepoPath {
						status = "❌ WRONG SYMLINK"
					}
				} else {
					status = "❌ CONFLICT (Real File)"
				}
			}
			fmt.Printf("  %-20s -> %-20s [%s]\n", repoPath, destPath, status)
		}
	}

	if len(cfg.Skips) > 0 {
		fmt.Println("\nSkiplist:")
		for _, s := range cfg.Skips {
			fmt.Printf("  - %s\n", s)
		}
	}
}

func handleSkip(cfg *Config, repoPath string) {
	for _, s := range cfg.Skips {
		if s == repoPath {
			fmt.Println("Already in skiplist.")
			return
		}
	}
	cfg.Skips = append(cfg.Skips, repoPath)
	saveConfig(cfg)
	fmt.Printf("Added to skiplist: %s\n", repoPath)
}

func handleDelete(cfg *Config, repoPath string) {
	modified := false

	// Remove from mappings
	if destPath, exists := cfg.Mappings[repoPath]; exists {
		delete(cfg.Mappings, repoPath)
		modified = true
		expandedDest := expandInstallPath(destPath)
		absRepo := filepath.Join(repoDir, repoPath)

		// Only remove the symlink if it points to OUR repo file
		info, err := os.Lstat(expandedDest)
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			if target, _ := os.Readlink(expandedDest); target == absRepo {
				os.Remove(expandedDest)
				fmt.Printf("Removed mapping and symlink for: %s\n", destPath)
			}
		} else {
			fmt.Printf("Removed mapping. Target at %s was not our symlink.\n", destPath)
		}
	}

	// Remove from skips
	newSkips := []string{}
	for _, s := range cfg.Skips {
		if s != repoPath {
			newSkips = append(newSkips, s)
		} else {
			modified = true
			fmt.Printf("Removed from skiplist: %s\n", repoPath)
		}
	}
	cfg.Skips = newSkips

	if modified {
		saveConfig(cfg)
	} else {
		fmt.Printf("Path not found in mappings or skips: %s\n", repoPath)
	}
}

// -- Core Mechanics --

func createSymlink(absRepoPath, expandedDest string, force bool) {
	parentDir := filepath.Dir(expandedDest)

	info, err := os.Stat(parentDir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				fatal("Failed to create parent dir %s: %v", parentDir, err)
			}
		} else {
			fatal("Error checking parent dir %s: %v", parentDir, err)
		}
	} else if !info.IsDir() {
		fatal("Target parent %s exists but is not a directory", parentDir)
	}

	info, err = os.Lstat(expandedDest)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			target, _ := os.Readlink(expandedDest)
			if target == absRepoPath {
				return // Idempotent success
			}
			if !force {
				fmt.Printf("[CONFLICT] %s is a symlink pointing elsewhere. Use -f to fix.\n", expandedDest)
				return
			}
			os.Remove(expandedDest) // Safe to remove wrong symlinks
		} else {
			// IDENTICAL FILE RESOLUTION (Feature #4)
			if filesIdentical(absRepoPath, expandedDest) {
				fmt.Printf("[FIX] %s is identical to repo file. Safely replacing with symlink.\n", expandedDest)
				os.Remove(expandedDest) // Safe to delete since it's identical!
			} else {
				fmt.Printf("[FATAL CONFLICT] %s is a regular file/dir differing from repo. Move it manually.\n", expandedDest)
				return
			}
		}
	}

	if err := os.Symlink(absRepoPath, expandedDest); err != nil {
		fmt.Printf("[ERROR] Failed to link %s: %v\n", expandedDest, err)
	} else {
		fmt.Printf("[OK] Linked: %s\n", expandedDest)
	}
}

// -- Utilities --

// makePortablePath prevents the shell from hardcoding absolute paths into tstow.yaml
func makePortablePath(provided string) string {
	if strings.HasPrefix(provided, "~/") {
		return provided
	}
	absProvided, err := filepath.Abs(provided)
	if err == nil && strings.HasPrefix(absProvided, installFolder) {
		rel, err := filepath.Rel(installFolder, absProvided)
		if err == nil {
			return "~/" + rel
		}
	}
	return provided
}

// filesIdentical compares two files byte-by-byte
func filesIdentical(file1, file2 string) bool {
	f1, err := os.ReadFile(file1)
	if err != nil {
		return false
	}
	f2, err := os.ReadFile(file2)
	if err != nil {
		return false
	}
	return bytes.Equal(f1, f2)
}

func loadRc() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(h, rcFile))
	if err == nil {
		var rc RcConfig
		if yaml.Unmarshal(data, &rc) == nil {
			return rc.RepoDir
		}
	}
	return ""
}

func saveRc(rDir string) {
	h, err := os.UserHomeDir()
	if err != nil {
		return
	}
	rc := RcConfig{RepoDir: rDir}
	data, err := yaml.Marshal(rc)
	if err == nil {
		os.WriteFile(filepath.Join(h, rcFile), data, 0644)
	}
}

func movePath(src, dst string) error {
	// Try the fast, kernel-level rename first
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// If it fails (likely an EXDEV cross-device link error), fallback to copy+delete
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		sourceFile.Close()
		return err
	}

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		sourceFile.Close()
		destFile.Close()
		return err
	}

	if info, err := os.Stat(src); err == nil {
		destFile.Chmod(info.Mode())
	}

	sourceFile.Close()
	destFile.Close()

	return os.Remove(src)
}

func loadConfig() *Config {
	cfg := &Config{Mappings: make(map[string]string), Skips: []string{}}
	path := filepath.Join(repoDir, mappingFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg
		}
		fatal("Error reading %s: %v", path, err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		fatal("Invalid YAML in %s: %v", path, err)
	}
	return cfg
}

func saveConfig(cfg *Config) {
	path := filepath.Join(repoDir, mappingFile)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		fatal("Failed to serialize config: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		fatal("Failed to write %s: %v", path, err)
	}
}

func expandInstallPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(installFolder, path[2:])
	}
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(installFolder, path)
}

func isSkipped(cfg *Config, repoPath string) bool {
	for _, s := range cfg.Skips {
		if repoPath == s || strings.HasPrefix(repoPath, s+"/") {
			return true
		}
	}
	return false
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// -- Demo Mode --

func runDemo() {
	fmt.Println("🎬 Initializing tstow demonstration environment...")
	time.Sleep(1 * time.Second)
	fmt.Println("$ echo 'alias ll=\"ls -l\"' > ~/.bash_aliases")
	fmt.Println("$ tstow add ~/.bash_aliases shell/aliases")
	time.Sleep(1 * time.Second)
	fmt.Println("Added: ~/.bash_aliases -> shell/aliases")
	fmt.Println("$ tstow show")
	time.Sleep(1 * time.Second)
	fmt.Println("Mappings:")
	fmt.Println("  shell/aliases        -> ~/.bash_aliases      [✅ OK]")
	fmt.Println("\n🎬 Demo complete. Ready for asciinema recording.")
}
