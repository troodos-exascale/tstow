tstow - The Troodos Exascale Dotfile Manager

[![Go Tests](https://github.com/troodos-exascale/tstow/actions/workflows/tests.yml/badge.svg)](https://github.com/troodos-exascale/tstow/actions/workflows/tests.yml)
[![GitHub tag](https://img.shields.io/github/v/tag/troodos-exascale/tstow)](https://github.com/troodos-exascale/tstow/tags)

`tstow` is an explicit, idempotent dotfile manager for managing system configuration files across multiple environments: 

 - its `tstow place` deployment functor installs symbolic links: pointing from the placement folder into the configuration repo
 - its adjoint `tstow ingest` functor grabs configuration files/folders from the placement folder and safely copies them into the repo
 
## The Core Philosophy: Explicit Boundaries

Unlike traditional managers that mirror directories implicitly and pollute your version control with application cache dumps, `tstow` operates strictly on **files and explicit rules**. 

It enforces a strict separation of state:
1. **Repository (`-r`)**: The pure, declarative state of your configuration.
2. **Placement Folder (`-p`)**: Your deployment folder (default `~`)

`tstow` will **never** silently delete a physical file or directory that isn't mapped, and it operates by copying underlying data to preserve your legacy setups. It enforces safety first.

## Installation

### macOS Installation:

Bash
```bash
brew install troodos-exascale/tap/tstow
Linux/WSL Installation:
Download the .deb or .rpm from the GitHub Releases page and install via dpkg -i or rpm -i.

Alpine Linux:

Download the .apk from the Releases page and install it using:
apk add --allow-untrusted ./tstow_arm64.apk

Die hard installation
Bash
make install
Core Workflows
Commands:

Explicit State: To prevent dangerous background mutations, mutating commands (ingest, place, undo) operate purely in the current directory (.) unless explicitly given the -r flag.

Global Inspection: Using explicit -r or -p flags writes those locations to ~/.tstowrc. The read-only tstow show command reads this file, allowing you to globally inspect your dotfile status from anywhere on the system.

Ingesting a config to the repository: Copy a file/directory into the repository and instantly link it. If the local file is already a symlink (e.g., from an old Stow setup), tstow safely adopts and copies the physical target data without harming the original.

Bash
tstow --place ~ ingest shell/.bashrc ~/.bashrc
Ingesting ignored/backup data: Keep useful data in your repository (like JSON exports or app backups) without symlinking them to your system using the -i (ignore) flag.

Bash
tstow ingest -i btt/exported-settings.json ~/Downloads/btt-export.json
Piping from Stdin: You can pipe data directly into your repository. This automatically acts as an "ignored" file since there is no local path to symlink.

Bash
echo "export AWS_REGION=us-east-1" | tstow ingest shell/aws_env
Reverting an ingestion: Made a mistake? The undo functor cleanly severs the symlink, restores the physical file back to your system, and erases the YAML mapping.

Bash
tstow undo shell/.bashrc
Recursive Placement:
Enforce the mapping state. You can place everything, or restrict it to a specific subfolder within your repo.

Bash
tstow place        # Enforces entire tstow.yaml
tstow place shell  # Recursively links any mapped file inside repo/shell/
(Safety First: If a local file conflicts with a mapped file, tstow halts to protect your data. However, if the conflicting local file is byte-for-byte identical to the repo file, tstow performs a "Safe Replace," automatically swapping it for a symlink to eliminate setup friction on new machines. Use tstow place -f to forcibly correct wrong symlinks).

Divergence and Exclusions: The Skip List
Sometimes a specific machine shouldn't link a file (e.g., macOS vs Linux). Add it to the skiplist so tstow place ignores it without deleting the configuration, or if you maintain stuff in your repo that you don't want placed (e.g. a package list), skip it.

Bash
tstow skip shell/linux_aliases
State Inspection
View exactly what is placed, what is skipped, and what is conflicting with local state.

Bash
tstow show
Acknowledgements
GNU stow is a fantastic solution that survived more than 20 years of evolution. tstow is deeply inspired by it.
