# tstow - The Troodos Exascale Dotfile Manager

[![Go Tests](https://github.com/troodos-exascale/tstow/actions/workflows/tests.yml/badge.svg)](https://github.com/troodos-exascale/tstow/actions/workflows/tests.yml)
[![GitHub tag](https://img.shields.io/github/v/tag/troodos-exascale/tstow)](https://github.com/troodos-exascale/tstow/tags)

`tstow` is an explicit, idempotent dotfile manager for managing system configuration files across multiple environments: 

 - its `tstow install` deployment functor installs symbolic links: the point from the installation folder into the configuration repo
 - its adjoint `tstow add` functor grabs configuration files/folders from the install folder and puts them in the repo
 
## The Core Philosophy: Explicit Boundaries

Unlike traditional managers that mirror directories implicitly and pollute your version control with application cache dumps, `tstow` operates strictly on **files and explicit rules**. 

It enforces a strict separation of state:
1. **Repository (`-r`)**: The pure, declarative state of your configuration.
2. **Installation Folder (`-i`)**: Your deployment folder (default `~`)

`tstow` will **never** delete a physical file or directory during an install, but it can move these to the repo. It enforces safety first.

## Installation

### macOS Installation:

Bash
```bash
brew install troodos-exascale/tap/tstow
```

### Linux/WSL Installation:

Download the .deb or .rpm from the [GitHub Releases](https://github.com/troodos-exascale/tstow/releases) page and install via dpkg -i or rpm -i.

**Alpine Linux:**

Download the `.apk` from the Releases page and install it using:
`apk add --allow-untrusted ./tstow_arm64.apk`

### Die hard installation

```bash
make install

## Core Workflows

Commands: 
- default places: -i is `~` and -r `.`.   You can override these to provision arbitrary systems, containers, or test environments.
- Adding a (config file) to the repository: Move a file/directory into the repository and instantly link it. Symlinks are rejected to prevent circular dependency loops.
```bash
tstow add ~/.bashrc shell/.bashrc
```
- Recursive Installation
Enforce the mapping state. You can install everything, or restrict it to a specific subfolder within your repo.
```bash
tstow install        # Enforces entire tstow.yaml
tstow install shell  # Recursively links any mapped file inside repo/shell/
```
(If a local file is conflicting with a symlink, tstow safely halts. Use tstow install -f to forcibly correct broken symlinks).
- Divergence and Exclusions: The Skip List
Sometimes a specific machine shouldn't link a file (e.g., macOS vs Linux). Add it to the skiplist so tstow install ignores it without deleting the configuration, or if you maintain stuff in your repo that you don't want installed (e.g. a package list), skip it. 
```bash
tstow skip shell/linux_aliases
```
- State Inspection
View exactly what is installed, what is skipped, and what is conflicting with local state.
```bash
tstow show
```
