.PHONY: build-tui build-tui-linux build-tui-darwin build-tui-windows
.PHONY: build-cli test lint clean

# Default binary directory
BIN_DIR=bin

# === TUI Builds ===
build-tui:
	go build -ldflags "-s -w" -o $(BIN_DIR)/janny-tui ./tui

build-tui-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o $(BIN_DIR)/janny-tui-linux ./tui

build-tui-darwin:
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o $(BIN_DIR)/janny-tui-darwin-amd64 ./tui
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o $(BIN_DIR)/janny-tui-darwin-arm64 ./tui

# Windows preparation
build-tui-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o $(BIN_DIR)/janny-tui.exe ./tui

build-all-tui: build-tui-linux build-tui-darwin build-tui-windows

# === Test ===
test:
	go test ./...

# === Clean ===
clean:
	if [ -d "$(BIN_DIR)" ]; then rm -rf $(BIN_DIR); fi
	@# For Windows compatibility in simple environments if rm fails
	@# rd /s /q bin 2>nul || exit 0

# === Lint ===
lint:
	go vet ./...
