# Goreleaser 配置更新说明

## 概述

已成功更新 `.goreleaser.yml` 配置文件，添加了前端构建钩子，确保在构建后端 Go 程序之前先构建前端。

## 修改内容

### `.goreleaser.yml` 配置更新

在 `builds` 部分添加了 `hooks.pre` 配置：

```yaml
builds:
  - env:
      - CGO_ENABLED=0
    ldflags:
      - -s -w -X github.com/futureharmony/storagebrowser/v2/version.Version={{ .Version }} -X github.com/futureharmony/storagebrowser/v2/version.CommitSHA={{ .ShortCommit }}
    main: main.go
    binary: storagebrowser
    goos:
      - darwin
      - linux
      - windows
      - freebsd
    goarch:
      - amd64
      - "386"
      - arm
      - arm64
      - riscv64
    goarm:
      - "5"
      - "6"
      - "7"
    ignore:
      - goos: darwin
        goarch: "386"
      - goos: freebsd
        goarch: arm
    hooks:
      pre: |
        cd frontend && pnpm build && cd ..
```

## 功能说明

### 前端构建钩子

- **位置**: `hooks.pre`
- **命令**: `cd frontend && pnpm build && cd ..`
- **执行时机**: 在每个 Go 构建目标之前执行
- **作用**: 确保前端资源在后端构建之前已经编译完成

### 构建流程

1. **预构建阶段** (`hooks.pre`):
   - 进入 `frontend` 目录
   - 执行 `pnpm build` 构建前端
   - 返回项目根目录

2. **Go 构建阶段**:
   - 执行 Go 程序构建
   - 生成二进制文件

3. **打包阶段**:
   - 将前端构建产物打包到最终的发布包中

## 验证

### 前端构建测试

已验证前端构建命令可以正常执行：

```bash
cd frontend && pnpm build
```

**输出结果**:
- ✅ 前端构建成功完成
- ✅ 生成了所有必要的静态文件
- ✅ TypeScript 类型检查通过
- ✅ Vite 构建过程无错误

### 与 Makefile 的一致性

当前的 goreleaser 配置与 `Makefile` 中的 `build-frontend` 目标保持一致：

```makefile
.PHONY: build-frontend
build-frontend: ## Build frontend
	$Q cd frontend && pnpm install --frozen-lockfile && pnpm run build
```

## 注意事项

1. **依赖关系**: 确保 `frontend` 目录存在且包含有效的 pnpm 项目
2. **Node.js 环境**: 构建机器需要安装 Node.js 和 pnpm
3. **网络连接**: 首次构建时需要下载 npm 依赖
4. **构建缓存**: 后续构建会利用缓存，速度更快

## 构建产物

执行 goreleaser 后，会生成：

- 前端静态文件 (HTML, CSS, JS, 字体文件等)
- Go 二进制文件 (支持多个平台和架构)
- 完整的发布包，包含前端和后端

## 相关文件

- `.goreleaser.yml` - 主要配置文件
- `Makefile` - 包含本地开发构建命令
- `frontend/package.json` - 前端项目配置
- `frontend/pnpm-lock.yaml` - pnpm 锁定文件

## 测试建议

1. 在 CI/CD 环境中测试完整的 goreleaser 流程
2. 验证生成的二进制文件包含正确的前端资源
3. 测试多平台构建是否正常工作
4. 确认版本信息正确嵌入到二进制文件中