# Makefile for blockchain project

.PHONY: build run test clean deps

# 构建目标
BINARY_NAME=chain_binary
BUILD_DIR=build

# 默认目标
all: deps build

# 安装依赖
deps:
	go mod tidy
	go mod download

# 构建项目
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/chain

# 运行项目
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME) --config config.yaml

# 测试
test:
	@echo "Running tests..."
	go test -v ./...

# 清理
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f chain_binary
	go clean

# 格式化代码
fmt:
	@echo "Formatting code..."
	go fmt ./...

# 代码检查
lint:
	@echo "Running linter..."
	golangci-lint run

# 生成文档
docs:
	@echo "Generating documentation..."
	godoc -http=:6060

# 帮助
help:
	@echo "Available targets:"
	@echo "  build   - Build the project"
	@echo "  run     - Build and run the project"
	@echo "  test    - Run tests"
	@echo "  clean   - Clean build artifacts"
	@echo "  deps    - Install dependencies"
	@echo "  fmt     - Format code"
	@echo "  lint    - Run linter"
	@echo "  docs    - Generate documentation"
	@echo "  help    - Show this help" 