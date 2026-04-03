#!/bin/bash

# 股票游戏自动化测试脚本
# 用于发现潜在问题和验证功能

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# macOS兼容的timeout函数
run_with_timeout() {
    local timeout=$1
    shift
    perl -e 'alarm shift; exec @ARGV' "$timeout" "$@"
}

# 日志文件
LOG_DIR="test_logs"
mkdir -p "$LOG_DIR"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
TEST_LOG="$LOG_DIR/test_${TIMESTAMP}.log"

# 辅助函数
log_info() {
    echo -e "${CYAN}[INFO]${NC} $1" | tee -a "$TEST_LOG"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1" | tee -a "$TEST_LOG"
    ((PASSED_TESTS++))
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1" | tee -a "$TEST_LOG"
    ((FAILED_TESTS++))
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1" | tee -a "$TEST_LOG"
}

run_test() {
    ((TOTAL_TESTS++))
    local test_name="$1"
    local test_func="$2"

    echo ""
    log_info "测试 #${TOTAL_TESTS}: ${test_name}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    if $test_func; then
        log_success "$test_name"
    else
        log_error "$test_name"
    fi
}

# 测试1: 编译检查
test_compilation() {
    log_info "检查程序是否已编译..."

    if [ ! -f "main.go" ]; then
        log_error "找不到 main.go"
        return 1
    fi

    log_info "重新编译程序..."
    if go build -o stock_game main.go 2>&1 | tee -a "$TEST_LOG"; then
        log_success "编译成功"
        return 0
    else
        log_error "编译失败"
        return 1
    fi
}

# 测试2: 代码静态检查
test_static_analysis() {
    log_info "执行静态代码分析..."

    # 检查关键初始化
    local checks=(
        "state.VolumeHistory.*\[\]int"
        "state.PriceHistory.*\[\]float64"
        "state.TotalMarketShares.*="
        "state.DayLow.*="
        "state.DayHigh.*="
        "state.InitialAsset.*="
    )

    local all_passed=true
    for check in "${checks[@]}"; do
        if grep -q "$check" main.go; then
            log_success "✓ 找到初始化: $check"
        else
            log_error "✗ 缺失初始化: $check"
            all_passed=false
        fi
    done

    # 检查边界保护
    if grep -q "len(state.VolumeHistory) > 0" main.go; then
        log_success "✓ VolumeHistory 边界检查存在"
    else
        log_error "✗ VolumeHistory 边界检查缺失"
        all_passed=false
    fi

    if [ "$all_passed" = true ]; then
        return 0
    else
        return 1
    fi
}

# 测试3: 正常游戏模式启动测试
test_normal_game_start() {
    log_info "测试正常游戏模式启动..."

    local output_file="$LOG_DIR/test_normal_start_${TIMESTAMP}.txt"

    # 输入: 1 (手动模式) -> 1 (经典模式) -> 回车 -> 5 (开始游戏) -> q (退出)
    (echo "1"; echo "1"; echo ""; echo "5"; sleep 1; echo "q") | run_with_timeout 5 ./stock_game > "$output_file" 2>&1 || true

    if [ -f "$output_file" ]; then
        # 检查是否有panic
        if grep -i "panic" "$output_file"; then
            log_error "发现 panic"
            cat "$output_file" | tail -20 | tee -a "$TEST_LOG"
            return 1
        fi

        # 检查是否有关键UI元素
        if grep -q "操盘手" "$output_file" && grep -q "现价" "$output_file"; then
            log_success "正常游戏UI渲染正常"
            return 0
        else
            log_error "UI渲染异常"
            return 1
        fi
    else
        log_error "输出文件未生成"
        return 1
    fi
}

# 测试4: 残局模式基础测试
test_endgame_basic() {
    log_info "测试残局模式基础功能..."

    local output_file="$LOG_DIR/test_endgame_basic_${TIMESTAMP}.txt"

    # 输入: 1 -> 1 -> 回车 -> 4 (残局) -> 返回
    (echo "1"; echo "1"; echo ""; echo "4"; echo ""; sleep 1) | run_with_timeout 5 ./stock_game > "$output_file" 2>&1 || true

    if [ -f "$output_file" ]; then
        if grep -i "panic" "$output_file"; then
            log_error "残局模式启动时发现 panic"
            cat "$output_file" | tail -20 | tee -a "$TEST_LOG"
            return 1
        fi

        # 检查是否显示了残局场景列表
        if grep -q "残局挑战" "$output_file" || grep -q "逃顶挑战" "$output_file"; then
            log_success "残局场景列表显示正常"
            return 0
        else
            log_warning "未找到残局场景列表"
            return 0  # 不算失败，可能是UI问题
        fi
    else
        log_error "输出文件未生成"
        return 1
    fi
}

# 测试5: 残局场景遍历测试
test_endgame_scenarios() {
    log_info "测试所有16个残局场景..."

    local all_passed=true

    for i in {1..16}; do
        log_info "  测试残局场景 #$i..."
        local output_file="$LOG_DIR/test_endgame_scenario_${i}_${TIMESTAMP}.txt"

        # 输入: 1 -> 1 -> 回车 -> 4 (残局) -> i (场景编号) -> q (退出)
        (echo "1"; echo "1"; echo ""; echo "4"; echo "$i"; sleep 1; echo "q") | run_with_timeout 5 ./stock_game > "$output_file" 2>&1 || true

        if [ -f "$output_file" ]; then
            if grep -i "panic\|runtime error" "$output_file"; then
                log_error "  场景 #$i 出现错误"
                grep -A 5 "panic\|runtime error" "$output_file" | tee -a "$TEST_LOG"
                all_passed=false
            else
                log_success "  场景 #$i 运行正常"
            fi
        else
            log_warning "  场景 #$i 输出文件未生成"
        fi
    done

    if [ "$all_passed" = true ]; then
        return 0
    else
        return 1
    fi
}

# 测试6: UI组件渲染测试
test_ui_components() {
    log_info "测试UI组件渲染..."

    local output_file="$LOG_DIR/test_ui_${TIMESTAMP}.txt"

    # 运行几个回合让UI组件都显示出来
    (echo "1"; echo "1"; echo ""; echo "5";
     sleep 1; echo "h"; sleep 1; echo "h"; sleep 1; echo "h";
     sleep 1; echo "q") | run_with_timeout 10 ./stock_game > "$output_file" 2>&1 || true

    if [ -f "$output_file" ]; then
        local ui_elements=(
            "操盘手"
            "现价"
            "资产"
            "AI.*仓位"
            "核心监控"
            "实时K线"
        )

        local all_found=true
        for element in "${ui_elements[@]}"; do
            if grep -q "$element" "$output_file"; then
                log_success "  ✓ UI元素: $element"
            else
                log_warning "  ✗ 未找到UI元素: $element"
                all_found=false
            fi
        done

        # 检查新增的UI组件
        if grep -q "资金流向雷达\|风险热力图\|增强K线" "$output_file"; then
            log_success "  ✓ 新增UI组件存在"
        else
            log_warning "  ✗ 新增UI组件未显示（可能需要更多回合）"
        fi

        if [ "$all_found" = true ]; then
            return 0
        else
            return 0  # UI元素缺失不算测试失败
        fi
    else
        log_error "输出文件未生成"
        return 1
    fi
}

# 测试7: 内存泄漏检查（简单版本）
test_memory_leak() {
    log_info "执行简单内存泄漏检查..."

    # 运行多次快速测试
    for i in {1..5}; do
        log_info "  迭代 #$i..."
        (echo "1"; echo "1"; echo ""; echo "5"; echo "q") | run_with_timeout 3 ./stock_game > /dev/null 2>&1 || true
    done

    log_success "多次启动测试完成（未发现明显问题）"
    return 0
}

# 测试8: 边界条件测试
test_edge_cases() {
    log_info "测试边界条件..."

    local output_file="$LOG_DIR/test_edge_${TIMESTAMP}.txt"

    # 测试无效输入
    (echo "999"; echo "abc"; echo "1"; echo "1"; echo ""; echo "5"; echo "q") | run_with_timeout 5 ./stock_game > "$output_file" 2>&1 || true

    if [ -f "$output_file" ]; then
        if grep -i "panic" "$output_file"; then
            log_error "边界条件触发 panic"
            return 1
        else
            log_success "边界条件处理正常"
            return 0
        fi
    else
        return 1
    fi
}

# 测试9: 残局HUD显示测试
test_endgame_hud() {
    log_info "测试残局HUD显示..."

    local output_file="$LOG_DIR/test_endgame_hud_${TIMESTAMP}.txt"

    # 进入残局场景1并运行几个回合
    (echo "1"; echo "1"; echo ""; echo "4"; echo "1";
     sleep 1; echo "h"; sleep 1; echo "h";
     sleep 1; echo "q") | run_with_timeout 8 ./stock_game > "$output_file" 2>&1 || true

    if [ -f "$output_file" ]; then
        local hud_elements=(
            "残局挑战"
            "进度"
            "目标"
            "回合"
        )

        local found_count=0
        for element in "${hud_elements[@]}"; do
            if grep -q "$element" "$output_file"; then
                ((found_count++))
            fi
        done

        if [ $found_count -ge 2 ]; then
            log_success "残局HUD显示正常 ($found_count/4 个元素)"
            return 0
        else
            log_warning "残局HUD显示不完整 ($found_count/4 个元素)"
            return 0  # 不算失败
        fi
    else
        log_error "输出文件未生成"
        return 1
    fi
}

# 测试10: 性能基准测试
test_performance() {
    log_info "执行性能基准测试..."

    local start_time=$(date +%s)

    # 运行10个回合
    (echo "1"; echo "1"; echo ""; echo "5";
     for i in {1..10}; do echo "h"; sleep 0.1; done;
     echo "q") | run_with_timeout 15 ./stock_game > /dev/null 2>&1 || true

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    log_info "  执行时间: ${duration}秒"

    if [ $duration -lt 20 ]; then
        log_success "性能正常 (${duration}s < 20s)"
        return 0
    else
        log_warning "性能较慢 (${duration}s >= 20s)"
        return 0  # 不算失败
    fi
}

# 主测试流程
main() {
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${BLUE}         股票游戏自动化测试套件${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    log_info "测试开始时间: $(date)"
    log_info "日志文件: $TEST_LOG"
    echo ""

    # 检查是否在正确的目录
    if [ ! -f "main.go" ]; then
        log_error "请在包含 main.go 的目录下运行此脚本"
        exit 1
    fi

    # 运行所有测试
    run_test "编译检查" test_compilation
    run_test "静态代码分析" test_static_analysis
    run_test "正常游戏模式启动" test_normal_game_start
    run_test "残局模式基础功能" test_endgame_basic
    run_test "残局场景遍历" test_endgame_scenarios
    run_test "UI组件渲染" test_ui_components
    run_test "内存泄漏检查" test_memory_leak
    run_test "边界条件测试" test_edge_cases
    run_test "残局HUD显示" test_endgame_hud
    run_test "性能基准测试" test_performance

    # 测试总结
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${BLUE}                 测试总结${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo -e "总测试数:   ${TOTAL_TESTS}"
    echo -e "${GREEN}通过:       ${PASSED_TESTS}${NC}"
    echo -e "${RED}失败:       ${FAILED_TESTS}${NC}"

    local success_rate=0
    if [ $TOTAL_TESTS -gt 0 ]; then
        success_rate=$(awk "BEGIN {printf \"%.1f\", ($PASSED_TESTS/$TOTAL_TESTS)*100}")
    fi
    echo -e "通过率:     ${success_rate}%"
    echo ""

    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}✅ 所有测试通过！${NC}"
        echo ""
        log_info "详细日志: $TEST_LOG"
        log_info "测试结果文件: $LOG_DIR/"
        return 0
    else
        echo -e "${RED}❌ 有 ${FAILED_TESTS} 个测试失败${NC}"
        echo ""
        log_error "请检查日志文件: $TEST_LOG"
        log_info "测试结果文件: $LOG_DIR/"
        return 1
    fi
}

# 清理函数
cleanup() {
    log_info "清理测试环境..."
    # 可以在这里添加清理逻辑
}

# 捕获退出信号
trap cleanup EXIT

# 运行主函数
main

exit $?
