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
