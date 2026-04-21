package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const mappingFile = "gs.yaml"

type Config map[string]string // repo_path: dest_path

var homeDir string

func main() {
	var forceInstall bool

	var rootCmd = &cobra.Command{
		Use:   "gstow",
		Short: "An explicit, idempotent deployment functor for dotfiles",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if homeDir == "" {
				h, err := os.UserHomeDir()
				if err != nil {
					return err
				}
				homeDir = h
			} else {
				abs, err := filepath.Abs(homeDir)
				if err != nil {
					return err
				}
				homeDir = abs
			}
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVarP(&homeDir, "dir", "d", "", "Override target home directory for ~ expansion (default is OS home)")

	var addCmd = &cobra.Command{
		Use:   "add <local_path> <repo_path>",
		Short: "Moves a local file/dir to the repo, adds mapping, and symlinks it",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			handleAdd(loadConfig(), args[0], args[1])
		},
	}

	var installCmd = &cobra.Command{
		Use:   "install",
		Short: "Enforces the mapping state (creates dirs, links files)",
		Run: func(cmd *cobra.Command, args []string) {
			handleInstall(loadConfig(), forceInstall)
		},
	}
	installCmd.Flags().BoolVarP(&forceInstall, "force", "f", false, "Force overwrite of existing files/symlinks")

	var rmlocCmd = &cobra.Command{
		Use:   "rmloc <repo_path>",
		Short: "Removes a mapping and its symlink (repo file remains intact)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			handleRmloc(loadConfig(), args[0])
		},
	}

	var showCmd = &cobra.Command{
		Use:   "show",
		Short: "Lists all current file mappings",
		Run: func(cmd *cobra.Command, args []string) {
			handleShow(loadConfig())
		},
	}

	rootCmd.AddCommand(addCmd, installCmd, rmlocCmd, showCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// -- Adjoint: Update Repo & State Mapping --

func handleAdd(cfg Config, localPath, repoPath string) {
	localExpanded := expandPath(localPath)

	info, err := os.Stat(localExpanded)
	if os.IsNotExist(err) {
		fatal("Local path does not exist: %s", localExpanded)
	}
	
	if info.IsDir() {
		fmt.Println("⚠️  WARNING: You are adding an entire directory.")
		fmt.Println("   This is safe for pure scripts (e.g., ~/scripts), but doing this")
		fmt.Println("   for app configs (e.g., Cursor, Emacs) will cause cache pollution.")
	}

	if _, err := os.Stat(repoPath); err == nil {
		fatal("Repo path already exists: %s", repoPath)
	}

	if err := os.MkdirAll(filepath.Dir(repoPath), 0755); err != nil {
		fatal("Failed to create repo directory: %v", err)
	}
	if err := os.Rename(localExpanded, repoPath); err != nil {
		fatal("Failed to move path to repo: %v", err)
	}

	cfg[repoPath] = localPath
	saveConfig(cfg)

	createSymlink(repoPath, localPath, false)
	fmt.Printf("Added: %s -> %s\n", localPath, repoPath)
}

// -- Deployment Functor: Enforce Desired State --

func handleInstall(cfg Config, force bool) {
	fmt.Println("Reconciling state...")
	for repoPath, destPath := range cfg {
		createSymlink(repoPath, destPath, force)
	}
	fmt.Println("Install complete.")
}

// -- Rmloc: Remove Mapping --

func handleRmloc(cfg Config, repoPath string) {
	destPath, exists := cfg[repoPath]
	if !exists {
		fatal("Mapping not found in %s for: %s", mappingFile, repoPath)
	}

	delete(cfg, repoPath)
	saveConfig(cfg)

	expandedDest := expandPath(destPath)
	if isSymlink(expandedDest) {
		os.Remove(expandedDest)
		fmt.Printf("Removed mapping and symlink for: %s\n", destPath)
	} else {
		fmt.Printf("Removed mapping. Target at %s was not a symlink.\n", destPath)
	}
	fmt.Printf("Repo file '%s' remains intact.\n", repoPath)
}

// -- Show: View State --

func handleShow(cfg Config) {
	if len(cfg) == 0 {
		fmt.Println("No mappings found. Use 'gstow add' to track files.")
		return
	}
	fmt.Println("Current Repository Mappings:")
	for repoPath, destPath := range cfg {
		fmt.Printf("  %s  ->  %s\n", repoPath, destPath)
	}
}

// -- Core Mechanics --

func createSymlink(repoPath, destPath string, force bool) {
	absRepoPath, _ := filepath.Abs(repoPath)
	expandedDest := expandPath(destPath)
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
				fmt.Printf("[SKIP] %s is a symlink pointing elsewhere. Use -f to overwrite.\n", destPath)
				return
			}
			os.Remove(expandedDest)
		} else {
			if !force {
				fmt.Printf("[SKIP] %s is a real file/dir. Use -f to overwrite.\n", destPath)
				return
			}
			os.RemoveAll(expandedDest)
		}
	}

	if err := os.Symlink(absRepoPath, expandedDest); err != nil {
		fmt.Printf("[ERROR] Failed to link %s: %v\n", destPath, err)
	} else {
		fmt.Printf("[OK] Linked: %s\n", destPath)
	}
}

// -- Utilities --

func loadConfig() Config {
	cfg := make(Config)
	data, err := os.ReadFile(mappingFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg
		}
		fatal("Error reading %s: %v", mappingFile, err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fatal("Invalid YAML in %s: %v", mappingFile, err)
	}
	return cfg
}

func saveConfig(cfg Config) {
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		fatal("Failed to serialize config: %v", err)
	}
	if err := os.WriteFile(mappingFile, data, 0644); err != nil {
		fatal("Failed to write %s: %v", mappingFile, err)
	}
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		return filepath.Join(homeDir, path[1:])
	}
	return path
}

func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode()&os.ModeSymlink != 0
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
