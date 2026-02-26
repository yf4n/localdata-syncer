# Makefile for alist-syncer

# 项目名称
APP_NAME := alist-syncer

# Go 命令
GO := go

# Docker 命令
DOCKER := docker

# Go 镜像
GO_IMAGE := golang:1.25

# 构建输出目录
OUTPUT_DIR := ./bin

# 构建标志
BUILD_FLAGS :=

# 运行标志
RUN_FLAGS :=

# 测试标志
TEST_FLAGS := -v

# 目标：默认
all: build

# 目标：构建
.PHONY: build
build: clean
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(OUTPUT_DIR)
	@$(GO) build $(BUILD_FLAGS) -o $(OUTPUT_DIR)/$(APP_NAME) .
	@echo "Build completed successfully!"

# 目标：使用Docker构建
.PHONY: docker-build
docker-build: clean
	@echo "Building $(APP_NAME) using Docker..."
	@mkdir -p $(OUTPUT_DIR)
	@$(DOCKER) run --rm -v "$(PWD):/app" -w "/app" $(GO_IMAGE) go build $(BUILD_FLAGS) -o $(OUTPUT_DIR)/$(APP_NAME) .
	@echo "Docker build completed successfully!"

# 目标：运行
.PHONY: run
run:
	@echo "Running $(APP_NAME)..."
	@$(GO) run $(RUN_FLAGS) .

# 目标：清理
.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -rf $(OUTPUT_DIR)
	@$(GO) clean
	@echo "Clean completed!"

# 目标：测试
.PHONY: test
test:
	@echo "Running tests..."
	@$(GO) test $(TEST_FLAGS) ./...

# 目标：格式化代码
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@$(GO) fmt ./...
	@echo "Format completed!"

# 目标：代码检查
.PHONY: vet
vet:
	@echo "Running vet..."
	@$(GO) vet ./...
	@echo "Vet completed!"

# 目标：安装依赖
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@$(GO) mod tidy
	@echo "Dependencies installed!"

# 目标：使用Docker安装依赖
.PHONY: docker-deps
docker-deps:
	@echo "Installing dependencies using Docker..."
	@$(DOCKER) run --rm -v "$(PWD):/app" -w "/app" $(GO_IMAGE) go mod tidy
	@echo "Docker dependencies installed!"

# 目标：检查依赖
.PHONY: deps-check
deps-check:
	@echo "Checking dependencies..."
	@$(GO) mod verify
	@echo "Dependencies check completed!"

# 目标：构建并运行
.PHONY: build-run
build-run: build
	@echo "Running $(APP_NAME)..."
	@$(OUTPUT_DIR)/$(APP_NAME) $(RUN_FLAGS)

# 目标：显示帮助
.PHONY: help
help:
	@echo "Makefile targets for $(APP_NAME):"
	@echo "  make all           - Build the application"
	@echo "  make build         - Build the application"
	@echo "  make docker-build  - Build using Docker (no Go environment needed)"
	@echo "  make run           - Run the application"
	@echo "  make clean         - Clean up build artifacts"
	@echo "  make test          - Run tests"
	@echo "  make fmt           - Format code"
	@echo "  make vet           - Run code vetting"
	@echo "  make deps          - Install dependencies"
	@echo "  make docker-deps   - Install dependencies using Docker"
	@echo "  make deps-check    - Check dependencies"
	@echo "  make build-run     - Build and run the application"
	@echo "  make help          - Show this help message"
