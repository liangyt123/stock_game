#!/bin/bash
# 快速测试残局模式

echo "=== 测试1: 检查编译是否成功 ==="
if [ ! -f "./stock_game" ]; then
    echo "✗ 游戏未编译"
    exit 1
fi
echo "✓ 游戏已编译"

echo ""
echo "=== 测试2: 检查代码中的关键初始化 ==="
if grep -q "state.VolumeHistory = \[\]int{5000}" main.go; then
    echo "✓ VolumeHistory 初始化已添加"
else
    echo "✗ VolumeHistory 初始化缺失"
    exit 1
fi

if grep -q "state.PriceHistory = \[\]float64{scenario.InitialPrice}" main.go; then
    echo "✓ PriceHistory 初始化已添加"
else
    echo "✗ PriceHistory 初始化缺失"
    exit 1
fi

if grep -q "state.TotalMarketShares = 100000" main.go; then
    echo "✓ TotalMarketShares 初始化已添加"
else
    echo "✗ TotalMarketShares 初始化缺失"
    exit 1
fi

if grep -q "state.DayLow = scenario.InitialPrice" main.go; then
    echo "✓ DayLow 初始化已添加"
else
    echo "✗ DayLow 初始化缺失"
    exit 1
fi

if grep -q "state.DayHigh = scenario.InitialPrice" main.go; then
    echo "✓ DayHigh 初始化已添加"
else
    echo "✗ DayHigh 初始化缺失"
    exit 1
fi

echo ""
echo "=== 测试3: 检查安全访问代码 ==="
if grep -q "if len(state.VolumeHistory) > 0 && state.TotalMarketShares > 0" main.go; then
    echo "✓ VolumeHistory 安全访问已添加"
else
    echo "✗ VolumeHistory 安全访问缺失"
    exit 1
fi

if grep -q "if len(state.PriceHistory) < 5 || len(state.VolumeHistory) < 1" main.go; then
    echo "✓ 增强K线边界检查已添加"
else
    echo "✗ 增强K线边界检查缺失"
    exit 1
fi

echo ""
echo "✅ 所有修复已正确应用！残局模式应该可以正常运行。"
echo ""
echo "你现在可以运行游戏并选择："
echo "  4 (残局挑战) → 选择任意场景 → 开始游戏"
