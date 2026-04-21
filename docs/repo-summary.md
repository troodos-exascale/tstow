# Repository Summary: /Users/braam/sw/gstow
* **Date:** Tue, 21 Apr 2026 14:13:33 UTC

FILE: .gitignore
# Build output
gstow
/bin/

# Go artifacts
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test artifacts (if generated outside tmp)
*.test
out/

# OS artifacts
.DS_Store

# Note: We do NOT ignore gs.yaml, as checking that file
# into git is the entire point of the system.
END: .gitignore

FILE: COPYRIGHT
Copyright 2026 troodos-exascale.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
END: COPYRIGHT

FILE: LICENSE
Apache License
Version 2.0, January 2004
http://www.apache.org/licenses/

TERMS AND CONDITIONS FOR USE, REPRODUCTION, AND DISTRIBUTION

1. Definitions.

   "License" shall mean the terms and conditions for use, reproduction,
   and distribution as defined by Sections 1 through 9 of this document.

   "Licensor" shall mean the copyright owner or entity authorized by
   the copyright owner that is granting the License.

   "Legal Entity" shall mean the union of the acting entity and all
   other entities that control, are controlled by, or are under common
   control with that entity. For the purposes of this definition,
   "control" means (i) the power, direct or indirect, to cause the
   direction or management of such entity, whether by contract or
   otherwise, or (ii) ownership of fifty percent (50%) or more of the
   outstanding shares, or (iii) beneficial ownership of such entity.

   "You" (or "Your") shall mean an individual or Legal Entity
   exercising permissions granted by this License.

   "Source" form shall mean the preferred form for making modifications,
   including but not limited to software source code, documentation
   source, and configuration files.

   "Object" form shall mean any form resulting from mechanical
   transformation or translation of a Source form, including but
   not limited to compiled object code, generated documentation,
   and conversions to other media types.

   "Work" shall mean the work of authorship, whether in Source or
   Object form, made available under the License, as indicated by a
   copyright notice that is included in or attached to the work
   (an example is provided in the Appendix below).

   "Derivative Works" shall mean any work, whether in Source or Object
   form, that is based on (or derived from) the Work and for which the
   editorial revisions, annotations, elaborations, or other modifications
   represent, as a whole, an original work of authorship. For the purposes
   of this License, Derivative Works shall not include works that remain
   separable from, or merely link (or bind by name) to the interfaces of,
   the Work and Derivative Works thereof.

   "Contribution" shall mean any work of authorship, including
   the original version of the Work and any modifications or additions
   to that Work or Derivative Works thereof, that is intentionally
   submitted to Licensor for inclusion in the Work by the copyright owner
   or by an individual or Legal Entity authorized to submit on behalf of
   the copyright owner. For the purposes of this definition, "submitted"
   means any form of electronic, verbal, or written communication sent
   to the Licensor or its representatives, including but not limited to
   communication on electronic mailing lists, source code control systems,
   and issue tracking systems that are managed by, or on behalf of, the
   Licensor for the purpose of discussing and improving the Work, but
   excluding communication that is conspicuously marked or otherwise
   designated in writing by the copyright owner as "Not a Contribution."

   "Contributor" shall mean Licensor and any individual or Legal Entity
   on behalf of whom a Contribution has been received by Licensor and
   subsequently incorporated within the Work.

2. Grant of Copyright License. Subject to the terms and conditions of
   this License, each Contributor hereby grants to You a perpetual,
   worldwide, non-exclusive, no-charge, royalty-free, irrevocable
   copyright license to reproduce, prepare Derivative Works of,
   publicly display, publicly perform, sublicense, and distribute the
   Work and such Derivative Works in Source or Object form.

3. Grant of Patent License. Subject to the terms and conditions of
   this License, each Contributor hereby grants to You a perpetual,
   worldwide, non-exclusive, no-charge, royalty-free, irrevocable
   (except as stated in this section) patent license to make, have made,
   use, offer to sell, sell, import, and otherwise transfer the Work,
   where such license applies only to those patent claims licensable
   by such Contributor that are necessarily infringed by their
   Contribution(s) alone or by combination of their Contribution(s)
   with the Work to which such Contribution(s) was submitted. If You
   institute patent litigation against any entity (including a
   cross-claim or counterclaim in a lawsuit) alleging that the Work
   or a Contribution incorporated within the Work constitutes direct
   or contributory patent infringement, then any patent licenses
   granted to You under this License for that Work shall terminate
   as of the date such litigation is filed.

4. Redistribution. You may reproduce and distribute copies of the
   Work or Derivative Works thereof in any medium, with or without
   modifications, and in Source or Object form, provided that You
   meet the following conditions:

   (a) You must give any other recipients of the Work or
       Derivative Works a copy of this License; and

   (b) You must cause any modified files to carry prominent notices
       stating that You changed the files; and

   (c) You must retain, in the Source form of any Derivative Works
       that You distribute, all copyright, patent, trademark, and
       attribution notices from the Source form of the Work,
       excluding those notices that do not pertain to any part of
       the Derivative Works; and

   (d) If the Work includes a "NOTICE" text file as part of its
       distribution, then any Derivative Works that You distribute must
       include a readable copy of the attribution notices contained
       within such NOTICE file, excluding those notices that do not
       pertain to any part of the Derivative Works, in at least one
       of the following places: within a NOTICE text file distributed
       as part of the Derivative Works; within the Source form or
       documentation, if provided along with the Derivative Works; or,
       within a display generated by the Derivative Works, if and
       wherever such third-party notices normally appear. The contents
       of the NOTICE file are for informational purposes only and
       do not modify the License. You may add Your own attribution
       notices within Derivative Works that You distribute, alongside
       or as an addendum to the NOTICE text from the Work, provided
       that such additional attribution notices cannot be construed
       as modifying the License.

   You may add Your own copyright statement to Your modifications and
   may provide additional or different license terms and conditions
   for use, reproduction, or distribution of Your modifications, or
   for any such Derivative Works as a whole, provided Your use,
   reproduction, and distribution of the Work otherwise complies with
   the conditions stated in this License.

5. Submission of Contributions. Unless You explicitly state otherwise,
   any Contribution intentionally submitted for inclusion in the Work
   by You to the Licensor shall be under the terms and conditions of
   this License, without any additional terms or conditions.
   Notwithstanding the above, nothing herein shall supersede or modify
   the terms of any separate license agreement you may have executed
   with Licensor regarding such Contributions.

6. Trademarks. This License does not grant permission to use the trade
   names, trademarks, service marks, or product names of the Licensor,
   except as required for reasonable and customary use in describing the
   origin of the Work and reproducing the content of the NOTICE file.

7. Disclaimer of Warranty. Unless required by applicable law or
   agreed to in writing, Licensor provides the Work (and each
   Contributor provides its Contributions) on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
   implied, including, without limitation, any warranties or conditions
   of TITLE, NON-INFRINGEMENT, MERCHANTABILITY, or FITNESS FOR A
   PARTICULAR PURPOSE. You are solely responsible for determining the
   appropriateness of using or redistributing the Work and assume any
   risks associated with Your exercise of permissions under this License.

8. Limitation of Liability. In no event and under no legal theory,
   whether in tort (including negligence), contract, or otherwise,
   unless required by applicable law (such as deliberate and grossly
   negligent acts) or agreed to in writing, shall any Contributor be
   liable to You for damages, including any direct, indirect, special,
   incidental, or consequential damages of any character arising as a
   result of this License or out of the use or inability to use the
   Work (including but not limited to damages for loss of goodwill,
   work stoppage, computer failure or malfunction, or any and all
   other commercial damages or losses), even if such Contributor
   has been advised of the possibility of such damages.

9. Accepting Warranty or Additional Liability. While redistributing
   the Work or Derivative Works thereof, You may choose to offer,
   and charge a fee for, acceptance of support, warranty, indemnity,
   or other liability obligations and/or rights consistent with this
   License. However, in accepting such obligations, You may act only
   on Your own behalf and on Your sole responsibility, not on behalf
   of any other Contributor, and only if You agree to indemnify,
   defend, and hold each Contributor harmless for any liability
   incurred by, or claims asserted against, such Contributor by reason
   of your accepting any such warranty or additional liability.

END OF TERMS AND CONDITIONS
END: LICENSE

FILE: Makefile
# Makefile for gstow

BINARY_NAME := gstow
INSTALL_DIR ?= /usr/local/bin

# Detect macOS vs Linux for the correct bash configuration file
BASH_FILE := $(if $(wildcard $(HOME)/.bash_profile),$(HOME)/.bash_profile,$(HOME)/.bashrc)
ZSH_FILE := $(HOME)/.zshrc

# User-local static completion directories
BASH_COMP_DIR := $(HOME)/.bash-completions
ZSH_COMP_DIR := $(HOME)/.zsh-completions

.PHONY: all init build test install completions clean

all: build

init:
	@echo "Initializing Go module..."
	@if [ ! -f go.mod ]; then go mod init dotfiles/gstow; fi
	@go get gopkg.in/yaml.v3
	@go get github.com/spf13/cobra@latest
	@if [ ! -f gs.yaml ]; then touch gs.yaml; echo "Created empty gs.yaml"; fi

build: init
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) main.go

test: init
	@echo "Running tests..."
	@go test -v

install: build completions
	@echo "Installing binary to $(INSTALL_DIR) (may prompt for sudo)..."
	@sudo install -d -m 755 $(INSTALL_DIR)
	@sudo install -m 755 $(BINARY_NAME) $(INSTALL_DIR)/
	@echo "✅ Install complete! $(BINARY_NAME) is now available globally."

completions: build
	@echo "Generating static shell completions..."
	@mkdir -p $(BASH_COMP_DIR) $(ZSH_COMP_DIR)
	@./$(BINARY_NAME) completion bash > $(BASH_COMP_DIR)/gstow.bash
	@./$(BINARY_NAME) completion zsh > $(ZSH_COMP_DIR)/gstow.zsh
	
	@echo "Wiring up bash completions..."
	@if ! grep -q "gstow.bash" $(BASH_FILE) 2>/dev/null; then \
		echo '\n# gstow bash completions\n[ -f $(BASH_COMP_DIR)/gstow.bash ] && source $(BASH_COMP_DIR)/gstow.bash' >> $(BASH_FILE); \
		echo "✅ Added bash completion hook to $(BASH_FILE)"; \
	else \
		echo "✅ Bash completion hook already exists in $(BASH_FILE)"; \
	fi
	
	@echo "Wiring up zsh completions..."
	@if [ -f $(ZSH_FILE) ]; then \
		if ! grep -q "gstow.zsh" $(ZSH_FILE) 2>/dev/null; then \
			echo '\n# gstow zsh completions\n[ -f $(ZSH_COMP_DIR)/gstow.zsh ] && source $(ZSH_COMP_DIR)/gstow.zsh' >> $(ZSH_FILE); \
			echo "✅ Added zsh completion hook to $(ZSH_FILE)"; \
		else \
			echo "✅ Zsh completion hook already exists in $(ZSH_FILE)"; \
		fi \
	fi

clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@rm -f go.mod go.sum
END: Makefile

FILE: README.md
# gstow

`gstow` is an explicit, idempotent deployment functor for your dotfiles. It maps configuration files stored in a Git repository to their specific locations on your system.

## The Core Philosophy: Explicit Tracking

You don't start by guessing how to arrange your repository. You start by adding files to it.

Most dotfile managers (like GNU Stow) rely on implicit directory mirroring. If you want to track your Emacs config, they force you to link `~/.emacs.d` directly to your Git repo. **This is a trap.** Modern apps treat their config directories as cache dumps. If you link `~/.emacs.d` or `~/Library/Application Support/Cursor`, those apps will silently dump hundreds of megabytes of compiled binaries and telemetry directly into your Git tree.

`gstow` solves this by tracking **individual files**, maintaining a strict, human-readable bipartite mapping in `gs.yaml`. You only map directories when you explicitly want to (like `~/scripts`).

## Installation

1. Initialize your dotfiles repository.
2. Run `make install` to build the binary.
3. Ensure `~/.local/bin` is in your `$PATH`.

*(Run `gstow completion -h` to set up your shell autocompletion).*

## Usage Workflow

### 1. Add files to the repository
To start tracking a file, use the `add` command. `gstow` will move the file into your repository, record the destination path in `gs.yaml`, and immediately create the symlink.
```bash
gstow add ~/.emacs.d/init.el emacs/init.el
gstow add ~/.tmux.conf tmux/.tmux.conf
END: README.md

FILE: go.mod
module dotfiles/gstow

go 1.25.0

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
END: go.mod

FILE: go.sum
github.com/cpuguy83/go-md2man/v2 v2.0.6/go.mod h1:oOW0eioCTA6cOiMLiUPZOpcVxMig6NIQQ7OS05n1F4g=
github.com/inconshreveable/mousetrap v1.1.0 h1:wN+x4NVGpMsO7ErUn/mUI3vEoE6Jt13X2s0bqwp9tc8=
github.com/inconshreveable/mousetrap v1.1.0/go.mod h1:vpF70FUmC8bwa3OWnCshd2FqLfsEA9PFc4w1p2J65bw=
github.com/russross/blackfriday/v2 v2.1.0/go.mod h1:+Rmxgy9KzJVeS9/2gXHxylqXiyQDYRxCVz55jmeOWTM=
github.com/spf13/cobra v1.10.2 h1:DMTTonx5m65Ic0GOoRY2c16WCbHxOOw6xxezuLaBpcU=
github.com/spf13/cobra v1.10.2/go.mod h1:7C1pvHqHw5A4vrJfjNwvOdzYu0Gml16OCs2GRiTUUS4=
github.com/spf13/pflag v1.0.9 h1:9exaQaMOCwffKiiiYk6/BndUBv+iRViNW+4lEMi0PvY=
github.com/spf13/pflag v1.0.9/go.mod h1:McXfInJRrz4CZXVZOBLb0bTZqETkiAhM9Iw0y3An2Bg=
go.yaml.in/yaml/v3 v3.0.4/go.mod h1:DhzuOOF2ATzADvBadXxruRBLzYTpT36CKvDb3+aBEFg=
gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405/go.mod h1:Co6ibVJAznAaIkqp8huTwlJQCZ016jof/cbN4VW5Yz0=
gopkg.in/yaml.v3 v3.0.1 h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=
gopkg.in/yaml.v3 v3.0.1/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=
END: go.sum

FILE: gs.yaml
END: gs.yaml

FILE: main.go
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
END: main.go

FILE: main_test.go
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
END: main_test.go



---

# Repository tree (ls -R)

COPYRIGHT
go.mod
go.sum
gs.yaml
gstow
LICENSE
main_test.go
main.go
Makefile
README.md
