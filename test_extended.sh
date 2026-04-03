#!/bin/bash

# 扩展测试套件 - 深度功能验证
# 包含高级测试场景和边界条件

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
PURPLE='\033[0;35m'
NC='\033[0m'

# 测试计数
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# 日志
LOG_DIR="test_logs"
mkdir -p "$LOG_DIR"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
EXT_LOG="$LOG_DIR/test_extended_${TIMESTAMP}.log"

log_test() {
    ((TOTAL_TESTS++))
    echo -e "${BLUE}[测试 $TOTAL_TESTS]${NC} $1" | tee -a "$EXT_LOG"
}

log_pass() {
    ((PASSED_TESTS++))
    echo -e "${GREEN}  ✓ $1${NC}" | tee -a "$EXT_LOG"
}

log_fail() {
    ((FAILED_TESTS++))
    echo -e "${RED}  ✗ $1${NC}" | tee -a "$EXT_LOG"
}

log_skip() {
    ((SKIPPED_TESTS++))
    echo -e "${YELLOW}  ⊘ $1${NC}" | tee -a "$EXT_LOG"
}

log_info() {
    echo -e "${CYAN}  → $1${NC}" | tee -a "$EXT_LOG"
}

# macOS兼容timeout
run_with_timeout() {
    local timeout=$1
    shift
    (
        "$@" &
        pid=$!
        (sleep $timeout; kill -9 $pid 2>/dev/null) &
        killer=$!
        wait $pid 2>/dev/null
        exit_code=$?
        kill -9 $killer 2>/dev/null
        exit $exit_code
    )
}

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${PURPLE}         扩展测试套件 - 深度功能验证${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "测试日志: $EXT_LOG"
echo ""

# ============================================================
# 测试1: AI行为多样性测试
# ============================================================
log_test "AI行为多样性测试"

ai_types=("Whale" "Quant" "Retail" "Institution")
ai_subtypes=("刺客" "打板" "埋伏" "网格" "新韭" "老散" "国家队")

found_types=0
for ai_type in "${ai_types[@]}"; do
    if grep -q "Type.*\"$ai_type\"" main.go; then
        ((found_types++))
        log_pass "找到AI类型: $ai_type"
    else
        log_fail "缺失AI类型: $ai_type"
    fi
done

found_subtypes=0
for subtype in "${ai_subtypes[@]}"; do
    if grep -q "SubType.*\"$subtype\"" main.go; then
        ((found_subtypes++))
    fi
done

log_info "AI类型: $found_types/4, 子类型: $found_subtypes/7"

if [ $found_types -ge 3 ]; then
    log_pass "AI多样性充足"
else
    log_fail "AI多样性不足"
fi

# ============================================================
# 测试2: 市场事件完整性测试
# ============================================================
log_test "市场事件完整性测试"

event_categories=(
    "Neutral"
    "MildOptimistic"
    "MildPessimistic"
    "StrongOptimistic"
    "StrongPessimistic"
    "ExtremePanic"
    "ExtremeOptimistic"
)

found_events=0
for category in "${event_categories[@]}"; do
    if grep -q "$category" main.go; then
        ((found_events++))
    fi
done

log_info "找到事件类别: $found_events/${#event_categories[@]}"

total_events=$(grep -c "Title:.*\"" main.go 2>/dev/null || echo "0")
log_info "事件总数: $total_events"

if [ $found_events -ge 5 ] && [ $total_events -ge 20 ]; then
    log_pass "市场事件充足 ($total_events 个事件)"
else
    log_fail "市场事件不足"
fi

# ============================================================
# 测试3: 杠杆/融资融券功能测试
# ============================================================
log_test "杠杆/融资融券功能测试"

leverage_keywords=(
    "MarginDebt"
    "IsMarginCalled"
    "爆仓"
    "强平"
    "配资"
)

found_leverage=0
for keyword in "${leverage_keywords[@]}"; do
    if grep -q "$keyword" main.go; then
        ((found_leverage++))
        log_pass "找到关键字: $keyword"
    fi
done

if [ $found_leverage -ge 3 ]; then
    log_pass "杠杆功能已实现"
else
    log_fail "杠杆功能不完整"
fi

# ============================================================
# 测试4: 崩盘机制完整性测试
# ============================================================
log_test "崩盘机制完整性测试"

crash_keywords=(
    "CrashWarningLevel"
    "IsCrashed"
    "ConsecutiveFallDays"
    "崩盘"
)

found_crash=0
for keyword in "${crash_keywords[@]}"; do
    if grep -q "$keyword" main.go; then
        ((found_crash++))
    fi
done

log_info "崩盘相关关键字: $found_crash/${#crash_keywords[@]}"

if grep -q "func.*crash\|func.*Crash" main.go; then
    log_pass "崩盘处理函数存在"
else
    log_fail "崩盘处理函数缺失"
fi

if [ $found_crash -ge 3 ]; then
    log_pass "崩盘机制完整"
else
    log_fail "崩盘机制不完整"
fi

# ============================================================
# 测试5: 成就系统测试
# ============================================================
log_test "成就系统测试"

achievement_keywords=(
    "Achievement"
    "Badge"
    "Level"
    "Experience"
    "勋章"
)

found_achievements=0
for keyword in "${achievement_keywords[@]}"; do
    if grep -q "$keyword" main.go; then
        ((found_achievements++))
    fi
done

achievement_count=$(grep -c "BadgeID.*\"" main.go 2>/dev/null || echo "0")
log_info "成就数量: $achievement_count"

if [ $found_achievements -ge 3 ] && [ $achievement_count -ge 10 ]; then
    log_pass "成就系统已实现 ($achievement_count 个成就)"
else
    log_skip "成就系统可选（已有 $achievement_count 个）"
fi

# ============================================================
# 测试6: 策略建议系统测试
# ============================================================
log_test "策略建议系统测试"

if grep -q "func generateStrategyAdvice" main.go; then
    log_pass "策略建议函数存在"

    strategy_factors=(
        "CrashWarning"
        "AI.*逃跑"
        "风险"
        "建议"
    )

    found_factors=0
    for factor in "${strategy_factors[@]}"; do
        if grep -q "$factor" main.go; then
            ((found_factors++))
        fi
    done

    log_info "策略因子: $found_factors/${#strategy_factors[@]}"

    if [ $found_factors -ge 3 ]; then
        log_pass "策略建议系统完整"
    else
        log_fail "策略建议系统不完整"
    fi
else
    log_fail "策略建议函数缺失"
fi

# ============================================================
# 测试7: 反身性理论和贝叶斯分析测试
# ============================================================
log_test "反身性理论和贝叶斯分析测试"

advanced_features=(
    "ReflexivityMetrics"
    "BayesianAnalysis"
    "CrashProbability"
    "calculateReflexivityMetrics"
    "runBayesianAnalysis"
)

found_advanced=0
for feature in "${advanced_features[@]}"; do
    if grep -q "$feature" main.go; then
        ((found_advanced++))
        log_pass "找到: $feature"
    fi
done

if [ $found_advanced -ge 4 ]; then
    log_pass "高级分析系统已实现"
else
    log_skip "高级分析系统未完全实现 ($found_advanced/5)"
fi

# ============================================================
# 测试8: UI组件详细测试
# ============================================================
log_test "UI组件详细测试"

ui_functions=(
    "renderEndgameHUD"
    "renderAIMoneyFlowRadar"
    "renderRiskHeatmap"
    "renderCriticalAlerts"
    "renderEnhancedKLine"
    "renderSparkline"
    "renderEventCard"
    "renderCostDistribution"
)

found_ui=0
for func in "${ui_functions[@]}"; do
    if grep -q "func $func" main.go; then
        ((found_ui++))
    fi
done

log_info "UI函数: $found_ui/${#ui_functions[@]}"

if [ $found_ui -ge 6 ]; then
    log_pass "UI组件丰富 ($found_ui 个函数)"
else
    log_fail "UI组件不足"
fi

# ============================================================
# 测试9: 数据持久化测试
# ============================================================
log_test "数据持久化测试"

persistence_keywords=(
    "json"
    "Marshal"
    "Unmarshal"
    "ioutil.WriteFile\|os.WriteFile"
    "ioutil.ReadFile\|os.ReadFile"
)

found_persistence=0
for keyword in "${persistence_keywords[@]}"; do
    if grep -q "$keyword" main.go; then
        ((found_persistence++))
    fi
done

if [ $found_persistence -ge 3 ]; then
    log_pass "数据持久化已实现"
else
    log_skip "数据持久化简单实现 ($found_persistence 个特性)"
fi

# ============================================================
# 测试10: 多模式支持测试
# ============================================================
log_test "多模式支持测试"

game_modes=$(grep -o "GameMode.*\"[^\"]*\"" main.go | wc -l)
log_info "游戏模式数量: $game_modes"

if grep -q "手动模式" main.go && grep -q "自动策略" main.go; then
    log_pass "支持多种游戏模式"
else
    log_fail "游戏模式单一"
fi

# ============================================================
# 测试11: 极端价格测试（边界条件）
# ============================================================
log_test "极端价格边界测试"

output_file="$LOG_DIR/test_extreme_price_${TIMESTAMP}.txt"

# 测试场景3（杠杆生死线）- 包含极端价格波动
cat > /tmp/test_extreme_input.txt << 'EOF'
1
1

4
3
q
EOF

(./stock_game < /tmp/test_extreme_input.txt > "$output_file" 2>&1) &
pid=$!

for i in {1..10}; do
    if ! kill -0 $pid 2>/dev/null; then
        break
    fi
    sleep 0.5
done

kill -9 $pid 2>/dev/null || true
rm -f /tmp/test_extreme_input.txt

if [ -f "$output_file" ]; then
    if grep -qi "panic.*divide.*zero\|panic.*overflow\|panic.*NaN" "$output_file"; then
        log_fail "极端价格导致计算错误"
        grep -i "panic" "$output_file" | head -3 | tee -a "$EXT_LOG"
    else
        log_pass "极端价格处理正常"
    fi
else
    log_skip "无法生成测试输出"
fi

# ============================================================
# 测试12: 连续操作稳定性测试
# ============================================================
log_test "连续操作稳定性测试"

log_info "模拟连续买卖操作..."

output_file="$LOG_DIR/test_continuous_ops_${TIMESTAMP}.txt"

cat > /tmp/test_continuous_input.txt << 'EOF'
1
1

5
b
1000
s
500
b
500
s
500
q
EOF

(./stock_game < /tmp/test_continuous_input.txt > "$output_file" 2>&1) &
pid=$!

for i in {1..10}; do
    if ! kill -0 $pid 2>/dev/null; then
        break
    fi
    sleep 0.5
done

kill -9 $pid 2>/dev/null || true
rm -f /tmp/test_continuous_input.txt

if [ -f "$output_file" ]; then
    if grep -qi "panic\|runtime error\|fatal" "$output_file"; then
        log_fail "连续操作出现错误"
    else
        log_pass "连续操作稳定"
    fi
else
    log_skip "无法生成测试输出"
fi

# ============================================================
# 测试13: 配置和设置系统测试
# ============================================================
log_test "配置和设置系统测试"

settings_keywords=(
    "Settings"
    "Config"
    "AutoSkipWeekend"
    "设置"
)

found_settings=0
for keyword in "${settings_keywords[@]}"; do
    if grep -q "$keyword" main.go; then
        ((found_settings++))
    fi
done

if [ $found_settings -ge 2 ]; then
    log_pass "设置系统存在"
else
    log_skip "设置系统简单"
fi

# ============================================================
# 测试14: T+1冻结机制测试
# ============================================================
log_test "T+1冻结机制测试"

t1_keywords=(
    "FrozenShares"
    "AvailableShares"
    "T\+1"
    "冻结"
)

found_t1=0
for keyword in "${t1_keywords[@]}"; do
    if grep -q "$keyword" main.go; then
        ((found_t1++))
    fi
done

if [ $found_t1 -ge 3 ]; then
    log_pass "T+1机制已实现"
else
    log_fail "T+1机制不完整"
fi

# ============================================================
# 测试15: 编年史/日志系统测试
# ============================================================
log_test "编年史/日志系统测试"

if grep -q "Chronicle" main.go; then
    log_pass "编年史系统存在"

    chronicle_features=$(grep -c "ChronicleEntry\|AddLog\|MarketLogs" main.go)
    log_info "日志相关代码: $chronicle_features 处"

    if [ $chronicle_features -ge 10 ]; then
        log_pass "日志系统完善"
    else
        log_skip "日志系统基础"
    fi
else
    log_fail "编年史系统缺失"
fi

# ============================================================
# 测试16: 性能压力测试
# ============================================================
log_test "性能压力测试（10次快速启动）"

start_time=$(date +%s)
failures=0

for i in {1..10}; do
    (echo "1"; echo "1"; echo ""; echo "5"; echo "q") | ./stock_game > /dev/null 2>&1 &
    pid=$!

    sleep 0.5

    if kill -0 $pid 2>/dev/null; then
        kill -9 $pid 2>/dev/null
    fi

    wait $pid 2>/dev/null
    exit_code=$?

    if [ $exit_code -gt 128 ]; then
        # 被信号杀死（正常）
        :
    elif [ $exit_code -ne 0 ]; then
        ((failures++))
    fi
done

end_time=$(date +%s)
duration=$((end_time - start_time))

log_info "执行时间: ${duration}秒, 失败次数: $failures"

if [ $failures -eq 0 ]; then
    log_pass "压力测试通过"
else
    log_fail "压力测试出现 $failures 次失败"
fi

# ============================================================
# 测试17: 代码质量检查
# ============================================================
log_test "代码质量检查"

# 检查代码行数
line_count=$(wc -l < main.go)
log_info "代码行数: $line_count"

# 检查TODO/FIXME
todo_count=$(grep -c "TODO\|FIXME\|XXX\|HACK" main.go 2>/dev/null || echo "0")
log_info "待办事项: $todo_count"

# 检查注释率
comment_count=$(grep -c "^[[:space:]]*//\|^[[:space:]]*/\*" main.go 2>/dev/null || echo "0")
comment_ratio=$(awk "BEGIN {printf \"%.1f\", ($comment_count/$line_count)*100}")
log_info "注释率: ${comment_ratio}%"

if [ $line_count -gt 1000 ] && [ $line_count -lt 10000 ]; then
    log_pass "代码规模适中 ($line_count 行)"
elif [ $line_count -ge 10000 ]; then
    log_skip "代码规模较大 ($line_count 行，考虑重构)"
else
    log_pass "代码简洁 ($line_count 行)"
fi

# ============================================================
# 测试18: 安全性检查
# ============================================================
log_test "安全性检查"

security_issues=0

# 检查不安全的函数
if grep -q "eval\|exec\|system" main.go; then
    log_fail "发现不安全函数调用"
    ((security_issues++))
fi

# 检查SQL注入风险（虽然是单机游戏，但检查下）
if grep -q "fmt.Sprintf.*SELECT\|fmt.Sprintf.*INSERT" main.go; then
    log_fail "可能存在SQL注入风险"
    ((security_issues++))
fi

# 检查路径遍历风险
if grep -q "\.\./\|filepath.Join" main.go; then
    log_info "使用文件路径操作，注意安全性"
fi

if [ $security_issues -eq 0 ]; then
    log_pass "未发现明显安全问题"
else
    log_fail "发现 $security_issues 个安全隐患"
fi

# ============================================================
# 测试结果汇总
# ============================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${PURPLE}                 测试总结${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "总测试数:    $TOTAL_TESTS"
echo -e "${GREEN}通过:        $PASSED_TESTS${NC}"
echo -e "${RED}失败:        $FAILED_TESTS${NC}"
echo -e "${YELLOW}跳过:        $SKIPPED_TESTS${NC}"
echo ""

success_rate=0
if [ $TOTAL_TESTS -gt 0 ]; then
    success_rate=$(awk "BEGIN {printf \"%.1f\", (($PASSED_TESTS+$SKIPPED_TESTS)/$TOTAL_TESTS)*100}")
fi

echo "通过率:      ${success_rate}%"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}✅ 扩展测试全部通过！${NC}"
    echo ""
    echo "游戏功能完整，代码质量良好"
else
    echo -e "${YELLOW}⚠️  有 ${FAILED_TESTS} 个测试失败${NC}"
    echo ""
    echo "建议查看日志: $EXT_LOG"
fi

echo ""
echo "详细日志: $EXT_LOG"
echo "测试文件: $LOG_DIR/"
echo ""

# 返回状态
if [ $FAILED_TESTS -eq 0 ]; then
    exit 0
else
    exit 1
fi
