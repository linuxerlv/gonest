# 贡献指南

感谢你对 gonest 项目的关注！

## 开发环境设置

### 前置要求

- Go 1.22 或更高版本
- Git

### 本地开发

1. Fork 本仓库
2. 克隆你的 fork：
   ```bash
   git clone https://github.com/YOUR_USERNAME/gonest.git
   cd gonest
   ```
3. 安装依赖：
   ```bash
   go mod download
   ```

## 代码风格

本项目使用 golangci-lint 进行代码检查。确保你的代码符合以下规范：

- 运行 `golangci-lint run` 检查代码
- 遵循 Go 官方代码规范
- 使用 gofmt 格式化代码

## PR 流程

1. 从 `develop` 分支创建新的 `feature/xxx` 分支
2. 进行开发和测试
3. 提交 PR 到 `develop` 分支
4. 等待 CI 检查通过和代码审查

## Commit Message 规范

使用 Conventional Commits 规范：

- `feat:` 新功能
- `fix:` 修复 bug
- `docs:` 文档更新
- `refactor:` 代码重构
- `test:` 测试相关
- `chore:` 构建/工具相关

示例：`feat: add rate limiting middleware`
