package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// ============================================================
// 情节链定义数据库
// 通过事件 Title 匹配启动条件，不需要修改 Event 结构体
// ============================================================

type ChainTriggerDef struct {
	ChainID     string
	Description string
	IsBullish   bool
	Momentum    int     // 初始动能（回合数）
	TruthLevel  float64 // 基础真实度（会有随机浮动）
	Successors  []string // 链内后续事件 Title（模糊匹配前缀）
}

// 能触发情节链的事件 Title → 链定义
var ChainTriggers = map[string]ChainTriggerDef{
	// ===== 牛市情节链 =====
	"国家级产业政策出台": {
		ChainID:     "政策牛市链",
		Description: "国家产业政策发酵，资金持续涌入，但监管随时可能收紧",
		IsBullish:   true,
		Momentum:    8,
		TruthLevel:  0.7,
		Successors:  []string{"外资疯狂扫货", "机构扎堆调研+研报唱多", "热门概念炒作风口", "技术面突破关键压力", "行业龙头超预期并购"},
	},
	"实控人真金白银增持": {
		ChainID:     "大股东增持链",
		Description: "大股东护盘，但主力可能利用此窗口出货",
		IsBullish:   true,
		Momentum:    5,
		TruthLevel:  0.8,
		Successors:  []string{"机构扎堆调研+研报唱多", "业绩预告大超预期", "知名游资进场扫货", "放量滞涨出货"},
	},
	"行业龙头超预期并购": {
		ChainID:     "产业整合链",
		Description: "并购重组题材发酵，整合预期推动估值重构",
		IsBullish:   true,
		Momentum:    6,
		TruthLevel:  0.6,
		Successors:  []string{"热门概念炒作风口", "外资疯狂扫货", "主力资金悄然撤离", "多空分歧加剧研报打架"},
	},
	"ChatGPT引爆AI革命": {
		ChainID:     "AI科技炒作链",
		Description: "AI概念爆发，板块轮动加速，但警惕泡沫破裂",
		IsBullish:   true,
		Momentum:    10,
		TruthLevel:  0.5,
		Successors:  []string{"国产芯片重大突破", "算力需求爆发", "机构抱团科技股", "技术路线之争", "产能过剩预警"},
	},
	"碳中和国家战略": {
		ChainID:     "新能源政策链",
		Description: "碳中和政策长期利好，但原材料涨价可能吞噬利润",
		IsBullish:   true,
		Momentum:    8,
		TruthLevel:  0.65,
		Successors:  []string{"销量暴增超预期", "巨头跨界入局", "补贴政策延续", "原材料价格飙涨"},
	},

	// ===== 熊市情节链 =====
	"监管突击约谈龙头企业": {
		ChainID:     "监管打压链",
		Description: "监管高压持续，市场信心崩溃，抄底需谨慎",
		IsBullish:   false,
		Momentum:    8,
		TruthLevel:  0.85,
		Successors:  []string{"财务造假疑云", "大股东违规减持曝光", "融资盘开始松动止损", "系统性金融风险"},
	},
	"大股东违规减持曝光": {
		ChainID:     "减持崩盘链",
		Description: "大股东出逃，散户接盘，股价面临持续重压",
		IsBullish:   false,
		Momentum:    6,
		TruthLevel:  0.9,
		Successors:  []string{"主力资金悄然撤离", "融资盘开始松动止损", "业绩不及预期传闻", "行业竞争加剧"},
	},
	"地缘政治极端事件": {
		ChainID:     "地缘危机链",
		Description: "地缘风险持续发酵，避险情绪主导市场",
		IsBullish:   false,
		Momentum:    7,
		TruthLevel:  0.75,
		Successors:  []string{"供应链断裂风险", "实体清单扩容", "汇率贬值利好", "全面对抗升级"},
	},
	"全球大流行": {
		ChainID:     "黑天鹅恐慌链",
		Description: "系统性风险爆发，流动性危机传导全市场",
		IsBullish:   false,
		Momentum:    10,
		TruthLevel:  0.95,
		Successors:  []string{"股市熔断", "经济停摆", "央行紧急救市", "疫情概念炒作"},
	},
	"系统性金融风险": {
		ChainID:     "流动性危机链",
		Description: "金融系统压力上升，连锁反应可能引发踩踏",
		IsBullish:   false,
		Momentum:    9,
		TruthLevel:  0.8,
		Successors:  []string{"股市熔断", "大股东违规减持曝光", "融资盘开始松动止损", "央行紧急救市"},
	},
}

// ============================================================
// 情节链引擎核心函数
// ============================================================

// 尝试触发情节链（在每天换事件时调用）
func tryActivateChain(state *GameState) {
	// 如果当前已有活跃链且动能充足，不重新触发
	if state.ActiveChain != nil && state.ActiveChain.Momentum > 3 {
		return
	}

	eventTitle := state.CurrentEvent.Title
	if def, ok := ChainTriggers[eventTitle]; ok {
		// 随机浮动真实度（±0.1），增加博弈不确定性
		truthLevel := def.TruthLevel + (rand.Float64()*0.2 - 0.1)
		if truthLevel > 1.0 {
			truthLevel = 1.0
		}
		if truthLevel < 0.0 {
			truthLevel = 0.0
		}

		state.ActiveChain = &EventChain{
			ChainID:     def.ChainID,
			StartDay:    state.Day,
			Momentum:    def.Momentum,
			TruthLevel:  truthLevel,
			IsBullish:   def.IsBullish,
			Description: def.Description,
		}

		chainColor := Green
		chainIcon := "🔗📈"
		if !def.IsBullish {
			chainColor = Red
			chainIcon = "🔗📉"
		}
		state.AddLog(fmt.Sprintf("%s%s 情节链触发: [%s]%s", chainColor, chainIcon, def.ChainID, Reset))
	}
}

// 每回合衰减情节链动能
func decayChainMomentum(state *GameState) {
	if state.ActiveChain == nil {
		return
	}
	state.ActiveChain.Momentum--
	if state.ActiveChain.Momentum <= 0 {
		// 链条结束时，根据 TruthLevel 决定结局
		if state.ActiveChain.TruthLevel >= 0.6 {
			if state.ActiveChain.IsBullish {
				state.AddLog(fmt.Sprintf("%s✅ 情节链兑现: [%s] 基本面支撑，利好成真%s", Green, state.ActiveChain.ChainID, Reset))
			} else {
				state.AddLog(fmt.Sprintf("%s☠️ 情节链兑现: [%s] 利空落地，风险出清%s", Red, state.ActiveChain.ChainID, Reset))
			}
		} else {
			state.AddLog(fmt.Sprintf("%s💨 情节链熄灭: [%s] 炒作退潮，回归现实%s", Yellow, state.ActiveChain.ChainID, Reset))
		}
		state.ActiveChain = nil
	}
}

// 情节链感知的事件选择：优先从链内 Successors 选择后续事件
func selectNextEventWithChain(state *GameState, eventPool []Event) Event {
	if state.ActiveChain != nil {
		// 找到当前链的触发定义，获取 Successors
		currentDef, found := ChainTriggers[state.CurrentEvent.Title]
		if !found {
			// 当前事件未在触发表里，但可能是链内事件——从当前链寻找
			for _, def := range ChainTriggers {
				if def.ChainID == state.ActiveChain.ChainID {
					currentDef = def
					found = true
					break
				}
			}
		}

		if found && len(currentDef.Successors) > 0 && rand.Float64() < 0.75 {
			// 75% 概率从链内后续事件中选
			successorTitle := currentDef.Successors[rand.Intn(len(currentDef.Successors))]
			for _, e := range eventPool {
				if e.Title == successorTitle {
					return e
				}
			}
		}
	}

	// 无链或链内找不到时，退回马尔科夫
	return selectNextEventByMarkov(state.CurrentEvent, eventPool)
}

// ============================================================
// MarketMood 引擎
// ============================================================

// 更新全局市场情绪值 MarketMood
// 每回合调用一次，带均值回归防止发散
func updateMarketMood(state *GameState) {
	sentiment := state.CurrentEvent.Sentiment

	// 情绪冲击：sentiment > 1 推高 mood, < 1 压低 mood
	impact := (sentiment - 1.0) * 40.0

	// 活跃链对情绪的额外放大
	if state.ActiveChain != nil {
		chainBoost := float64(state.ActiveChain.Momentum) * 2.0
		if !state.ActiveChain.IsBullish {
			chainBoost = -chainBoost
		}
		impact += chainBoost * state.ActiveChain.TruthLevel
	}

	state.MarketMood += impact

	// 均值回归：每回合向 0 衰减 8%（防止无限发散）
	state.MarketMood *= 0.92

	// 硬性 clamp [-100, 100]
	state.MarketMood = math.Max(-100, math.Min(100, state.MarketMood))
}

// 获取 MarketMood 对价格的基础修正量（注入 processTurn 的价格引擎）
func getMarketMoodPriceBonus(state *GameState) float64 {
	// MarketMood / 1500 换算成价格变动百分比
	// ±100 mood → ±6.7% 额外价格偏移
	return state.MarketMood / 1500.0
}

// ============================================================
// 舆情帖子生成系统
// ============================================================

var whaleAuthors = []string{
	"江浙刺客游资", "宁波涨停敢死队", "游资大佬老王", "华鑫证券席位",
}
var retailAuthors = []string{
	"散户007", "A股老韭菜", "梭哈战士", "抄底小能手", "空仓等机会",
}
var noiseAuthors = []string{
	"财经小编", "匿名股友", "某机构内部人士?", "消息灵通人士",
}
var analystAuthors = []string{
	"蚂蚁券商研究所", "知名分析师张教授", "华泰策略团队",
}

// 牛市舆情模板
var bullishTemplates = []string{
	"🚀 今天的涨势完全没问题，主力在洗盘，明天必然大涨！",
	"📈 筹码结构非常健康，量能配合良好，强烈看多！",
	"💎 这种回调就是上车机会，不买会后悔一辈子",
	"🔥 大资金在悄悄建仓，散户还不知道，快跑！",
	"📊 技术面已经蓄势完毕，突破在即，目标价+30%%！",
	"🌟 国家政策支持，行业景气向上，长期持有无压力！",
}

// 熊市舆情模板
var bearishTemplates = []string{
	"📉 这个走势很危险，我已经清仓了，大家自己判断",
	"⚠️ 主力已经跑了，还有散户在追高，太可怜了",
	"🚨 量价背离，小心接盘，这种行情我见过太多了",
	"💀 有人在砸盘，明天止损，不要犹豫",
	"😰 感觉风向不对，今晚有大消息，先撤为敬",
	"🔻 散户都在追高，机构在出货，这是教科书级别的陷阱",
}

// 噪音舆情模板（干扰信息）
var noiseTemplates = []string{
	"🎲 今日龙虎榜数据异常，可能有大资金在暗中操盘",
	"🕵️ 听说某大佬已经满仓了，不知真假，自己判断",
	"💬 股吧里有人说明天要停牌，消息来源不明",
	"🤔 这个位置我也看不懂，既不敢卖也不敢买",
	"📰 媒体开始唱多了，这往往是见顶信号...",
	"🎯 今日成交量放大，多空分歧加剧，谨慎操作",
}

// 中性分析舆情
var neutralTemplates = []string{
	"📐 从技术面看，关键支撑在 $%.2f，跌破要重新评估",
	"⚖️ 当前市盈率偏高，但行业景气度支撑估值，中性看待",
	"🔍 换手率%.1f%%，属于正常水平，无异常波动",
	"📋 综合来看，短期震荡，中期偏多，长期看基本面",
}

// 情节链相关舆情（透露链条信息）
var chainBullishHints = []string{
	"🔒 消息人士透露，这次政策利好还有后续，耐心持有",
	"📡 据悉几家顶级机构已在讨论，链条还没走完",
	"💡 这次的行情和上次XX事件很相似，当时最终涨了XX%%",
	"🎯 有人在悄悄抄底，知道内情的人都在加仓",
}
var chainBearishHints = []string{
	"😱 不妙，消息说利空还在发酵，后面还有更多...",
	"🚨 监管的动作还没结束，小心继续有雷",
	"🔥 资金在快速撤离，这种情况历史上都跌得很惨",
	"⚡ 传言后续还有更重磅的利空，先止损保命",
}

// 生成每回合的舆情帖子
func generateSocialFeed(state *GameState) []SocialPost {
	posts := []SocialPost{}
	numPosts := 2 + rand.Intn(3) // 2-4 条

	for i := 0; i < numPosts; i++ {
		var post SocialPost
		roll := rand.Float64()

		if roll < 0.15 && state.ActiveChain != nil {
			// 15%：情节链内幕帖（有参考价值）
			post.TruthLevel = 0.6 + rand.Float64()*0.3
			post.IsNoise = false
			if state.ActiveChain.IsBullish {
				post.AuthorType = "whale"
				post.Author = whaleAuthors[rand.Intn(len(whaleAuthors))]
				post.Content = chainBullishHints[rand.Intn(len(chainBullishHints))]
			} else {
				post.AuthorType = "analyst"
				post.Author = analystAuthors[rand.Intn(len(analystAuthors))]
				post.Content = chainBearishHints[rand.Intn(len(chainBearishHints))]
			}
		} else if roll < 0.40 {
			// 25%：基于当前事件情绪的真实观点
			post.IsNoise = false
			if state.CurrentEvent.Sentiment > 1.1 {
				post.AuthorType = "whale"
				post.Author = whaleAuthors[rand.Intn(len(whaleAuthors))]
				post.Content = bullishTemplates[rand.Intn(len(bullishTemplates))]
				post.TruthLevel = 0.5 + rand.Float64()*0.3
			} else if state.CurrentEvent.Sentiment < 0.9 {
				post.AuthorType = "retail"
				post.Author = retailAuthors[rand.Intn(len(retailAuthors))]
				post.Content = bearishTemplates[rand.Intn(len(bearishTemplates))]
				post.TruthLevel = 0.4 + rand.Float64()*0.4
			} else {
				post.AuthorType = "analyst"
				post.Author = analystAuthors[rand.Intn(len(analystAuthors))]
				tmpl := neutralTemplates[rand.Intn(len(neutralTemplates))]
				// 填充模板中的数值
				if len(tmpl) > 0 {
					switch rand.Intn(4) {
					case 0:
						post.Content = fmt.Sprintf(tmpl, state.Price*0.95)
					case 1:
						post.Content = fmt.Sprintf(tmpl, state.Price*1.05)
					case 2:
						turnover := 0.0
						if len(state.VolumeHistory) > 0 && state.TotalMarketShares > 0 {
							turnover = float64(state.VolumeHistory[len(state.VolumeHistory)-1]) / float64(state.TotalMarketShares) * 100
						}
						post.Content = fmt.Sprintf(tmpl, turnover)
					default:
						post.Content = tmpl
					}
				}
				post.TruthLevel = 0.6 + rand.Float64()*0.2
			}
		} else if roll < 0.70 {
			// 30%：散户情绪（跟风，参考价值低）
			post.AuthorType = "retail"
			post.Author = retailAuthors[rand.Intn(len(retailAuthors))]
			post.IsNoise = false
			if state.MarketMood > 20 {
				post.Content = bullishTemplates[rand.Intn(len(bullishTemplates))]
			} else if state.MarketMood < -20 {
				post.Content = bearishTemplates[rand.Intn(len(bearishTemplates))]
			} else {
				post.Content = noiseTemplates[rand.Intn(len(noiseTemplates))]
			}
			post.TruthLevel = 0.2 + rand.Float64()*0.3
		} else {
			// 30%：噪音/误导信息
			post.AuthorType = "noise"
			post.Author = noiseAuthors[rand.Intn(len(noiseAuthors))]
			post.IsNoise = true
			post.TruthLevel = rand.Float64() * 0.3
			// 反转当前情绪（制造噪音）
			if state.CurrentEvent.Sentiment > 1.0 {
				post.Content = bearishTemplates[rand.Intn(len(bearishTemplates))]
			} else {
				post.Content = bullishTemplates[rand.Intn(len(bullishTemplates))]
			}
		}

		if post.Content != "" {
			posts = append(posts, post)
		}
	}

	return posts
}

// ============================================================
// 龙虎榜系统
// ============================================================

// 席位标签映射（AI名称 → 龙虎榜标签）
var seatLabels = map[string]string{
	"江浙刺客游资": "章盟主（顶级短线席位）",
	"内资长线底仓": "沪股通外资席位",
	"高频打板量化": "知名量化机构",
	"幻方网格量化": "幻方量化（百亿级）",
	"发财梦新韭菜": "拉萨天团散户基地",
	"装死死扛老散": "宁波成指型老散",
	"平准托底基金": "⭐国家队（平准基金）",
}

// 在清算阶段累计龙虎榜数据
func recordDragonTigerTrade(state *GameState, name string, buyAmt float64, sellAmt float64, isPlayer bool) {
	if state.TurnBuyAmt == nil {
		state.TurnBuyAmt = make(map[string]float64)
	}
	if state.TurnSellAmt == nil {
		state.TurnSellAmt = make(map[string]float64)
	}
	state.TurnBuyAmt[name] += buyAmt
	state.TurnSellAmt[name] += sellAmt
	if isPlayer {
		state.TurnBuyAmt["【玩家】"] += buyAmt
		state.TurnSellAmt["【玩家】"] += sellAmt
	}
}

// 尾盘结算龙虎榜（每天 advanceTime 进入次日早盘时调用）
func settleDragonTigerList(state *GameState) {
	if state.TurnBuyAmt == nil && state.TurnSellAmt == nil {
		return
	}

	// 汇总所有席位
	allNames := make(map[string]bool)
	for name := range state.TurnBuyAmt {
		allNames[name] = true
	}
	for name := range state.TurnSellAmt {
		allNames[name] = true
	}

	entries := []DragonTigerEntry{}
	for name := range allNames {
		if name == "" {
			continue
		}
		label, ok := seatLabels[name]
		if !ok {
			if name == "【玩家】" {
				label = "你（操盘玩家）"
			} else {
				label = name
			}
		}
		entry := DragonTigerEntry{
			Name:     name,
			Label:    label,
			BuyAmt:   state.TurnBuyAmt[name],
			SellAmt:  state.TurnSellAmt[name],
			IsPlayer: name == "【玩家】",
		}
		entries = append(entries, entry)
	}

	// 按总成交金额降序排序
	sort.Slice(entries, func(i, j int) bool {
		ti := entries[i].BuyAmt + entries[i].SellAmt
		tj := entries[j].BuyAmt + entries[j].SellAmt
		return ti > tj
	})

	// 只保留前 5
	if len(entries) > 5 {
		entries = entries[:5]
	}

	state.DragonTigerList = entries

	// 重置当日累计数据
	state.TurnBuyAmt = make(map[string]float64)
	state.TurnSellAmt = make(map[string]float64)

	// 将龙虎榜入日志
	if len(entries) > 0 {
		state.AddLog(fmt.Sprintf("%s🐉 今日龙虎榜已出炉，共 %d 席位上榜%s", Yellow, len(entries), Reset))
	}
}

// 渲染龙虎榜
func renderDragonTigerList(state *GameState) {
	if len(state.DragonTigerList) == 0 {
		return
	}
	fmt.Printf("\n%s┌────────────────── 🐉 今日龙虎榜 ──────────────────────────┐%s\n", Yellow, Reset)
	for i, entry := range state.DragonTigerList {
		netFlow := entry.BuyAmt - entry.SellAmt
		netColor := Green
		netIcon := "▲买超"
		if netFlow < 0 {
			netColor = Red
			netIcon = "▼卖超"
		}
		nameColor := Cyan
		if entry.IsPlayer {
			nameColor = Purple
		}
		rank := fmt.Sprintf("#%d", i+1)
		fmt.Printf("  %s%s%s %s%-20s%s  买:$%6.0f  卖:$%6.0f  %s%s$%.0f%s\n",
			Yellow, rank, Reset,
			nameColor, entry.Label, Reset,
			entry.BuyAmt, entry.SellAmt,
			netColor, netIcon, math.Abs(netFlow), Reset)
	}
	fmt.Printf("%s└──────────────────────────────────────────────────────────────┘%s\n", Yellow, Reset)
}

// 渲染舆情 Feed（替代原来的纯文本日志区域）
func renderSocialFeed(state *GameState) {
	if len(state.SocialFeed) == 0 {
		return
	}
	fmt.Printf("\n%s┌────────────────── 📱 股吧舆情 Feed ────────────────────────┐%s\n", Cyan, Reset)

	// 显示活跃链条横幅
	if state.ActiveChain != nil {
		chainColor := Green
		chainIcon := "📈"
		if !state.ActiveChain.IsBullish {
			chainColor = Red
			chainIcon = "📉"
		}
		fmt.Printf("  %s%s [情节链] %s 动能:%d 回合剩余%s\n",
			chainColor, chainIcon, state.ActiveChain.ChainID,
			state.ActiveChain.Momentum, Reset)
		fmt.Printf("  %s↳ %s%s\n", chainColor, state.ActiveChain.Description, Reset)
	}
	fmt.Printf("%s  ─────────────────────────────────────────────%s\n", Cyan, Reset)

	// MarketMood 指示器
	moodBar := renderMoodBar(state.MarketMood)
	moodLabel := "中性"
	moodColor := Yellow
	if state.MarketMood > 40 {
		moodLabel = "情绪亢奋"
		moodColor = Red
	} else if state.MarketMood > 15 {
		moodLabel = "偏多"
		moodColor = Green
	} else if state.MarketMood < -40 {
		moodLabel = "极度悲观"
		moodColor = Red
	} else if state.MarketMood < -15 {
		moodLabel = "偏空"
		moodColor = Red
	}
	fmt.Printf("  %s全市情绪: %s%s [%s] %.0f%s\n", moodColor, moodBar, moodColor, moodLabel, state.MarketMood, Reset)
	fmt.Printf("  %s─────────────────────────────────────────────%s\n", Cyan, Reset)

	// 帖子列表
	for _, post := range state.SocialFeed {
		authorColor := Cyan
		authorIcon := "💬"
		switch post.AuthorType {
		case "whale":
			authorColor = Purple
			authorIcon = "🐋"
		case "retail":
			authorColor = Green
			authorIcon = "🌱"
		case "noise":
			authorColor = Gray
			authorIcon = "🎭"
		case "analyst":
			authorColor = Blue
			authorIcon = "📊"
		}
		fmt.Printf("  %s%s %s%s%s: %s\n",
			authorColor, authorIcon, authorColor, post.Author, Reset,
			post.Content)
	}
	fmt.Printf("  %s%s─────────────────────────────────────────────%s\n", Gray, Gray, Reset)
	fmt.Printf("  %s💡 提示: 🐋=大资金 📊=机构 🌱=散户 🎭=噪音%s\n", Gray, Reset)
	fmt.Printf("%s└──────────────────────────────────────────────────────────────┘%s\n", Cyan, Reset)
}

// 渲染情绪条
func renderMoodBar(mood float64) string {
	barLen := 20
	filled := int((mood + 100) / 200.0 * float64(barLen))
	if filled < 0 {
		filled = 0
	}
	if filled > barLen {
		filled = barLen
	}
	bar := ""
	for i := 0; i < barLen; i++ {
		if i == 10 {
			bar += "|" // 中间基准线
		}
		if i < filled {
			if mood > 0 {
				bar += "█"
			} else {
				bar += "░"
			}
		} else {
			if mood > 0 {
				bar += "░"
			} else {
				bar += "█"
			}
		}
	}
	return bar
}

// ============================================================
// 情报系统（接入 IntelPoints）
// ============================================================

// 使用情报点查看当前情节链内幕
func useIntelOnChain(state *GameState) string {
	data := loadAchievementData()
	if data.IntelPoints < 2 {
		return fmt.Sprintf("%s❌ 情报点不足（需要2点，当前%d点）%s", Red, data.IntelPoints, Reset)
	}
	if state.ActiveChain == nil {
		return fmt.Sprintf("%s🔍 当前无活跃情节链，无法使用情报%s", Yellow, Reset)
	}
	if state.IntelUsedThisTurn {
		return fmt.Sprintf("%s⚠️ 本回合已使用情报，请下回合再试%s", Yellow, Reset)
	}

	// 扣除情报点
	data.IntelPoints -= 2
	saveAchievementData(data)
	state.IntelUsedThisTurn = true

	chain := state.ActiveChain
	truthDesc := ""
	truthColor := Yellow
	if chain.TruthLevel >= 0.75 {
		truthDesc = fmt.Sprintf("高度可信（基本面支撑强，真实度 %.0f%%）", chain.TruthLevel*100)
		truthColor = Green
	} else if chain.TruthLevel >= 0.5 {
		truthDesc = fmt.Sprintf("存疑（半真半假，真实度 %.0f%%）", chain.TruthLevel*100)
		truthColor = Yellow
	} else {
		truthDesc = fmt.Sprintf("纯炒作泡沫（基本面不支持，真实度仅 %.0f%%）", chain.TruthLevel*100)
		truthColor = Red
	}

	direction := "上涨"
	if !chain.IsBullish {
		direction = "下跌"
	}

	return fmt.Sprintf(`
%s╔══════════ 🕵️ 机密情报（已消耗2点情报点）══════════╗%s
  %s情节链: %s [%s方向]%s
  %s评估: %s%s
  %s动能剩余: %d 回合（约 %d 天）%s
  %s策略建议: %s%s
%s╚════════════════════════════════════════════════════╝%s`,
		Purple, Reset,
		Cyan, chain.ChainID, direction, Reset,
		truthColor, truthDesc, Reset,
		Yellow, chain.Momentum, chain.Momentum/2, Reset,
		Green, getIntelAdvice(chain), Reset,
		Purple, Reset)
}

func getIntelAdvice(chain *EventChain) string {
	if chain.IsBullish && chain.TruthLevel >= 0.65 {
		return "链条真实，可考虑趁回调加仓，目标持有到链条兑现"
	} else if chain.IsBullish && chain.TruthLevel < 0.65 {
		return "炒作成分大，建议在高点分批减仓，不要贪心"
	} else if !chain.IsBullish && chain.TruthLevel >= 0.65 {
		return "利空为真，建议尽快减仓或清仓，等链条结束后再评估"
	} else {
		return "利空可能过度，可小仓位试探性抄底，但注意止损"
	}
}
