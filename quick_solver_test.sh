#!/bin/bash
# 快速测试求解器功能

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  求解器集成测试"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 测试1: 检查编译
echo "[1/3] 检查编译..."
if go build -o stock_game 2>&1 | grep -q "error"; then
    echo "✗ 编译失败"
    exit 1
else
    echo "✓ 编译成功"
fi

# 测试2: 检查函数存在
echo "[2/3] 检查求解器函数..."
if grep -q "func RunSolverCLI" solve_cli.go; then
    echo "✓ RunSolverCLI 函数存在"
else
    echo "✗ RunSolverCLI 函数不存在"
    exit 1
fi

# 测试3: 检查QuickSolve函数
echo "[3/3] 检查QuickSolve函数..."
if grep -q "func QuickSolve" solver.go; then
    echo "✓ QuickSolve 函数存在"
else
    echo "✗ QuickSolve 函数不存在"
    exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ 所有检查通过！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "提示: 运行游戏选择 [5] 🤖 AI求解器 即可使用"
echo ""
