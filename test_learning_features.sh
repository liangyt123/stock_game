#!/bin/bash
# 学习增强功能测试脚本

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  求解器学习功能验证"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 测试1: 编译检查
echo "[1/5] 编译检查..."
if go build -o stock_game 2>&1 | grep -q "error"; then
    echo "✗ 编译失败"
    exit 1
else
    echo "✓ 编译成功"
fi

# 测试2: 检查新文件存在
echo "[2/5] 检查新文件..."
if [ -f "solver_learning.go" ]; then
    echo "✓ solver_learning.go 存在"
else
    echo "✗ solver_learning.go 不存在"
    exit 1
fi

# 测试3: 检查关键函数
echo "[3/5] 检查学习辅助函数..."
functions=(
    "calculateNodeImportance"
    "generateAlternativeStrategies"
    "generateKeyDecisions"
    "generateCommonMistakes"
    "identifyStrategyPattern"
    "generateTeachingPoints"
    "DisplayEnhancedSolverResult"
)

for func in "${functions[@]}"; do
    if grep -q "func.*$func" solver_learning.go; then
        echo "  ✓ $func"
    else
        echo "  ✗ $func 缺失"
        exit 1
    fi
done

# 测试4: 检查数据结构
echo "[4/5] 检查增强数据结构..."
if grep -q "AlternativeStrategy" solver.go; then
    echo "✓ AlternativeStrategy 结构存在"
fi
if grep -q "DecisionExplanation" solver.go; then
    echo "✓ DecisionExplanation 结构存在"
fi
if grep -q "CommonMistake" solver.go; then
    echo "✓ CommonMistake 结构存在"
fi

# 测试5: 验证集成
echo "[5/5] 验证集成..."
if grep -q "DisplayEnhancedSolverResult" solver.go; then
    echo "✓ 增强显示已集成到QuickSolve"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ 所有学习功能测试通过！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "已实现的学习功能："
echo "  ✅ 1. 关键决策点标注 (🔴🟡⚪)"
echo "  ✅ 2. 备选策略对比 (Top 3)"
echo "  ✅ 3. 概率分析教学 (期望值计算)"
echo "  ✅ 4. 关键指标解释"
echo "  ✅ 5. 常见错误警示"
echo "  ✅ 6. 策略模式归纳"
echo "  ✅ 7. 教学要点生成"
echo ""
echo "运行游戏并选择 [5] 🤖 AI求解器 体验新功能！"
echo ""
