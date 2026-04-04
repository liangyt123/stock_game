package main

import (
	"bufio"
	"fmt"
	"os"
)

// ============================================================
// 渐进式提示系统 - 三级提示引导思考
// ============================================================

// HintLevel 提示等级
type HintLevel int

const (
	DirectionHint HintLevel = 1 // 方向提示
	QuantityHint  HintLevel = 2 // 数量提示
	ReasonHint    HintLevel = 3 // 原因分析
)

// ProgressiveHintSystem 渐进式提示系统
type ProgressiveHintSystem struct {
	Scenario      *EndgameScenario
	CurrentTurn   int
	OptimalAction *SolverAction
	State         *SolverState
	Reader        *bufio.Reader
}

// NewProgressiveHintSystem 创建提示系统
func NewProgressiveHintSystem(scenario *EndgameScenario, turn int, optimalAction *SolverAction, state *SolverState) *ProgressiveHintSystem {
	return &ProgressiveHintSystem{
		Scenario:      scenario,
		CurrentTurn:   turn,
		OptimalAction: optimalAction,
		State:         state,
		Reader:        bufio.NewReader(os.Stdin),
	}
}

// RunProgressiveHints 运行渐进式提示
func (h *ProgressiveHintSystem) RunProgressiveHints() {
	fmt.Printf("\n%s╔═══════════════ 🎓 渐进式提示模式 ═══════════════╗%s\n", Purple, Reset)
	fmt.Printf("%s║%s  场景: %s%s%s\n", Purple, Reset, Cyan, h.Scenario.Name, Reset)
	fmt.Printf("%s║%s  当前: 第%d回合\n", Purple, Reset, h.CurrentTurn)
	fmt.Printf("%s║%s  目标: 通过逐步思考，找到最优操作\n", Purple, Reset)
	fmt.Printf("%s╚═══════════════════════════════════════════════════╝%s\n", Purple, Reset)

	// 显示当前市场状态
	h.displayMarketState()

	// 第一级提示：方向
	if !h.askDirectionHint() {
		return // 用户放弃
	}

	fmt.Println()

	// 第二级提示：数量
	if !h.askQuantityHint() {
		return
	}

	fmt.Println()

	// 第三级提示：原因分析
	h.showReasonAnalysis()

	fmt.Printf("\n%s按回车键继续...%s ", Cyan, Reset)
	h.Reader.ReadString('\n')
}

// displayMarketState 显示当前市场状态
func (h *ProgressiveHintSystem) displayMarketState() {
	fmt.Printf("\n%s【当前市场状态】%s\n", Yellow, Reset)
	fmt.Printf("  价格: $%.2f\n", h.State.Price)
	fmt.Printf("  AI逃离率: %.0f%%\n", h.State.EscapedAIRatio*100)
	fmt.Printf("  崩盘警告: %d级\n", h.State.CrashWarning)
	fmt.Printf("  你的持仓: %d股\n", h.State.PlayerShares)
	fmt.Printf("  可用现金: $%.0f\n", h.State.PlayerCash)
	if h.State.MarginDebt > 0 {
		fmt.Printf("  负债: $%.0f\n", h.State.MarginDebt)
	}
}

// askDirectionHint 第一级：方向提示
func (h *ProgressiveHintSystem) askDirectionHint() bool {
	fmt.Printf("\n%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", Cyan, Reset)
	fmt.Printf("%s【第一级提示 - 方向性】%s\n", Yellow, Reset)

	// 生成方向提示
	hint := h.generateDirectionHint()
	fmt.Printf("\n%s💡 提示: %s%s\n\n", Yellow, hint, Reset)

	// 使用交互式菜单
	items := []MenuItem{
		{Label: "买入（加仓/建仓）", Value: "1", Icon: "📈"},
		{Label: "卖出（减仓/清仓）", Value: "2", Icon: "📉"},
		{Label: "持有（观望）", Value: "3", Icon: "⏸️"},
		{Label: "放弃提示，直接查看答案", Value: "0", Icon: "💡"},
	}

	menu := NewInteractiveMenu("❓ 当前应该采取什么操作？", items)
	menu.ShowNumbers = true
	value, _ := menu.Show()

	if value == "0" || value == "" {
		fmt.Printf("\n%s直接查看答案...%s\n", Yellow, Reset)
		h.showDirectAnswer()
		return false
	}

	// 检查答案
	correctDirection := h.getCorrectDirection()
	userDirection := 0
	fmt.Sscanf(value, "%d", &userDirection)

	if userDirection == correctDirection {
		fmt.Printf("\n%s✓ 方向正确！%s\n", Green, Reset)
		return true
	} else {
		fmt.Printf("\n%s✗ 方向不太对...%s\n", Red, Reset)
		fmt.Printf("%s正确答案是: %s%s\n", Yellow, h.getDirectionName(correctDirection), Reset)
		fmt.Printf("%s让我们看看为什么...%s\n", Cyan, Reset)
		return true
	}
}

// askQuantityHint 第二级：数量提示
func (h *ProgressiveHintSystem) askQuantityHint() bool {
	correctDirection := h.getCorrectDirection()

	if correctDirection == 3 { // 持有
		fmt.Printf("%s持有操作无需选择数量%s\n", Cyan, Reset)
		return true
	}

	fmt.Printf("\n%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", Cyan, Reset)
	fmt.Printf("%s【第二级提示 - 数量级】%s\n", Yellow, Reset)

	hint := h.generateQuantityHint()
	fmt.Printf("\n%s💡 提示: %s%s\n\n", Yellow, hint, Reset)

	var items []MenuItem

	if correctDirection == 1 { // 买入
		// 根据可用资金计算选项
		maxShares := int(h.State.PlayerCash / h.State.Price)
		option1 := maxShares / 4
		option2 := maxShares / 2
		option3 := maxShares * 3 / 4

		if option1 < 100 {
			option1 = 100
		}
		if option2 < 500 {
			option2 = 500
		}
		if option3 < 1000 {
			option3 = 1000
		}

		items = []MenuItem{
			{Label: fmt.Sprintf("%d股", option1), Description: "试探性，约25%资金", Value: "1"},
			{Label: fmt.Sprintf("%d股", option2), Description: "半仓，约50%资金", Value: "2"},
			{Label: fmt.Sprintf("%d股", option3), Description: "重仓，约75%资金", Value: "3"},
		}

	} else { // 卖出
		currentShares := h.State.PlayerShares
		option1 := currentShares / 3
		option2 := currentShares / 2
		option3 := currentShares

		items = []MenuItem{
			{Label: fmt.Sprintf("%d股", option1), Description: "部分止盈，约33%", Value: "1"},
			{Label: fmt.Sprintf("%d股", option2), Description: "半仓离场，50%", Value: "2"},
			{Label: fmt.Sprintf("%d股", option3), Description: "全部清仓，100%", Value: "3"},
		}
	}

	var menuTitle string
	if correctDirection == 1 {
		menuTitle = "❓ 应该买入多少股？"
	} else {
		menuTitle = "❓ 应该卖出多少股？"
	}

	menu := NewInteractiveMenu(menuTitle, items)
	menu.ShowNumbers = true
	value, _ := menu.Show()

	userChoice := 0
	fmt.Sscanf(value, "%d", &userChoice)

	correctChoice := h.getCorrectQuantityChoice()

	if userChoice == correctChoice {
		fmt.Printf("\n%s✓ 数量选择正确！%s\n", Green, Reset)
	} else {
		fmt.Printf("\n%s✗ 数量不是最优...%s\n", Yellow, Reset)
		fmt.Printf("%s更好的选择是: 选项%d%s\n", Yellow, correctChoice, Reset)
	}

	return true
}

// showReasonAnalysis 第三级：原因分析
func (h *ProgressiveHintSystem) showReasonAnalysis() {
	fmt.Printf("\n%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", Cyan, Reset)
	fmt.Printf("%s【第三级提示 - 深度分析】%s\n", Yellow, Reset)

	fmt.Printf("\n%s✓ 最优操作: %s%s%s\n", Green, Cyan, h.OptimalAction.Description, Reset)

	// 原因分析
	fmt.Printf("\n%s📊 为什么这么做？%s\n\n", Yellow, Reset)

	reasons := h.generateReasons()
	for i, reason := range reasons {
		fmt.Printf("  %d. %s\n", i+1, reason)
	}

	// 关键指标
	fmt.Printf("\n%s📈 关键指标分析:%s\n", Yellow, Reset)
	indicators := h.analyzeKeyIndicators()
	for _, indicator := range indicators {
		fmt.Printf("  • %s\n", indicator)
	}

	// 风险评估
	fmt.Printf("\n%s⚠️  风险评估:%s\n", Yellow, Reset)
	risks := h.assessRisks()
	for _, risk := range risks {
		fmt.Printf("  • %s\n", risk)
	}

	// 预期结果
	fmt.Printf("\n%s🎯 预期结果:%s\n", Yellow, Reset)
	fmt.Printf("  如果市场按预期发展，这个操作可以带来最优收益\n")
	fmt.Printf("  但要注意设置止损，防止意外情况\n")
}

// 辅助方法

func (h *ProgressiveHintSystem) getCorrectDirection() int {
	switch h.OptimalAction.Type {
	case "Buy":
		return 1
	case "Sell":
		return 2
	case "Hold":
		return 3
	default:
		return 3
	}
}

func (h *ProgressiveHintSystem) getDirectionName(direction int) string {
	switch direction {
	case 1:
		return "买入"
	case 2:
		return "卖出"
	case 3:
		return "持有"
	default:
		return "未知"
	}
}

func (h *ProgressiveHintSystem) generateDirectionHint() string {
	state := h.State

	// 根据市场状态生成提示
	if state.CrashWarning >= 3 {
		return "观察崩盘警告等级"
	} else if state.EscapedAIRatio > 0.5 {
		return "注意AI逃离的速度和比例"
	} else if state.EscapedAIRatio < 0.3 && state.CrashWarning < 2 {
		return "市场情绪如何？是否有建仓机会？"
	} else {
		return "综合考虑AI行为和风险等级"
	}
}

func (h *ProgressiveHintSystem) generateQuantityHint() string {
	state := h.State

	if h.OptimalAction.Type == "Buy" {
		if state.CrashWarning >= 2 {
			return "风险较高，考虑保守一点"
		} else {
			return "考虑风险承受能力和反转信号的强度"
		}
	} else {
		if state.CrashWarning >= 4 {
			return "崩盘风险极高时应该如何？"
		} else {
			return "已经获得不错收益时，如何保护利润？"
		}
	}
}

func (h *ProgressiveHintSystem) getCorrectQuantityChoice() int {
	// 简化：根据操作类型和风险等级决定
	state := h.State

	if h.OptimalAction.Type == "Buy" {
		if state.CrashWarning < 2 && state.EscapedAIRatio < 0.3 {
			return 2 // 半仓
		} else {
			return 1 // 试探性
		}
	} else { // Sell
		if state.CrashWarning >= 4 {
			return 3 // 全部清仓
		} else if state.CrashWarning >= 3 {
			return 2 // 半仓
		} else {
			return 1 // 部分止盈
		}
	}
}

func (h *ProgressiveHintSystem) generateReasons() []string {
	reasons := []string{}
	state := h.State

	switch h.OptimalAction.Type {
	case "Buy":
		if state.EscapedAIRatio < 0.3 {
			reasons = append(reasons, "AI逃离率低(<30%)，说明市场信心较强")
		}
		if state.CrashWarning < 2 {
			reasons = append(reasons, "崩盘风险低，当前是相对安全的环境")
		}
		reasons = append(reasons, "基于概率分析，价格上涨的可能性大于下跌")

	case "Sell":
		if state.CrashWarning >= 3 {
			reasons = append(reasons, fmt.Sprintf("崩盘警告已达%d级，风险急剧上升", state.CrashWarning))
		}
		if state.EscapedAIRatio > 0.4 {
			reasons = append(reasons, fmt.Sprintf("AI逃离率%.0f%%，主力资金正在撤离", state.EscapedAIRatio*100))
		}
		reasons = append(reasons, "及时止盈/止损，保护已有收益")

	case "Hold":
		reasons = append(reasons, "当前时机不明朗，等待更好的入场/出场机会")
		reasons = append(reasons, "避免频繁交易，降低手续费成本")
	}

	return reasons
}

func (h *ProgressiveHintSystem) analyzeKeyIndicators() []string {
	indicators := []string{}
	state := h.State

	indicators = append(indicators, fmt.Sprintf("当前价格: $%.2f", state.Price))
	indicators = append(indicators, fmt.Sprintf("AI逃离率: %.0f%% %s",
		state.EscapedAIRatio*100, h.getAIEscapeComment(state.EscapedAIRatio)))
	indicators = append(indicators, fmt.Sprintf("崩盘警告: %d级 %s",
		state.CrashWarning, h.getCrashWarningComment(state.CrashWarning)))

	if state.MarginDebt > 0 {
		indicators = append(indicators, fmt.Sprintf("⚠️  负债: $%.0f (注意杠杆风险)", state.MarginDebt))
	}

	return indicators
}

func (h *ProgressiveHintSystem) getAIEscapeComment(ratio float64) string {
	if ratio < 0.2 {
		return "(低，市场信心强)"
	} else if ratio < 0.4 {
		return "(中等)"
	} else if ratio < 0.6 {
		return "(高，警惕)"
	} else {
		return "(极高，危险)"
	}
}

func (h *ProgressiveHintSystem) getCrashWarningComment(level int) string {
	if level < 2 {
		return "(安全)"
	} else if level < 3 {
		return "(注意)"
	} else if level < 4 {
		return "(警告)"
	} else {
		return "(危险！)"
	}
}

func (h *ProgressiveHintSystem) assessRisks() []string {
	risks := []string{}
	state := h.State

	if h.OptimalAction.Type == "Buy" {
		if state.CrashWarning >= 2 {
			risks = append(risks, "⚠️  存在一定崩盘风险，建议设置止损")
		}
		if state.MarginDebt > 0 {
			risks = append(risks, "⚠️  已有负债，避免过度加杠杆")
		}
		if state.EscapedAIRatio > 0.3 {
			risks = append(risks, "⚠️  部分AI正在撤离，需谨慎")
		}

		if len(risks) == 0 {
			risks = append(risks, "✓ 风险可控，但仍需设置止损线")
		}
	} else if h.OptimalAction.Type == "Sell" {
		risks = append(risks, "✓ 离场操作本身风险低")
		risks = append(risks, "⚠️  但要注意不要过早离场错失后续收益")
	} else {
		risks = append(risks, "✓ 观望风险低，但可能错过机会")
	}

	return risks
}

func (h *ProgressiveHintSystem) showDirectAnswer() {
	fmt.Printf("\n%s【直接答案】%s\n", Yellow, Reset)
	fmt.Printf("  最优操作: %s%s%s\n", Cyan, h.OptimalAction.Description, Reset)

	reasons := h.generateReasons()
	fmt.Printf("\n  原因:\n")
	for _, reason := range reasons {
		fmt.Printf("    • %s\n", reason)
	}
}

// RunInteractiveSolver 交互式求解器（整合渐进式提示）
func RunInteractiveSolver(scenarioIndex int) {
	if scenarioIndex < 0 || scenarioIndex >= len(EndgameScenarios) {
		fmt.Printf("%s错误: 场景索引无效%s\n", Red, Reset)
		return
	}

	scenario := &EndgameScenarios[scenarioIndex]

	fmt.Printf("\n%s════════════════════════════════════════════════════════════%s\n", Purple, Reset)
	fmt.Printf("%s          🎓 交互式学习模式%s\n", Purple, Reset)
	fmt.Printf("%s════════════════════════════════════════════════════════════%s\n", Purple, Reset)

	fmt.Printf("\n%s场景: %s%s (%s)%s\n", Yellow, Cyan, scenario.Name, scenario.Difficulty, Reset)
	fmt.Printf("%s目标: 盈利 %.1f%%  |  时限: %d回合%s\n",
		Yellow, scenario.TargetProfit*100, scenario.TimeLimit, Reset)

	// 先求解获得最优路径
	fmt.Printf("\n%s正在分析场景...%s ", Cyan, Reset)
	solver := NewEndgameSolver(scenario)
	result := solver.Solve()
	fmt.Printf("%s完成！%s\n", Green, Reset)

	// 使用交互式菜单选择模式
	modeItems := []MenuItem{
		{Label: "渐进式提示（引导思考）", Description: "通过三级提示引导你分析盘面，适合学习", Value: "1", Icon: "🎓"},
		{Label: "直接查看完整答案", Description: "立即显示 AI 计算出的最优操作序列", Value: "2", Icon: "📊"},
	}

	modeMenu := NewInteractiveMenu("请选择求解模式:", modeItems)
	input, _ := modeMenu.Show()

	if input == "" {
		input = "2" // 默认显示答案
	}

	if input == "1" {
		// 渐进式提示模式 - 选择一个关键回合进行提示
		if len(result.OptimalPath) > 1 {
			// 选择第一个有操作的回合
			for i, node := range result.OptimalPath {
				if node.Action != nil && node.Action.Type != "Hold" {
					hints := NewProgressiveHintSystem(scenario, i+1, node.Action, node.State)
					hints.RunProgressiveHints()
					break
				}
			}
		}

		// 最后显示完整答案
		fmt.Printf("\n%s查看完整最优解...%s\n", Cyan, Reset)
		DisplayEnhancedSolverResult(result)

	} else {
		// 直接显示完整答案
		DisplayEnhancedSolverResult(result)
	}

	fmt.Println()
}
