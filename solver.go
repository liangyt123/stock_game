package main

import (
	"fmt"
	"math"
)

// ============================================================
// 残局求解器 - 使用动态规划找到最优策略
// ============================================================

// SolverAction 表示一个可能的行动
type SolverAction struct {
	Type        string  // "Buy", "Sell", "Hold"
	Shares      int     // 交易股数
	Description string  // 行动描述
}

// SolverState 表示求解器中的游戏状态
type SolverState struct {
	Day            int
	Session        string
	Price          float64
	PlayerShares   int
	PlayerCash     float64
	MarginDebt     float64
	EscapedAIRatio float64
	CrashWarning   int
	TurnNumber     int
}

// SolverNode 表示决策树中的节点
type SolverNode struct {
	State          *SolverState
	Action         *SolverAction
	ExpectedValue  float64 // 期望收益
	BestChildPath  []*SolverNode
	Probability    float64 // 到达此节点的概率
	Depth          int
	Importance     string  // 🔴关键 / 🟡重要 / ⚪普通
	ImportanceReason string // 为什么重要
	ValueImpact    float64 // 对最终收益的影响
}

// AlternativeStrategy 备选策略
type AlternativeStrategy struct {
	Name           string
	Path           []*SolverNode
	ExpectedReturn float64
	SuccessRate    float64
	WorstCase      float64
	RiskLevel      string
	SuitableFor    string
	Pros           []string
	Cons           []string
}

// DecisionExplanation 决策解释
type DecisionExplanation struct {
	Turn           int
	Action         string
	AlternativeActions []string
	WhyOptimal     string
	ProbabilityBreakdown map[string]float64 // 每种情况的概率
	ExpectedValueCalc    string // 期望值计算过程
	KeyIndicators        map[string]interface{} // 关键指标
}

// CommonMistake 常见错误
type CommonMistake struct {
	MistakeType  string
	Description  string
	Impact       string
	CorrectAction string
	CaseStudy    string
}

// SolverResult 求解结果
type SolverResult struct {
	OptimalPath        []*SolverNode
	AlternativePaths   []AlternativeStrategy // 备选策略
	ExpectedReturn     float64
	SuccessProbability float64
	WorstCase          float64
	BestCase           float64
	Strategy           string
	AnalysisSteps      []string
	KeyDecisions       []DecisionExplanation // 关键决策详解
	CommonMistakes     []CommonMistake       // 该场景常见错误
	StrategyPattern    string                // 策略模式（抄底/逃顶/震荡）
	TeachingPoints     []string              // 教学要点
}

// EndgameSolver 残局求解器
type EndgameSolver struct {
	Scenario      *EndgameScenario
	MaxDepth      int // 最大搜索深度
	memo          map[string]float64 // 记忆化缓存
	exploredNodes int
}

// NewEndgameSolver 创建求解器
func NewEndgameSolver(scenario *EndgameScenario) *EndgameSolver {
	return &EndgameSolver{
		Scenario:  scenario,
		MaxDepth:  scenario.TimeLimit, // 搜索到时间限制
		memo:      make(map[string]float64),
	}
}

// Solve 求解最优策略
func (s *EndgameSolver) Solve() *SolverResult {
	fmt.Printf("\n%s╔═══════════════════════ 🤖 AI求解器 ═══════════════════════╗%s\n", Cyan, Reset)
	fmt.Printf("%s║%s  正在分析残局场景: %s%s%s\n", Cyan, Reset, Yellow, s.Scenario.Name, Reset)
	fmt.Printf("%s║%s  搜索深度: %s%d回合%s  |  难度: %s%s%s\n",
		Cyan, Reset, Yellow, s.MaxDepth, Reset, Yellow, s.Scenario.Difficulty, Reset)
	fmt.Printf("%s╚═══════════════════════════════════════════════════════════╝%s\n", Cyan, Reset)
	fmt.Println()

	// 初始状态
	initialState := &SolverState{
		Day:            s.Scenario.StartDay,
		Session:        s.Scenario.StartSession,
		Price:          s.Scenario.InitialPrice,
		PlayerShares:   s.Scenario.PlayerShares,
		PlayerCash:     s.Scenario.PlayerCash,
		MarginDebt:     s.Scenario.MarginDebt,
		EscapedAIRatio: s.Scenario.AIExitRatio,
		CrashWarning:   s.Scenario.CrashWarning,
		TurnNumber:     0,
	}

	// 递归搜索最优路径
	fmt.Printf("%s[搜索中]%s ", Yellow, Reset)
	bestNode := s.searchBestPath(initialState, 0)
	fmt.Printf(" %s完成！%s\n", Green, Reset)
	fmt.Printf("%s探索了 %d 个节点%s\n\n", Cyan, s.exploredNodes, Reset)

	// 构建结果
	result := s.buildResult(bestNode, initialState)

	return result
}

// searchBestPath 递归搜索最优路径（带记忆化）
func (s *EndgameSolver) searchBestPath(state *SolverState, depth int) *SolverNode {
	s.exploredNodes++

	// 显示进度
	if s.exploredNodes%100 == 0 {
		fmt.Printf(".")
	}

	// 终止条件
	if depth >= s.MaxDepth || s.isTerminalState(state) {
		return &SolverNode{
			State:         state,
			ExpectedValue: s.evaluateTerminalValue(state),
			Depth:         depth,
		}
	}

	// 检查记忆化缓存
	stateKey := s.stateToKey(state)
	if cachedValue, exists := s.memo[stateKey]; exists {
		return &SolverNode{
			State:         state,
			ExpectedValue: cachedValue,
			Depth:         depth,
		}
	}

	// 生成所有可能的行动
	actions := s.generateActions(state)

	var bestNode *SolverNode
	bestValue := -math.MaxFloat64

	// 对每个行动，模拟下一状态并递归
	for _, action := range actions {
		// 预测下一回合的可能状态
		possibleStates := s.predictNextStates(state, action)

		// 计算期望价值（考虑概率）
		expectedValue := 0.0
		var bestChildPath []*SolverNode

		for _, nextStatePair := range possibleStates {
			nextState := nextStatePair.state
			probability := nextStatePair.probability

			// 递归搜索
			childNode := s.searchBestPath(nextState, depth+1)

			expectedValue += probability * childNode.ExpectedValue

			if len(bestChildPath) == 0 {
				bestChildPath = append([]*SolverNode{childNode}, childNode.BestChildPath...)
			}
		}

		// 更新最佳节点
		if expectedValue > bestValue {
			bestValue = expectedValue
			bestNode = &SolverNode{
				State:          state,
				Action:         action,
				ExpectedValue:  expectedValue,
				BestChildPath:  bestChildPath,
				Depth:          depth,
			}
		}
	}

	// 缓存结果
	s.memo[stateKey] = bestValue

	return bestNode
}

// stateProbabilityPair 状态-概率对
type stateProbabilityPair struct {
	state       *SolverState
	probability float64
}

// predictNextStates 预测下一回合的可能状态
func (s *EndgameSolver) predictNextStates(currentState *SolverState, action *SolverAction) []stateProbabilityPair {
	// 执行行动后的状态
	newState := s.applyAction(currentState, action)

	// 预测价格变化（基于历史和场景特征）
	priceChanges := s.predictPriceChanges(newState)

	var results []stateProbabilityPair

	for _, change := range priceChanges {
		nextState := s.copyState(newState)
		nextState.Price = nextState.Price * (1 + change.changeRate)
		nextState.EscapedAIRatio = s.predictAIExitRatio(nextState)
		nextState.CrashWarning = s.predictCrashWarning(nextState)
		nextState.TurnNumber++
		s.advanceSession(nextState)

		results = append(results, stateProbabilityPair{
			state:       nextState,
			probability: change.probability,
		})
	}

	return results
}

// priceChangeProbability 价格变化-概率对
type priceChangeProbability struct {
	changeRate  float64 // 变化率 (-0.1 = -10%)
	probability float64
}

// predictPriceChanges 预测价格变化概率分布
func (s *EndgameSolver) predictPriceChanges(state *SolverState) []priceChangeProbability {
	// 基础场景：3种可能（涨/平/跌）
	baseProb := 0.33

	// 根据场景特征调整概率
	upProb := baseProb
	flatProb := baseProb
	downProb := baseProb

	// 因素1: 崩盘预警
	if state.CrashWarning >= 3 {
		downProb += 0.3
		upProb -= 0.15
		flatProb -= 0.15
	}

	// 因素2: AI逃跑
	if state.EscapedAIRatio > 0.5 {
		downProb += 0.2
		upProb -= 0.1
		flatProb -= 0.1
	}

	// 因素3: 场景特定
	if s.Scenario.IsMonsterStock {
		upProb += 0.2
		downProb -= 0.1
		flatProb -= 0.1
	}

	// 归一化
	total := upProb + flatProb + downProb
	upProb /= total
	flatProb /= total
	downProb /= total

	// 重要：根据场景事件的情绪值 (Sentiment) 动态调整步幅
	// 如果事件情绪极佳 (Sentiment > 1.2), 涨幅从 5% 提升到 8%
	stepSize := 0.05
	if s.Scenario.EventPreset != nil {
		if s.Scenario.EventPreset.Sentiment > 1.2 {
			stepSize = 0.08
			upProb += 0.1 // 增加上涨权重
		} else if s.Scenario.EventPreset.Sentiment > 1.05 {
			stepSize = 0.06
			upProb += 0.05
		}
		
		// 再次归一化
		norm := upProb + flatProb + downProb
		upProb /= norm
		flatProb /= norm
		downProb /= norm
	}

	return []priceChangeProbability{
		{changeRate: stepSize, probability: upProb},   // 动态涨幅
		{changeRate: 0.0, probability: flatProb},      // 持平
		{changeRate: -stepSize, probability: downProb}, // 动态跌幅
	}
}

// applyAction 应用行动到状态
func (s *EndgameSolver) applyAction(state *SolverState, action *SolverAction) *SolverState {
	newState := s.copyState(state)

	switch action.Type {
	case "Buy":
		cost := float64(action.Shares) * state.Price
		if cost <= newState.PlayerCash {
			newState.PlayerShares += action.Shares
			newState.PlayerCash -= cost
		}

	case "Sell":
		sellShares := action.Shares
		if sellShares > newState.PlayerShares {
			sellShares = newState.PlayerShares
		}
		newState.PlayerShares -= sellShares
		newState.PlayerCash += float64(sellShares) * state.Price

	case "Hold":
		// 不做操作
	}

	return newState
}

// generateActions 生成所有可能的行动
func (s *EndgameSolver) generateActions(state *SolverState) []*SolverAction {
	actions := []*SolverAction{
		{Type: "Hold", Shares: 0, Description: "持有"},
	}

	// 买入选项（如果有现金）
	if state.PlayerCash > state.Price*100 {
		maxBuy := int(state.PlayerCash / state.Price / 100) * 100 // 100股为单位
		if maxBuy > 0 {
			// 小量买入
			actions = append(actions, &SolverAction{
				Type:        "Buy",
				Shares:      min_int(500, maxBuy),
				Description: fmt.Sprintf("买入%d股", min_int(500, maxBuy)),
			})
			// 大量买入
			if maxBuy > 1000 {
				actions = append(actions, &SolverAction{
					Type:        "Buy",
					Shares:      maxBuy / 2,
					Description: fmt.Sprintf("买入%d股", maxBuy/2),
				})
			}
		}
	}

	// 卖出选项（如果持有股票）
	if state.PlayerShares > 0 {
		// 部分卖出
		actions = append(actions, &SolverAction{
			Type:        "Sell",
			Shares:      state.PlayerShares / 2,
			Description: fmt.Sprintf("卖出一半(%d股)", state.PlayerShares/2),
		})
		// 全部卖出
		actions = append(actions, &SolverAction{
			Type:        "Sell",
			Shares:      state.PlayerShares,
			Description: fmt.Sprintf("清仓(%d股)", state.PlayerShares),
		})
	}

	return actions
}

// evaluateTerminalValue 评估终局状态的价值
func (s *EndgameSolver) evaluateTerminalValue(state *SolverState) float64 {
	// 计算最终资产
	totalAsset := float64(state.PlayerShares)*state.Price + state.PlayerCash - state.MarginDebt

	// 初始资产同步：使用游戏统一的 $10.0 平均成本作为基准
	// 这样 solver 的目标评估将与 main.go 的 HUD 完全对齐
	const defaultAvgCost = 10.0
	initialAsset := float64(s.Scenario.PlayerShares)*defaultAvgCost +
		s.Scenario.PlayerCash - s.Scenario.MarginDebt

	// 收益率
	returnRate := (totalAsset - initialAsset) / initialAsset

	// 考虑风险惩罚
	riskPenalty := 0.0

	// 惩罚1: 爆仓风险
	if state.MarginDebt > 0 && totalAsset < state.MarginDebt*1.2 {
		riskPenalty += 0.2 // -20%惩罚
	}

	// 惩罚2: 未达到目标
	if returnRate < s.Scenario.TargetProfit {
		gap := s.Scenario.TargetProfit - returnRate
		riskPenalty += gap * 0.5 // 差距的50%作为惩罚
	}

	// 奖励：超额完成
	bonus := 0.0
	if returnRate > s.Scenario.TargetProfit {
		excess := returnRate - s.Scenario.TargetProfit
		bonus = excess * 0.3 // 超额的30%作为奖励
	}

	return returnRate - riskPenalty + bonus
}

// isTerminalState 判断是否为终局状态
func (s *EndgameSolver) isTerminalState(state *SolverState) bool {
	// 时间到了
	if state.TurnNumber >= s.MaxDepth {
		return true
	}

	// 爆仓了
	if state.MarginDebt > 0 {
		netAsset := float64(state.PlayerShares)*state.Price + state.PlayerCash
		if netAsset <= state.MarginDebt {
			return true
		}
	}

	// 已经清仓且无法再买入
	if state.PlayerShares == 0 && state.PlayerCash < state.Price*100 {
		return true
	}

	return false
}

// 辅助函数
func (s *EndgameSolver) stateToKey(state *SolverState) string {
	return fmt.Sprintf("%d_%s_%.2f_%d_%.0f_%.0f",
		state.Day, state.Session, state.Price,
		state.PlayerShares, state.PlayerCash, state.MarginDebt)
}

func (s *EndgameSolver) copyState(state *SolverState) *SolverState {
	return &SolverState{
		Day:            state.Day,
		Session:        state.Session,
		Price:          state.Price,
		PlayerShares:   state.PlayerShares,
		PlayerCash:     state.PlayerCash,
		MarginDebt:     state.MarginDebt,
		EscapedAIRatio: state.EscapedAIRatio,
		CrashWarning:   state.CrashWarning,
		TurnNumber:     state.TurnNumber,
	}
}

func (s *EndgameSolver) advanceSession(state *SolverState) {
	sessions := []string{"早盘", "午盘", "尾盘", "夜盘"}
	for i, sess := range sessions {
		if state.Session == sess {
			if i == len(sessions)-1 {
				state.Day++
				state.Session = sessions[0]
			} else {
				state.Session = sessions[i+1]
			}
			return
		}
	}
}

func (s *EndgameSolver) predictAIExitRatio(state *SolverState) float64 {
	// 简化：AI逃跑率随崩盘预警增加
	baseRatio := state.EscapedAIRatio
	increase := float64(state.CrashWarning) * 0.05
	return math.Min(baseRatio+increase, 0.9)
}

func (s *EndgameSolver) predictCrashWarning(state *SolverState) int {
	// 简化：警告等级随AI逃跑增加
	if state.EscapedAIRatio > 0.7 {
		return 4
	} else if state.EscapedAIRatio > 0.5 {
		return 3
	} else if state.EscapedAIRatio > 0.3 {
		return 2
	}
	return state.CrashWarning
}

// buildResult 构建求解结果
func (s *EndgameSolver) buildResult(bestNode *SolverNode, initialState *SolverState) *SolverResult {
	// 提取最优路径
	path := []*SolverNode{bestNode}
	path = append(path, bestNode.BestChildPath...)

	// 计算统计
	initialAsset := float64(initialState.PlayerShares)*initialState.Price +
		initialState.PlayerCash - initialState.MarginDebt

	finalState := path[len(path)-1].State
	finalAsset := float64(finalState.PlayerShares)*finalState.Price +
		finalState.PlayerCash - finalState.MarginDebt

	expectedReturn := (finalAsset - initialAsset) / initialAsset

	// 生成策略描述
	strategy := s.generateStrategyDescription(path)

	// 生成分析步骤
	analysisSteps := s.generateAnalysisSteps(path)

	// ===== 新增：学习辅助功能 =====

	// 1. 计算每个节点的重要性
	pathValues := make([]float64, len(path))
	for i, node := range path {
		pathValues[i] = node.ExpectedValue
	}
	for _, node := range path {
		s.calculateNodeImportance(node, pathValues)
	}

	// 2. 生成备选策略
	alternativePaths := s.generateAlternativeStrategies(path, initialState)

	// 3. 生成关键决策解释
	keyDecisions := s.generateKeyDecisions(path)

	// 4. 生成常见错误
	commonMistakes := s.generateCommonMistakes(s.Scenario)

	// 5. 识别策略模式
	strategyPattern := s.identifyStrategyPattern(s.Scenario, path)

	// 6. 生成教学要点
	teachingPoints := s.generateTeachingPoints(strategyPattern, s.Scenario)

	return &SolverResult{
		OptimalPath:        path,
		AlternativePaths:   alternativePaths,
		ExpectedReturn:     expectedReturn,
		SuccessProbability: s.calculateSuccessProbability(path),
		WorstCase:          expectedReturn * 0.7,
		BestCase:           expectedReturn * 1.3,
		Strategy:           strategy,
		AnalysisSteps:      analysisSteps,
		KeyDecisions:       keyDecisions,
		CommonMistakes:     commonMistakes,
		StrategyPattern:    strategyPattern,
		TeachingPoints:     teachingPoints,
	}
}

func (s *EndgameSolver) generateStrategyDescription(path []*SolverNode) string {
	// 分析路径特征
	holdCount := 0
	buyCount := 0
	sellCount := 0

	for _, node := range path {
		if node.Action != nil {
			switch node.Action.Type {
			case "Hold":
				holdCount++
			case "Buy":
				buyCount++
			case "Sell":
				sellCount++
			}
		}
	}

	// 生成描述
	if sellCount > buyCount {
		return "保守离场策略 - 优先保护利润"
	} else if buyCount > sellCount {
		return "激进加仓策略 - 追求高收益"
	} else {
		return "平衡策略 - 稳健操作"
	}
}

func (s *EndgameSolver) generateAnalysisSteps(path []*SolverNode) []string {
	steps := []string{}

	for i, node := range path {
		if node.Action == nil || i >= 10 { // 只显示前10步
			break
		}

		step := fmt.Sprintf("第%d回合: %s (预期收益: %.1f%%)",
			i+1, node.Action.Description, node.ExpectedValue*100)
		steps = append(steps, step)
	}

	return steps
}

func (s *EndgameSolver) calculateSuccessProbability(path []*SolverNode) float64 {
	// 简化：基于期望收益和目标的比较
	expectedReturn := path[0].ExpectedValue
	targetReturn := s.Scenario.TargetProfit

	if expectedReturn >= targetReturn {
		return 0.8 + math.Min(0.15, (expectedReturn-targetReturn)*0.5)
	} else {
		return 0.5 - math.Min(0.3, (targetReturn-expectedReturn)*0.5)
	}
}

// DisplayResult 显示求解结果
func DisplaySolverResult(result *SolverResult) {
	fmt.Printf("\n%s╔═══════════════════════ 📊 求解结果 ═══════════════════════╗%s\n", Green, Reset)

	// 总体评估
	returnColor := Green
	if result.ExpectedReturn < 0 {
		returnColor = Red
	}

	fmt.Printf("%s║%s  【最优策略】%s%s%s\n", Green, Reset, Yellow, result.Strategy, Reset)
	fmt.Printf("%s║%s  期望收益率: %s%.2f%%%s\n", Green, Reset, returnColor, result.ExpectedReturn*100, Reset)
	fmt.Printf("%s║%s  成功概率:   %s%.1f%%%s\n", Green, Reset, Cyan, result.SuccessProbability*100, Reset)
	fmt.Printf("%s║%s  最坏情况:   %s%.2f%%%s  |  最好情况: %s%.2f%%%s\n",
		Green, Reset, Red, result.WorstCase*100, Reset, Green, result.BestCase*100, Reset)

	// 策略步骤
	if len(result.AnalysisSteps) > 0 {
		fmt.Printf("%s║%s\n", Green, Reset)
		fmt.Printf("%s║%s  【关键决策点】\n", Green, Reset)

		displayCount := min_int(5, len(result.AnalysisSteps))
		for i := 0; i < displayCount; i++ {
			fmt.Printf("%s║%s    %s\n", Green, Reset, result.AnalysisSteps[i])
		}

		if len(result.AnalysisSteps) > 5 {
			fmt.Printf("%s║%s    ... (还有%d步)\n", Green, Reset, len(result.AnalysisSteps)-5)
		}
	}

	fmt.Printf("%s╚═══════════════════════════════════════════════════════════╝%s\n", Green, Reset)
}

// 快速求解接口 - 用于测试
func QuickSolve(scenarioIndex int) {
	if scenarioIndex < 0 || scenarioIndex >= len(EndgameScenarios) {
		fmt.Printf("%s错误: 场景索引无效%s\n", Red, Reset)
		return
	}

	scenario := &EndgameScenarios[scenarioIndex]

	fmt.Printf("\n%s════════════════════════════════════════════════════════════%s\n", Cyan, Reset)
	fmt.Printf("%s          残局求解器 - 寻找最优策略%s\n", Cyan, Reset)
	fmt.Printf("%s════════════════════════════════════════════════════════════%s\n", Cyan, Reset)

	fmt.Printf("\n%s场景: %s%s (%s)%s\n", Yellow, Cyan, scenario.Name, scenario.Difficulty, Reset)
	fmt.Printf("%s目标: 盈利 %.1f%%  |  时限: %d回合%s\n",
		Yellow, scenario.TargetProfit*100, scenario.TimeLimit, Reset)

	solver := NewEndgameSolver(scenario)
	result := solver.Solve()

	// 使用增强版显示（包含学习辅助功能）
	DisplayEnhancedSolverResult(result)

	// 策略建议
	fmt.Printf("\n%s💡 AI建议:%s\n", Yellow, Reset)
	if result.ExpectedReturn >= scenario.TargetProfit {
		fmt.Printf("  %s✓ 该场景有可行的最优解%s\n", Green, Reset)
		fmt.Printf("  %s→ 严格按照上述步骤操作，成功概率 %.0f%%%s\n",
			Cyan, result.SuccessProbability*100, Reset)
	} else {
		fmt.Printf("  %s! 该场景极具挑战性%s\n", Red, Reset)
		fmt.Printf("  %s→ 即使采用最优策略，达标概率也较低%s\n", Cyan, Reset)
		fmt.Printf("  %s→ 建议重点关注风险控制而非追求目标%s\n", Cyan, Reset)
	}

	fmt.Println()
}
