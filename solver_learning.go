package main

import (
	"fmt"
	"math"
)

// ============================================================
// 求解器学习辅助功能
// ============================================================

// ===== 1. 重要性标注系统 =====

// calculateNodeImportance 计算节点的重要性
func (s *EndgameSolver) calculateNodeImportance(node *SolverNode, pathValues []float64) {
	if node.Action == nil {
		node.Importance = "⚪"
		node.ImportanceReason = "初始状态"
		node.ValueImpact = 0
		return
	}

	// 计算该节点对最终收益的影响
	valueImpact := 0.0
	if len(pathValues) > node.Depth {
		currentValue := pathValues[node.Depth]
		if node.Depth > 0 {
			prevValue := pathValues[node.Depth-1]
			valueImpact = currentValue - prevValue
		}
	}

	node.ValueImpact = valueImpact

	// 根据影响程度标注重要性
	absImpact := math.Abs(valueImpact)

	if absImpact > 0.10 { // 影响>10%
		node.Importance = "🔴"
		node.ImportanceReason = fmt.Sprintf("关键决策点！该操作影响最终收益%.1f%%", valueImpact*100)
	} else if absImpact > 0.05 { // 影响5-10%
		node.Importance = "🟡"
		node.ImportanceReason = fmt.Sprintf("重要决策，影响收益%.1f%%", valueImpact*100)
	} else {
		node.Importance = "⚪"
		node.ImportanceReason = "常规操作"
	}

	// 特殊情况标注
	state := node.State
	if state.CrashWarning >= 4 && node.Action.Type == "Sell" {
		node.Importance = "🔴"
		node.ImportanceReason = "关键！崩盘前及时离场，避免巨额损失"
	} else if state.EscapedAIRatio > 0.5 && node.Action.Type == "Buy" {
		node.Importance = "🔴"
		node.ImportanceReason = "风险！在AI大量逃离时买入，需要格外谨慎"
	}
}

// ===== 2. 备选策略生成 =====

// generateAlternativeStrategies 生成Top 3备选策略
func (s *EndgameSolver) generateAlternativeStrategies(optimalPath []*SolverNode, initialState *SolverState) []AlternativeStrategy {
	strategies := []AlternativeStrategy{}

	// 策略A: 最优策略（激进）
	strategies = append(strategies, s.analyzeStrategy(optimalPath, initialState, "激进博弈", "高风险高回报"))

	// 策略B: 保守策略（提前止盈）
	conservativePath := s.generateConservativeStrategy(optimalPath)
	strategies = append(strategies, s.analyzeStrategy(conservativePath, initialState, "稳健跟随", "中等风险"))

	// 策略C: 极保守策略（快进快出）
	safePath := s.generateSafeStrategy(optimalPath)
	strategies = append(strategies, s.analyzeStrategy(safePath, initialState, "保守观望", "低风险"))

	return strategies
}

func (s *EndgameSolver) analyzeStrategy(path []*SolverNode, initialState *SolverState, name string, riskLevel string) AlternativeStrategy {
	initialAsset := float64(initialState.PlayerShares)*initialState.Price +
		initialState.PlayerCash - initialState.MarginDebt

	finalState := path[len(path)-1].State
	finalAsset := float64(finalState.PlayerShares)*finalState.Price +
		finalState.PlayerCash - finalState.MarginDebt

	expectedReturn := (finalAsset - initialAsset) / initialAsset

	strategy := AlternativeStrategy{
		Name:           name,
		Path:           path,
		ExpectedReturn: expectedReturn,
		RiskLevel:      riskLevel,
	}

	// 根据策略类型设置特征
	if name == "激进博弈" {
		strategy.SuccessRate = 0.65
		strategy.WorstCase = expectedReturn * 0.5
		strategy.SuitableFor = "经验丰富、风险承受能力强的玩家"
		strategy.Pros = []string{"收益最高", "把握最佳时机", "充分利用市场波动"}
		strategy.Cons = []string{"风险大", "容错率低", "需要精准判断"}
	} else if name == "稳健跟随" {
		strategy.SuccessRate = 0.85
		strategy.WorstCase = expectedReturn * 0.7
		strategy.SuitableFor = "稳健型玩家"
		strategy.Pros = []string{"风险可控", "成功率高", "心理压力小"}
		strategy.Cons = []string{"收益较低", "可能错过最高点"}
	} else {
		strategy.SuccessRate = 0.90
		strategy.WorstCase = expectedReturn * 0.9
		strategy.SuitableFor = "保守型、新手玩家"
		strategy.Pros = []string{"几乎无风险", "适合学习", "心态平和"}
		strategy.Cons = []string{"收益很低", "可能完全错过机会"}
	}

	return strategy
}

func (s *EndgameSolver) generateConservativeStrategy(optimalPath []*SolverNode) []*SolverNode {
	// 简化：提前止盈版本（在达到目标80%时就离场）
	conservativePath := []*SolverNode{}
	for _, node := range optimalPath {
		conservativePath = append(conservativePath, node)
		// 如果收益达到目标的80%，提前卖出
		if node.ExpectedValue >= s.Scenario.TargetProfit*0.8 && len(conservativePath) > 2 {
			break
		}
	}
	return conservativePath
}

func (s *EndgameSolver) generateSafeStrategy(optimalPath []*SolverNode) []*SolverNode {
	// 简化：只取前半段路径（早进早出）
	safePath := []*SolverNode{}
	maxLen := len(optimalPath) / 2
	if maxLen < 2 {
		maxLen = 2
	}
	for i := 0; i < maxLen && i < len(optimalPath); i++ {
		safePath = append(safePath, optimalPath[i])
	}
	return safePath
}

// ===== 3. 关键决策解释 =====

// generateKeyDecisions 生成关键决策的详细解释
func (s *EndgameSolver) generateKeyDecisions(path []*SolverNode) []DecisionExplanation {
	explanations := []DecisionExplanation{}

	for i, node := range path {
		if node.Importance == "🔴" || node.Importance == "🟡" {
			exp := DecisionExplanation{
				Turn:   i + 1,
				Action: node.Action.Description,
			}

			// 概率分解
			exp.ProbabilityBreakdown = s.calculateProbabilityBreakdown(node.State)

			// 期望值计算
			exp.ExpectedValueCalc = s.explainExpectedValue(node)

			// 为什么最优
			exp.WhyOptimal = s.explainWhyOptimal(node)

			// 关键指标
			exp.KeyIndicators = s.extractKeyIndicators(node.State)

			// 备选操作
			exp.AlternativeActions = s.getAlternativeActions(node)

			explanations = append(explanations, exp)

			// 最多显示3个关键决策
			if len(explanations) >= 3 {
				break
			}
		}
	}

	return explanations
}

func (s *EndgameSolver) calculateProbabilityBreakdown(state *SolverState) map[string]float64 {
	breakdown := make(map[string]float64)

	// 基于市场状态估算概率
	upProb := 0.33
	flatProb := 0.34
	downProb := 0.33

	// 调整概率
	if state.CrashWarning >= 4 {
		downProb = 0.60
		flatProb = 0.25
		upProb = 0.15
	} else if state.CrashWarning >= 3 {
		downProb = 0.45
		flatProb = 0.30
		upProb = 0.25
	}

	if state.EscapedAIRatio > 0.5 {
		downProb += 0.10
		upProb -= 0.10
	}

	// 归一化
	total := upProb + flatProb + downProb
	breakdown["上涨"] = upProb / total
	breakdown["持平"] = flatProb / total
	breakdown["下跌"] = downProb / total

	return breakdown
}

func (s *EndgameSolver) explainExpectedValue(node *SolverNode) string {
	if node.Action == nil {
		return ""
	}

	probs := s.calculateProbabilityBreakdown(node.State)

	// 简化的期望值说明
	return fmt.Sprintf("基于当前市场状态:\n"+
		"  上涨概率%.0f%% → 预期收益+%.1f%%\n"+
		"  持平概率%.0f%% → 预期收益 0%%\n"+
		"  下跌概率%.0f%% → 预期损失-%.1f%%\n"+
		"  综合期望值: %.1f%%",
		probs["上涨"]*100, node.ExpectedValue*150,
		probs["持平"]*100,
		probs["下跌"]*100, node.ExpectedValue*50,
		node.ExpectedValue*100)
}

func (s *EndgameSolver) explainWhyOptimal(node *SolverNode) string {
	if node.Action == nil {
		return ""
	}

	state := node.State
	reasons := []string{}

	// 根据市场状态给出解释
	if node.Action.Type == "Buy" {
		if state.EscapedAIRatio < 0.3 {
			reasons = append(reasons, "AI逃离率低(<30%)，市场信心较强")
		}
		if state.CrashWarning < 2 {
			reasons = append(reasons, "崩盘风险低，安全环境")
		}
		reasons = append(reasons, "预期价格上涨概率大，买入时机好")
	} else if node.Action.Type == "Sell" {
		if state.CrashWarning >= 3 {
			reasons = append(reasons, fmt.Sprintf("崩盘警告%d级，及时止盈避险", state.CrashWarning))
		}
		if state.EscapedAIRatio > 0.4 {
			reasons = append(reasons, fmt.Sprintf("AI逃离率%.0f%%，主力资金撤离信号明显", state.EscapedAIRatio*100))
		}
		if node.ExpectedValue >= s.Scenario.TargetProfit*0.8 {
			reasons = append(reasons, "已接近目标收益，落袋为安")
		}
	} else { // Hold
		reasons = append(reasons, "当前时机不明朗，观望等待更好机会")
	}

	if len(reasons) == 0 {
		return "基于期望值计算的最优选择"
	}

	result := ""
	for i, reason := range reasons {
		if i > 0 {
			result += "; "
		}
		result += reason
	}
	return result
}

func (s *EndgameSolver) extractKeyIndicators(state *SolverState) map[string]interface{} {
	indicators := make(map[string]interface{})

	indicators["当前价格"] = fmt.Sprintf("$%.2f", state.Price)
	indicators["AI逃离率"] = fmt.Sprintf("%.0f%%", state.EscapedAIRatio*100)
	indicators["崩盘警告"] = state.CrashWarning
	indicators["负债水平"] = fmt.Sprintf("$%.0f", state.MarginDebt)

	return indicators
}

func (s *EndgameSolver) getAlternativeActions(node *SolverNode) []string {
	alternatives := []string{}

	if node.Action == nil {
		return alternatives
	}

	switch node.Action.Type {
	case "Buy":
		alternatives = append(alternatives, "持有观望", "卖出部分仓位")
	case "Sell":
		alternatives = append(alternatives, "继续持有", "只卖出一半")
	case "Hold":
		alternatives = append(alternatives, "买入建仓", "卖出止盈")
	}

	return alternatives
}

// ===== 4. 常见错误库 =====

// generateCommonMistakes 根据场景生成常见错误
func (s *EndgameSolver) generateCommonMistakes(scenario *EndgameScenario) []CommonMistake {
	mistakes := []CommonMistake{}

	// 根据场景特征添加对应的常见错误

	// 如果是下跌场景
	if scenario.PriceChange < -0.10 {
		mistakes = append(mistakes, CommonMistake{
			MistakeType:   "❌ 抄底太早",
			Description:   "看到价格下跌-10%就全仓买入，期待反弹",
			Impact:        "价格继续下跌至-25%，被迫止损或爆仓",
			CorrectAction: "等待反转信号：AI出逃放缓、恐慌指数回落、出现小阳线",
			CaseStudy:     "错误操作收益-18% vs 最优操作+12%，差距30个百分点",
		})
	}

	// 如果崩盘风险高
	if scenario.CrashWarning >= 3 {
		mistakes = append(mistakes, CommonMistake{
			MistakeType:   "❌ 贪顶不止盈",
			Description:   "已经盈利+30%，但看到还在涨，继续持有想赚更多",
			Impact:        "遇到崩盘，最终只赚+5%甚至亏损",
			CorrectAction: "分批止盈：+20%卖30%，+30%卖50%，剩余设止损",
			CaseStudy:     "贪顶操作+5% vs 分批止盈+26%",
		})

		mistakes = append(mistakes, CommonMistake{
			MistakeType:   "❌ 忽视警告信号",
			Description:   fmt.Sprintf("崩盘警告已达%d级，但认为\"还能涨\"", scenario.CrashWarning),
			Impact:        "突然崩盘，损失50-70%",
			CorrectAction: "警告≥3级时立即清仓或至少减仓50%",
			CaseStudy:     "忽视警告-60% vs 及时离场+20%",
		})
	}

	// 如果AI逃离率高
	if scenario.AIExitRatio > 0.4 {
		mistakes = append(mistakes, CommonMistake{
			MistakeType:   "❌ 追高接盘",
			Description:   fmt.Sprintf("看到价格暴涨，FOMO冲进去（此时AI已逃离%.0f%%）", scenario.AIExitRatio*100),
			Impact:        "买在最高点，次日暴跌-20%",
			CorrectAction: "永远不追高！等待回调或放弃这次机会",
			CaseStudy:     "追高操作-22% vs 等待回调+8%",
		})
	}

	// 通用错误
	mistakes = append(mistakes, CommonMistake{
		MistakeType:   "❌ 满仓梭哈",
		Description:   "看到机会就全仓买入，不留余地",
		Impact:        "遇到意外情况无法应对，被动承受损失",
		CorrectAction: "分批建仓：首次50%，确认趋势后加仓，保留20%应急",
		CaseStudy:     "满仓操作风险爆表 vs 分批建仓进退自如",
	})

	mistakes = append(mistakes, CommonMistake{
		MistakeType:   "❌ 逆势加仓（补仓摊平成本）",
		Description:   "亏损-10%后，想\"补仓降成本\"",
		Impact:        "越补越亏，最终-30%以上",
		CorrectAction: "趋势错了就认错止损！保留本金比摊平更重要",
		CaseStudy:     "补仓摊平-35% vs 及时止损-10%",
	})

	// 限制最多5个
	if len(mistakes) > 5 {
		mistakes = mistakes[:5]
	}

	return mistakes
}

// ===== 5. 策略模式识别 =====

// identifyStrategyPattern 识别策略模式
func (s *EndgameSolver) identifyStrategyPattern(scenario *EndgameScenario, path []*SolverNode) string {
	// 分析场景和策略特征

	// 抄底模式
	if scenario.PriceChange < -0.15 && scenario.ConsecutiveFall >= 3 {
		return "抄底反弹"
	}

	// 逃顶模式
	if scenario.ConsecutiveRise >= 4 && scenario.CrashWarning >= 3 {
		return "高位逃顶"
	}

	// 震荡模式
	if math.Abs(scenario.PriceChange) < 0.10 && scenario.CrashWarning < 2 {
		return "震荡网格"
	}

	// 快进快出
	if len(path) <= 3 {
		return "快进快出"
	}

	// 长期持有
	buyCount := 0
	for _, node := range path {
		if node.Action != nil && node.Action.Type == "Buy" {
			buyCount++
		}
	}
	if buyCount > len(path)/2 {
		return "长期持有"
	}

	return "综合策略"
}

// ===== 6. 教学要点生成 =====

// generateTeachingPoints 生成教学要点
func (s *EndgameSolver) generateTeachingPoints(pattern string, scenario *EndgameScenario) []string {
	points := []string{}

	switch pattern {
	case "抄底反弹":
		points = append(points, "【抄底三要素】连续下跌≥3天 + AI出逃放缓 + 恐慌指数回落")
		points = append(points, "【分批建仓】不要一次性买入，先试探30-50%，确认反弹后加仓")
		points = append(points, "【及时止盈】反弹达到目标后立即离场，避免二次下跌")
		points = append(points, "【止损纪律】如果抄底失败，价格继续跌破-8%，果断止损")

	case "高位逃顶":
		points = append(points, "【逃顶信号】崩盘警告≥3级 + AI逃离>40% + 散户FOMO")
		points = append(points, "【分批离场】不要等最高点，分批卖出：+20%卖30%，+30%卖50%")
		points = append(points, "【宁少赚不亏】在高风险环境下，保护利润比博弈更高点重要")
		points = append(points, "【警惕逆向思维陷阱】\"别人恐慌我贪婪\"不适用于崩盘前夕")

	case "震荡网格":
		points = append(points, "【震荡识别】价格在±10%范围波动 + 无明显趋势")
		points = append(points, "【网格操作】低买高卖：价格<均价-5%买入，>均价+5%卖出")
		points = append(points, "【保持底仓】维持50%基础仓位，避免踏空")
		points = append(points, "【突破应对】如果突破震荡区间，立即改变策略")

	case "快进快出":
		points = append(points, "【时机把握】抓住确定性高的短期机会")
		points = append(points, "【快速决断】进场要快，离场更要快，不拖泥带水")
		points = append(points, "【风险控制】单次收益目标适中（5-10%），积少成多")

	default:
		points = append(points, "【观察先行】进场前充分观察市场信号")
		points = append(points, "【仓位管理】永远不要满仓，保留应急资金")
		points = append(points, "【止盈止损】提前设定目标和止损线，严格执行")
		points = append(points, "【情绪控制】不要FOMO，不要贪婪，不要恐慌")
	}

	// 通用要点
	points = append(points, "【核心原则】顺势而为，不要逆势硬抗")

	return points
}

// ===== 7. 增强的显示系统 =====

// DisplayEnhancedSolverResult 展示增强的求解结果（包含所有学习辅助功能）
func DisplayEnhancedSolverResult(result *SolverResult) {
	fmt.Printf("\n%s╔═══════════════════ 📊 求解结果 (学习增强版) ═══════════════════╗%s\n", Green, Reset)
	
	// 总体评估
	returnColor := Green
	if result.ExpectedReturn < 0 {
		returnColor = Red
	}
	
	// 策略模式标识
	patternIcon := "📈"
	switch result.StrategyPattern {
	case "抄底反弹":
		patternIcon = "📉→📈"
	case "高位逃顶":
		patternIcon = "🔥→💰"
	case "震荡网格":
		patternIcon = "↕️"
	case "快进快出":
		patternIcon = "⚡"
	}
	
	fmt.Printf("%s║%s  【策略模式】%s %s%s%s\n", Green, Reset, patternIcon, Yellow, result.StrategyPattern, Reset)
	fmt.Printf("%s║%s  【最优策略】%s%s%s\n", Green, Reset, Yellow, result.Strategy, Reset)
	fmt.Printf("%s║%s  期望收益率: %s%.2f%%%s\n", Green, Reset, returnColor, result.ExpectedReturn*100, Reset)
	fmt.Printf("%s║%s  成功概率:   %s%.1f%%%s\n", Green, Reset, Cyan, result.SuccessProbability*100, Reset)
	fmt.Printf("%s║%s  风险评估:   最坏%s%.1f%%%s  |  最好%s%.1f%%%s\n",
		Green, Reset, Red, result.WorstCase*100, Reset, Green, result.BestCase*100, Reset)
	fmt.Printf("%s╚═══════════════════════════════════════════════════════════════╝%s\n", Green, Reset)
	
	// ===== 1. 关键决策点（带重要性标注） =====
	if len(result.OptimalPath) > 0 {
		fmt.Printf("\n%s╔═══════════════ 🎯 最优操作路径 (重要性标注) ═══════════════╗%s\n", Cyan, Reset)
		
		displayCount := min_int_solver(8, len(result.OptimalPath))
		for i := 0; i < displayCount; i++ {
			node := result.OptimalPath[i]
			if node.Action == nil {
				continue
			}
			
			importance := node.Importance
			if importance == "" {
				importance = "⚪"
			}
			
			fmt.Printf("%s║%s  %s 回合%d: %s%s%s\n",
				Cyan, Reset, importance, i+1, Yellow, node.Action.Description, Reset)
			
			if node.ImportanceReason != "" && importance != "⚪" {
				fmt.Printf("%s║%s     └─ %s\n", Cyan, Reset, node.ImportanceReason)
			}
		}
		
		if len(result.OptimalPath) > 8 {
			fmt.Printf("%s║%s  ... (还有%d步操作)\n", Cyan, Reset, len(result.OptimalPath)-8)
		}
		
		fmt.Printf("%s╚═══════════════════════════════════════════════════════════════╝%s\n", Cyan, Reset)
	}
	
	// ===== 2. 备选策略对比 =====
	if len(result.AlternativePaths) > 0 {
		fmt.Printf("\n%s╔═══════════════ 📊 多策略对比 (Top %d) ═══════════════════╗%s\n",
			Purple, len(result.AlternativePaths), Reset)
		
		for i, alt := range result.AlternativePaths {
			icon := "🥇"
			if i == 1 {
				icon = "🥈"
			} else if i == 2 {
				icon = "🥉"
			}
			
			returnColor := Green
			if alt.ExpectedReturn < 0 {
				returnColor = Red
			}
			
			fmt.Printf("%s║%s\n", Purple, Reset)
			fmt.Printf("%s║%s  %s【策略%s: %s】\n", Purple, Reset, icon, string(rune('A'+i)), alt.Name)
			fmt.Printf("%s║%s     期望收益: %s%.2f%%%s  |  成功率: %.0f%%  |  风险: %s\n",
				Purple, Reset, returnColor, alt.ExpectedReturn*100, Reset, alt.SuccessRate*100, alt.RiskLevel)
			fmt.Printf("%s║%s     适合: %s\n", Purple, Reset, alt.SuitableFor)
			
			if len(alt.Pros) > 0 {
				fmt.Printf("%s║%s     ✓ 优点: %s\n", Purple, Reset, alt.Pros[0])
			}
			if len(alt.Cons) > 0 {
				fmt.Printf("%s║%s     ✗ 缺点: %s\n", Purple, Reset, alt.Cons[0])
			}
		}
		
		fmt.Printf("%s╚═══════════════════════════════════════════════════════════════╝%s\n", Purple, Reset)
	}
	
	// ===== 3. 关键决策解释（概率分析） =====
	if len(result.KeyDecisions) > 0 {
		fmt.Printf("\n%s╔═══════════════ 🧠 关键决策详解 (概率推演) ═══════════════╗%s\n", Yellow, Reset)
		
		for i, decision := range result.KeyDecisions {
			if i >= 3 { // 最多显示3个
				break
			}
			
			fmt.Printf("%s║%s\n", Yellow, Reset)
			fmt.Printf("%s║%s  【回合%d】%s\n", Yellow, Reset, decision.Turn, decision.Action)
			
			// 概率分解
			if len(decision.ProbabilityBreakdown) > 0 {
				fmt.Printf("%s║%s  概率分析:\n", Yellow, Reset)
				for scenario, prob := range decision.ProbabilityBreakdown {
					fmt.Printf("%s║%s    • %s: %.0f%%\n", Yellow, Reset, scenario, prob*100)
				}
			}
			
			// 为什么最优
			if decision.WhyOptimal != "" {
				fmt.Printf("%s║%s  💡 为什么最优:\n", Yellow, Reset)
				fmt.Printf("%s║%s     %s\n", Yellow, Reset, decision.WhyOptimal)
			}
			
			// 备选操作
			if len(decision.AlternativeActions) > 0 {
				fmt.Printf("%s║%s  备选: %s\n", Yellow, Reset, decision.AlternativeActions[0])
			}
		}
		
		fmt.Printf("%s╚═══════════════════════════════════════════════════════════════╝%s\n", Yellow, Reset)
	}
	
	// ===== 4. 常见错误警示 =====
	if len(result.CommonMistakes) > 0 {
		fmt.Printf("\n%s╔═══════════════ ⚠️  常见错误避坑指南 ═══════════════════╗%s\n", Red, Reset)
		
		displayCount := min_int_solver(3, len(result.CommonMistakes))
		for i := 0; i < displayCount; i++ {
			mistake := result.CommonMistakes[i]
			
			fmt.Printf("%s║%s\n", Red, Reset)
			fmt.Printf("%s║%s  %s\n", Red, Reset, mistake.MistakeType)
			fmt.Printf("%s║%s  场景: %s\n", Red, Reset, mistake.Description)
			fmt.Printf("%s║%s  后果: %s\n", Red, Reset, mistake.Impact)
			fmt.Printf("%s║%s  ✓ 正确做法: %s\n", Red, Reset, mistake.CorrectAction)
		}
		
		if len(result.CommonMistakes) > 3 {
			fmt.Printf("%s║%s  ... (还有%d个常见错误)\n", Red, Reset, len(result.CommonMistakes)-3)
		}
		
		fmt.Printf("%s╚═══════════════════════════════════════════════════════════════╝%s\n", Red, Reset)
	}
	
	// ===== 5. 教学要点 =====
	if len(result.TeachingPoints) > 0 {
		fmt.Printf("\n%s╔═══════════════ 🎓 核心教学要点 ═══════════════════════╗%s\n", Cyan, Reset)
		
		for _, point := range result.TeachingPoints {
			fmt.Printf("%s║%s  • %s\n", Cyan, Reset, point)
		}
		
		fmt.Printf("%s╚═══════════════════════════════════════════════════════════════╝%s\n", Cyan, Reset)
	}
	
	fmt.Println()
}

// min_int_solver 辅助函数（避免与main.go中的min_int冲突）
func min_int_solver(a, b int) int {
	if a < b {
		return a
	}
	return b
}
