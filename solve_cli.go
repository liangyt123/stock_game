package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

// 求解器测试和演示程序
func RunSolverCLI() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("\033[H\033[2J") // 清屏

	fmt.Printf("%s╔═══════════════════════════════════════════════════════════╗%s\n", Purple, Reset)
	fmt.Printf("%s║               🤖 残局求解器 - AI最优策略分析              ║%s\n", Purple, Reset)
	fmt.Printf("%s╚═══════════════════════════════════════════════════════════╝%s\n", Purple, Reset)
	fmt.Println()

	fmt.Printf("%s这个工具使用动态规划算法，搜索残局场景的最优策略%s\n", Cyan, Reset)
	fmt.Printf("%s可以帮助你了解：%s\n", Cyan, Reset)
	fmt.Println("  • 理论上的最佳操作序列")
	fmt.Println("  • 预期收益和成功概率")
	fmt.Println("  • 不同决策点的关键选择")
	fmt.Println()

	for {
		fmt.Printf("%s═══════════════════════════════════════════════════════════%s\n", Cyan, Reset)

		// 创建菜单项
		items := make([]MenuItem, len(EndgameScenarios)+2) // +2 为退出和复盘

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
				Label: scenario.Name,
				Description: fmt.Sprintf("%s %s | 目标: %.0f%% | 时限: %d回合",
					diffIcon, scenario.Difficulty,
					scenario.TargetProfit*100,
					scenario.TimeLimit),
				Value: fmt.Sprintf("%d", i+1),
			}
		}

		// 添加特殊选项
		items[len(EndgameScenarios)] = MenuItem{
			Label:       "📝 实战复盘分析（对比你的操作）",
			Description: "",
			Value:       "99",
			Icon:        "📝",
		}

		items[len(EndgameScenarios)+1] = MenuItem{
			Label:       "退出程序",
			Description: "",
			Value:       "0",
		}

		menu := NewInteractiveMenu("请选择要求解的残局场景：", items)
		input, _ := menu.Show()

		if input == "" || input == "0" {
			fmt.Printf("\n%s感谢使用！%s\n\n", Cyan, Reset)
			break
		}

		if input == "99" {
			fmt.Print("\033[H\033[2J") // 清屏
			RunReplayAnalyzer()
			fmt.Printf("\n%s按回车键继续...%s ", Cyan, Reset)
			reader.ReadString('\n')
			fmt.Print("\033[H\033[2J") // 清屏
			continue
		}

		sceneNum, err := strconv.Atoi(input)
		if err != nil || sceneNum < 1 || sceneNum > len(EndgameScenarios) {
			fmt.Printf("\n%s无效的场景编号，请重新输入%s\n\n", Red, Reset)
			continue
		}

		// 选择求解模式（使用交互式菜单）
		modeItems := []MenuItem{
			{Label: "交互式学习模式（渐进式提示）", Value: "1", Icon: "🎓"},
			{Label: "直接查看完整答案", Value: "2", Icon: "📊"},
		}

		modeMenu := NewInteractiveMenu("请选择求解模式:", modeItems)
		modeValue, _ := modeMenu.Show()

		// 执行求解
		fmt.Print("\033[H\033[2J") // 清屏

		if modeValue == "1" {
			RunInteractiveSolver(sceneNum - 1)
		} else if modeValue == "2" {
			QuickSolve(sceneNum - 1)
		} else {
			// ESC退出，默认显示答案
			QuickSolve(sceneNum - 1)
		}

		// 等待用户查看结果
		fmt.Printf("\n%s按回车键继续...%s ", Cyan, Reset)
		reader.ReadString('\n')
		fmt.Print("\033[H\033[2J") // 清屏
	}
}
