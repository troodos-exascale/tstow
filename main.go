package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const mappingFile = "tstow.yaml"

type Config struct {
	Mappings map[string]string `yaml:"mappings"`
	Skips    []string          `yaml:"skips,omitempty"`
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

			// Resolve repository directory
			absRepo, err := filepath.Abs(repoDir)
			if err != nil {
				return err
			}
			repoDir = absRepo

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
		Short: "Removes a location pair or skip rule",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			handleDelete(loadConfig(), args[0])
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

	rootCmd.AddCommand(addCmd, installCmd, showCmd, skipCmd, deleteCmd, demoCmd)

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
	if err := os.Rename(localExpanded, absRepoPath); err != nil {
		fatal("Failed to move path to repo: %v", err)
	}

	if cfg.Mappings == nil {
		cfg.Mappings = make(map[string]string)
	}
	cfg.Mappings[repoPath] = localPath
	saveConfig(cfg)

	createSymlink(absRepoPath, localExpanded, false)
	fmt.Printf("Added: %s -> %s\n", localPath, repoPath)
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
			// RULE: NEVER remove regular file or dir
			fmt.Printf("[FATAL CONFLICT] %s is a regular file/dir. tstow will NEVER remove it. Move it manually.\n", expandedDest)
			return
		}
	}

	if err := os.Symlink(absRepoPath, expandedDest); err != nil {
		fmt.Printf("[ERROR] Failed to link %s: %v\n", expandedDest, err)
	} else {
		fmt.Printf("[OK] Linked: %s\n", expandedDest)
	}
}

// -- Utilities --

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
