# Makefile for tstow

BINARY_NAME := tstow
INSTALL_DIR ?= /usr/local/bin

BASH_FILE := $(if $(wildcard $(HOME)/.bash_profile),$(HOME)/.bash_profile,$(HOME)/.bashrc)
ZSH_FILE := $(HOME)/.zshrc

BASH_COMP_DIR := $(HOME)/.bash-completions
ZSH_COMP_DIR := $(HOME)/.zsh-completions

.PHONY: all init build test install completions clean

all: build

tidy:
	@go mod tidy

vet:
	@go vet .

fmt:
	@go fmt ./...

init:
	@echo "Initializing Go module..."
	@if [ ! -f go.mod ]; then go mod init dotfiles/tstow; fi
	@go get gopkg.in/yaml.v3
	@go get github.com/spf13/cobra@latest
	@if [ ! -f tstow.yaml ]; then touch tstow.yaml; echo "Created empty tstow.yaml"; fi

build: fmt tidy vet init
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
	@./$(BINARY_NAME) completion bash > $(BASH_COMP_DIR)/tstow.bash
	@./$(BINARY_NAME) completion zsh > $(ZSH_COMP_DIR)/tstow.zsh
	@echo "Wiring up bash completions..."
	@if ! grep -q "tstow.bash" $(BASH_FILE) 2>/dev/null; then \
		echo '\n# tstow bash completions\n[ -f $(BASH_COMP_DIR)/tstow.bash ] && source $(BASH_COMP_DIR)/tstow.bash' >> $(BASH_FILE); \
		echo "✅ Added bash completion hook to $(BASH_FILE)"; \
	else \
		echo "✅ Bash completion hook already exists in $(BASH_FILE)"; \
	fi
	@echo "Wiring up zsh completions..."
	@if [ -f $(ZSH_FILE) ]; then \
		if ! grep -q "tstow.zsh" $(ZSH_FILE) 2>/dev/null; then \
			echo '\n# tstow zsh completions\n[ -f $(ZSH_COMP_DIR)/tstow.zsh ] && source $(ZSH_COMP_DIR)/tstow.zsh' >> $(ZSH_FILE); \
			echo "✅ Added zsh completion hook to $(ZSH_FILE)"; \
		else \
			echo "✅ Zsh completion hook already exists in $(ZSH_FILE)"; \
		fi \
	fi

clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@rm -f go.mod go.sum
