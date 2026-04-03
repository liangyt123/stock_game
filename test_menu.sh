#!/bin/bash

# 交互式菜单测试脚本

echo "🎮 交互式菜单系统 - 快速测试"
echo "================================"
echo ""
echo "✅ 编译检查..."

if go build -o stock_game 2>&1 | grep -q "error"; then
    echo "❌ 编译失败"
    exit 1
else
    echo "✅ 编译成功"
fi

echo ""
echo "📋 已改进的界面："
echo ""
echo "1. 主菜单 - 使用方向键选择"
echo "2. AI求解器模式选择 - 使用方向键选择"
echo "3. 渐进式提示（方向选择） - 使用方向键选择"
echo "4. 渐进式提示（数量选择） - 使用方向键选择"
echo ""
echo "⌨️  按键说明："
echo "  ↑/↓ 或 k/j  - 上下移动"
echo "  Enter       - 确认选择"
echo "  ESC 或 q    - 取消/退出"
echo "  1-9         - 数字快捷键"
echo ""
echo "🚀 启动游戏测试..."
echo ""
echo "提示: 进入游戏后，选择 [5] AI求解器 来体验完整的交互式菜单系统"
echo ""
read -p "按回车键启动游戏... "

./stock_game
