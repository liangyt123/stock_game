package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ============================================================
// 实战复盘分析工具 - 对比玩家操作 vs 最优解
// ============================================================

// PlayerAction 玩家操作记录
type PlayerAction struct {
	Turn        int
	ActionType  string // "Buy", "Sell", "Hold"
	Shares      int
	Price       float64
	Description string
}

// ReplayAnalyzer 复盘分析器
type ReplayAnalyzer struct {
	Scenario       *EndgameScenario
	PlayerActions  []PlayerAction
	OptimalActions []*SolverNode
	Reader         *bufio.Reader
}

// RunReplayAnalyzer 运行复盘分析工具
func RunReplayAnalyzer() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\n%s╔═══════════════════════════════════════════════════════════╗%s\n", Purple, Reset)
	fmt.Printf("%s║               📝 实战复盘分析工具                          ║%s\n", Purple, Reset)
	fmt.Printf("%s╚═══════════════════════════════════════════════════════════╝%s\n", Purple, Reset)
	fmt.Println()

	fmt.Printf("%s这个工具帮你对比自己的操作和AI最优解，找出差距%s\n", Cyan, Reset)
	fmt.Println()

	// 使用交互式菜单选择场景
	items := make([]MenuItem, len(EndgameScenarios)+1)
	for i, scenario := range EndgameScenarios {
		diffIcon := "🟢"
		if scenario.Difficulty == "中等" {
			diffIcon = "🟡"
		} else if scenario.Difficulty == "困难" {
			diffIcon = "🔴"
		} else if scenario.Difficulty == "地狱" {
			diffIcon = "💀"
		}

		items[i] = MenuItem{
			Label:       scenario.Name,
			Description: fmt.Sprintf("%s %s | 目标: +%.0f%% | %d时段", diffIcon, scenario.Difficulty, scenario.TargetProfit*100, scenario.TimeLimit),
			Value:       fmt.Sprintf("%d", i+1),
		}
	}
	items[len(EndgameScenarios)] = MenuItem{Label: "放弃复盘", Value: "0", Icon: "↩️"}

	menu := NewInteractiveMenu("请选择你刚才挑战的场景：", items)
	menu.Reader = reader
	value, _ := menu.Show()

	if value == "" || value == "0" {
		return
	}

	sceneNum, _ := strconv.Atoi(value)
	scenario := &EndgameScenarios[sceneNum-1]

	fmt.Printf("\n%s已选择场景: %s%s%s\n", Yellow, Cyan, scenario.Name, Reset)

	// 询问最终结果
	fmt.Printf("\n%s你的最终收益率是多少？(输入数字，如 15 表示+15%%): %s", Green, Reset)
	profitInput, _ := reader.ReadString('\n')
	profitInput = strings.TrimSpace(profitInput)

	playerProfit := 0.0
	fmt.Sscanf(profitInput, "%f", &playerProfit)
	playerProfit = playerProfit / 100.0

	// 计算最优解
	fmt.Printf("\n%s正在计算最优解...%s ", Cyan, Reset)
	solver := NewEndgameSolver(scenario)
	result := solver.Solve()
	fmt.Printf("%s完成！%s\n", Green, Reset)

	// 对比分析
	fmt.Printf("\n%s╔═══════════════ 📊 实战复盘分析 ═══════════════════╗%s\n", Green, Reset)

	profitDiff := playerProfit - result.ExpectedReturn
	profitColor := Green
	if profitDiff < 0 {
		profitColor = Red
	}

	fmt.Printf("%s║%s  【收益对比】\n", Green, Reset)
	fmt.Printf("%s║%s  你的收益:   %s%.2f%%%s\n", Green, Reset, getColorForProfit(playerProfit), playerProfit*100, Reset)
	fmt.Printf("%s║%s  最优收益:   %s%.2f%%%s\n", Green, Reset, Green, result.ExpectedReturn*100, Reset)
	fmt.Printf("%s║%s  差距:       %s%.2f%%%s\n", Green, Reset, profitColor, profitDiff*100, Reset)
	fmt.Printf("%s║%s\n", Green, Reset)

	// 表现评级
	performance := evaluatePerformance(playerProfit, result.ExpectedReturn)
	fmt.Printf("%s║%s  【表现评级】%s%s%s\n", Green, Reset, Yellow, performance.Grade, Reset)
	fmt.Printf("%s║%s  %s\n", Green, Reset, performance.Comment)
	fmt.Printf("%s║%s\n", Green, Reset)

	// 可能的问题
	fmt.Printf("%s║%s  【可能的问题】\n", Green, Reset)
	problems := analyzeProbableProblems(playerProfit, result.ExpectedReturn, scenario)
	for _, problem := range problems {
		fmt.Printf("%s║%s  • %s\n", Green, Reset, problem)
	}

	fmt.Printf("%s╚═══════════════════════════════════════════════════════════╝%s\n", Green, Reset)

	// 显示最优策略
	fmt.Printf("\n%s查看最优策略以对比你的操作...%s\n", Cyan, Reset)
	DisplayEnhancedSolverResult(result)
}

// Performance 表现评级
type Performance struct {
	Grade   string
	Comment string
}

func getColorForProfit(profit float64) string {
	if profit >= 0.2 {
		return Green
	} else if profit >= 0 {
		return Yellow
	} else {
		return Red
	}
}

func evaluatePerformance(playerProfit, optimalProfit float64) Performance {
	ratio := playerProfit / optimalProfit
	diff := playerProfit - optimalProfit

	if ratio >= 0.95 { // 达到最优的95%以上
		return Performance{
			Grade:   "S级 - 近乎完美！",
			Comment: "你的操作非常接近理论最优解，继续保持！",
		}
	} else if ratio >= 0.80 { // 80-95%
		return Performance{
			Grade:   "A级 - 优秀",
			Comment: "很不错的表现，只有小幅优化空间",
		}
	} else if ratio >= 0.60 { // 60-80%
		return Performance{
			Grade:   "B级 - 良好",
			Comment: "整体方向正确，但时机把握还有提升空间",
		}
	} else if ratio >= 0.40 { // 40-60%
		return Performance{
			Grade:   "C级 - 及格",
			Comment: "基本策略可行，但存在明显失误",
		}
	} else if ratio >= 0 { // 0-40%
		return Performance{
			Grade:   "D级 - 需改进",
			Comment: "策略选择或执行出现较大问题",
		}
	} else if diff > -0.15 { // 小幅亏损
		return Performance{
			Grade:   "F级 - 亏损",
			Comment: "出现亏损，需要反思操作逻辑",
		}
	} else { // 大幅亏损
		return Performance{
			Grade:   "F- - 重大失误",
			Comment: "严重亏损，可能犯了致命错误",
		}
	}
}

func analyzeProbableProblems(playerProfit, optimalProfit float64, scenario *EndgameScenario) []string {
	problems := []string{}
	diff := optimalProfit - playerProfit

	// 根据差距推测问题
	if diff > 0.20 { // 差距超过20%
		problems = append(problems, "可能错过了关键买入/卖出时机")

		if scenario.CrashWarning >= 3 {
			problems = append(problems, "可能在高风险时期没有及时离场")
		}

		if scenario.ConsecutiveFall >= 3 {
			problems = append(problems, "可能在抄底反弹时过于保守")
		}
	} else if diff > 0.10 { // 差距10-20%
		problems = append(problems, "时机把握有偏差，可以更精准")

		if scenario.AIExitRatio > 0.4 {
			problems = append(problems, "可能没有充分重视AI逃离信号")
		}
	} else if diff > 0 && diff <= 0.10 { // 差距0-10%
		problems = append(problems, "整体不错，细节可以更完美")
	}

	// 亏损情况
	if playerProfit < -0.10 {
		problems = []string{} // 清空，用更严重的问题
		problems = append(problems, "❌ 可能犯了致命错误：")

		if scenario.CrashWarning >= 4 {
			problems = append(problems, "  → 崩盘前没有离场（贪顶）")
		} else if scenario.ConsecutiveFall >= 3 {
			problems = append(problems, "  → 抄底太早或逆势加仓")
		} else {
			problems = append(problems, "  → 方向判断错误或止损不及时")
		}
	}

	if len(problems) == 0 {
		problems = append(problems, "✓ 没有发现明显问题，表现优秀！")
	}

	return problems
}

// QuickReplayFromInput 快速复盘（简化版，不需要逐回合输入）
func QuickReplayFromInput(scenarioIndex int, playerProfit float64) {
	if scenarioIndex < 0 || scenarioIndex >= len(EndgameScenarios) {
		return
	}

	scenario := &EndgameScenarios[scenarioIndex]

	// 计算最优解
	solver := NewEndgameSolver(scenario)
	result := solver.Solve()

	// 显示对比
	fmt.Printf("\n%s╔═══════════════ 📊 快速复盘 ═══════════════════╗%s\n", Cyan, Reset)
	fmt.Printf("%s║%s  场景: %s\n", Cyan, Reset, scenario.Name)
	fmt.Printf("%s║%s  你的收益: %.2f%%  vs  最优: %.2f%%\n",
		Cyan, Reset, playerProfit*100, result.ExpectedReturn*100)

	diff := playerProfit - result.ExpectedReturn
	if diff >= -0.05 {
		fmt.Printf("%s║%s  ✓ 表现优秀！\n", Cyan, Reset)
	} else {
		fmt.Printf("%s║%s  可以做得更好，查看最优策略↓\n", Cyan, Reset)
	}

	fmt.Printf("%s╚═══════════════════════════════════════════════════════════╝%s\n", Cyan, Reset)

	// 显示最优策略
	DisplayEnhancedSolverResult(result)
}
