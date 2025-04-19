.PHONY: all build run clean test

# Binary name and output directory
BINARY_NAME=dotman
OUT_DIR=out

# Default target
all: run

# Create output directory
$(OUT_DIR):
	mkdir -p $(OUT_DIR)

# Build the application
build: $(OUT_DIR)
	go build -o $(OUT_DIR)/$(BINARY_NAME) .

# Run the application
run: build
	./$(OUT_DIR)/$(BINARY_NAME)

# Clean build artifacts
clean:
	go clean
	rm -rf $(OUT_DIR)

# Run tests
test:
	go test ./... 