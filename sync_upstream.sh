#!/bin/bash

# rclone官方代码同步脚本
# 使用方法: ./sync_upstream.sh

set -e  # 遇到错误立即退出

echo "🔄 开始同步官方rclone代码..."

# 1. 获取官方最新代码
echo "📥 获取官方最新代码..."
git fetch origin

# 2. 切换到master并同步到官方最新
echo "🔄 更新master分支..."
git checkout master
git reset --hard origin/master

# 3. 推送更新后的master到fork
echo "📤 推送master到fork仓库..."
git push fork master --force

# 4. 将自定义功能分支变基到最新master
echo "🔧 更新自定义功能分支..."
git checkout feature/custom-filename-obfuscation

# 检查是否需要rebase
if git merge-base --is-ancestor master HEAD; then
    echo "✅ 功能分支已经是最新的"
else
    echo "🔄 正在将功能分支变基到最新master..."
    if git rebase master; then
        echo "✅ 变基成功"
        # 推送更新后的功能分支
        echo "📤 推送功能分支到fork仓库..."
        git push fork feature/custom-filename-obfuscation --force
    else
        echo "❌ 变基过程中出现冲突，请手动解决冲突后运行:"
        echo "   git add <解决冲突的文件>"
        echo "   git rebase --continue"
        echo "   git push fork feature/custom-filename-obfuscation --force"
        exit 1
    fi
fi

# 5. 显示当前状态
echo "📊 当前状态:"
echo "   Master分支: $(git log --oneline origin/master -1)"
echo "   功能分支: $(git log --oneline feature/custom-filename-obfuscation -1)"

echo "🎉 同步完成!"
echo ""
echo "🔗 查看您的代码:"
echo "   Master: https://github.com/Hittlert/rclone/tree/master"
echo "   功能分支: https://github.com/Hittlert/rclone/tree/feature/custom-filename-obfuscation"