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
	RepoDir     string `yaml:"repo_dir"`
	PlaceFolder string `yaml:"place_folder"`
}

var (
	repoDir     string
	placeFolder string
)

func main() {
	var forcePlace bool
	var ignoreIngest bool

	var rootCmd = &cobra.Command{
		Use:   "tstow",
		Short: "Troodos Exascale dotfile manager",
		Long: `tstow is an explicit, idempotent deployment functor for dotfiles. 
It maps configuration files from a repository to a placement folder.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			isShow := cmd.Name() == "show"
			explicitPlace := cmd.Flags().Changed("place")
			explicitRepo := cmd.Flags().Changed("repo-dir")

			rcRepo, rcPlace := loadRc()

			// Resolve Place Folder
			if !explicitPlace {
				if isShow && rcPlace != "" {
					placeFolder = rcPlace
				} else {
					h, err := os.UserHomeDir()
					if err != nil {
						return err
					}
					placeFolder = h
				}
			} else {
				if placeFolder == "~" {
					h, err := os.UserHomeDir()
					if err != nil {
						return err
					}
					placeFolder = h
				} else {
					abs, err := filepath.Abs(placeFolder)
					if err != nil {
						return err
					}
					placeFolder = abs
				}
			}

			// Resolve Repo Directory
			if !explicitRepo {
				if isShow && rcRepo != "" {
					repoDir = rcRepo
				} else {
					repoDir = "."
				}
			}

			absRepo, err := filepath.Abs(repoDir)
			if err != nil {
				return err
			}
			repoDir = absRepo

			if explicitRepo || explicitPlace {
				saveRc(repoDir, placeFolder)
			}

			return nil
		},
	}

	rootCmd.PersistentFlags().StringVarP(&repoDir, "repo-dir", "r", ".", "Target repository directory")
	rootCmd.PersistentFlags().StringVarP(&placeFolder, "place", "p", "~", "Target placement folder")

	var ingestCmd = &cobra.Command{
		Use:   "ingest [-i] <repo_path> [<local_path>]",
		Short: "Ingests a file into the repo. Reads from stdin if local_path is omitted.",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			localPath := ""
			if len(args) == 2 {
				localPath = args[1]
			}
			handleIngest(loadConfig(), args[0], localPath, ignoreIngest)
		},
	}
	ingestCmd.Flags().BoolVarP(&ignoreIngest, "ignore", "i", false, "Ignore mapping: securely copies the file/dir to repo without mapping, symlinking, or deleting the source")

	var placeCmd = &cobra.Command{
		Use:   "place [repo_folder]",
		Short: "Enforces the mapping state. Optionally restrict to a specific repo folder.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			handlePlace(loadConfig(), forcePlace, args)
		},
	}
	placeCmd.Flags().BoolVarP(&forcePlace, "force", "f", false, "Fix wrong symbolic links (NEVER deletes regular files)")

	var showCmd = &cobra.Command{
		Use:   "show",
		Short: "Lists placed configurations and conflicts",
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
		Short: "Reverts an ingest: restores the physical file to the placement dir and deletes the mapping",
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

	rootCmd.AddCommand(ingestCmd, placeCmd, showCmd, skipCmd, deleteCmd, undoCmd, demoCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// -- Core Commands --

func handleIngest(cfg *Config, repoPath, localPath string, ignore bool) {
	if filepath.IsAbs(repoPath) {
		fatal("Repo path must be relative: %s", repoPath)
	}

	absRepoPath := filepath.Join(repoDir, repoPath)

	// Helper to automatically add to skip list
	addSkip := func(path string) {
		for _, s := range cfg.Skips {
			if s == path {
				return
			}
		}
		cfg.Skips = append(cfg.Skips, path)
		saveConfig(cfg)
	}

	// 1. Handle Stdin
	if localPath == "" {
		if err := os.MkdirAll(filepath.Dir(absRepoPath), 0755); err != nil {
			fatal("Failed to create repo directory: %v", err)
		}

		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			fatal("Error: Missing <local_path> and no data piped to stdin.")
		}

		outFile, err := os.Create(absRepoPath)
		if err != nil {
			fatal("Failed to create file %s: %v", absRepoPath, err)
		}
		defer outFile.Close()

		if _, err := io.Copy(outFile, os.Stdin); err != nil {
			fatal("Failed to read from stdin: %v", err)
		}

		// Treat stdin ingestions as ignored by default (no placement link)
		addSkip(repoPath)
		fmt.Printf("Ingested (from stdin): saved to %s and added to skiplist\n", repoPath)
		return
	}

	// 2. Resolve strictly relative to CWD
	absLocal := resolveCliPath(localPath)

	info, err := os.Lstat(absLocal)
	if os.IsNotExist(err) {
		fatal("Local path does not exist: %s", absLocal)
	}

	// 3. Check if it lives under the placement folder (Safety Boundary)
	isUnderPlacement := false
	if rel, err := filepath.Rel(placeFolder, absLocal); err == nil && !strings.HasPrefix(rel, "..") {
		isUnderPlacement = true
	}

	actualSource := absLocal
	isSymlink := info.Mode()&os.ModeSymlink != 0

	// 4. Resolve legacy symlinks to the real physical data
	if isSymlink {
		realPath, err := filepath.EvalSymlinks(absLocal)
		if err != nil {
			fatal("Failed to resolve symlink %s: %v", absLocal, err)
		}

		infoReal, _ := os.Stat(realPath)
		if infoReal != nil && infoReal.IsDir() {
			fmt.Printf("Migrating legacy directory symlink: %s\n", absLocal)
		} else {
			target, _ := os.Readlink(absLocal)
			if target == absRepoPath {
				fmt.Printf("Already in place: %s\n", repoPath)
				if !ignore && isUnderPlacement {
					if cfg.Mappings == nil {
						cfg.Mappings = make(map[string]string)
					}
					portableDest := makePortablePath(absLocal)
					if cfg.Mappings[repoPath] != portableDest {
						cfg.Mappings[repoPath] = portableDest
						saveConfig(cfg)
					}
				}
				return
			}
		}
		actualSource = realPath
	}

	// 5. Verify Directory Merge Validity
	sourceInfo, err := os.Stat(actualSource)
	if err != nil {
		fatal("Failed to stat source: %v", err)
	}

	if destInfo, err := os.Stat(absRepoPath); err == nil {
		if !destInfo.IsDir() || !sourceInfo.IsDir() {
			fatal("Repo path already exists and cannot be cleanly merged: %s", absRepoPath)
		}
	} else if err := os.MkdirAll(filepath.Dir(absRepoPath), 0755); err != nil {
		fatal("Failed to create repo directory: %v", err)
	}

	// 6. ALWAYS COPY. Never move the underlying asset.
	if err := copyPath(actualSource, absRepoPath); err != nil {
		fatal("Failed to copy path: %v", err)
	}

	// 7. If ignored, we add to skips, do not map, do not symlink, and DO NOT delete the source
	if ignore {
		addSkip(repoPath)
		fmt.Printf("Ingested (ignored): %s securely copied to %s and added to skiplist\n", actualSource, repoPath)
		return
	}

	// 8. Gather individual files to process (Never map a directory)
	var files []string
	if sourceInfo.IsDir() {
		err = filepath.Walk(actualSource, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !fi.IsDir() {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			fatal("Error walking directory: %v", err)
		}
	} else {
		files = append(files, actualSource)
	}

	// If we are migrating a legacy directory symlink from the placement folder,
	// delete the link and create a real directory so we can place files inside it.
	if isSymlink && sourceInfo.IsDir() && isUnderPlacement {
		os.Remove(absLocal)
		os.MkdirAll(absLocal, 0755)
	}

	if cfg.Mappings == nil {
		cfg.Mappings = make(map[string]string)
	}

	// 9. Process File-by-File Mappings
	for _, srcFile := range files {
		relPath, _ := filepath.Rel(actualSource, srcFile)

		fileRepoRel := filepath.Join(repoPath, relPath)
		fileRepoAbs := filepath.Join(absRepoPath, relPath)
		fileLocalAbs := filepath.Join(absLocal, relPath)

		// Only mutate the system if it's under the placement folder
		if isUnderPlacement {
			os.MkdirAll(filepath.Dir(fileLocalAbs), 0755)
			os.Remove(fileLocalAbs) // Safely overwrite old file or link
			createSymlink(fileRepoAbs, fileLocalAbs, false)

			portableDest := makePortablePath(fileLocalAbs)
			cfg.Mappings[fileRepoRel] = portableDest
			fmt.Printf("Ingested: %s -> %s\n", portableDest, fileRepoRel)
		} else {
			// If it's imported from OUTSIDE the placement folder, we don't automatically symlink it
			// into an external folder. We just copy it to the repo.
			fmt.Printf("Imported from external source: %s -> %s\n", srcFile, fileRepoRel)
		}
	}

	if isUnderPlacement {
		saveConfig(cfg)
	}
}

func handlePlace(cfg *Config, force bool, args []string) {
	fmt.Println("Reconciling state...")

	targetPrefix := ""
	if len(args) > 0 {
		targetPrefix = args[0]
	}

	for repoPath, destPath := range cfg.Mappings {
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

		expandedDest := expandPlacePath(destPath)
		createSymlink(absRepoPath, expandedDest, force)
	}
	fmt.Println("Placement complete.")
}

func handleUndo(cfg *Config, repoPath string) {
	destPath, exists := cfg.Mappings[repoPath]
	if !exists {
		fatal("Mapping not found for: %s", repoPath)
	}

	absRepo := filepath.Join(repoDir, repoPath)
	expandedDest := expandPlacePath(destPath)

	info, err := os.Lstat(expandedDest)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		if target, _ := os.Readlink(expandedDest); target == absRepo {
			os.Remove(expandedDest)
		}
	}

	// Restore physical data
	if err := copyPath(absRepo, expandedDest); err != nil {
		fatal("Failed to restore physical file to %s: %v", expandedDest, err)
	}

	// Clean up repo
	os.RemoveAll(absRepo)

	delete(cfg.Mappings, repoPath)
	saveConfig(cfg)

	fmt.Printf("✅ Undo complete: %s restored to %s\n", repoPath, destPath)
}

func handleShow(cfg *Config) {
	fmt.Printf("State for Placement Folder: %s\n", placeFolder)
	fmt.Printf("                Repository: %s\n\n", repoDir)

	if len(cfg.Mappings) == 0 && len(cfg.Skips) == 0 {
		fmt.Println("No mappings or skips found.")
		return
	}

	if len(cfg.Mappings) > 0 {
		fmt.Println("Mappings:")
		for repoPath, destPath := range cfg.Mappings {
			absRepoPath := filepath.Join(repoDir, repoPath)
			expandedDest := expandPlacePath(destPath)

			status := "✅ OK"
			if isSkipped(cfg, repoPath) {
				status = "⏭️  SKIPPED"
			} else if _, err := os.Stat(absRepoPath); os.IsNotExist(err) {
				status = "❌ MISSING IN REPO"
			} else {
				info, err := os.Lstat(expandedDest)
				if os.IsNotExist(err) {
					status = "⚠️  NOT PLACED"
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
			fmt.Printf("  %s\n", s)
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

	if destPath, exists := cfg.Mappings[repoPath]; exists {
		delete(cfg.Mappings, repoPath)
		modified = true
		expandedDest := expandPlacePath(destPath)
		absRepo := filepath.Join(repoDir, repoPath)

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
				return
			}
			if !force {
				fmt.Printf("[CONFLICT] %s is a symlink pointing elsewhere. Use -f to fix.\n", expandedDest)
				return
			}
			os.Remove(expandedDest)
		} else {
			if filesIdentical(absRepoPath, expandedDest) {
				fmt.Printf("[FIX] %s is identical to repo file. Safely replacing with symlink.\n", expandedDest)
				os.Remove(expandedDest)
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

func resolveCliPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(placeFolder, path[2:])
	}
	abs, err := filepath.Abs(path)
	if err == nil {
		return abs
	}
	return path
}

func makePortablePath(absProvided string) string {
	if strings.HasPrefix(absProvided, placeFolder) {
		rel, err := filepath.Rel(placeFolder, absProvided)
		if err == nil {
			return "~/" + rel
		}
	}
	return absProvided
}

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

func loadRc() (string, string) {
	h, err := os.UserHomeDir()
	if err != nil {
		return "", ""
	}
	data, err := os.ReadFile(filepath.Join(h, rcFile))
	if err == nil {
		var rc RcConfig
		if yaml.Unmarshal(data, &rc) == nil {
			return rc.RepoDir, rc.PlaceFolder
		}
	}
	return "", ""
}

func saveRc(rDir string, pDir string) {
	h, err := os.UserHomeDir()
	if err != nil {
		return
	}
	rc := RcConfig{RepoDir: rDir, PlaceFolder: pDir}
	data, err := yaml.Marshal(rc)
	if err == nil {
		os.WriteFile(filepath.Join(h, rcFile), data, 0644)
	}
}

func copyPath(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return copyDir(src, dst, info.Mode())
	}
	return copyFile(src, dst, info.Mode())
}

func copyDir(src string, dst string, mode os.FileMode) error {
	if err := os.MkdirAll(dst, mode); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			subInfo, _ := os.Stat(srcPath)
			if err := copyDir(srcPath, dstPath, subInfo.Mode()); err != nil {
				return err
			}
		} else {
			subInfo, _ := os.Stat(srcPath)
			if err := copyFile(srcPath, dstPath, subInfo.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return os.Chmod(dst, mode)
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

func expandPlacePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(placeFolder, path[2:])
	}
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(placeFolder, path)
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

func runDemo() {
	fmt.Println("🎬 Initializing tstow demonstration environment...")
	time.Sleep(1 * time.Second)
	fmt.Println("$ echo 'alias ll=\"ls -l\"' > ~/.bash_aliases")
	fmt.Println("$ tstow ingest shell/aliases ~/.bash_aliases")
	time.Sleep(1 * time.Second)
	fmt.Println("Ingested: ~/.bash_aliases -> shell/aliases")
	fmt.Println("$ tstow show")
	time.Sleep(1 * time.Second)
	fmt.Println("Mappings:")
	fmt.Println("  shell/aliases        -> ~/.bash_aliases      [✅ OK]")
	fmt.Println("\n🎬 Demo complete. Ready for asciinema recording.")
}
