# Makefile for tstow

BINARY_NAME := tstow
INSTALL_DIR ?= /usr/local/bin

# Ask Go directly for the OS and Architecture
OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)

# Dynamically parse the latest version from changelog.yaml
VERSION := $(shell grep -m 1 'version:' changelog.yaml | cut -d '"' -f 2)

# Dynamic output directory
BUILD_DIR := build/$(ARCH)/$(OS)
COMPILED_BIN := $(BUILD_DIR)/$(BINARY_NAME)

# User-local shell configurations
BASH_FILE := $(if $(wildcard $(HOME)/.bash_profile),$(HOME)/.bash_profile,$(HOME)/.bashrc)
ZSH_FILE := $(HOME)/.zshrc
BASH_COMP_DIR := $(HOME)/.bash-completions
ZSH_COMP_DIR := $(HOME)/.zsh-completions

.PHONY: all tidy vet fmt init build test install completions package release clean

all: build

init:
	@echo "Initializing Go module..."
	@if [ ! -f go.mod ]; then go mod init dotfiles/tstow; fi
	@go get gopkg.in/yaml.v3
	@go get github.com/spf13/cobra@latest
	@if [ ! -f tstow.yaml ]; then touch tstow.yaml; echo "Created empty tstow.yaml"; fi

tidy: init
	@go mod tidy

vet: init
	@go vet .

fmt: init
	@go fmt ./...

build: fmt tidy vet
	@echo "Building $(BINARY_NAME) for $(OS)/$(ARCH)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(COMPILED_BIN) main.go

test: init
	@echo "Running tests..."
	@go test -v

install: build completions
	@echo "Installing binary to $(INSTALL_DIR) (may prompt for sudo)..."
	@sudo install -d -m 755 $(INSTALL_DIR)
	@sudo install -m 755 $(COMPILED_BIN) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✅ Install complete! $(BINARY_NAME) is now available globally."

completions: build
	@echo "Generating static shell completions..."
	@mkdir -p $(BASH_COMP_DIR) $(ZSH_COMP_DIR)
	@./$(COMPILED_BIN) completion bash > $(BASH_COMP_DIR)/tstow.bash
	@./$(COMPILED_BIN) completion zsh > $(ZSH_COMP_DIR)/tstow.zsh
	
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

ifeq ($(OS),linux)
package: build
	@echo "📦 Building Linux packages via fpm (v$(VERSION))..."
	@mkdir -p out dist/usr/local/bin
	@cp $(COMPILED_BIN) dist/usr/local/bin/$(BINARY_NAME)
	@fpm -f -s dir -t deb -n $(BINARY_NAME) -v $(VERSION) -C dist -p out/$(BINARY_NAME)_$(VERSION)_$(ARCH).deb usr/local/bin/$(BINARY_NAME)
	@fpm -f -s dir -t rpm -n $(BINARY_NAME) -v $(VERSION) -C dist -p out/$(BINARY_NAME)_$(VERSION)_$(ARCH).rpm usr/local/bin/$(BINARY_NAME)
	@fpm -f -s dir -t apk -n $(BINARY_NAME) -v $(VERSION) -C dist -p out/$(BINARY_NAME)_$(VERSION)_$(ARCH).apk usr/local/bin/$(BINARY_NAME)
	@rm -rf dist
	@echo "🔐 Generating cryptographic checksums..."
	@cd out && sha256sum * > $(BINARY_NAME)_$(VERSION)_$(ARCH)_checksums.txt
else
package: build
	@echo "⚠️  Not on Linux (detected $(OS)). Skipping fpm package generation."
endif

release: build test package
	@echo "🚀 Executing automated changelog release..."
	@go run scripts/release.go

clean:
	@echo "Cleaning up..."
	@rm -rf build out dist
