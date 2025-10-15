.PHONY: build run clean

BIN_DIR := .bin
BIN_FILE := $(BIN_DIR)/main
MAIN_FILE := cmd/api/main.go

build:
	@echo "🔧 Building..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_FILE) $(MAIN_FILE)
	@echo "✅ Build completed: $(BIN_FILE)"

run:
	@echo "🚀 Running..."
	-@go run $(MAIN_FILE) || true
	@echo "🛑 Run stopped (possibly by Ctrl+C)"

clean:
	@echo "🧹 Cleaning up..."
	@rm -rf $(BIN_DIR)
	@echo "✅ Cleaned."