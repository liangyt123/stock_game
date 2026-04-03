package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	runewidth "github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

// 颜色和样式常量
const (
	Bold     = "\033[1m"
	DarkGray = "\033[90m"
)

// MenuItem 菜单项
type MenuItem struct {
	Label       string // 显示文本
	Description string // 描述（可选）
	Value       string // 返回值
	Icon        string // 图标（可选）
}

// InteractiveMenu 交互式菜单配置
type InteractiveMenu struct {
	Title       string        // 菜单标题
	Items       []MenuItem    // 菜单项
	Selected    int           // 当前选中项
	ShowNumbers bool          // 是否显示序号
	UseArrows   bool          // 是否使用箭头指示（true=箭头，false=高亮）
	Reader      *bufio.Reader // 输入读取器（可选，如果不提供则直接读取os.Stdin）
	LastHeight  int           // 上次渲染的行数 (用于 Windows 下平滑刷新)
}

// NewInteractiveMenu 创建新菜单
func NewInteractiveMenu(title string, items []MenuItem) *InteractiveMenu {
	return &InteractiveMenu{
		Title:       title,
		Items:       items,
		Selected:    0,
		ShowNumbers: true,
		UseArrows:   true,
		Reader:      nil,
	}
}

// Show 显示菜单并获取用户选择
func (m *InteractiveMenu) Show() (string, int) {
	// 1. 隐藏光标
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h") // 恢复光标

	// 2. 初始高度设为0
	m.LastHeight = 0

	// 保存当前终端状态
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return m.fallbackToNumberInput()
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	for {
		m.render()

		key := m.readKey()

		switch key {
		case "ctrl-c":
			m.clearMenu()
			fmt.Print("\r\n程序已退出\r\n")
			os.Exit(0)
		case "up", "k":
			if m.Selected > 0 {
				m.Selected--
			} else if len(m.Items) > 0 {
				m.Selected = len(m.Items) - 1
			}
		case "down", "j":
			if len(m.Items) > 0 {
				if m.Selected < len(m.Items)-1 {
					m.Selected++
				} else {
					m.Selected = 0
				}
			}
		case "enter":
			m.clearMenu()
			if len(m.Items) > 0 && m.Selected >= 0 && m.Selected < len(m.Items) {
				return m.Items[m.Selected].Value, m.Selected
			}
			return "", -1
		case "esc", "q", "backspace":
			m.clearMenu()
			return "", -1
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			num := int(key[0] - '0')
			if num > 0 && num <= len(m.Items) {
				m.Selected = num - 1
				m.clearMenu()
				return m.Items[m.Selected].Value, m.Selected
			}
		}
	}
}

// render 渲染菜单
func (m *InteractiveMenu) render() {
	var b strings.Builder
	currentLines := 0

	// 1. 如果之前渲染过，先将光标向上移动到菜单起始位置
	if m.LastHeight > 0 {
		b.WriteString(fmt.Sprintf("\033[%dA", m.LastHeight))
	}
	// 2. 清除光标之后的所有行
	b.WriteString("\033[J")
	b.WriteString("\r\n")

	if m.Title != "" {
		b.WriteString(fmt.Sprintf("%s%s%s\r\n", Cyan+Bold, m.Title, Reset))
		b.WriteString(fmt.Sprintf("%s%s%s\r\n\r\n", Cyan, strings.Repeat("─", displayWidth(m.Title)), Reset))
		currentLines += 3
	}
	currentLines++ // 对应 \033[J 后面的 \r\n

	maxWidth := 0
	for _, item := range m.Items {
		w := displayWidth(item.Label)
		if item.Icon != "" {
			w += displayWidth(item.Icon) + 1
		}
		if w > maxWidth {
			maxWidth = w
		}
	}
	maxWidth += 4

	for i, item := range m.Items {
		// 在打印每一行之前，先强制清除该行的旧内容 (\033[2K) 并确保光标在行首 (\r)
		b.WriteString("\033[2K\r")

		prefix := "  "
		if m.Selected == i {
			if m.UseArrows {
				prefix = Green + "▶ " + Reset
			}
		}

		number := ""
		if m.ShowNumbers {
			number = fmt.Sprintf("%s[%d]%s ", Yellow, i+1, Reset)
		}

		icon := ""
		if item.Icon != "" {
			icon = item.Icon + " "
		}

		label := item.Label
		if m.Selected == i {
			label = Green + Bold + label + Reset
		}

		currentW := displayWidth(item.Label)
		if item.Icon != "" {
			currentW += displayWidth(item.Icon) + 1
		}
		padding := strings.Repeat(" ", maxWidth-currentW)

		desc := ""
		if item.Description != "" {
			if m.Selected == i {
				desc = Gray + " " + item.Description + Reset
			} else {
				desc = DarkGray + " " + item.Description + Reset
			}
		}

		b.WriteString(prefix + number + icon + label + padding + desc + "\r\n")
		currentLines++
	}

	b.WriteString("\r\n" + DarkGray + "↑↓ 选择  |  数字快选  |  Enter 确认  |  ESC 退出" + Reset)
	currentLines++ // 页脚前的空行已经在最后一个 item 的 \r\n 中了，这里只算页脚文本这一行

	// LastHeight 代表我们要向上回退的次数。从第 N 行回到第 1 行需要移动 N-1 次。
	m.LastHeight = currentLines
	fmt.Print(b.String())
	os.Stdout.Sync()
}

func displayWidth(s string) int {
	return runewidth.StringWidth(s)
}

func (m *InteractiveMenu) clearMenu() {
	// 彻底清理 alternate buffer
	fmt.Print("\033[H\033[J")
}

func (m *InteractiveMenu) readKey() string {
	buf := make([]byte, 3)
	var n int
	var err error

	if m.Reader != nil {
		n, err = m.Reader.Read(buf)
	} else {
		n, err = os.Stdin.Read(buf)
	}

	if err != nil || n == 0 {
		return ""
	}

	// Enter (适配 Windows 连按或长按产生多字节的情况)
	if n > 0 && (buf[0] == 10 || buf[0] == 13) {
		return "enter"
	}

	// Arrows
	if n == 3 && buf[0] == 27 && buf[1] == 91 {
		switch buf[2] {
		case 65: return "up"
		case 66: return "down"
		case 67: return "right"
		case 68: return "left"
		}
	}

	// Single keys
	if n == 1 {
		switch buf[0] {
		case 3: return "ctrl-c"
		case 27: return "esc"
		case 127: return "backspace"
		case 'k', 'K': return "up"
		case 'j', 'J': return "down"
		case 'h', 'H': return "left"
		case 'l', 'L': return "right"
		case 'q', 'Q': return "esc"
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return string(buf[0])
		}
	}
	return ""
}

func (m *InteractiveMenu) fallbackToNumberInput() (string, int) {
	fmt.Printf("\n%s\n", m.Title)
	for i, item := range m.Items {
		fmt.Printf("[%d] %s %s\n", i+1, item.Label, item.Description)
	}
	fmt.Printf("\n请输入编号: ")
	var input string
	if m.Reader != nil {
		input, _ = m.Reader.ReadString('\n')
	} else {
		input, _ = bufio.NewReader(os.Stdin).ReadString('\n')
	}
	input = strings.TrimSpace(input)
	for i, item := range m.Items {
		if input == fmt.Sprintf("%d", i+1) || input == item.Value {
			return item.Value, i
		}
	}
	return "", -1
}

func SimpleMenu(title string, options []string) (int, string) {
	items := make([]MenuItem, len(options))
	for i, opt := range options {
		items[i] = MenuItem{Label: opt, Value: fmt.Sprintf("%d", i+1)}
	}
	menu := NewInteractiveMenu(title, items)
	_, index := menu.Show()
	return index, options[index]
}

func YesNoMenu(question string) bool {
	items := []MenuItem{{Label: "是", Value: "yes"}, {Label: "否", Value: "no"}}
	menu := NewInteractiveMenu(question, items)
	v, _ := menu.Show()
	return v == "yes"
}

func NumberedMenu(title string, items []MenuItem) (string, int) {
	menu := NewInteractiveMenu(title, items)
	return menu.Show()
}
