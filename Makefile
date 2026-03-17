# NapSec Makefile
BINARY_NAME := napsec
BUILD_DIR   := build
VERSION     := 0.1.0
LDFLAGS     := -ldflags "-X main.Version=$(VERSION)"

.PHONY: all build test clean install lint run

## all: 编译 + 测试
all: test build

## build: 编译二进制文件
build:
	@echo "▶ 编译 NapSec $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/napsec/
	@echo "✅ 编译完成: $(BUILD_DIR)/$(BINARY_NAME)"

## build-all: 交叉编译（Linux / macOS / Windows）
build-all:
	@echo "▶ 交叉编译..."
	GOOS=linux   GOARCH=amd64  go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64   ./cmd/napsec/
	GOOS=darwin  GOARCH=amd64  go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64  ./cmd/napsec/
	GOOS=darwin  GOARCH=arm64  go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64  ./cmd/napsec/
	GOOS=windows GOARCH=amd64  go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/napsec/
	@echo "✅ 交叉编译完成"

## test: 运行所有测试
test:
	@echo "▶ 运行测试..."
	go test ./... -v -timeout 30s
	@echo "✅ 测试完成"

## lint: 代码检查
lint:
	@which golangci-lint > /dev/null || \
		(echo "安装 golangci-lint..." && \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

## clean: 清理构建产物
clean:
	@echo "▶ 清理..."
	rm -rf $(BUILD_DIR)
	@echo "✅ 清理完成"

## install: 安装到系统
install: build
	@echo "▶ 安装 NapSec..."
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "✅ 已安装到 /usr/local/bin/$(BINARY_NAME)"

## uninstall: 从系统卸载
uninstall:
	rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "✅ 已卸载"

## run: 快速启动（演习模式）
run: build
	./$(BUILD_DIR)/$(BINARY_NAME) start . --dry-run

## help: 显示帮助
help:
	@echo "NapSec 构建命令:"
	@grep -E '^## ' Makefile | sed 's/## /  /'