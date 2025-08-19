#!/bin/bash

# rclone 自定义版本发布脚本
# 使用方法: 
#   ./create_release.sh                    # 自动基于官方版本 (推荐)
#   ./create_release.sh auto               # 同上
#   ./create_release.sh v1.70.3-custom    # 手动指定版本

set -e

# 获取官方最新版本
get_official_version() {
    # 首先尝试从 VERSION 文件获取
    if [ -f "VERSION" ]; then
        local version=$(cat VERSION)
        # 检查是否已经有 v 前缀
        if [[ "$version" =~ ^v ]]; then
            echo "$version"
        else
            echo "v$version"
        fi
    else
        # 如果没有 VERSION 文件，从 git tag 获取最新的官方版本
        git tag --sort=-version:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -1
    fi
}

# 处理版本号参数
if [ -n "$1" ]; then
    # 用户提供了版本号
    if [[ "$1" == "auto" ]]; then
        # 自动模式：基于官方版本 + custom 后缀
        OFFICIAL_VERSION=$(get_official_version)
        VERSION="${OFFICIAL_VERSION}-custom"
        echo "🔍 检测到官方版本: $OFFICIAL_VERSION"
        echo "📝 将发布自定义版本: $VERSION"
    else
        # 手动指定版本
        VERSION="$1"
        # 验证版本号格式 (支持 vX.Y.Z 和 vX.Y.Z-suffix)
        if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-.*)?$ ]]; then
            echo "❌ 错误: 版本号格式不正确"
            echo "支持格式: vX.Y.Z 或 vX.Y.Z-suffix (例如: v1.70.3-custom)"
            exit 1
        fi
    fi
else
    # 默认使用自动模式
    OFFICIAL_VERSION=$(get_official_version)
    VERSION="${OFFICIAL_VERSION}-custom"
    echo "🔍 自动检测官方版本: $OFFICIAL_VERSION"
    echo "📝 将发布自定义版本: $VERSION"
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

# 获取基础版本信息
if [[ "$VERSION" =~ ^(.+)-custom$ ]]; then
    BASE_VERSION="${BASH_REMATCH[1]}"
    VERSION_INFO="Based on official rclone $BASE_VERSION"
else
    VERSION_INFO="Custom build"
fi

git tag -a "$VERSION" -m "rclone Custom Build $VERSION

$VERSION_INFO

Features:
- ✨ Custom filename obfuscation for Windows compatibility
- 🔧 Enhanced crypt backend with CJK character set
- 🎲 Deterministic obfuscation with Fisher-Yates shuffling
- 🔄 Full backward compatibility with existing modes

Build Info:
- Branch: $CURRENT_BRANCH  
- Commit: $(git rev-parse --short HEAD)
- Date: $(date -u)
- Base Version: ${BASE_VERSION:-N/A}"

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