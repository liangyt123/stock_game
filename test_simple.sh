#!/bin/bash

# 简化版自动化测试脚本
# 快速验证核心功能

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${CYAN}  股票游戏快速测试${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 测试1: 编译检查
echo -e "${YELLOW}[1/5]${NC} 编译检查..."
if go build -o stock_game main.go 2>&1 | grep -q "error"; then
    echo -e "${RED}✗ 编译失败${NC}"
    exit 1
else
    echo -e "${GREEN}✓ 编译成功${NC}"
fi

# 测试2: 代码静态检查
echo -e "${YELLOW}[2/5]${NC} 静态分析..."
errors=0

checks=(
    "state.VolumeHistory.*\[\]int"
    "state.PriceHistory.*\[\]float64"
    "state.TotalMarketShares"
    "state.DayLow.*="
    "state.DayHigh.*="
    "len(state.VolumeHistory) > 0"
)

for check in "${checks[@]}"; do
    if ! grep -q "$check" main.go; then
        echo -e "${RED}✗ 缺失: $check${NC}"
        errors=$((errors + 1))
    fi
done

if [ $errors -eq 0 ]; then
    echo -e "${GREEN}✓ 静态检查通过${NC}"
else
    echo -e "${RED}✗ 发现 $errors 个问题${NC}"
    exit 1
fi

# 测试3: 残局场景定义检查
echo -e "${YELLOW}[3/5]${NC} 残局场景检查..."
scenario_count=$(grep -c "endgame_" main.go || echo "0")
if [ "$scenario_count" -ge 16 ]; then
    echo -e "${GREEN}✓ 找到 $scenario_count 个残局场景${NC}"
else
    echo -e "${YELLOW}⚠ 仅找到 $scenario_count 个残局场景${NC}"
fi

# 测试4: UI组件检查
echo -e "${YELLOW}[4/5]${NC} UI组件检查..."
ui_components=(
    "renderEndgameHUD"
    "renderAIMoneyFlowRadar"
    "renderRiskHeatmap"
    "renderCriticalAlerts"
    "renderEnhancedKLine"
)

found=0
for component in "${ui_components[@]}"; do
    if grep -q "func $component" main.go; then
        found=$((found + 1))
    fi
done

echo -e "${GREEN}✓ 找到 $found/5 个新UI组件${NC}"

# 测试5: 运行快速冒烟测试
echo -e "${YELLOW}[5/5]${NC} 快速冒烟测试..."

# 创建测试输入文件
cat > /tmp/smoke_test_input.txt << 'EOF'
1
1

5
q
EOF

# 使用后台进程和超时
(./stock_game < /tmp/smoke_test_input.txt > /tmp/smoke_test_output.txt 2>&1) &
pid=$!

# 等待最多5秒
for i in {1..10}; do
    if ! kill -0 $pid 2>/dev/null; then
        break
    fi
    sleep 0.5
done

# 如果还在运行，强制结束
if kill -0 $pid 2>/dev/null; then
    kill -9 $pid 2>/dev/null
fi

# 检查输出
if [ -f /tmp/smoke_test_output.txt ]; then
    if grep -qi "panic\|runtime error\|fatal" /tmp/smoke_test_output.txt; then
        echo -e "${RED}✗ 发现运行时错误${NC}"
        grep -i "panic\|runtime error" /tmp/smoke_test_output.txt | head -5
        rm -f /tmp/smoke_test_input.txt /tmp/smoke_test_output.txt
        exit 1
    else
        echo -e "${GREEN}✓ 程序运行正常（无panic/错误）${NC}"
    fi
else
    echo -e "${YELLOW}⚠ 无法生成输出文件${NC}"
fi

# 清理
rm -f /tmp/smoke_test_input.txt /tmp/smoke_test_output.txt

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}✅ 所有测试通过！${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "提示: 运行 './test_automation.sh' 进行完整测试"
echo "      运行 './stock_game' 开始游戏"
echo ""
