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
