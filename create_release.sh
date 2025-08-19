#!/bin/bash

# rclone 自定义版本发布脚本
# 使用方法: ./create_release.sh v1.0.0

set -e

VERSION="$1"

if [ -z "$VERSION" ]; then
    echo "❌ 错误: 请提供版本号"
    echo "使用方法: $0 v1.0.0"
    exit 1
fi

# 验证版本号格式
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "❌ 错误: 版本号格式不正确，应该是 vX.Y.Z 格式 (例如: v1.0.0)"
    exit 1
fi

echo "🚀 准备发布 rclone 自定义版本: $VERSION"

# 检查当前分支
CURRENT_BRANCH=$(git branch --show-current)
echo "📍 当前分支: $CURRENT_BRANCH"

# 确保在正确的分支上
if [ "$CURRENT_BRANCH" != "feature/custom-filename-obfuscation" ]; then
    echo "⚠️  警告: 建议在 feature/custom-filename-obfuscation 分支上发布"
    read -p "是否继续? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# 检查工作区状态
if [ -n "$(git status --porcelain)" ]; then
    echo "❌ 错误: 工作区有未提交的更改，请先提交或暂存"
    git status --short
    exit 1
fi

# 同步最新代码
echo "🔄 同步最新代码..."
git fetch fork
git push fork "$CURRENT_BRANCH"

# 创建并推送标签
echo "🏷️  创建标签: $VERSION"
git tag -a "$VERSION" -m "rclone Custom Build $VERSION

Features:
- Custom filename obfuscation for Windows compatibility
- Enhanced crypt backend with CJK character set
- Deterministic obfuscation with Fisher-Yates shuffling
- Full backward compatibility

Build Info:
- Branch: $CURRENT_BRANCH  
- Commit: $(git rev-parse --short HEAD)
- Date: $(date -u)"

echo "📤 推送标签到远程仓库..."
git push fork "$VERSION"

echo "✅ 发布完成!"
echo ""
echo "🔗 GitHub Actions 将自动触发构建，请查看:"
echo "   主要构建: https://github.com/Hittlert/rclone/actions/workflows/build.yml"
echo "   Release 构建: https://github.com/Hittlert/rclone/actions/workflows/custom-release.yml"
echo ""
echo "📦 发布页面 (构建完成后可用):"
echo "   https://github.com/Hittlert/rclone/releases/tag/$VERSION"
echo ""
echo "ℹ️  构建说明:"
echo "   - build.yml: 基于官方工作流，进行完整的测试和构建"
echo "   - custom-release.yml: 专门用于创建 GitHub Release 和上传二进制文件"
echo "   - 构建通常需要 10-15 分钟完成"