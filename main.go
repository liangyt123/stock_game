package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
)

// 事件情绪分类（用于马尔科夫链）
type EventCategory int

const (
	ExtremeOptimistic EventCategory = iota // 极度乐观
	MildOptimistic                         // 温和乐观
	Neutral                                // 中性
	MildPessimistic                        // 温和悲观
	ExtremePanic                           // 极度恐慌
)

type Event struct {
	Title        string
	Desc         string
	Sentiment    float64       // >1乐观, <1恐慌
	FearModifier float64       // 恐慌乘数
	Category     EventCategory // 事件情绪分类（马尔科夫链用）
}

// 事件库 - 经典模式（25个事件）
var ClassicEvents = []Event{
	// 极度乐观 (Sentiment > 1.3) - 5个
	{"国家级产业政策出台", "类似新能源国家战略、半导体扶持，板块瞬间狂欢！但记住，当年光伏531政策后龙头三天暴跌40%...", 1.5, 0.5, ExtremeOptimistic},
	{"实控人真金白银增持", "董事长以个人名义豪掷数亿增持，这可不是嘴上说说。但也要警惕：增持当天见顶的案例比比皆是。", 1.4, 0.4, ExtremeOptimistic},
	{"央行意外释放流动性", "类似降准降息超预期组合拳，资金面宽松信号明确。但历史证明，放水不一定流入股市，更可能去炒房...", 1.45, 0.55, ExtremeOptimistic},
	{"行业龙头超预期并购", "类似芯片龙头突然宣布海外并购，市场解读为国产替代加速。游资闻风而动，封板资金超10亿！", 1.42, 0.48, ExtremeOptimistic},
	{"外资疯狂扫货", "北向资金单日净流入破百亿，陆股通持股比例逼近上限。聪明钱在抄底还是接盘？", 1.38, 0.52, ExtremeOptimistic},

	// 温和乐观 (Sentiment 1.1-1.3) - 5个
	{"机构扎堆调研+研报唱多", "多家顶级券商密集发布看多报告，目标价上调30%。散户蜂拥而入，但机构是在调研还是在出货？", 1.25, 0.7, MildOptimistic},
	{"业绩预告大超预期", "公司提前披露业绩快报，净利润同比暴增200%！但要警惕'业绩兑现后无行情'的A股魔咒...", 1.2, 0.75, MildOptimistic},
	{"热门概念炒作风口", "类似元宇宙/ChatGPT/华为鸿蒙概念爆发，市场FOMO情绪浓厚。每个人都怕错过十倍股，但泡沫何时破裂？", 1.15, 0.8, MildOptimistic},
	{"知名游资进场扫货", "龙虎榜显示炒作大佬席位大举买入，带动跟风资金蜂拥。但游资有进有出，你能跑得过他们吗？", 1.18, 0.78, MildOptimistic},
	{"技术面突破关键压力", "放量突破长期压力位，MACD金叉，技术派集体唱多。但技术分析在A股真的有用吗？", 1.12, 0.82, MildOptimistic},

	// 中性平稳 (Sentiment 0.9-1.1) - 5个
	{"盘整蓄势，方向未明", "多空双方势均力敌，成交量萎缩。技术派说要突破，价值派说要回调，到底听谁的？", 1.0, 1.0, Neutral},
	{"外部扰动暂时缓解", "美联储暂停加息，地缘冲突传来和谈消息。但这是真缓和还是暴风雨前的宁静？", 0.95, 1.1, Neutral},
	{"资金观望情绪浓厚", "北向资金小幅净流入，两融余额持平。大家都在等待一个明确的方向信号，就像2015年6月初...", 1.05, 1.05, Neutral},
	{"消息真空期", "没有重大利好也没有重大利空，市场进入自我消化阶段。无聊的横盘往往孕育着大行情。", 1.02, 1.02, Neutral},
	{"震荡反复磨底", "股价在箱体内反复震荡，散户被来回打脸。主力在洗盘还是真的出不了货？", 0.98, 1.08, Neutral},

	// 温和悲观 (Sentiment 0.7-0.9) - 5个
	{"主力资金悄然撤离", "龙虎榜显示知名游资席位大举卖出，但散户还在追高接盘。这是洗盘还是真出货？历史告诉你：十有八九是后者。", 0.85, 1.4, MildPessimistic},
	{"多空分歧加剧研报打架", "某券商喊出目标价30元，另一家却说只值15元。市场陷入混乱，不确定性是杀跌的最大推手。", 0.9, 1.8, MildPessimistic},
	{"融资盘开始松动止损", "两融余额连续下降，部分杠杆资金扛不住了开始平仓。技术面破位，多头信心开始动摇。", 0.75, 1.6, MildPessimistic},
	{"业绩不及预期传闻", "市场传言财报数据可能低于预期，分析师纷纷下调评级。虽未证实，但资金已用脚投票。", 0.8, 1.5, MildPessimistic},
	{"行业竞争加剧", "新进入者疯狂抢市场份额，价格战一触即发。龙头企业护城河正在被蚕食，但股价还没反应过来。", 0.82, 1.45, MildPessimistic},

	// 极度恐慌 (Sentiment < 0.7) - 5个
	{"监管突击约谈龙头企业", "类似教培/游戏/平台经济行业被监管，核心公司被紧急约谈。市场恐慌抛售，流动性瞬间枯竭！", 0.6, 2.2, ExtremePanic},
	{"大股东违规减持曝光", "实控人通过亲属账户突击减持数亿，涉嫌信息披露违规。证监会立案调查，市场信心彻底崩塌！", 0.55, 2.5, ExtremePanic},
	{"地缘政治极端事件", "类似战争突然升级/核心供应链断裂/石油禁运。全球恐慌指数VIX飙升50%，A股避险情绪全面蔓延。", 0.4, 3.0, ExtremePanic},
	{"财务造假疑云", "做空机构发布重磅报告，质疑公司财务数据真实性。虽然公司否认，但股价已按跌停排队...", 0.5, 2.6, ExtremePanic},
	{"系统性金融风险", "类似2015股灾、2008次贷危机，银行间拆借利率飙升，流动性危机传导至股市。所有人都在夺路而逃！", 0.35, 3.2, ExtremePanic},
}

// 科技股狂潮主题（AI芯片概念炒作）
var TechBoomEvents = []Event{
	{"ChatGPT引爆AI革命", "OpenAI发布划时代产品，全球科技股暴涨。A股AI概念一字板，但多少是真AI，多少是蹭热度？", 1.55, 0.45, ExtremeOptimistic},
	{"国产芯片重大突破", "某芯片厂商宣布7nm制程量产，打破封锁！但良品率多少？产能几何？市场选择先炒为敬。", 1.48, 0.5, ExtremeOptimistic},
	{"科技巨头All in AI", "互联网大厂宣布千亿投入大模型，产业链全线受益。但烧钱大战何时见底？", 1.4, 0.55, ExtremeOptimistic},
	{"算力需求爆发", "AI训练需求暴增，算力板块供不应求。但这是刚需还是炒作泡沫？记住2000年互联网泡沫...", 1.35, 0.6, ExtremeOptimistic},
	{"机构抱团科技股", "公募基金重仓科技股比例创新高，北向资金疯狂扫货。但抱团的尽头往往是踩踏...", 1.25, 0.75, MildOptimistic},
	{"技术路线之争", "不同技术方案激烈竞争，市场陷入选择困难。押错宝的公司将被淘汰出局。", 1.0, 1.0, Neutral},
	{"产能过剩预警", "多家企业同时扩产，分析师警告供需失衡风险。但资金还在追高...", 0.85, 1.4, MildPessimistic},
	{"美国收紧技术出口", "核心设备禁售清单更新，国产替代遭遇技术瓶颈。地缘政治风险升温。", 0.65, 2.0, ExtremePanic},
	{"泡沫破裂信号", "明星科技股突然暴跌，市场惊觉估值过高。类似2000年纳斯达克崩盘前夕...", 0.5, 2.5, ExtremePanic},
	{"行业寒冬降临", "订单大幅下滑，去库存周期开启。昔日龙头开始裁员降薪，股价腰斩再腰斩。", 0.4, 2.8, ExtremePanic},
}

// 新能源泡沫主题（2020-2021新能源车疯狂）
var NewEnergyEvents = []Event{
	{"碳中和国家战略", "国家宣布2060碳中和目标，新能源产业链迎来黄金时代。资金疯狂涌入，龙头连拉涨停。", 1.52, 0.48, ExtremeOptimistic},
	{"销量暴增超预期", "新能源车渗透率突破20%，拐点已至！产业链订单爆满，扩产都来不及。", 1.45, 0.52, ExtremeOptimistic},
	{"巨头跨界入局", "传统车企、科技公司纷纷宣布造车计划。百万亿市场，谁都想分一杯羹。", 1.38, 0.58, ExtremeOptimistic},
	{"锂电池技术突破", "固态电池/钠离子电池获突破，续航里程翻倍。相关材料股一字涨停。", 1.3, 0.65, ExtremeOptimistic},
	{"补贴政策延续", "新能源补贴超预期不退坡，行业景气度持续。但补贴依赖症何时能戒掉？", 1.2, 0.75, MildOptimistic},
	{"原材料价格飙涨", "锂/钴/镍价格暴涨，中游电池厂利润被挤压。成本传导引发产业链博弈。", 0.9, 1.3, Neutral},
	{"特斯拉大幅降价", "行业龙头突然降价10%，引发价格战。谁能扛住？谁会出局？", 0.8, 1.5, MildPessimistic},
	{"补贴退坡加速", "政府宣布补贴提前退出，行业面临阵痛期。消费需求能否接棒？", 0.7, 1.8, MildPessimistic},
	{"自燃事故频发", "多起新能源车自燃事故，安全性遭质疑。监管趋严，行业信心受挫。", 0.55, 2.3, ExtremePanic},
	{"产能过剩危机", "疯狂扩产后需求不及预期，库存高企。类似2018年光伏531后的惨烈...", 0.45, 2.7, ExtremePanic},
}

// 贸易战惊魂主题（2018-2019地缘博弈）
var TradeWarEvents = []Event{
	{"贸易谈判积极信号", "双方释放善意，关税减免有望。出口链公司集体反弹，但协议何时能签？", 1.3, 0.65, ExtremeOptimistic},
	{"科技自主可控", "国产替代成为主线，自主芯片/软件迎来政策支持。但技术差距能快速追赶吗？", 1.25, 0.7, MildOptimistic},
	{"汇率贬值利好", "人民币意外贬值，出口企业竞争力提升。但资本外流压力随之而来...", 1.15, 0.8, MildOptimistic},
	{"市场暂时平静", "双方暂停加征关税，企业抓紧窗口期出货。但这只是台风眼，暴风雨还在后面。", 1.0, 1.0, Neutral},
	{"关税再度加码", "新一轮加征关税清单公布，涉及商品规模扩大。出口企业订单骤减。", 0.75, 1.6, MildPessimistic},
	{"供应链断裂风险", "核心零部件断供，产业链遭遇卡脖子。寻找替代方案需要时间和成本。", 0.65, 2.0, ExtremePanic},
	{"实体清单扩容", "多家科技公司被列入黑名单，无法采购美国技术和设备。股价连续跌停。", 0.5, 2.5, ExtremePanic},
	{"金融脱钩威胁", "传言中概股可能被迫退市，外资撤离A股。市场恐慌情绪蔓延。", 0.45, 2.7, ExtremePanic},
	{"全面对抗升级", "从贸易到科技到金融，冲突全面升级。类似冷战格局，避险资产暴涨。", 0.38, 3.0, ExtremePanic},
	{"汇率崩盘恐慌", "破7、破8的担忧蔓延，资本大规模外逃。央行紧急干预，但市场信心已崩。", 0.35, 3.2, ExtremePanic},
}

// 疫情黑天鹅主题（2020初突发疫情）
var PandemicEvents = []Event{
	{"疫情初现警报", "新型病毒出现，传染性未明。医药股异动，但大多数人还在过节...", 1.1, 0.9, MildOptimistic},
	{"封城!史无前例", "一线城市宣布封城，交通停运。市场震惊，但还没意识到严重性。", 0.8, 1.4, MildPessimistic},
	{"全球大流行", "WHO宣布全球大流行，多国进入紧急状态。A股开盘跌停上千只！", 0.4, 3.0, ExtremePanic},
	{"股市熔断", "美股连续熔断，全球股市暴跌。类似2008年，流动性危机爆发。", 0.3, 3.5, ExtremePanic},
	{"央行紧急救市", "各国央行联合放水，利率降到零。无限QE开启，股市V型反转。", 1.4, 0.55, ExtremeOptimistic},
	{"疫情概念炒作", "口罩、呼吸机、疫苗概念疯涨，20CM涨停天天见。但多少是真实业绩？", 1.35, 0.6, ExtremeOptimistic},
	{"经济停摆", "企业停工停产，失业率飙升。盈利暴跌但股市却在新高，估值泡沫隐现。", 0.85, 1.35, MildPessimistic},
	{"疫情反复", "变异株不断出现，防控政策摇摆。不确定性持续，市场剧烈波动。", 0.9, 1.5, Neutral},
	{"通胀高企", "放水后遗症：原材料暴涨，通胀失控。央行面临加息压力，股市承压。", 0.7, 1.8, MildPessimistic},
	{"政策转向", "宽松政策退出，流动性收紧。疫情红利消失，高估值股票暴跌。", 0.55, 2.3, ExtremePanic},
}

// 主题模式定义
type ThemeMode struct {
	Name        string
	Description string
	Events      []Event
	Difficulty  string
}

var AllThemes = []ThemeMode{
	{
		Name:        "经典模式",
		Description: "综合各种市场场景，适合新手学习市场规律",
		Events:      ClassicEvents,
		Difficulty:  "⭐⭐⭐⭐",
	},
	{
		Name:        "科技股狂潮",
		Description: "重现AI芯片概念炒作，体验科技泡沫的疯狂与崩塌",
		Events:      TechBoomEvents,
		Difficulty:  "⭐⭐⭐⭐⭐",
	},
	{
		Name:        "新能源泡沫",
		Description: "新能源车黄金时代，从万人追捧到产能过剩",
		Events:      NewEnergyEvents,
		Difficulty:  "⭐⭐⭐⭐",
	},
	{
		Name:        "贸易战惊魂",
		Description: "地缘政治博弈，感受供应链断裂的恐怖",
		Events:      TradeWarEvents,
		Difficulty:  "⭐⭐⭐⭐⭐",
	},
	{
		Name:        "疫情黑天鹅",
		Description: "重温2020年全球股灾，从熔断到史诗级反弹",
		Events:      PandemicEvents,
		Difficulty:  "⭐⭐⭐⭐⭐",
	},
}

// 当前使用的事件集（默认为经典模式）
var Events = ClassicEvents
var CurrentTheme = &AllThemes[0]

// 马尔科夫链转移概率矩阵
// 行：当前事件类别，列：下一个事件类别
// [ExtremeOptimistic, MildOptimistic, Neutral, MildPessimistic, ExtremePanic]
var MarkovTransitionMatrix = [][]float64{
	// 当前：极度乐观 → 容易继续群体狂欢涨停潮，也会有乐极生悲的突发跳水
	{0.25, 0.35, 0.20, 0.10, 0.10}, // ExtremeOptimistic
	// 当前：温和乐观 → 情绪发酵或回归平淡
	{0.20, 0.40, 0.25, 0.10, 0.05}, // MildOptimistic
	// 当前：中性 → 风平浪静，酝酿方向
	{0.10, 0.30, 0.40, 0.15, 0.05}, // Neutral
	// 当前：温和悲观 → 情绪转差，开始阴跌
	{0.05, 0.15, 0.30, 0.40, 0.10}, // MildPessimistic
	// 当前：极度恐慌 → 地狱难度：较高概率产生"连环踩踏跌停"，或少概率产生"深V反弹"救市
	{0.15, 0.05, 0.15, 0.30, 0.35}, // ExtremePanic
}

// 根据马尔科夫链选择下一个事件
func selectNextEventByMarkov(currentEvent Event, eventPool []Event) Event {
	currentCategory := currentEvent.Category
	transitionProbs := MarkovTransitionMatrix[currentCategory]

	// 按类别分组事件
	eventsByCategory := make(map[EventCategory][]Event)
	for _, event := range eventPool {
		eventsByCategory[event.Category] = append(eventsByCategory[event.Category], event)
	}

	// 使用轮盘赌算法选择下一个类别
	randValue := rand.Float64()
	cumProb := 0.0
	selectedCategory := Neutral // 默认中性

	for category, prob := range transitionProbs {
		cumProb += prob
		if randValue < cumProb {
			selectedCategory = EventCategory(category)
			break
		}
	}

	// 从选中的类别中随机选择一个事件
	categoryEvents := eventsByCategory[selectedCategory]
	if len(categoryEvents) == 0 {
		// 如果该类别没有事件，降级到全随机
		return eventPool[rand.Intn(len(eventPool))]
	}

	return categoryEvents[rand.Intn(len(categoryEvents))]
}

// 战绩记录系统
type GameRecord struct {
	Timestamp   string  `json:"timestamp"`    // 游戏时间
	Theme       string  `json:"theme"`        // 主题模式
	FinalProfit float64 `json:"final_profit"` // 最终收益率
	Grade       string  `json:"grade"`        // 评级
	Score       int     `json:"score"`        // 总分
	IsCrashed   bool    `json:"is_crashed"`   // 是否崩盘
	PlayerSold  bool    `json:"player_sold"`  // 玩家是否卖出
	SoldDay     int     `json:"sold_day"`     // 卖出日期(如果卖出)
	SoldSession string  `json:"sold_session"` // 卖出时段(如果卖出)
	ProfitScore int     `json:"profit_score"` // 收益分
	TimingScore int     `json:"timing_score"` // 时机分
	RiskScore   int     `json:"risk_score"`   // 风控分
}

// 成就系统
type Achievement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsUnlocked  bool   `json:"is_unlocked"`
	UnlockDate  string `json:"unlock_date"`
	Level       int    `json:"level"` // 0:未解锁, 1:铜, 2:银, 3:金
}

type AchievementData struct {
	Achievements    map[string]Achievement `json:"achievements"`
	TotalGames      int                    `json:"total_games"`
	TotalProfit     float64                `json:"total_profit"`
	MaxSingleProfit float64                `json:"max_single_profit"`
	CrashesEscaped  int                    `json:"crashes_escaped"`
	BustTimes       int                    `json:"bust_times"`
	TotalSRanks     int                    `json:"total_sranks"`
	ThemesMastered  map[string]bool        `json:"themes_mastered"` // 存储已获得A级及以上的主题
	Level           int                    `json:"level"`           // 交易员等级
	Experience      int                    `json:"experience"`      // 经验值
	IntelPoints     int                    `json:"intel_points"`    // 情报点数 (可用于偷看底牌)
}

type TradePoint struct {
	Day           int
	Session       string
	Action        string // "Buy" / "Sell"
	Price         float64
	Shares        int
	WhaleStatus   string // 当时游资的状态 (如 "出货", "洗盘", "Spoofing")
	MarketContext string // 当时大盘的状态 (如 "妖股狂热", "高位震荡")
}

type RecordHistory struct {
	Records []GameRecord `json:"records"`
}

const (
	recordFilePath      = "game_records.json"
	achievementFilePath = "achievements.json"
)

// 预设成就列表 (单次解锁类)
var DefaultAchievements = []Achievement{
	{ID: "FIRST_PROFIT", Name: "初露锋芒", Description: "第一次在游戏中获得正收益率"},
	{ID: "FIRST_S_RANK", Name: "股神降临", Description: "获得一次 S 级评价"},
	{ID: "CRASH_SURVIVOR", Name: "劫后余生", Description: "在崩盘中成功逃顶，空仓避险"},
	{ID: "HUNDRED_PERCENT", Name: "翻倍大神", Description: "单场收益率超过 100%"},
	{ID: "LEVERAGE_MASTER", Name: "刀尖舔血", Description: "使用 3 倍杠杆并获得 S 级评价"},
	{ID: "BANKRUPT", Name: "交学费", Description: "第一次爆仓（净资产归零）"},
	{ID: "DIAMOND_HANDS", Name: "铁头功", Description: "格局到底：从未卖出且最终获利超过 50%"},
	{ID: "THEME_MASTER", Name: "全能选手", Description: "在 3 个不同的主题模式中获得 A 级及以上评价"},
}

// 阶梯勋章配置
type TieredMedal struct {
	ID          string
	Name        string
	Description string
	Thresholds  []float64 // 铜、银、金的阈值
}

var TieredMedalConfigs = []TieredMedal{
	{ID: "MEDAL_TRADER", Name: "职业交易员", Description: "累计完成的游戏场次", Thresholds: []float64{5, 20, 50}},
	{ID: "MEDAL_SURVIVOR", Name: "避险专家", Description: "累计崩盘逃顶成功次数", Thresholds: []float64{1, 5, 15}},
	{ID: "MEDAL_PROFIT", Name: "盈利大师", Description: "累计获得的收益率总量", Thresholds: []float64{1.0, 5.0, 20.0}},
	{ID: "MEDAL_S_RANK", Name: "传奇操盘手", Description: "累计获得的 S 级评价次数", Thresholds: []float64{1, 5, 10}},
}

type AI struct {
	ID             string
	Type           string // "Whale", "Quant", "Retail", "Institution"
	SubType        string // "刺客", "打板", "新韭", "老散", "国家队"
	Name           string
	Shares         int
	Cost           float64
	Cash           float64 // 新增：可用于买入的现金
	TargetProfit   float64
	FearBasis      float64
	HasSold        bool
	LastOpinion    string
	HasInsiderInfo bool // 是否有内幕消息（提前知道下一个事件）

	// 当前回合的订单
	OrderType   string  // "Buy", "Sell", ""
	OrderShares int     // 计划卖出几股 / 计划买入几股 (买入时根据Cash算)
	OrderCash   float64 // 计划投入多少钱买
	MarginDebt  float64 // 新增：游资自身场外配资
	StatusFlag  string  // 新增：记录当前行为状态 "Spoofing", "ForcedLiquidation", "GridTrading" 等
}

// 自动交易策略
type AutoStrategy struct {
	Name           string  // 策略名称
	StopProfit     float64 // 止盈线（例如0.2表示+20%止盈）
	StopLoss       float64 // 止损线（例如-0.1表示-10%止损）
	RebuyThreshold float64 // 回买阈值（价格跌到多少时考虑买回，例如0.9表示跌10%）
	EnableRebuy    bool    // 是否启用回买
}

type ChronicleEntry struct {
	Day          int
	Session      string
	Price        float64
	Change       float64
	Event        string
	PlayerAction string // "买入", "卖出", "持仓", "空仓"
	WhaleAction  string // "吸筹", "拉升", "出货", "洗盘", "观望"
	MarketMood   string // "正常", "狂热", "恐慌"
}

type GameState struct {
	Day                   int
	Session               string // "早盘" 或 "尾盘"
	MaxDays               int
	GameMode              string // "fixed" 固定回合 或 "auto" 自动结束
	Price                 float64
	LastPrice             float64
	PlayerShares          int
	PlayerAvailableShares int     // 可用筹码 (T+1机制)
	PlayerFrozenShares    int     // 冻结筹码（今天新买入的，明天可用）
	PlayerCash            float64 // 当前现金
	PlayerAvgCost         float64 // 玩家持仓均价
	MarginDebt            float64 // 融资负债
	InitialAsset          float64 // 初始总资产（用于计算收益）
	IsMarginCalled        bool    // 是否被爆仓强平
	PlayerSoldDay         int
	PlayerSoldSession     string
	PriceHistory          []float64 // 价格历史，用于绘制走势图
	IsGameOver            bool
	IsCrashed             bool
	CurrentEvent          Event
	NextEvent             Event // 下一个事件（信息不对称）
	AIs                   []*AI
	TotalActiveShare      int
	CrashWarningLevel     int           // 崩盘预警等级 0=正常, 1=预警, 2=危险, 3=高危, 4=极限
	ConsecutiveFallDays   int           // 连续下跌天数
	TotalEscapedAIShares  int           // 已逃跑的AI持股总数
	TradeHistory          []string      // 交易历史记录
	AutoTradeStrategy     *AutoStrategy // 自动交易策略（nil表示手动）
	TotalMarketShares     int           // 流通盘总股数 (如 100,000)
	ConsecutiveGrowthDays int           // 新增：连涨天数用于触发妖股模型
	IsMonsterStock        bool          // 新增：妖股狂热状态

	// 博弈数据（零和撮合）
	TotalBuyDemandCash    float64 // 本回合全场涌入的总买单金额
	TotalBuyDemandShares  int     // 本回合全场需求多少股 (预估)
	TotalSellSupplyShares int     // 本回合全场砸出多少股
	FakeSellPressure      int     // 新增：游资假抛单
	FakeBuyPressure       int     // 新增：托单

	BuyPressure       int // 显示用的买压（原逻辑兼容）
	SellPressure      int // 显示用的卖压（原逻辑兼容）
	WhaleBuying       int
	WhaleSelling      int
	RetailBuying      int
	RetailSelling     int
	LastActionMessage string // 上一回合操作反馈（显示在仪表盘中）

	// 玩家在输入控制台提交的暂存订单
	PlayerOrderType   string // "Buy", "Sell", ""
	PlayerOrderShares int
	PlayerOrderCash   float64

	TradePoints       []TradePoint       // 上帝视角复盘记录点
	MarketLogs        []string           // 实时市场动态日志
	Chronicle         []ChronicleEntry   // 操盘编年史
	DailyFortune      string             // 今日运势
	InitialAIAssets   map[string]float64 // AI 初始资产
	IntelUsedThisTurn bool               // 本回合是否已使用情报
	DayHigh           float64            // 今日最高价
	DayLow            float64            // 今日最低价
	VolumeHistory     []int              // 成交量历史

	// 残局模式专属
	IsEndgameMode    bool             // 是否为残局模式
	EndgameScenario  *EndgameScenario // 残局场景信息
	EndgameStartTurn int              // 残局开始时的总回合数
}

// ===== 反身性分析系统 =====

type ReflexivityMetrics struct {
	SentimentPriceCorrelation float64 // -1 to 1: 情绪与价格相关性
	PanicSellIntensity        float64 // 0-10: 恐慌性抛售强度
	FOMOBuyIntensity          float64 // 0-10: FOMO追涨强度
	WhaleHerdingEffect        float64 // 0-1: 大资金羊群效应
	RetailChasingEffect       float64 // 0-1: 散户追涨效应
	FundamentalDisconnect     float64 // 0-10: 价格与基本面脱节度
	MomentumDecay             float64 // 动量衰减率

	// 历史窗口数据
	RecentPriceChanges    []float64
	RecentSentimentScores []float64
	RecentAIExitRatios    []float64
	RecentWhaleActions    []string
}

type ReflexivitySignal struct {
	Type        string  // "恐慌踩踏", "FOMO狂热", "大资金出逃", "散户接盘"
	Strength    float64 // 0-10
	Description string
	IsBullish   bool
}

// ===== 贝叶斯概率引擎 =====

type BayesianCrashModel struct {
	Prior           float64            // 先验崩盘概率
	Posterior       float64            // 后验崩盘概率
	EvidenceWeights map[string]float64 // 证据权重
	CurrentEvidence map[string]float64 // 当前证据值
	CrashProbByDay  map[int]float64    // 各天崩盘概率分布
}

type BayesianReversalModel struct {
	RecoveryProbability         float64 // 反转上涨概率
	ContinuedDeclineProbability float64 // 继续下跌概率
	OversoldIndicator           float64 // 超卖指标 (0-1)
	VolumeExhaustion            float64 // 成交量枯竭度 (0-1)
	WhaleReentrySignals         int     // 大资金回流信号数
}

type AIBehaviorModel struct {
	TraderID        string
	TraderType      string
	ProbNextSell    float64  // 下回合卖出概率
	ProbNextBuy     float64  // 下回合买入概率
	ProbHold        float64  // 下回合持有概率
	BehaviorHistory []string // 最近10次行为记录
	ProfitThreshold float64  // 预估止盈阈值
	FearThreshold   float64  // 预估恐慌阈值
}

type BayesianAnalysis struct {
	CrashProbability       BayesianCrashModel
	ExitTimingDistribution map[int]float64 // 天数 -> 最优离场概率
	BestExitDay            int
	BestExitConfidence     float64
	ReversalProbability    BayesianReversalModel
	AIBehaviorPrediction   map[string]*AIBehaviorModel
}

// ===== 高级分析输出 =====

type AdvancedAnalysis struct {
	BasicAdvice        StrategyAdvice
	ReflexivityMetrics ReflexivityMetrics
	ReflexivitySignals []ReflexivitySignal
	BayesianAnalysis   BayesianAnalysis
	OverallRiskScore   float64  // 0-100 综合风险
	KeyInsights        []string // 核心洞察 (最多3条)
}

// ===== 历史状态缓存 =====

type HistoricalStateCache struct {
	States []GameStateSnapshot
}

// ===== 分析缓存（性能优化） =====

type AnalysisCache struct {
	LastAnalysis   AdvancedAnalysis
	LastUpdateTurn int  // 上次更新的回合数（Day + Session）
	IsValid        bool // 缓存是否有效
}

// ===== 游戏统计系统 =====

type GameStats struct {
	TotalGames              int            // 总游戏场次
	CrashPredictionHits     int            // 崩盘预测命中次数
	CrashPredictionTotal    int            // 崩盘预测总次数
	ExitTimingErrors        []int          // 离场时机误差（实际离场日 - 建议离场日）
	ReflexivityPatternsSeen map[string]int // 见过的反身性模式
}

// ===== 游戏设置 =====

type GameSettings struct {
	ShowAdvancedAnalysis bool   // 是否显示高级分析
	AnalysisDetail       string // "simple", "detailed", "expert"
	ShowEducation        bool   // 是否显示教学提示
}

// ===== 残局模式 =====

type EndgameScenario struct {
	ID              string
	Name            string
	Description     string
	Difficulty      string // "简单", "中等", "困难", "地狱"
	StartDay        int
	StartSession    string
	InitialPrice    float64
	PriceChange     float64 // 累计涨跌幅（相对开盘价10元）
	PlayerCash      float64
	PlayerShares    int
	PlayerFrozen    int      // 冻结股数
	AIExitRatio     float64  // 已逃跑AI比例
	WhaleExitCount  int      // 已逃跑大资金数量
	CrashWarning    int      // 崩盘预警等级
	ConsecutiveFall int      // 连续下跌天数
	ConsecutiveRise int      // 连续上涨天数
	EventPreset     *Event   // 预设当前事件
	MarginDebt      float64  // 配资欠款
	IsMonsterStock  bool     // 是否妖股
	TargetProfit    float64  // 目标收益率（完美通关标准）
	TimeLimit       int      // 剩余时段数（0=不限制）
	TeachingPoints  []string // 教学要点
}

// 全局设置和统计
var GlobalSettings = GameSettings{
	ShowAdvancedAnalysis: true,
	AnalysisDetail:       "detailed",
	ShowEducation:        true,
}

var GlobalStats = GameStats{
	ReflexivityPatternsSeen: make(map[string]int),
}

var GlobalCache = AnalysisCache{
	IsValid: false,
}

// ===== 残局场景定义 =====

var EndgameScenarios = []EndgameScenario{
	{
		ID:              "endgame_1",
		Name:            "逃顶挑战",
		Description:     "Day 8早盘，股价已暴涨80%，70%大资金已出逃。崩盘概率极高，你能在雪崩前安全离场吗？",
		Difficulty:      "中等",
		StartDay:        8,
		StartSession:    "早盘",
		InitialPrice:    18.0,
		PriceChange:     0.80,
		PlayerCash:      20000,
		PlayerShares:    4444, // 约8万市值
		PlayerFrozen:    0,
		AIExitRatio:     0.70,
		WhaleExitCount:  5,
		CrashWarning:    4,
		ConsecutiveFall: 0,
		ConsecutiveRise: 4,
		EventPreset:     &Event{"监管层喊话风险", "市场监管部门发布风险提示，警告妖股炒作行为。但散户依然狂热。", 0.88, 2.2, MildPessimistic},
		MarginDebt:      0,
		IsMonsterStock:  true,
		TargetProfit:    0.60, // 60%收益率为完美通关
		TimeLimit:       4,    // 仅剩4个时段
		TeachingPoints:  []string{"识别顶部信号", "克服贪婪情绪", "反身性FOMO陷阱"},
	},
	{
		ID:              "endgame_2",
		Name:            "抄底陷阱",
		Description:     "Day 6午盘，股价已暴跌40%，超卖指标极高。是抄底良机还是接飞刀？",
		Difficulty:      "困难",
		StartDay:        6,
		StartSession:    "午盘",
		InitialPrice:    6.0,
		PriceChange:     -0.40,
		PlayerCash:      100000,
		PlayerShares:    0,
		PlayerFrozen:    0,
		AIExitRatio:     0.50,
		WhaleExitCount:  3,
		CrashWarning:    3,
		ConsecutiveFall: 3,
		ConsecutiveRise: 0,
		EventPreset:     &Event{"恐慌性抛售蔓延", "恐慌情绪主导市场，获利盘踩踏式出逃，成交量放大3倍。", 0.75, 2.8, ExtremePanic},
		MarginDebt:      0,
		IsMonsterStock:  false,
		TargetProfit:    0.20, // 20%收益率为完美通关（抄底成功）
		TimeLimit:       8,    // 剩余8个时段
		TeachingPoints:  []string{"区分反弹与反转", "成交量枯竭信号", "贝叶斯反转概率"},
	},
	{
		ID:              "endgame_3",
		Name:            "杠杆生死线",
		Description:     "Day 5午盘，你已使用3倍配资满仓，价格波动-15%。距离爆仓线仅5%，反身性恐慌已开始。",
		Difficulty:      "地狱",
		StartDay:        5,
		StartSession:    "午盘",
		InitialPrice:    8.5,
		PriceChange:     -0.15,
		PlayerCash:      0,
		PlayerShares:    35294, // 约30万市值
		PlayerFrozen:    0,
		AIExitRatio:     0.40,
		WhaleExitCount:  2,
		CrashWarning:    3,
		ConsecutiveFall: 2,
		ConsecutiveRise: 0,
		EventPreset:     &Event{"利空消息突发", "行业监管政策收紧，机构纷纷下调评级。", 0.82, 2.5, ExtremePanic},
		MarginDebt:      200000, // 20万配资欠款，爆仓线约8.1元
		IsMonsterStock:  false,
		TargetProfit:    0.0, // 只要不爆仓就算成功
		TimeLimit:       10,
		TeachingPoints:  []string{"杠杆风险管理", "止损纪律", "恐慌中的理性决策"},
	},
	{
		ID:              "endgame_4",
		Name:            "FOMO狂热",
		Description:     "Day 4尾盘，连续3天涨停，妖股氛围浓厚。散户疯狂追涨，但龙虎榜显示大资金正悄然出货。",
		Difficulty:      "中等",
		StartDay:        4,
		StartSession:    "尾盘",
		InitialPrice:    13.5,
		PriceChange:     0.35,
		PlayerCash:      50000,
		PlayerShares:    3704, // 约5万市值
		PlayerFrozen:    0,
		AIExitRatio:     0.30,
		WhaleExitCount:  2,
		CrashWarning:    2,
		ConsecutiveFall: 0,
		ConsecutiveRise: 3,
		EventPreset:     &Event{"散户追涨狂潮", "股吧、论坛一片看多声浪，散户跑步进场。但知名游资席位在悄然减仓。", 1.35, 0.85, ExtremeOptimistic},
		MarginDebt:      0,
		IsMonsterStock:  true,
		TargetProfit:    0.40, // 40%收益率（需在顶部前离场）
		TimeLimit:       12,
		TeachingPoints:  []string{"识破羊群效应", "区分散户与主力行为", "反身性FOMO信号"},
	},
	{
		ID:              "endgame_5",
		Name:            "最后48小时",
		Description:     "Day 14早盘，游戏即将结束（仅剩4个时段）。你满仓持股，价格稳定，但时间紧迫。",
		Difficulty:      "简单",
		StartDay:        14,
		StartSession:    "早盘",
		InitialPrice:    12.5,
		PriceChange:     0.25,
		PlayerCash:      10000,
		PlayerShares:    7200, // 约9万市值
		PlayerFrozen:    0,
		AIExitRatio:     0.25,
		WhaleExitCount:  1,
		CrashWarning:    1,
		ConsecutiveFall: 0,
		ConsecutiveRise: 1,
		EventPreset:     &Event{"市场情绪中性", "市场进入观望状态，多空博弈胶着。", 1.0, 1.0, Neutral},
		MarginDebt:      0,
		IsMonsterStock:  false,
		TargetProfit:    0.25, // 25%收益率
		TimeLimit:       4,
		TeachingPoints:  []string{"时间压力下的决策", "见好就收", "最优离场时机"},
	},
	{
		ID:              "endgame_6",
		Name:            "反身性连环踩踏",
		Description:     "Day 7早盘，恐慌踩踏已开始，价格每时段-10%，卖盘涌出。反身性下跌螺旋正在形成。",
		Difficulty:      "困难",
		StartDay:        7,
		StartSession:    "早盘",
		InitialPrice:    7.2,
		PriceChange:     -0.28,
		PlayerCash:      30000,
		PlayerShares:    9722, // 约7万市值
		PlayerFrozen:    0,
		AIExitRatio:     0.60,
		WhaleExitCount:  4,
		CrashWarning:    4,
		ConsecutiveFall: 3,
		ConsecutiveRise: 0,
		EventPreset:     &Event{"恐慌性踩踏", "恐慌情绪主导市场，卖盘如潮水涌出，反身性螺旋加速。", 0.70, 3.0, ExtremePanic},
		MarginDebt:      0,
		IsMonsterStock:  false,
		TargetProfit:    -0.10, // 仅亏损10%以内即为成功（保存实力）
		TimeLimit:       8,
		TeachingPoints:  []string{"识别反身性螺旋", "在崩溃中止损", "对抗恐慌情绪"},
	},
	{
		ID:              "endgame_7",
		Name:            "高开低走陷阱",
		Description:     "Day 3早盘，开盘暴涨12%诱多，但大资金正在出货。识破诱多陷阱，避免高位接盘。",
		Difficulty:      "中等",
		StartDay:        3,
		StartSession:    "早盘",
		InitialPrice:    11.2,
		PriceChange:     0.12,
		PlayerCash:      100000,
		PlayerShares:    0,
		PlayerFrozen:    0,
		AIExitRatio:     0.15,
		WhaleExitCount:  1,
		CrashWarning:    1,
		ConsecutiveFall: 0,
		ConsecutiveRise: 2,
		EventPreset:     &Event{"市场传闻利好", "市场传闻公司将有重大利好消息，但龙虎榜显示游资正在出货。", 1.25, 1.1, MildOptimistic},
		MarginDebt:      0,
		IsMonsterStock:  false,
		TargetProfit:    0.15, // 识破陷阱，避免损失
		TimeLimit:       6,
		TeachingPoints:  []string{"识破高开诱多", "量价背离分析", "龙虎榜解读"},
	},
	{
		ID:              "endgame_8",
		Name:            "尾盘跳水惊魂",
		Description:     "Day 9尾盘，前期稳定运行，突然有传闻引发尾盘跳水-8%。是恐慌性错杀还是真利空？",
		Difficulty:      "困难",
		StartDay:        9,
		StartSession:    "尾盘",
		InitialPrice:    9.2,
		PriceChange:     -0.08,
		PlayerCash:      15000,
		PlayerShares:    8696, // 约8万市值
		PlayerFrozen:    0,
		AIExitRatio:     0.35,
		WhaleExitCount:  2,
		CrashWarning:    2,
		ConsecutiveFall: 1,
		ConsecutiveRise: 0,
		EventPreset:     &Event{"尾盘突发利空传闻", "尾盘突然传出行业监管传闻，引发恐慌性抛售，真实性待确认。", 0.85, 2.2, MildPessimistic},
		MarginDebt:      0,
		IsMonsterStock:  false,
		TargetProfit:    0.10,
		TimeLimit:       5,
		TeachingPoints:  []string{"尾盘异动识别", "传闻真伪判断", "恐慌性错杀机会"},
	},
	{
		ID:              "endgame_9",
		Name:            "地天板反转",
		Description:     "Day 5午盘，早盘跌停-10%，但午盘突然放量拉升至-2%。是抄底良机还是诱多反弹？",
		Difficulty:      "地狱",
		StartDay:        5,
		StartSession:    "午盘",
		InitialPrice:    9.8,
		PriceChange:     -0.02,
		PlayerCash:      80000,
		PlayerShares:    0,
		PlayerFrozen:    0,
		AIExitRatio:     0.55,
		WhaleExitCount:  3,
		CrashWarning:    3,
		ConsecutiveFall: 2,
		ConsecutiveRise: 0,
		EventPreset:     &Event{"跌停板打开", "早盘跌停，但突然有神秘资金强势拉升，跌停板被打开。", 0.92, 2.0, MildPessimistic},
		MarginDebt:      0,
		IsMonsterStock:  false,
		TargetProfit:    0.30, // 成功把握反转
		TimeLimit:       6,
		TeachingPoints:  []string{"地天板形态识别", "成交量变化", "短线博弈技巧"},
	},
	{
		ID:              "endgame_10",
		Name:            "温水煮青蛙",
		Description:     "Day 10早盘，价格缓慢阴跌，每天-2%看似温和，但已累计跌20%。何时止损？",
		Difficulty:      "简单",
		StartDay:        10,
		StartSession:    "早盘",
		InitialPrice:    8.0,
		PriceChange:     -0.20,
		PlayerCash:      10000,
		PlayerShares:    11250, // 约9万市值
		PlayerFrozen:    0,
		AIExitRatio:     0.50,
		WhaleExitCount:  3,
		CrashWarning:    2,
		ConsecutiveFall: 5,
		ConsecutiveRise: 0,
		EventPreset:     &Event{"市场情绪低迷", "市场进入缓慢下行通道，成交量萎缩，多头无力反击。", 0.88, 1.6, MildPessimistic},
		MarginDebt:      0,
		IsMonsterStock:  false,
		TargetProfit:    -0.05, // 及时止损，控制损失
		TimeLimit:       5,
		TeachingPoints:  []string{"缓慢下跌止损", "趋势判断", "避免深套"},
	},
	{
		ID:              "endgame_11",
		Name:            "放量滞涨出货",
		Description:     "Day 6尾盘，成交量暴增3倍，但价格仅涨2%。典型的主力出货特征。",
		Difficulty:      "中等",
		StartDay:        6,
		StartSession:    "尾盘",
		InitialPrice:    14.3,
		PriceChange:     0.43,
		PlayerCash:      25000,
		PlayerShares:    5245, // 约7.5万市值
		PlayerFrozen:    0,
		AIExitRatio:     0.40,
		WhaleExitCount:  3,
		CrashWarning:    2,
		ConsecutiveFall: 0,
		ConsecutiveRise: 3,
		EventPreset:     &Event{"成交量异常放大", "成交量突然放大，但价格涨幅有限，龙虎榜显示知名游资减仓。", 1.12, 1.3, MildOptimistic},
		MarginDebt:      0,
		IsMonsterStock:  true,
		TargetProfit:    0.35,
		TimeLimit:       7,
		TeachingPoints:  []string{"量价背离", "主力出货识别", "高位减仓"},
	},
	{
		ID:              "endgame_12",
		Name:            "假突破诱多",
		Description:     "Day 7早盘，突破前高18元创新高，散户疯狂追涨。但这是真突破还是假突破？",
		Difficulty:      "困难",
		StartDay:        7,
		StartSession:    "早盘",
		InitialPrice:    18.5,
		PriceChange:     0.85,
		PlayerCash:      30000,
		PlayerShares:    3784, // 约7万市值
		PlayerFrozen:    0,
		AIExitRatio:     0.25,
		WhaleExitCount:  2,
		CrashWarning:    1,
		ConsecutiveFall: 0,
		ConsecutiveRise: 4,
		EventPreset:     &Event{"突破前高", "股价突破前期高点，技术派欢呼雀跃，但成交量未能有效放大。", 1.40, 0.9, ExtremeOptimistic},
		MarginDebt:      0,
		IsMonsterStock:  true,
		TargetProfit:    0.50,
		TimeLimit:       6,
		TeachingPoints:  []string{"真假突破判断", "成交量配合", "技术陷阱识别"},
	},
	{
		ID:              "endgame_13",
		Name:            "破位反抽陷阱",
		Description:     "Day 8午盘，跌破10元支撑位后反弹至10.5元。是止跌企稳还是诱多出货？",
		Difficulty:      "困难",
		StartDay:        8,
		StartSession:    "午盘",
		InitialPrice:    10.5,
		PriceChange:     0.05,
		PlayerCash:      50000,
		PlayerShares:    4762, // 约5万市值
		PlayerFrozen:    0,
		AIExitRatio:     0.45,
		WhaleExitCount:  3,
		CrashWarning:    3,
		ConsecutiveFall: 2,
		ConsecutiveRise: 0,
		EventPreset:     &Event{"破位后反弹", "跌破重要支撑位后出现反弹，成交量萎缩，技术上呈现弱反弹特征。", 0.95, 1.8, Neutral},
		MarginDebt:      0,
		IsMonsterStock:  false,
		TargetProfit:    0.05,
		TimeLimit:       6,
		TeachingPoints:  []string{"破位后反抽", "支撑位有效性", "弱反弹识别"},
	},
	{
		ID:              "endgame_14",
		Name:            "T+0日内套利",
		Description:     "Day 11早盘，价格在12-13元区间震荡，波动率大。考验你的日内高抛低吸能力。",
		Difficulty:      "简单",
		StartDay:        11,
		StartSession:    "早盘",
		InitialPrice:    12.5,
		PriceChange:     0.25,
		PlayerCash:      50000,
		PlayerShares:    4000, // 约5万市值，可T+0操作
		PlayerFrozen:    0,
		AIExitRatio:     0.20,
		WhaleExitCount:  1,
		CrashWarning:    1,
		ConsecutiveFall: 0,
		ConsecutiveRise: 1,
		EventPreset:     &Event{"震荡行情", "市场进入震荡整理，日内波动加大，适合短线交易。", 1.05, 1.2, Neutral},
		MarginDebt:      0,
		IsMonsterStock:  false,
		TargetProfit:    0.35, // 通过T+0套利增强收益
		TimeLimit:       4,
		TeachingPoints:  []string{"日内波段操作", "高抛低吸", "T+0策略"},
	},
	{
		ID:              "endgame_15",
		Name:            "题材炒作末期",
		Description:     "Day 12尾盘，ST摘帽题材已炒作5天，累计涨幅70%。题材即将退潮，何时离场？",
		Difficulty:      "中等",
		StartDay:        12,
		StartSession:    "尾盘",
		InitialPrice:    17.0,
		PriceChange:     0.70,
		PlayerCash:      15000,
		PlayerShares:    5000, // 约8.5万市值
		PlayerFrozen:    0,
		AIExitRatio:     0.35,
		WhaleExitCount:  2,
		CrashWarning:    2,
		ConsecutiveFall: 0,
		ConsecutiveRise: 5,
		EventPreset:     &Event{"题材炒作降温", "ST摘帽题材持续多日，市场开始出现分歧，部分资金开始撤离。", 1.18, 1.4, MildOptimistic},
		MarginDebt:      0,
		IsMonsterStock:  true,
		TargetProfit:    0.60,
		TimeLimit:       3,
		TeachingPoints:  []string{"题材炒作周期", "题材退潮信号", "投机时机把握"},
	},
	{
		ID:              "endgame_16",
		Name:            "机构对倒识破",
		Description:     "Day 5午盘，盘面出现大量对倒单，成交活跃但价格波动小。识破机构对倒行为。",
		Difficulty:      "地狱",
		StartDay:        5,
		StartSession:    "午盘",
		InitialPrice:    11.8,
		PriceChange:     0.18,
		PlayerCash:      60000,
		PlayerShares:    3390, // 约4万市值
		PlayerFrozen:    0,
		AIExitRatio:     0.30,
		WhaleExitCount:  2,
		CrashWarning:    2,
		ConsecutiveFall: 0,
		ConsecutiveRise: 2,
		EventPreset:     &Event{"盘面异常活跃", "盘中出现大量对倒盘，成交量放大但价格窄幅震荡，疑似机构对倒吸引跟风盘。", 1.08, 1.5, Neutral},
		MarginDebt:      0,
		IsMonsterStock:  false,
		TargetProfit:    0.25,
		TimeLimit:       8,
		TeachingPoints:  []string{"对倒盘识别", "盘口语言解读", "主力行为分析"},
	},
}

// ===== 教学系统 =====

var SeenSignals = make(map[string]bool) // 记录已看过的信号类型

type GameStateSnapshot struct {
	Day                 int
	Session             string
	Price               float64
	EscapedAIRatio      float64
	WhaleEscapeCount    int
	CrashWarningLevel   int
	ConsecutiveFallDays int
	EventSentiment      float64
	EventFear           float64
	PlayerAction        string
	TotalBuyPressure    int
	TotalSellPressure   int
}

// 获取交易员头衔
func getTraderTitle(level int) string {
	titles := []string{"", "新手韭菜", "入门散户", "资深股民", "职业交易员", "短线猎人", "趋势专家", "传奇作手", "市场主宰者"}
	if level < len(titles) {
		return titles[level]
	}
	return "无上庄家"
}

// 添加市场日志
func (s *GameState) AddLog(msg string) {
	s.MarketLogs = append(s.MarketLogs, msg)
	if len(s.MarketLogs) > 8 {
		s.MarketLogs = s.MarketLogs[1:]
	}
}

// 叙述性日志输出，让回合结算有动态滚动感
func (s *GameState) SpeakLog(msg string) {
	s.AddLog(msg)
	fmt.Printf("  %s%s%s\n", Gray, msg, Reset)
	time.Sleep(150 * time.Millisecond)
}

// 记录编年史
func (s *GameState) AddChronicle() {
	playerAction := "空仓"
	if s.PlayerShares > 0 {
		playerAction = "持仓"
	}
	// 特殊动作覆盖
	if s.PlayerOrderType == "Buy" {
		playerAction = "买入"
	} else if s.PlayerOrderType == "Sell" {
		playerAction = "卖出"
	}

	whaleAction := "观望"
	for _, ai := range s.AIs {
		if ai.Type == "Whale" && ai.StatusFlag != "" {
			switch ai.StatusFlag {
			case "Spoofing":
				whaleAction = "诱多"
			case "Shakeout":
				whaleAction = "洗盘"
			case "Bailout":
				whaleAction = "救市"
			case "ForcedLiquidation":
				whaleAction = "出局"
			}
		}
	}
	if s.WhaleBuying > s.WhaleSelling {
		whaleAction = "拉升"
	} else if s.WhaleSelling > s.WhaleBuying {
		whaleAction = "砸盘"
	}

	mood := "正常"
	if s.IsMonsterStock {
		mood = "狂热"
	} else if s.CrashWarningLevel >= 3 {
		mood = "恐慌"
	}

	change := (s.Price - s.LastPrice) / s.LastPrice

	s.Chronicle = append(s.Chronicle, ChronicleEntry{
		Day:          s.Day,
		Session:      s.Session,
		Price:        s.Price,
		Change:       change,
		Event:        s.CurrentEvent.Title,
		PlayerAction: playerAction,
		WhaleAction:  whaleAction,
		MarketMood:   mood,
	})
}

// 加载历史战绩
func loadRecordHistory() RecordHistory {
	history := RecordHistory{Records: []GameRecord{}}

	file, err := os.ReadFile(recordFilePath)
	if err != nil {
		// 文件不存在或无法读取,返回空历史
		return history
	}

	json.Unmarshal(file, &history)
	return history
}

// 保存历史战绩
func saveGameRecord(record GameRecord) error {
	history := loadRecordHistory()

	// 将新记录添加到最前面
	history.Records = append([]GameRecord{record}, history.Records...)

	// 最多保留50条记录
	if len(history.Records) > 50 {
		history.Records = history.Records[:50]
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(recordFilePath, data, 0644)
}

// 显示历史战绩
func showRecordHistory(reader *bufio.Reader) {
	history := loadRecordHistory()

	if len(history.Records) == 0 {
		fmt.Println(Yellow + "\n暂无历史战绩记录。开始第一场游戏吧！\n" + Reset)
		fmt.Print("按回车键继续...")
		waitEnter(reader)
		return
	}

	fmt.Print("\033[H\033[2J") // 清屏
	fmt.Println(Purple + "╔══════════════════════════════════════════════════════╗")
	fmt.Println("║             历史战绩 - 你的成长轨迹                ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝" + Reset)

	// 统计数据
	totalGames := len(history.Records)
	sGrades := 0
	aGrades := 0
	avgProfit := 0.0
	crashedEscaped := 0 // 崩盘时成功逃顶的次数
	totalCrashed := 0   // 崩盘总次数

	for _, record := range history.Records {
		if record.Grade == "S" {
			sGrades++
		} else if record.Grade == "A" {
			aGrades++
		}
		avgProfit += record.FinalProfit
		if record.IsCrashed {
			totalCrashed++
			if record.PlayerSold {
				crashedEscaped++
			}
		}
	}
	avgProfit /= float64(totalGames)

	// 显示统计信息
	fmt.Println("\n" + Cyan + "📊 总体统计" + Reset)
	fmt.Printf("  总场次: %d   S级: %d   A级: %d   平均收益: %+.1f%%\n", totalGames, sGrades, aGrades, avgProfit*100)
	if totalCrashed > 0 {
		escapedRate := float64(crashedEscaped) / float64(totalCrashed) * 100
		fmt.Printf("  崩盘逃顶成功率: %.0f%% (%d/%d)\n", escapedRate, crashedEscaped, totalCrashed)
	}

	// 显示最近10场战绩（从最新到最旧）
	fmt.Println("\n" + Cyan + "📜 最近战绩 (最多显示10场，从新到旧)" + Reset)
	fmt.Println(Cyan + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + Reset)

	displayCount := min(10, len(history.Records))
	// 从最新的记录开始显示（索引0是最新的）
	for i := 0; i < displayCount; i++ {
		record := history.Records[i]

		// 评级颜色
		gradeColor := Reset
		switch record.Grade {
		case "S":
			gradeColor = Purple
		case "A":
			gradeColor = Blue
		case "B":
			gradeColor = Cyan
		case "C":
			gradeColor = Yellow
		case "D", "F":
			gradeColor = Red
		}

		// 收益颜色
		profitColor := Green
		if record.FinalProfit < 0 {
			profitColor = Red
		}

		// 崩盘标志
		crashIcon := ""
		if record.IsCrashed {
			if record.PlayerSold {
				crashIcon = " 🎉逃顶"
			} else {
				crashIcon = " ☠️被埋"
			}
		}

		fmt.Printf("\n%s[%d] %s%s "+gradeColor+"[%s级 %d分]"+Reset+"%s\n",
			Yellow, i+1, Reset, record.Timestamp, record.Grade, record.Score, crashIcon)
		fmt.Printf("    主题: %s  |  收益: %s%+.1f%%%s  |  ",
			record.Theme, profitColor, record.FinalProfit*100, Reset)

		if record.PlayerSold {
			fmt.Printf("第%d天%s离场\n", record.SoldDay, record.SoldSession)
		} else {
			fmt.Printf("格局到底\n")
		}

		fmt.Printf("    详细评分: 💰%d分 ⏱️%d分 🛡️%d分\n",
			record.ProfitScore, record.TimingScore, record.RiskScore)
	}

	fmt.Println("\n" + Cyan + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + Reset)
	fmt.Print("\n按回车键返回...")
	waitEnter(reader)
}

// 加载成就数据
func loadAchievementData() AchievementData {
	data := AchievementData{
		Achievements:   make(map[string]Achievement),
		ThemesMastered: make(map[string]bool),
	}

	// 初始化默认成就
	for _, ach := range DefaultAchievements {
		data.Achievements[ach.ID] = ach
	}
	// 初始化阶梯勋章
	for _, cfg := range TieredMedalConfigs {
		data.Achievements[cfg.ID] = Achievement{
			ID:          cfg.ID,
			Name:        cfg.Name,
			Description: cfg.Description,
			Level:       0,
		}
	}

	file, err := os.ReadFile(achievementFilePath)
	if err == nil {
		json.Unmarshal(file, &data)
		if data.Achievements == nil {
			data.Achievements = make(map[string]Achievement)
		}
		// 确保所有配置都在 map 中
		for _, ach := range DefaultAchievements {
			if _, exists := data.Achievements[ach.ID]; !exists {
				data.Achievements[ach.ID] = ach
			}
		}
		for _, cfg := range TieredMedalConfigs {
			if _, exists := data.Achievements[cfg.ID]; !exists {
				data.Achievements[cfg.ID] = Achievement{
					ID:          cfg.ID,
					Name:        cfg.Name,
					Description: cfg.Description,
					Level:       0,
				}
			}
		}
		if data.ThemesMastered == nil {
			data.ThemesMastered = make(map[string]bool)
		}
		if data.Level == 0 {
			data.Level = 1
		}
	} else {
		data.Level = 1
	}

	return data
}

// 保存成就数据
func saveAchievementData(data AchievementData) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(achievementFilePath, jsonData, 0644)
}

// 检测成就
func checkAchievements(state *GameState, rank PlayerRank) []Achievement {
	data := loadAchievementData()
	newUnlocked := []Achievement{}
	now := time.Now().Format("2006-01-02 15:04:05")

	finalAsset := float64(state.PlayerShares)*state.Price + state.PlayerCash - state.MarginDebt
	finalProfit := (finalAsset - state.InitialAsset) / state.InitialAsset

	// 更新统计数据
	data.TotalGames++
	data.TotalProfit += finalProfit
	if finalProfit > data.MaxSingleProfit {
		data.MaxSingleProfit = finalProfit
	}
	if rank.Grade == "S" {
		data.TotalSRanks++
	}

	// 经验值计算
	xpGained := rank.Score * 5
	if finalProfit > 0 {
		xpGained += int(finalProfit * 1000)
	}
	if state.IsCrashed && state.PlayerShares == 0 {
		xpGained += 200 // 逃顶奖
	}
	data.Experience += xpGained

	// 升级逻辑
	xpRequired := data.Level * 1000
	if data.Experience >= xpRequired {
		data.Level++
		data.IntelPoints += 3 // 升级奖励情报点
		fmt.Printf("\n%s🚀 【交易等级提升】 %s -> %s %s(获得3点情报点)%s\n", Yellow, getTraderTitle(data.Level-1), getTraderTitle(data.Level), Cyan, Reset)
	}
	fmt.Printf("\n%s📈 本场经验: +%d | 等级进度: %d/%d%s\n", Cyan, xpGained, data.Experience, xpRequired, Reset)

	// 辅助函数：尝试解锁单次成就
	unlock := func(id string) {
		ach := data.Achievements[id]
		if !ach.IsUnlocked {
			ach.IsUnlocked = true
			ach.UnlockDate = now
			data.Achievements[id] = ach
			newUnlocked = append(newUnlocked, ach)
		}
	}

	// 1. 初露锋芒
	if finalProfit > 0 {
		unlock("FIRST_PROFIT")
	}

	// 2. 股神降临
	if rank.Grade == "S" {
		unlock("FIRST_S_RANK")
	}

	// 3. 劫后余生
	if state.IsCrashed && state.PlayerShares == 0 && state.PlayerSoldDay > 0 {
		unlock("CRASH_SURVIVOR")
		data.CrashesEscaped++
	}

	// 4. 翻倍大神
	if finalProfit >= 1.0 {
		unlock("HUNDRED_PERCENT")
	}

	// 5. 刀尖舔血
	if state.MarginDebt > 0 && rank.Grade == "S" {
		unlock("LEVERAGE_MASTER")
	}

	// 6. 交学费
	if state.IsMarginCalled || (state.IsCrashed && finalAsset <= 0) {
		unlock("BANKRUPT")
		data.BustTimes++
	}

	// 7. 铁头功
	if len(state.TradeHistory) == 0 && finalProfit >= 0.5 {
		unlock("DIAMOND_HANDS")
	}

	// 8. 全能选手
	if rank.Grade == "S" || rank.Grade == "A" {
		data.ThemesMastered[CurrentTheme.Name] = true
		if len(data.ThemesMastered) >= 3 {
			unlock("THEME_MASTER")
		}
	}

	// --- 阶梯勋章逻辑 ---
	checkMedalTier := func(id string, currentVal float64, thresholds []float64) {
		ach := data.Achievements[id]
		oldLevel := ach.Level
		newLevel := 0
		for i, t := range thresholds {
			if currentVal >= t {
				newLevel = i + 1
			}
		}
		if newLevel > oldLevel {
			ach.Level = newLevel
			ach.IsUnlocked = true
			ach.UnlockDate = now
			// 为了通知显示，我们临时修改名称
			tierNames := []string{"", "铜", "银", "金"}
			notifyAch := ach
			notifyAch.Name = fmt.Sprintf("%s勋章 [%s级]", ach.Name, tierNames[newLevel])
			newUnlocked = append(newUnlocked, notifyAch)
			data.Achievements[id] = ach
		}
	}

	for _, cfg := range TieredMedalConfigs {
		val := 0.0
		switch cfg.ID {
		case "MEDAL_TRADER":
			val = float64(data.TotalGames)
		case "MEDAL_SURVIVOR":
			val = float64(data.CrashesEscaped)
		case "MEDAL_PROFIT":
			val = data.TotalProfit
		case "MEDAL_S_RANK":
			val = float64(data.TotalSRanks)
		}
		checkMedalTier(cfg.ID, val, cfg.Thresholds)
	}

	saveAchievementData(data)
	return newUnlocked
}

// 定义勋章颜色
const (
	Gray   = "\033[90m"
	Gold   = "\033[33m"
	Silver = "\033[37m"
	Bronze = "\033[31m"
)

// 显示成就墙
func showAchievements(reader *bufio.Reader) {
	data := loadAchievementData()

	fmt.Print("\033[H\033[2J") // 清屏
	fmt.Println(Purple + "╔══════════════════════════════════════════════════════╗")
	fmt.Println("║             荣誉殿堂 - 勋章墙                       ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝" + Reset)

	// 统计信息
	unlockedCount := 0
	for _, ach := range data.Achievements {
		if ach.IsUnlocked {
			unlockedCount++
		}
	}

	fmt.Println("\n" + Cyan + "🏆 总体勋章概览" + Reset)
	fmt.Printf("  已解锁成就/勋章: %d/%d   总场次: %d   总S级: %d\n",
		unlockedCount, len(data.Achievements), data.TotalGames, data.TotalSRanks)
	fmt.Printf("  累计收益: %+.1f%%   单场最高: %+.1f%%\n", data.TotalProfit*100, data.MaxSingleProfit*100)
	fmt.Printf("  崩盘逃顶: %d 次   爆仓次数: %d 次\n", data.CrashesEscaped, data.BustTimes)

	fmt.Println("\n" + Cyan + "🎖️ 阶梯勋章 (进步永无止境)" + Reset)
	fmt.Println(Cyan + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + Reset)

	for _, cfg := range TieredMedalConfigs {
		ach := data.Achievements[cfg.ID]
		color := Gray
		tierName := "未获得"
		switch ach.Level {
		case 1:
			color = Bronze
			tierName = "铜级"
		case 2:
			color = Silver
			tierName = "银级"
		case 3:
			color = Gold
			tierName = "金级"
		}

		currentVal := 0.0
		unit := ""
		switch cfg.ID {
		case "MEDAL_TRADER":
			currentVal = float64(data.TotalGames)
			unit = "场"
		case "MEDAL_SURVIVOR":
			currentVal = float64(data.CrashesEscaped)
			unit = "次"
		case "MEDAL_PROFIT":
			currentVal = data.TotalProfit * 100
			unit = "%"
		case "MEDAL_S_RANK":
			currentVal = float64(data.TotalSRanks)
			unit = "次"
		}

		// 计算下一级进度
		nextThreshold := 0.0
		if ach.Level < 3 {
			nextThreshold = cfg.Thresholds[ach.Level]
			if cfg.ID == "MEDAL_PROFIT" {
				nextThreshold *= 100
			}
		}

		fmt.Printf("\n%s● [%s] %s%s  %s\n", color, tierName, ach.Name, Reset, ach.Description)
		if ach.Level < 3 {
			progress := currentVal / nextThreshold * 10
			if progress > 10 {
				progress = 10
			}
			bar := strings.Repeat("█", int(progress)) + strings.Repeat("░", 10-int(progress))
			fmt.Printf("    进度: [%s] %.0f/%.0f %s\n", bar, currentVal, nextThreshold, unit)
		} else {
			fmt.Printf("    %s✨ 已达成最高荣誉：金级勋章 ✨%s\n", Gold, Reset)
		}
	}

	fmt.Println("\n" + Cyan + "📜 基础成就 (里程碑)" + Reset)
	fmt.Println(Cyan + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + Reset)

	// 按预设顺序显示
	for _, defaultAch := range DefaultAchievements {
		ach := data.Achievements[defaultAch.ID]
		if ach.IsUnlocked {
			fmt.Printf("\n%s🌟 [%s]%s  %s\n", Gold, ach.Name, Reset, ach.Description)
			fmt.Printf("    %s解锁时间: %s%s\n", Green, ach.UnlockDate, Reset)
		} else {
			fmt.Printf("\n%s🔒 [???]%s  %s\n", Gray, Reset, ach.Description)
			fmt.Printf("    %s尚未解锁%s\n", Gray, Reset)
		}
	}

	fmt.Println("\n" + Cyan + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + Reset)
	fmt.Print("\n按回车键返回...")
	waitEnter(reader)
}

// 设置菜单
func showSettingsMenu(reader *bufio.Reader) {
	for {
		fmt.Print("\033[H\033[2J") // 清屏
		fmt.Println(Cyan + "╔══════════════════════════════════════════════════════╗")
		fmt.Println("║                ⚙️  游戏设置                         ║")
		fmt.Println("╚══════════════════════════════════════════════════════╝" + Reset)
		fmt.Println()

		// 显示当前设置
		fmt.Println(Yellow + "【当前设置】" + Reset)
		fmt.Printf("  [1] 高级分析显示: %s\n", boolToOnOff(GlobalSettings.ShowAdvancedAnalysis))
		fmt.Printf("  [2] 分析详细度: %s (%s)\n", GlobalSettings.AnalysisDetail, detailLevelDesc(GlobalSettings.AnalysisDetail))
		fmt.Printf("  [3] 教学提示: %s\n", boolToOnOff(GlobalSettings.ShowEducation))
		fmt.Println()
		fmt.Println(Green + "  [0] 返回主菜单" + Reset)
		fmt.Println()
		fmt.Print(Green + "请选择要修改的设置项 (0-3): " + Reset)

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "0":
			return // 返回主菜单
		case "1":
			GlobalSettings.ShowAdvancedAnalysis = !GlobalSettings.ShowAdvancedAnalysis
			if GlobalSettings.ShowAdvancedAnalysis {
				fmt.Println(Green + "\n✅ 已开启高级分析显示" + Reset)
			} else {
				fmt.Println(Yellow + "\n⚠️ 已关闭高级分析显示" + Reset)
			}
			time.Sleep(800 * time.Millisecond)
		case "2":
			// 切换详细度: simple -> detailed -> expert -> simple
			switch GlobalSettings.AnalysisDetail {
			case "simple":
				GlobalSettings.AnalysisDetail = "detailed"
			case "detailed":
				GlobalSettings.AnalysisDetail = "expert"
			case "expert":
				GlobalSettings.AnalysisDetail = "simple"
			default:
				GlobalSettings.AnalysisDetail = "detailed"
			}
			fmt.Printf(Green+"\n✅ 分析详细度已切换为: %s (%s)"+Reset+"\n", GlobalSettings.AnalysisDetail, detailLevelDesc(GlobalSettings.AnalysisDetail))
			time.Sleep(800 * time.Millisecond)
		case "3":
			GlobalSettings.ShowEducation = !GlobalSettings.ShowEducation
			if GlobalSettings.ShowEducation {
				fmt.Println(Green + "\n✅ 已开启教学提示" + Reset)
			} else {
				fmt.Println(Yellow + "\n⚠️ 已关闭教学提示" + Reset)
			}
			time.Sleep(800 * time.Millisecond)
		}
	}
}

// 辅助函数：布尔值转开关显示
func boolToOnOff(b bool) string {
	if b {
		return Green + "开启" + Reset
	}
	return Gray + "关闭" + Reset
}

// 辅助函数：详细度描述
func detailLevelDesc(level string) string {
	switch level {
	case "simple":
		return "简洁版"
	case "detailed":
		return "详细版"
	case "expert":
		return "专家版 (最详细)"
	default:
		return "未知"
	}
}

// 辅助函数：等待回车（更健壮地处理 \r, \n, \r\n）
func waitEnter(reader *bufio.Reader) {
	if reader == nil {
		reader = bufio.NewReader(os.Stdin)
	}
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return
		}
		// 只要读到 \r (13) 或 \n (10) 就视作回车
		if b == 13 || b == 10 {
			// 智能清理：如果缓冲区里紧跟着另一个换行符(如 \r 后面的 \n)，一并清理
			// 使用 Buffered() 确保 Peek 不会触发新的阻塞读取
			for reader.Buffered() > 0 {
				peek, _ := reader.Peek(1)
				if len(peek) > 0 && (peek[0] == 10 || peek[0] == 13) {
					_, _ = reader.ReadByte()
				} else {
					break
				}
			}
			return
		}
	}
}

// 残局场景选择
func selectEndgameScenario(reader *bufio.Reader) *EndgameScenario {
	fmt.Print("\033[H\033[2J") // 清屏
	fmt.Println(Purple + "╔══════════════════════════════════════════════════════╗")
	fmt.Println("║              🎯 残局挑战选择                        ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝" + Reset)
	fmt.Println()
	fmt.Println(Cyan + "残局模式：从关键市场情境开始，专注训练特定决策能力" + Reset)
	fmt.Println()

	// 动态创建菜单项
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
			Label: scenario.Name,
			Description: fmt.Sprintf("%s %s | 目标: +%.0f%% | %d时段 | %s",
				diffIcon, scenario.Difficulty,
				scenario.TargetProfit*100,
				scenario.TimeLimit,
				strings.Join(scenario.TeachingPoints, ", ")),
			Value: fmt.Sprintf("%d", i+1),
		}
	}

	// 添加返回选项
	items[len(EndgameScenarios)] = MenuItem{
		Label:       "返回主菜单",
		Description: "",
		Value:       "0",
	}

	menu := NewInteractiveMenu("请选择残局场景：", items)
	menu.Reader = reader
	value, _ := menu.Show()

	choice := 0
	fmt.Sscanf(value, "%d", &choice)

	if choice == 0 || value == "" {
		return nil // 返回主菜单
	}

	if choice < 1 || choice > len(EndgameScenarios) {
		return nil
	}

	selected := &EndgameScenarios[choice-1]

	// 显示场景详情
	fmt.Print("\033[H\033[2J")
	fmt.Println(Purple + "╔══════════════════════════════════════════════════════╗")
	fmt.Printf("║          残局挑战: %s%-30s║\n", Yellow, selected.Name+Reset)
	fmt.Println("╚══════════════════════════════════════════════════════╝" + Reset)
	fmt.Println()
	fmt.Printf(Cyan+"【场景描述】"+Reset+"\n%s\n\n", selected.Description)
	fmt.Printf(Yellow + "【起始条件】" + Reset + "\n")
	fmt.Printf("  Day %d %s | 当前股价: $%.2f (累计%+.0f%%)\n",
		selected.StartDay, selected.StartSession, selected.InitialPrice, selected.PriceChange*100)
	fmt.Printf("  玩家持仓: %d股 + $%.0f现金\n", selected.PlayerShares, selected.PlayerCash)
	if selected.MarginDebt > 0 {
		fmt.Printf("  %s配资欠款: $%.0f%s (爆仓风险！)\n", Red, selected.MarginDebt, Reset)
	}
	fmt.Printf("  AI逃跑率: %.0f%% | 大资金出逃: %d个\n", selected.AIExitRatio*100, selected.WhaleExitCount)
	fmt.Printf("  崩盘预警等级: %s%d/5%s\n", Red, selected.CrashWarning, Reset)
	fmt.Println()
	fmt.Printf(Green+"【通关目标】"+Reset+" 收益率达到 %s%+.0f%%%s\n", Yellow, selected.TargetProfit*100, Reset)
	fmt.Printf(Cyan+"【教学重点】"+Reset+" %s\n\n", strings.Join(selected.TeachingPoints, ", "))

	fmt.Print(Green + "按回车键开始挑战..." + Reset)
	waitEnter(reader)

	return selected
}

func selectGameMode(reader *bufio.Reader) (string, int, *AutoStrategy) {
	fmt.Print("\033[H\033[2J") // 清屏
	fmt.Println(Purple + "╔══════════════════════════════════════════════════════╗")
	fmt.Println("║       妖股搏杀 - 游戏模式选择                       ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝" + Reset)
	fmt.Println()

	items := []MenuItem{
		{
			Label:       "手动模式",
			Description: "最多15天，崩盘结束",
			Value:       "1",
			Icon:        "🎮",
		},
		{
			Label:       "手动60回合",
			Description: "固定60回合，无崩盘",
			Value:       "2",
			Icon:        "🎯",
		},
		{
			Label:       "AI策略测试",
			Description: "自动止盈止损",
			Value:       "3",
			Icon:        "🤖",
		},
	}

	menu := NewInteractiveMenu("游戏模式", items)
	menu.Reader = reader
	value, _ := menu.Show()

	choice := 1
	fmt.Sscanf(value, "%d", &choice)

	switch choice {
	case 2:
		// 60回合固定模式
		return "fixed", 30, nil // 30天 = 60回合
	case 3:
		// 自动策略模拟
		strategy := selectAutoStrategy(reader)
		return "auto", 15, strategy
	default:
		// 默认：自动结束模式
		return "auto", 15, nil
	}
}

func selectAutoStrategy(reader *bufio.Reader) *AutoStrategy {
	fmt.Print("\033[H\033[2J") // 清屏
	fmt.Println(Purple + "╔══════════════════════════════════════════════════════╗")
	fmt.Println("║       自动策略选择 - 0-1博弈测试                    ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝" + Reset)
	fmt.Println()

	strategies := []AutoStrategy{
		{"激进策略", 0.15, -0.05, 0.85, false},
		{"稳健策略", 0.25, -0.08, 0.9, false},
		{"波段策略", 0.20, -0.10, 0.85, true},
		{"格局策略", 0.40, -0.15, 0.80, true},
	}

	items := []MenuItem{
		{
			Label:       "激进策略",
			Description: "止盈+15%, 止损-5%, 不回买",
			Value:       "1",
		},
		{
			Label:       "稳健策略",
			Description: "止盈+25%, 止损-8%, 不回买",
			Value:       "2",
		},
		{
			Label:       "波段策略",
			Description: "止盈+20%, 止损-10%, 跌15%回买",
			Value:       "3",
		},
		{
			Label:       "格局策略",
			Description: "止盈+40%, 止损-15%, 跌20%回买",
			Value:       "4",
		},
	}

	menu := NewInteractiveMenu("预设策略：", items)
	menu.Reader = reader
	value, _ := menu.Show()

	choice := 1
	fmt.Sscanf(value, "%d", &choice)

	if choice < 1 || choice > 4 {
		choice = 1
	}

	strategy := strategies[choice-1]
	fmt.Printf("\n"+Cyan+"已选择: %s\n"+Reset, strategy.Name)
	fmt.Printf("止盈: %+.0f%%  止损: %.0f%%", strategy.StopProfit*100, strategy.StopLoss*100)
	if strategy.EnableRebuy {
		fmt.Printf("  回买: 价格跌%.0f%%时\n", (1-strategy.RebuyThreshold)*100)
	} else {
		fmt.Printf("  不回买\n")
	}
	fmt.Print("\n按回车键确认...")
	waitEnter(reader)

	return &strategy
}

func selectTheme(reader *bufio.Reader) *ThemeMode {
	fmt.Print("\033[H\033[2J") // 清屏
	fmt.Println(Purple + "╔══════════════════════════════════════════════════════╗")
	fmt.Println("║       妖股搏杀 - 主题模式选择                       ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝" + Reset)
	fmt.Println()

	// 动态创建菜单项
	items := make([]MenuItem, len(AllThemes))
	for i, theme := range AllThemes {
		items[i] = MenuItem{
			Label:       theme.Name,
			Description: fmt.Sprintf("%s | 难度: %s", theme.Description, theme.Difficulty),
			Value:       fmt.Sprintf("%d", i+1),
		}
	}

	menu := NewInteractiveMenu("请选择你想体验的主题模式：", items)
	menu.Reader = reader
	value, _ := menu.Show()

	choice := 1 // 默认经典模式
	fmt.Sscanf(value, "%d", &choice)

	if choice < 1 || choice > len(AllThemes) {
		choice = 1
	}

	selectedTheme := &AllThemes[choice-1]

	fmt.Printf("\r\n"+Cyan+"你选择了: %s%s%s\n"+Reset, Yellow, selectedTheme.Name, Cyan)
	fmt.Printf("%s\r\n", selectedTheme.Description)
	fmt.Print("\r\n按回车键确认开始...")
	waitEnter(reader)

	return selectedTheme
}

func renderSplash() {
	fmt.Print("\033[H\033[2J")
	logo := `
  %s███████╗ ██████╗  ██████╗ ██╗  ██╗     ██████╗  █████╗ ███╗   ███╗███████╗
  ██╔════╝██╔═══██╗██╔════╝ ██║ ██╔╝     ██╔════╝ ██╔══██╗████╗ ████║██╔════╝
  ███████╗██║   ██║██║      █████╔╝      ██║  ███╗███████║██╔████╔██║█████╗  
  ╚════██║██║   ██║██║      ██╔═██╗      ██║   ██║██╔══██║██║╚██╔╝██║██╔══╝  
  ███████║╚██████╔╝╚██████╗ ██║  ██╗     ╚██████╔╝██║  ██║██║ ╚═╝ ██║███████╗
  ╚══════╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═╝      ╚═════╝ ╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝%s
                                                                             
           %s>> 妖股搏杀：从韭菜到庄家的修罗场 <<%s
           %s[ 模拟器版本 v2.5 - 勋章/编年史版 ]%s
	`
	fmt.Printf(logo, Red, Reset, Yellow, Reset, Cyan, Reset)
	fmt.Println("\n  " + strings.Repeat("—", 65))
}

func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Print("\033[H\033[2J\033[3J") // 程序刚启动时大清屏
	reader := bufio.NewReader(os.Stdin)

	// 选择游戏模式
	gameMode, maxDays, autoStrategy := selectGameMode(reader)

	// 选择主题模式
	CurrentTheme = selectTheme(reader)
	Events = CurrentTheme.Events

	// 询问是否查看历史战绩或成就
	var selectedEndgame *EndgameScenario
	for {
		// 每次返回主菜单时大清屏，确保干净
		fmt.Print("\033[H\033[2J\033[3J")
		renderSplash()

		// 使用交互式菜单
		items := []MenuItem{
			{Label: "查看历史战绩", Value: "1"},
			{Label: "查看成就墙 (荣誉殿堂)", Value: "2"},
			{Label: "设置", Value: "3", Icon: "⚙️"},
			{Label: "残局挑战", Value: "4", Icon: "🎯"},
			{Label: "AI求解器 (动态规划寻优)", Value: "5", Icon: "🤖"},
			{Label: "直接开始游戏", Value: "6", Icon: "🎮"},
		}

		menu := NewInteractiveMenu("欢迎来到妖股搏杀！请选择操作：", items)
		menu.Reader = reader
		value, _ := menu.Show()

		if value == "" {
			// ESC退出
			break
		}

		switch value {
		case "1":
			showRecordHistory(reader)
		case "2":
			showAchievements(reader)
		case "3":
			showSettingsMenu(reader)
		case "4":
			selectedEndgame = selectEndgameScenario(reader)
			if selectedEndgame != nil {
				break // 进入残局模式
			}
		case "5":
			RunSolverCLI() // 运行AI求解器
		default:
			break // 开始普通游戏
		}

		if value == "4" && selectedEndgame != nil {
			break
		}
		if value == "6" {
			break
		}
	}

	var state *GameState
	if selectedEndgame != nil {
		state = initEndgameState(selectedEndgame)
	} else {
		state = initGame(gameMode, maxDays, autoStrategy)
	}

	// 初始化历史状态缓存
	histCache := &HistoricalStateCache{
		States: []GameStateSnapshot{},
	}

	renderSplash()
	fmt.Printf("       %s当前选择主题：%s%s%s\n", Cyan, Yellow, CurrentTheme.Name, Reset)
	fmt.Printf("       %s【地狱修罗难度：全周期零和博弈】%s\n", Red, Reset)
	fmt.Println("  " + strings.Repeat("—", 65))
	fmt.Println(Red + "  【财富掠夺逻辑已开启】" + Reset)
	fmt.Println("  在这个市场中，没有新资金进入。每一分钱的赚取都意味着另一个对手盘的破产。")
	fmt.Println("  1. 股价由买卖盘实时撮合决定，没有绝对的对错，只有对手盘。")
	fmt.Println("  2. 注意观察游资(Whale)的动作，他们可能是你的轿夫，也可能是屠夫。")
	fmt.Println("  3. 勋章墙记录你的长期成长，编年史记录你的每一次贪婪。")
	fmt.Print("\n  按回车键，踏入修罗场...")
	waitEnter(reader)
	fmt.Println("\n" + Yellow + strings.Repeat("=", 25) + " 🚀 游戏正式开始 " + strings.Repeat("=", 25) + Reset + "\n")

	// 翻牌动画标记：第一天早盘不做动画，之后每次新的早盘做

	// 翻牌动画标记：第一天早盘不做动画，之后每次新的早盘做
	prevSession := ""
	for !state.IsGameOver {
		generateOpinions(state)
		// 新的一天早盘时，做翻牌动画
		if state.Session == "早盘" && state.Day > 1 && prevSession == "尾盘" {
			renderEventCard(state.CurrentEvent, true)
		}
		prevSession = state.Session
		renderFrame(state, histCache)

		if state.Session == "盘中上午" || state.Session == "盘中下午" {
			state.PlayerOrderType = ""
			state.PlayerOrderShares = 0
			state.PlayerOrderCash = 0

			fmt.Printf("\n"+Purple+"  🕰️ 【%d天-%s】 %s === AI 正在全速撮合处理中 === %s"+Reset+"\n", state.Day, state.Session, Yellow, Reset)
			time.Sleep(1000 * time.Millisecond)
			processTurn(state)
			fmt.Println("\n" + Cyan + strings.Repeat("-", 20) + " 本时段处理完毕 " + strings.Repeat("-", 20) + Reset)

			// 捕获当前状态快照
			snapshot := captureGameStateSnapshot(state)
			histCache.States = append(histCache.States, snapshot)
			if len(histCache.States) > 20 {
				histCache.States = histCache.States[1:] // 保留最近20个
			}
			continue
		}

		// 自动策略模式 vs 手动模式
		if state.AutoTradeStrategy != nil {
			// 自动执行策略
			executeAutoStrategy(state)
			time.Sleep(500 * time.Millisecond) // 暂停让玩家看到
		} else {
			// 手动模式 - 使用交互式菜单
			var menuItems []MenuItem
			data := loadAchievementData()

			// 选项1: 观望
			menuItems = append(menuItems, MenuItem{
				Label: "观望",
				Value: "1",
			})

			// 卖出选项 (2-4)
			if state.PlayerAvailableShares > 0 {
				sellThird := state.PlayerAvailableShares / 3
				sellHalf := state.PlayerAvailableShares / 2
				menuItems = append(menuItems, MenuItem{
					Label:       "减仓1/3",
					Description: fmt.Sprintf("%d股", sellThird),
					Value:       "2",
				})
				menuItems = append(menuItems, MenuItem{
					Label:       "卖出一半",
					Description: fmt.Sprintf("%d股", sellHalf),
					Value:       "3",
				})
				menuItems = append(menuItems, MenuItem{
					Label:       "清仓",
					Description: fmt.Sprintf("%d股", state.PlayerAvailableShares),
					Value:       "4",
					Icon:        "💣",
				})
			}

			// 买入选项 (5-7)
			canBuyShares := int(state.PlayerCash / state.Price)
			if canBuyShares > 0 {
				buyThird := canBuyShares / 3
				buyHalf := canBuyShares / 2
				menuItems = append(menuItems, MenuItem{
					Label:       "建仓1/3",
					Description: fmt.Sprintf("~%d股", buyThird),
					Value:       "5",
				})
				menuItems = append(menuItems, MenuItem{
					Label:       "半仓买入",
					Description: fmt.Sprintf("~%d股", buyHalf),
					Value:       "6",
				})
				menuItems = append(menuItems, MenuItem{
					Label:       "满仓",
					Description: fmt.Sprintf("~%d股", canBuyShares),
					Value:       "7",
					Icon:        "🚀",
				})
			}

			// 配资选项 (8)
			if state.MarginDebt == 0 && (state.PlayerCash > 0 || state.PlayerShares > 0) {
				menuItems = append(menuItems, MenuItem{
					Label:       "3倍杠杆",
					Description: "配资满仓",
					Value:       "8",
					Icon:        "🎲",
				})
			}

			// 情报选项 (9)
			if data.IntelPoints > 0 && !state.IntelUsedThisTurn {
				menuItems = append(menuItems, MenuItem{
					Label:       "内幕情报",
					Description: fmt.Sprintf("剩余%d点", data.IntelPoints),
					Value:       "9",
					Icon:        "🕵️",
				})
			}

			// 显示菜单并获取选择
			menu := NewInteractiveMenu("【交易指令台】", menuItems)
			menu.Reader = reader
			input, _ := menu.Show()

			// 如果用户取消（ESC），默认为观望
			if input == "" {
				input = "1"
			}

			if input == "9" && data.IntelPoints > 0 && !state.IntelUsedThisTurn {
				data.IntelPoints--
				saveAchievementData(data)
				state.IntelUsedThisTurn = true
				state.LastActionMessage = fmt.Sprintf("🕵️ 【绝密内幕】 明天预测事件: %s (%s)", state.NextEvent.Title, state.NextEvent.Desc)
				state.AddLog(fmt.Sprintf("%s 🕵️ 你动用关系获取了明天情报: %s%s", Cyan, state.NextEvent.Title, Reset))
				renderFrame(state, histCache) // 刷新一次以显示日志
				continue                      // 继续本回合操作
			}

			switch input {
			case "2", "3", "4":
				if state.PlayerAvailableShares > 0 {
					var sellShares int
					if input == "2" {
						sellShares = state.PlayerAvailableShares / 3
					} else if input == "3" {
						sellShares = state.PlayerAvailableShares / 2
					} else {
						sellShares = state.PlayerAvailableShares
					}
					if sellShares == 0 {
						sellShares = state.PlayerAvailableShares
					}
					state.PlayerOrderType = "Sell"
					state.PlayerOrderShares = sellShares
					state.LastActionMessage = fmt.Sprintf("📝 挂单：全服限价卖出 %d 股 (等待盘后撮合...)", sellShares)
					// 预扣除，防止重复下单
					state.PlayerAvailableShares -= sellShares
				} else {
					state.LastActionMessage = "❌ 无可用筹码（T+1冻结中）"
					state.PlayerOrderType = ""
				}
			case "5", "6", "7":
				canBuyShares := int(state.PlayerCash / state.Price)
				if canBuyShares > 0 {
					var buyShares int
					if input == "5" {
						buyShares = canBuyShares / 3
					} else if input == "6" {
						buyShares = canBuyShares / 2
					} else {
						buyShares = canBuyShares
					}
					if buyShares == 0 {
						buyShares = canBuyShares
					}
					buyAmount := float64(buyShares) * state.Price
					state.PlayerOrderType = "Buy"
					state.PlayerOrderCash = buyAmount
					state.LastActionMessage = fmt.Sprintf("📝 挂单：全仓限价买入约 %d 股，预算 $%.2f (等待盘后撮合...)", buyShares, buyAmount)
					// 预先冻结资金
					state.PlayerCash -= buyAmount
				} else {
					state.LastActionMessage = "❌ 现金不足，无法买入"
					state.PlayerOrderType = ""
				}
			case "8":
				if state.MarginDebt == 0 && (state.PlayerCash > 0 || state.PlayerShares > 0) {
					netAsset := float64(state.PlayerShares)*state.Price + state.PlayerCash
					borrowAmount := netAsset * 2.0
					state.MarginDebt += borrowAmount
					state.PlayerCash += borrowAmount
					canBuyLeveraged := int(state.PlayerCash / state.Price)
					if canBuyLeveraged > 0 {
						buyAmount := float64(canBuyLeveraged) * state.Price
						state.PlayerOrderType = "Buy"
						state.PlayerOrderCash = buyAmount
						state.LastActionMessage = fmt.Sprintf("🎲 借入场外高利贷$%.2f，挂单梭哈买入中！(等待撮合)", borrowAmount)
						state.PlayerCash -= buyAmount
					}
				} else {
					state.LastActionMessage = "❌ 已有配资负债，无法再次开杠杆"
					state.PlayerOrderType = ""
				}
			default:
				state.LastActionMessage = "👁  你选择默默挂机观望…"
				state.PlayerOrderType = ""
			}
		}

		fmt.Printf("\n"+Blue+"  ⚡ 【%d天-%s】 %s === 玩家指令已提交，系统正在撮合 === %s"+Reset+"\n", state.Day, state.Session, Yellow, Reset)
		processTurn(state)
		fmt.Println("\n" + Green + strings.Repeat("=", 65) + Reset)

		// 捕获当前状态快照
		snapshot := captureGameStateSnapshot(state)
		histCache.States = append(histCache.States, snapshot)
		if len(histCache.States) > 20 {
			histCache.States = histCache.States[1:] // 保留最近20个
		}
	}

	renderGameOver(state, reader)

	// 生成复盘报告
	generatePostGameReport(state, histCache)
}

// 辅助函数：计算总回合数（用于残局进度追踪）
func calculateTurnNumberHelper(day int, session string) int {
	// 每天4个session（集合竞价、早盘、午盘、尾盘）
	baseTurns := (day - 1) * 4
	sessionTurn := sessionToInt(session)
	return baseTurns + sessionTurn
}

// 初始化残局模式游戏状态
func initEndgameState(scenario *EndgameScenario) *GameState {
	// 基础状态
	state := &GameState{
		Day:                   scenario.StartDay,
		Session:               scenario.StartSession,
		MaxDays:               15, // 固定15天
		GameMode:              "手动",
		Price:                 scenario.InitialPrice,
		LastPrice:             scenario.InitialPrice * 0.98, // 模拟前一价格
		PlayerShares:          scenario.PlayerShares,
		PlayerAvailableShares: scenario.PlayerShares,
		PlayerFrozenShares:    scenario.PlayerFrozen,
		PlayerCash:            scenario.PlayerCash,
		PlayerAvgCost:         10.0, // 假设初始成本为开盘价
		MarginDebt:            scenario.MarginDebt,
		IsGameOver:            false,
		IsCrashed:             false,
		IsMarginCalled:        false,
		CurrentEvent:          *scenario.EventPreset,
		NextEvent:             Events[rand.Intn(len(Events))],
		BuyPressure:           0,
		SellPressure:          0,
		RetailBuying:          0,
		RetailSelling:         0,
		CrashWarningLevel:     scenario.CrashWarning,
		ConsecutiveFallDays:   scenario.ConsecutiveFall,
		ConsecutiveGrowthDays: scenario.ConsecutiveRise,
		IsMonsterStock:        scenario.IsMonsterStock,
		IntelUsedThisTurn:     false,
		AutoTradeStrategy:     nil,
		MarketLogs:            []string{},
		Chronicle:             []ChronicleEntry{},
		PlayerSoldDay:         0,
		PlayerSoldSession:     "",
		PlayerOrderType:       "",
		PlayerOrderShares:     0,
		PlayerOrderCash:       0,
		LastActionMessage:     fmt.Sprintf("🎯 残局挑战：%s", scenario.Name),
		IsEndgameMode:         true,
		EndgameScenario:       scenario,
		EndgameStartTurn:      calculateTurnNumberHelper(scenario.StartDay, scenario.StartSession),
	}

	// 初始化AI交易员（根据场景调整）
	totalAICount := 20
	whaleCount := 7
	quantCount := 6
	retailCount := 7

	state.AIs = []*AI{}

	// 创建AI并标记部分已出逃
	aiEscapedCount := int(float64(totalAICount) * scenario.AIExitRatio)
	whaleEscapedCount := scenario.WhaleExitCount

	// Whales
	for i := 0; i < whaleCount; i++ {
		ai := &AI{
			ID:           fmt.Sprintf("Whale-%d", i+1),
			Type:         "Whale",
			SubType:      []string{"刺客", "打板", "埋伏"}[rand.Intn(3)],
			Shares:       10000 + rand.Intn(5000),
			Cost:         10.0,
			Cash:         0,
			HasSold:      i < whaleEscapedCount, // 前N个已出逃
			TargetProfit: 0.30 + float64(rand.Intn(20))*0.01,
			FearBasis:    2.0 + float64(rand.Intn(10))*0.1,
		}
		if ai.HasSold {
			ai.Cash = float64(ai.Shares) * scenario.InitialPrice * 0.9 // 假设以稍低价格卖出
			ai.Shares = 0
		}
		state.AIs = append(state.AIs, ai)
	}

	// Quants
	remainingEscaped := aiEscapedCount - whaleEscapedCount
	for i := 0; i < quantCount; i++ {
		ai := &AI{
			ID:           fmt.Sprintf("Quant-%d", i+1),
			Type:         "Quant",
			SubType:      []string{"网格", "趋势", "对冲"}[rand.Intn(3)],
			Shares:       5000 + rand.Intn(3000),
			Cost:         10.0,
			Cash:         0,
			HasSold:      i < remainingEscaped,
			TargetProfit: 0.05 + float64(rand.Intn(10))*0.01,
			FearBasis:    1.8,
		}
		if ai.HasSold {
			ai.Cash = float64(ai.Shares) * scenario.InitialPrice * 0.92
			ai.Shares = 0
			remainingEscaped--
		}
		state.AIs = append(state.AIs, ai)
	}

	// Retail
	for i := 0; i < retailCount; i++ {
		ai := &AI{
			ID:           fmt.Sprintf("Retail-%d", i+1),
			Type:         "Retail",
			SubType:      []string{"新韭", "老韭", "佛系"}[rand.Intn(3)],
			Shares:       1000 + rand.Intn(2000),
			Cost:         10.0,
			Cash:         0,
			HasSold:      remainingEscaped > 0 && i < remainingEscaped,
			TargetProfit: 0.50,
			FearBasis:    2.5,
		}
		if ai.HasSold {
			ai.Cash = float64(ai.Shares) * scenario.InitialPrice * 0.88
			ai.Shares = 0
			remainingEscaped--
		}
		state.AIs = append(state.AIs, ai)
	}

	// 添加国家队（不会逃跑）
	state.AIs = append(state.AIs, &AI{
		ID:           "Institution-1",
		Type:         "Institution",
		SubType:      "国家队",
		Shares:       20000,
		Cost:         10.0,
		Cash:         100000,
		HasSold:      false,
		TargetProfit: 0.10,
		FearBasis:    5.0,
	})

	// 初始化缺失的关键字段
	state.TotalMarketShares = 100000 // 全服流通盘
	state.InitialAsset = float64(state.PlayerShares)*state.Price + state.PlayerCash - state.MarginDebt

	// 初始化历史数据（模拟之前的数据）
	state.PriceHistory = []float64{scenario.InitialPrice}
	state.VolumeHistory = []int{5000} // 初始成交量
	state.DayLow = scenario.InitialPrice
	state.DayHigh = scenario.InitialPrice
	state.TradeHistory = []string{}

	// 初始化博弈数据
	state.TotalBuyDemandCash = 0
	state.TotalBuyDemandShares = 0
	state.TotalSellSupplyShares = 0
	state.InitialAIAssets = make(map[string]float64)

	// 计算AI初始资产
	for _, ai := range state.AIs {
		state.InitialAIAssets[ai.ID] = ai.Cash + float64(ai.Shares)*state.Price
	}

	// 添加初始日志
	state.AddLog(fmt.Sprintf("🎯 残局模式：%s", scenario.Name))
	state.AddLog(fmt.Sprintf("起始: Day %d %s, 股价 $%.2f", scenario.StartDay, scenario.StartSession, scenario.InitialPrice))
	state.AddLog(fmt.Sprintf("目标收益率: %+.0f%%", scenario.TargetProfit*100))

	return state
}

func initGame(gameMode string, maxDays int, autoStrategy *AutoStrategy) *GameState {
	initialShares := 4000 // 占总流通盘的 4%
	initialPrice := 10.0

	// 随机挑选一个中性或温和的事件作为开局，防止一上来就是黑天鹅熔断体验太差
	var startEvent Event
	for {
		startEvent = Events[rand.Intn(len(Events))]
		if startEvent.Category == Neutral || startEvent.Category == MildOptimistic || startEvent.Category == MildPessimistic {
			break
		}
	}

	state := &GameState{
		Day:                   1,
		Session:               "早盘",
		MaxDays:               maxDays,
		GameMode:              gameMode,
		Price:                 initialPrice,
		LastPrice:             initialPrice,
		PlayerShares:          initialShares,
		PlayerAvailableShares: initialShares, // 初始即可全卖
		PlayerFrozenShares:    0,
		PlayerCash:            0.0, // 初始全仓股票，无现金
		PlayerAvgCost:         initialPrice,
		MarginDebt:            0.0, // 无融资负债
		IsMarginCalled:        false,
		InitialAsset:          float64(initialShares) * initialPrice, // 记录初始总资产
		AIs:                   make([]*AI, 0),
		CurrentEvent:          startEvent,
		NextEvent:             selectNextEventByMarkov(startEvent, Events), // 使用马尔科夫链预生成下一个事件
		TradeHistory:          []string{},
		PriceHistory:          []float64{initialPrice}, // 初始化一条K线基础记录
		AutoTradeStrategy:     autoStrategy,
		TotalMarketShares:     100000, // 全服只有 10 万股，零和博弈
		// 博弈数据初始化
		TotalBuyDemandCash:    0,
		TotalBuyDemandShares:  0,
		TotalSellSupplyShares: 0,
		BuyPressure:           0,
		SellPressure:          0,
		WhaleBuying:           0,
		WhaleSelling:          0,
		RetailBuying:          0,
		RetailSelling:         0,
		DailyFortune:          "宜：空仓观望，忌：满仓梭哈", // 初始默认
		InitialAIAssets:       make(map[string]float64),
	}

	// 真正的零和市场生态 (Total = 100,000 股)
	// Player = 4000 股
	// createAIsAdvanced 参数：类型, 子类型, 名称, 数量, 单个持股, 备用资金, 成本, 目标倍数, 恐慌系数

	// Whale (总 27000 股)
	createAIsAdvanced(state, "Whale", "刺客", "江浙刺客游资", 1, 15000, 600000.0, 9.5, 1.4, 0.6)
	createAIsAdvanced(state, "Whale", "机构", "内资长线底仓", 1, 12000, 1200000.0, 8.5, 2.0, 0.1) // 极度铁头

	// Quant (总 26000 股)
	createAIsAdvanced(state, "Quant", "打板", "高频打板量化", 1, 14000, 300000.0, 10.1, 1.2, 1.5)
	createAIsAdvanced(state, "Quant", "网格", "幻方网格量化", 1, 12000, 350000.0, 10.0, 1.15, 1.7)

	// Retail (总 43000 股)
	createAIsAdvanced(state, "Retail", "新韭", "发财梦新韭菜", 2, 8000, 80000.0, 10.8, 1.15, 2.8)
	createAIsAdvanced(state, "Retail", "老散", "装死死扛老散", 3, 9000, 15000.0, 13.0, 1.05, 0.2) // 强迫症装死

	// 隐藏国家队 (0 股，500万备用金准备托底)
	createAIsAdvanced(state, "Institution", "国家队", "平准托底基金", 1, 0, 5000000.0, 10.0, 1.05, 0.0)

	// Whale有内幕消息优势（提前知道下一个事件），刺客有场外配资负债
	for _, ai := range state.AIs {
		state.InitialAIAssets[ai.ID] = ai.Cash + float64(ai.Shares)*state.Price
		if ai.Type == "Whale" {
			ai.HasInsiderInfo = true
			if ai.SubType == "刺客" {
				ai.MarginDebt = 300000.0 // 刺客借了30万高利贷拉盘
			}
		}
	}

	return state
}

// 执行自动交易策略
func executeAutoStrategy(state *GameState) {
	strategy := state.AutoTradeStrategy
	// 计算总资产 (市值 + 现金 - 负债)
	totalAsset := float64(state.PlayerShares)*state.Price + state.PlayerCash - state.MarginDebt
	profitRatio := (totalAsset - state.InitialAsset) / state.InitialAsset

	// 持仓时检查止盈止损
	if state.PlayerShares > 0 {
		shouldSell := false
		reason := ""

		// 检查止盈
		if profitRatio >= strategy.StopProfit {
			shouldSell = true
			reason = fmt.Sprintf("触发止盈 (收益%.1f%% >= 目标%.0f%%)", profitRatio*100, strategy.StopProfit*100)
		}

		// 检查止损
		if profitRatio <= strategy.StopLoss {
			shouldSell = true
			reason = fmt.Sprintf("触发止损 (亏损%.1f%% <= 止损%.0f%%)", profitRatio*100, strategy.StopLoss*100)
		}

		if shouldSell {
			if state.PlayerAvailableShares > 0 {
				sellShares := state.PlayerAvailableShares
				sellAmount := float64(sellShares) * state.Price
				state.PlayerCash += sellAmount
				state.PlayerAvailableShares = 0
				state.PlayerShares -= sellShares
				tradeLog := fmt.Sprintf("第%d天%s: [策略] 卖出%d股 @$%.2f，获得现金$%.2f [%s]",
					state.Day, state.Session, sellShares, state.Price, sellAmount, reason)
				state.TradeHistory = append(state.TradeHistory, tradeLog)
				state.LastActionMessage = fmt.Sprintf("🤖 [策略] %s → 卖出 %d 股，入账 $%.2f", reason, sellShares, sellAmount)
				state.PlayerSoldDay = state.Day
				state.PlayerSoldSession = state.Session
			} else {
				state.LastActionMessage = fmt.Sprintf("🤖 [策略] %s → 但筹码 T+1 冻结，无法执行！", reason)
			}
		} else {
			state.LastActionMessage = fmt.Sprintf("🤖 [策略] 持仓中，当前收益 %+.1f%%", profitRatio*100)
		}
	} else {
		if strategy.EnableRebuy {
			priceChangeRatio := (state.Price - 10.0) / 10.0
			if priceChangeRatio <= (strategy.RebuyThreshold - 1.0) {
				canBuyShares := int(state.PlayerCash / state.Price)
				if canBuyShares > 0 {
					buyAmount := float64(canBuyShares) * state.Price
					state.PlayerFrozenShares += canBuyShares
					state.PlayerShares += canBuyShares
					state.PlayerCash -= buyAmount
					tradeLog := fmt.Sprintf("第%d天%s: [策略] 买入%d股 @$%.2f，花费$%.2f [触发回买]",
						state.Day, state.Session, canBuyShares, state.Price, buyAmount)
					state.TradeHistory = append(state.TradeHistory, tradeLog)
					state.LastActionMessage = fmt.Sprintf("🤖 [策略] 价格跌 %.1f%%，触发回买 %d 股（T+1冻结）", priceChangeRatio*100, canBuyShares)
				}
			} else {
				state.LastActionMessage = fmt.Sprintf("🤖 [策略] 空仓观望，等待跌破 %.0f%%", (strategy.RebuyThreshold-1.0)*100)
			}
		} else {
			state.LastActionMessage = "🤖 [策略] 空仓，不回买"
		}
	}
}

func createAIsAdvanced(state *GameState, aiType string, subType string, namePrefix string, count int, baseShares int, baseCash float64, baseCost float64, baseTarget float64, baseFear float64) {
	for i := 0; i < count; i++ {
		sharesMut := int(float64(baseShares) * (0.8 + rand.Float64()*0.4))
		cashMut := baseCash * (0.8 + rand.Float64()*0.4)
		targetMut := baseTarget * (0.9 + rand.Float64()*0.2)
		costMut := baseCost * (0.97 + rand.Float64()*0.06) // 成本有±3%的波动

		name := namePrefix
		if count > 1 {
			name = fmt.Sprintf("%s-%d", namePrefix, i+1)
		}

		state.AIs = append(state.AIs, &AI{
			ID:           fmt.Sprintf("%s-%d", aiType, len(state.AIs)),
			Type:         aiType,
			SubType:      subType,
			Name:         name,
			Shares:       sharesMut,
			Cash:         cashMut,
			Cost:         costMut,
			TargetProfit: targetMut,
			FearBasis:    baseFear,
			HasSold:      false,
			OrderType:    "",
			OrderShares:  0,
			OrderCash:    0,
		})
	}
}

func generateOpinions(state *GameState) {
	for _, ai := range state.AIs {
		avatar := "🥬"
		if ai.Type == "Whale" {
			avatar = "🐋"
		} else if ai.Type == "Quant" {
			avatar = "🤖"
		}

		if ai.HasSold {
			if ai.Type == "Whale" {
				ai.LastOpinion = avatar + " \"利润落袋，今天去洗浴中心。\""
			} else if ai.Type == "Quant" {
				ai.LastOpinion = avatar + " [SYSTEM] 已按均价出清，进入休眠休假模式。"
			} else {
				ai.LastOpinion = avatar + " \"呼...终于斩仓割肉了，这辈子不玩股票了。\""
			}
			continue
		}

		profitRatio := state.Price / ai.Cost
		eventFear := state.CurrentEvent.FearModifier

		if ai.Type == "Whale" {
			// Whale有内幕消息，可以提前反应下一个事件
			if ai.HasInsiderInfo {
				nextEventSentiment := state.NextEvent.Sentiment
				nextEventFear := state.NextEvent.FearModifier

				if nextEventFear > 2.0 {
					ai.LastOpinion = avatar + " 【内幕】:\"风向不对！明天必有核按钮，先挂满跌停跑路！\""
				} else if nextEventSentiment > 1.3 {
					ai.LastOpinion = avatar + " 【内幕】:\"听说明天有重磅利好，悄悄吃货，拉起来再出给小散。\""
				} else if eventFear > 1.8 {
					ai.LastOpinion = avatar + " 咬牙切齿:\"这雷太大！一键断头铡出货！\""
				} else if profitRatio > ai.TargetProfit*0.95 {
					ai.LastOpinion = avatar + " 冷酷无情:\"轿子太重了。拉高诱多，准备派发筹码。\""
				} else {
					ai.LastOpinion = avatar + " 沉吟思考:\"下方没承接，只能做个T降降成本了...\""
				}
			} else {
				if eventFear > 1.8 {
					ai.LastOpinion = avatar + " 咬牙切齿:\"这雷太大！一键断头铡出货！\""
				} else if profitRatio > ai.TargetProfit*0.95 {
					ai.LastOpinion = avatar + " 冷酷无情:\"利润吃够了，准备派发筹码。\""
				} else {
					ai.LastOpinion = avatar + " 沉吟思考:\"量能跟不上，随时准备撤退...\""
				}
			}
		} else if ai.Type == "Quant" {
			if eventFear > 1.2 {
				ai.LastOpinion = avatar + " [高危红警] 流动性枯竭预警！风控止损模块强制启动..."
			} else if profitRatio > ai.TargetProfit {
				ai.LastOpinion = avatar + " [微利平仓] 微薄利润已达标，开始自动市价砸盘套现。"
			} else {
				ai.LastOpinion = avatar + " [动量跟踪] 高频监控板上封单，震荡滤波中。"
			}
		} else { // Retail
			if eventFear > 1.8 {
				ai.LastOpinion = avatar + " 魂飞魄散:\"为什么软件卡死排不进去了！放我出去啊！\""
			} else if eventFear > 1.3 {
				ai.LastOpinion = avatar + " 瑟瑟发抖:\"这分时图要被A杀了，老婆叫我赶紧割肉保本...\""
			} else if profitRatio > 1.3 {
				ai.LastOpinion = avatar + " 患得患失:\"家人门，赚了30%！要是明天跌了咋办？\""
			} else if profitRatio < 0.8 {
				ai.LastOpinion = avatar + " 躺板板:\"主力哥哥别砸了，我已经装死了，打死不卖！\""
			} else {
				ai.LastOpinion = avatar + " 陷入癫狂:\"无脑加仓！明天连板赚别墅靠大海！\""
			}
		}
	}
}

// 策略建议系统
type StrategyAdvice struct {
	Action      string  // "HOLD" 或 "SELL"
	Confidence  string  // "强烈建议", "建议", "可考虑"
	Reason      string  // 主要原因
	RiskLevel   string  // "低风险", "中风险", "高风险", "极高风险"
	ProfitRatio float64 // 当前盈利比例
}

func generateStrategyAdvice(state *GameState) StrategyAdvice {
	advice := StrategyAdvice{}

	// 计算当前净资产和盈亏 (Gross - MarginDebt)
	netAsset := float64(state.PlayerShares)*state.Price + state.PlayerCash - state.MarginDebt
	profitRatio := (netAsset - state.InitialAsset) / state.InitialAsset
	advice.ProfitRatio = profitRatio

	// 统计AI出货情况
	escapedCount := 0
	whaleEscaped := 0
	for _, ai := range state.AIs {
		if ai.HasSold {
			escapedCount++
			if ai.Type == "Whale" {
				whaleEscaped++
			}
		}
	}
	escapedRatio := float64(escapedCount) / float64(len(state.AIs))

	// 风险评分系统 (0-10分，越高越危险)
	riskScore := 0.0

	// 1. 崩盘预警等级影响 (0-4分)
	riskScore += float64(state.CrashWarningLevel)

	// 2. 连续下跌影响 (0-2分)
	if state.ConsecutiveFallDays >= 2 {
		riskScore += 1.0
	}
	if state.ConsecutiveFallDays >= 3 {
		riskScore += 1.0
	}

	// 3. AI逃跑比例影响 (0-2分)
	if escapedRatio > 0.2 {
		riskScore += 1.0
	}
	if escapedRatio > 0.4 {
		riskScore += 1.0
	}

	// 4. Whale逃跑影响 (0-2分)
	riskScore += float64(whaleEscaped)

	// 5. 事件情绪影响 (0-2分)
	if state.CurrentEvent.FearModifier > 2.0 {
		riskScore += 1.0
	}
	if state.CurrentEvent.FearModifier > 2.5 {
		riskScore += 1.0
	}

	// 6. 游戏后期风险 (0-1分)
	if state.Day > 12 {
		riskScore += 1.0
	}

	// 确定风险等级
	if riskScore >= 7.0 {
		advice.RiskLevel = "极高风险"
	} else if riskScore >= 5.0 {
		advice.RiskLevel = "高风险"
	} else if riskScore >= 3.0 {
		advice.RiskLevel = "中风险"
	} else {
		advice.RiskLevel = "低风险"
	}

	// 决策逻辑
	reasons := []string{}
	sellScore := 0 // 卖出倾向分数
	holdScore := 0 // 持有倾向分数

	// 分析1: 盈利情况
	if profitRatio > 0.25 {
		sellScore += 3
		reasons = append(reasons, fmt.Sprintf("已获利%.0f%%，利润丰厚", profitRatio*100))
	} else if profitRatio > 0.15 {
		sellScore += 2
		reasons = append(reasons, fmt.Sprintf("已获利%.0f%%，可考虑落袋为安", profitRatio*100))
	} else if profitRatio > 0.05 {
		holdScore += 1
		reasons = append(reasons, fmt.Sprintf("小幅盈利%.0f%%，可再等等", profitRatio*100))
	} else if profitRatio < -0.05 {
		sellScore += 2
		reasons = append(reasons, fmt.Sprintf("已亏损%.0f%%，止损保本", -profitRatio*100))
	} else {
		holdScore += 1
		reasons = append(reasons, "当前基本持平，可观望")
	}

	// 分析2: 崩盘预警
	if state.CrashWarningLevel >= 3 {
		sellScore += 4
		reasons = append(reasons, "崩盘预警达到高危级别！")
	} else if state.CrashWarningLevel >= 2 {
		sellScore += 2
		reasons = append(reasons, "崩盘预警信号出现")
	}

	// 分析3: Whale动向
	if whaleEscaped >= 2 {
		sellScore += 3
		reasons = append(reasons, "大资金已开始撤离！")
	} else if whaleEscaped >= 1 {
		sellScore += 2
		reasons = append(reasons, "有游资开始出货")
	}

	// 分析4: 连续下跌
	if state.ConsecutiveFallDays >= 3 {
		sellScore += 2
		reasons = append(reasons, fmt.Sprintf("连续下跌%d时段，趋势转坏", state.ConsecutiveFallDays))
	}

	// 分析5: 事件情绪
	if state.CurrentEvent.FearModifier > 2.5 {
		sellScore += 2
		reasons = append(reasons, "市场极度恐慌")
	} else if state.CurrentEvent.Sentiment > 1.3 {
		holdScore += 2
		reasons = append(reasons, "市场情绪乐观，可继续持有")
	}

	// 分析6: AI逃跑比例
	if escapedRatio > 0.4 {
		sellScore += 3
		reasons = append(reasons, fmt.Sprintf("%.0f%%的AI已出逃，羊群踩踏风险！", escapedRatio*100))
	} else if escapedRatio > 0.2 {
		sellScore += 1
		reasons = append(reasons, "部分AI开始出货")
	}

	// 分析7: 游戏时间
	if state.Day > 13 {
		sellScore += 2
		reasons = append(reasons, "游戏接近尾声，落袋为安")
	} else if state.Day <= 5 {
		holdScore += 1
		reasons = append(reasons, "游戏初期，还有机会")
	}

	// 综合决策
	if sellScore >= holdScore+3 {
		advice.Action = "SELL"
		advice.Confidence = "强烈建议"
	} else if sellScore >= holdScore+1 {
		advice.Action = "SELL"
		advice.Confidence = "建议"
	} else if sellScore > holdScore {
		advice.Action = "SELL"
		advice.Confidence = "可考虑"
	} else if holdScore > sellScore+2 {
		advice.Action = "HOLD"
		advice.Confidence = "建议"
	} else {
		advice.Action = "HOLD"
		advice.Confidence = "可考虑"
	}

	// 选择最重要的原因（最多2条）
	if len(reasons) > 2 {
		advice.Reason = reasons[0] + "；" + reasons[1]
	} else if len(reasons) > 0 {
		advice.Reason = reasons[0]
	} else {
		advice.Reason = "市场平稳"
	}

	return advice
}

// ===== Phase 2: 反身性 + 贝叶斯分析函数 =====

// 捕获当前游戏状态快照
func captureGameStateSnapshot(state *GameState) GameStateSnapshot {
	escapedCount := 0
	whaleEscaped := 0
	for _, ai := range state.AIs {
		if ai.HasSold {
			escapedCount++
			if ai.Type == "Whale" {
				whaleEscaped++
			}
		}
	}
	escapedRatio := float64(escapedCount) / float64(len(state.AIs))

	playerAction := "空仓"
	if state.PlayerShares > 0 {
		playerAction = "持仓"
	}
	if state.PlayerOrderType == "Buy" {
		playerAction = "买入"
	} else if state.PlayerOrderType == "Sell" {
		playerAction = "卖出"
	}

	return GameStateSnapshot{
		Day:                 state.Day,
		Session:             state.Session,
		Price:               state.Price,
		EscapedAIRatio:      escapedRatio,
		WhaleEscapeCount:    whaleEscaped,
		CrashWarningLevel:   state.CrashWarningLevel,
		ConsecutiveFallDays: state.ConsecutiveFallDays,
		EventSentiment:      state.CurrentEvent.Sentiment,
		EventFear:           state.CurrentEvent.FearModifier,
		PlayerAction:        playerAction,
		TotalBuyPressure:    state.BuyPressure,
		TotalSellPressure:   state.SellPressure,
	}
}

// 计算反身性指标
func calculateReflexivityMetrics(state *GameState, histCache *HistoricalStateCache) ReflexivityMetrics {
	metrics := ReflexivityMetrics{}

	if len(histCache.States) < 3 {
		return metrics // 数据不足
	}

	// 提取历史数据
	var prices, sentiments, exitRatios []float64
	var whaleActions []string

	for _, snap := range histCache.States {
		prices = append(prices, snap.Price)
		sentiments = append(sentiments, snap.EventSentiment)
		exitRatios = append(exitRatios, snap.EscapedAIRatio)

		// 推断whale行为
		whaleAction := "观望"
		if snap.WhaleEscapeCount > 0 {
			whaleAction = "出货"
		}
		whaleActions = append(whaleActions, whaleAction)
	}

	// 保留最近10个数据点
	if len(prices) > 10 {
		prices = prices[len(prices)-10:]
		sentiments = sentiments[len(sentiments)-10:]
		exitRatios = exitRatios[len(exitRatios)-10:]
		whaleActions = whaleActions[len(whaleActions)-10:]
	}

	metrics.RecentPriceChanges = prices
	metrics.RecentSentimentScores = sentiments
	metrics.RecentAIExitRatios = exitRatios
	metrics.RecentWhaleActions = whaleActions

	// 1. 计算情绪-价格相关性 (Pearson)
	metrics.SentimentPriceCorrelation = calculatePearsonCorrelation(prices, sentiments)

	// 2. 检测恐慌性抛售强度
	metrics.PanicSellIntensity = calculatePanicIntensity(state, exitRatios)

	// 3. 检测FOMO追涨强度
	metrics.FOMOBuyIntensity = calculateFOMOIntensity(state, prices)

	// 4. Whale羊群效应
	metrics.WhaleHerdingEffect = calculateWhaleHerding(whaleActions)

	// 5. 散户追涨效应
	metrics.RetailChasingEffect = calculateRetailChasing(state, prices)

	// 6. 基本面脱节度
	openPrice := histCache.States[0].Price
	currentPrice := state.Price
	fairValue := openPrice // 简化：以开盘价为公允价值
	metrics.FundamentalDisconnect = math.Abs(currentPrice-fairValue) / fairValue * 10
	if metrics.FundamentalDisconnect > 10 {
		metrics.FundamentalDisconnect = 10
	}

	// 7. 动量衰减
	if len(prices) >= 5 {
		recentMomentum := (prices[len(prices)-1] - prices[len(prices)-3]) / prices[len(prices)-3]
		olderMomentum := (prices[len(prices)-3] - prices[len(prices)-5]) / prices[len(prices)-5]
		metrics.MomentumDecay = olderMomentum - recentMomentum // 正值=动量衰减
	}

	return metrics
}

// Pearson相关系数计算
func calculatePearsonCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}

	n := float64(len(x))
	var sumX, sumY, sumXY, sumX2, sumY2 float64

	for i := 0; i < len(x); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}

	numerator := n*sumXY - sumX*sumY
	denominator := math.Sqrt((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

// 计算恐慌抛售强度
func calculatePanicIntensity(state *GameState, exitRatios []float64) float64 {
	intensity := 0.0

	// 因子1: 连续下跌
	if state.ConsecutiveFallDays >= 3 {
		intensity += 4.0
	} else if state.ConsecutiveFallDays >= 2 {
		intensity += 2.0
	}

	// 因子2: AI逃跑加速
	if len(exitRatios) >= 3 {
		recentExit := exitRatios[len(exitRatios)-1]
		olderExit := exitRatios[len(exitRatios)-3]
		acceleration := (recentExit - olderExit) / 0.01 // 每1%加速度
		intensity += math.Min(acceleration, 3.0)
	}

	// 因子3: 恐慌事件
	if state.CurrentEvent.FearModifier > 2.5 {
		intensity += 3.0
	} else if state.CurrentEvent.FearModifier > 2.0 {
		intensity += 1.5
	}

	if intensity > 10 {
		intensity = 10
	}
	return intensity
}

// 计算FOMO追涨强度
func calculateFOMOIntensity(state *GameState, _ []float64) float64 {
	intensity := 0.0
	// prices 参数保留以便未来基于价格波动分析FOMO强度

	// 因子1: 妖股狂热
	if state.IsMonsterStock {
		intensity += 5.0
	}

	// 因子2: 连涨天数
	if state.ConsecutiveGrowthDays >= 3 {
		intensity += 3.0
	} else if state.ConsecutiveGrowthDays >= 2 {
		intensity += 1.5
	}

	// 因子3: 情绪极度乐观
	if state.CurrentEvent.Sentiment > 1.4 {
		intensity += 2.0
	} else if state.CurrentEvent.Sentiment > 1.3 {
		intensity += 1.0
	}

	if intensity > 10 {
		intensity = 10
	}
	return intensity
}

// 计算Whale羊群效应
func calculateWhaleHerding(whaleActions []string) float64 {
	if len(whaleActions) < 3 {
		return 0
	}

	recent := whaleActions[len(whaleActions)-3:]
	sellCount := 0
	for _, action := range recent {
		if action == "出货" {
			sellCount++
		}
	}

	return float64(sellCount) / float64(len(recent))
}

// 计算散户追涨效应
func calculateRetailChasing(state *GameState, prices []float64) float64 {
	// 简化：如果价格上涨且买压大于卖压，认为散户在追涨
	if len(prices) < 2 {
		return 0
	}

	priceRising := prices[len(prices)-1] > prices[len(prices)-2]
	retailBuyingHigh := state.RetailBuying > state.RetailSelling

	if priceRising && retailBuyingHigh {
		return 0.8
	} else if priceRising {
		return 0.4
	}

	return 0.0
}

// 生成反身性信号
func generateReflexivitySignals(metrics ReflexivityMetrics, _ *GameState) []ReflexivitySignal {
	signals := []ReflexivitySignal{}
	// state 参数保留以便未来基于游戏状态生成更精准的信号

	// 信号1: 恐慌踩踏
	if metrics.PanicSellIntensity > 7 && metrics.WhaleHerdingEffect > 0.6 {
		signals = append(signals, ReflexivitySignal{
			Type:        "恐慌踩踏",
			Strength:    metrics.PanicSellIntensity,
			Description: "大资金集体出逃引发连锁抛售，市场进入自我强化的下跌螺旋",
			IsBullish:   false,
		})
	}

	// 信号2: FOMO狂热
	if metrics.FOMOBuyIntensity > 7 && metrics.RetailChasingEffect > 0.7 {
		signals = append(signals, ReflexivitySignal{
			Type:        "FOMO狂热",
			Strength:    metrics.FOMOBuyIntensity,
			Description: "散户疯狂追涨，情绪推动价格脱离基本面，泡沫风险积聚",
			IsBullish:   false, // 短期看涨，但风险信号
		})
	}

	// 信号3: 大资金出逃
	if metrics.WhaleHerdingEffect > 0.8 {
		signals = append(signals, ReflexivitySignal{
			Type:        "大资金出逃",
			Strength:    metrics.WhaleHerdingEffect * 10,
			Description: "游资形成羊群效应，主力资金正在有序撤离",
			IsBullish:   false,
		})
	}

	// 信号4: 散户接盘
	if metrics.RetailChasingEffect > 0.8 && metrics.WhaleHerdingEffect > 0.5 {
		signals = append(signals, ReflexivitySignal{
			Type:        "散户接盘",
			Strength:    8.0,
			Description: "散户高位追涨接盘，而大资金正在出货，典型的反身性陷阱",
			IsBullish:   false,
		})
	}

	// 信号5: 基本面严重脱节
	if metrics.FundamentalDisconnect > 7 {
		signals = append(signals, ReflexivitySignal{
			Type:        "价格泡沫",
			Strength:    metrics.FundamentalDisconnect,
			Description: fmt.Sprintf("价格与基本面脱节度%.1f/10，反身性推动的泡沫", metrics.FundamentalDisconnect),
			IsBullish:   false,
		})
	}

	return signals
}

// 运行贝叶斯分析
func runBayesianAnalysis(state *GameState, histCache *HistoricalStateCache, reflexMetrics ReflexivityMetrics) BayesianAnalysis {
	analysis := BayesianAnalysis{}

	// 1. 崩盘概率模型
	analysis.CrashProbability = updateCrashProbability(state, reflexMetrics)

	// 2. 最优离场时机
	analysis.ExitTimingDistribution = make(map[int]float64)
	analysis.BestExitDay, analysis.BestExitConfidence = calculateExitTiming(state, analysis.CrashProbability)

	// 3. 价格反转概率
	analysis.ReversalProbability = calculateReversalProbability(state, histCache)

	// 4. AI行为预测
	analysis.AIBehaviorPrediction = make(map[string]*AIBehaviorModel)
	for _, ai := range state.AIs {
		if ai.Type == "Whale" || (ai.Shares > 1000 && !ai.HasSold) {
			analysis.AIBehaviorPrediction[ai.ID] = predictAIBehavior(ai, state)
		}
	}

	return analysis
}

// 更新崩盘概率 (贝叶斯)
func updateCrashProbability(state *GameState, reflexMetrics ReflexivityMetrics) BayesianCrashModel {
	model := BayesianCrashModel{
		EvidenceWeights: make(map[string]float64),
		CurrentEvidence: make(map[string]float64),
		CrashProbByDay:  make(map[int]float64),
	}

	// 先验概率 (根据主题难度动态调整)
	model.Prior = 0.15 // 默认15%

	// 根据主题调整先验概率
	switch CurrentTheme.Name {
	case "经典模式":
		model.Prior = 0.15 // 平衡风险
	case "科技股狂潮":
		model.Prior = 0.25 // 泡沫破裂风险高
	case "新能源泡沫":
		model.Prior = 0.20 // 政策依赖风险
	case "疫情周期":
		model.Prior = 0.22 // 不确定性高
	case "贸易战周期":
		model.Prior = 0.18 // 中等风险
	default:
		model.Prior = 0.15
	}

	// 证据与似然比
	likelihoodRatios := []float64{}

	// E1: AI逃跑比例
	escapedCount := 0
	for _, ai := range state.AIs {
		if ai.HasSold {
			escapedCount++
		}
	}
	escapedRatio := float64(escapedCount) / float64(len(state.AIs))

	if escapedRatio > 0.4 {
		likelihoodRatios = append(likelihoodRatios, 3.0)
	} else if escapedRatio > 0.2 {
		likelihoodRatios = append(likelihoodRatios, 1.8)
	} else {
		likelihoodRatios = append(likelihoodRatios, 0.8)
	}

	// E2: 连续下跌天数
	if state.ConsecutiveFallDays >= 3 {
		likelihoodRatios = append(likelihoodRatios, 4.0)
	} else if state.ConsecutiveFallDays == 2 {
		likelihoodRatios = append(likelihoodRatios, 2.5)
	}

	// E3: 崩盘预警等级
	if state.CrashWarningLevel >= 4 {
		likelihoodRatios = append(likelihoodRatios, 5.0)
	} else if state.CrashWarningLevel >= 3 {
		likelihoodRatios = append(likelihoodRatios, 3.0)
	} else if state.CrashWarningLevel >= 2 {
		likelihoodRatios = append(likelihoodRatios, 1.5)
	}

	// E4: 恐慌情绪
	if state.CurrentEvent.FearModifier > 2.5 {
		likelihoodRatios = append(likelihoodRatios, 3.5)
	} else if state.CurrentEvent.FearModifier > 2.0 {
		likelihoodRatios = append(likelihoodRatios, 2.0)
	}

	// E5: 基本面脱节 (反身性)
	if reflexMetrics.FundamentalDisconnect > 7 {
		likelihoodRatios = append(likelihoodRatios, 2.5)
	} else if reflexMetrics.FundamentalDisconnect > 4 {
		likelihoodRatios = append(likelihoodRatios, 1.5)
	}

	// E6: 恐慌抛售强度 (反身性)
	if reflexMetrics.PanicSellIntensity > 8 {
		likelihoodRatios = append(likelihoodRatios, 4.0)
	} else if reflexMetrics.PanicSellIntensity > 5 {
		likelihoodRatios = append(likelihoodRatios, 2.0)
	}

	// 贝叶斯序贯更新
	posterior := model.Prior
	for _, lr := range likelihoodRatios {
		posterior = (lr * posterior) / (lr*posterior + (1 - posterior))
	}

	// 限制范围
	if posterior < 0.01 {
		posterior = 0.01
	}
	if posterior > 0.99 {
		posterior = 0.99
	}

	model.Posterior = posterior

	// 计算各天崩盘概率分布
	for d := state.Day + 1; d <= state.MaxDays; d++ {
		dayDiff := d - state.Day
		// 越接近后期，崩盘概率越高
		timeFactor := 1.0 + float64(d-10)*0.1 // Day 10之后风险递增
		if timeFactor < 0.5 {
			timeFactor = 0.5
		}
		model.CrashProbByDay[d] = posterior * timeFactor * math.Pow(0.95, float64(dayDiff))
	}

	return model
}

// 计算最优离场时机
func calculateExitTiming(state *GameState, crashModel BayesianCrashModel) (int, float64) {
	utilities := make(map[int]float64)

	currentAsset := float64(state.PlayerShares)*state.Price + state.PlayerCash

	for d := state.Day + 1; d <= state.MaxDays; d++ {
		// 预期收益
		growthRate := 0.01 // 默认每天1%
		if state.IsMonsterStock {
			growthRate = 0.05
		} else if state.CurrentEvent.Sentiment > 1.2 {
			growthRate = 0.03
		} else if state.CurrentEvent.Sentiment < 0.9 {
			growthRate = -0.02
		}

		daysAhead := d - state.Day
		expectedReturn := currentAsset * math.Pow(1+growthRate, float64(daysAhead))

		// 累计崩盘风险
		cumulativeCrashProb := 0.0
		for day := state.Day + 1; day <= d; day++ {
			if prob, ok := crashModel.CrashProbByDay[day]; ok {
				cumulativeCrashProb += prob
			}
		}
		if cumulativeCrashProb > 1.0 {
			cumulativeCrashProb = 1.0
		}

		// 崩盘损失
		crashLoss := currentAsset * 0.7 // 崩盘损失70%

		// 效用 = 预期收益 - 风险损失
		utility := expectedReturn*(1-cumulativeCrashProb) - crashLoss*cumulativeCrashProb
		utilities[d] = utility
	}

	// Softmax转换为概率分布
	maxUtility := -1e9
	for _, u := range utilities {
		if u > maxUtility {
			maxUtility = u
		}
	}

	expSum := 0.0
	expValues := make(map[int]float64)
	for d, u := range utilities {
		expVal := math.Exp((u - maxUtility) / 1000) // 温度参数=1000
		expValues[d] = expVal
		expSum += expVal
	}

	bestDay := state.Day + 1
	bestProb := 0.0

	for d, expVal := range expValues {
		prob := expVal / expSum
		if prob > bestProb {
			bestProb = prob
			bestDay = d
		}
	}

	return bestDay, bestProb
}

// 计算价格反转概率
func calculateReversalProbability(state *GameState, histCache *HistoricalStateCache) BayesianReversalModel {
	model := BayesianReversalModel{}

	// 1. 超卖指标
	if len(histCache.States) >= 5 {
		highestPrice := 0.0
		for _, snap := range histCache.States {
			if snap.Price > highestPrice {
				highestPrice = snap.Price
			}
		}

		// 避免除零
		if highestPrice > 0 {
			drawdown := (highestPrice - state.Price) / highestPrice
			if drawdown > 0.3 {
				model.OversoldIndicator = 1.0
			} else if drawdown > 0.2 {
				model.OversoldIndicator = 0.7
			} else if drawdown > 0.1 {
				model.OversoldIndicator = 0.4
			}
		}
	}

	// 2. 成交量枯竭
	if len(histCache.States) >= 5 {
		recentVolume := histCache.States[len(histCache.States)-1].TotalSellPressure
		avgVolume := 0
		for i := len(histCache.States) - 5; i < len(histCache.States)-1; i++ {
			avgVolume += histCache.States[i].TotalSellPressure
		}
		avgVolume /= 4

		// 避免除零
		if avgVolume > 0 {
			if recentVolume < avgVolume/2 {
				model.VolumeExhaustion = 0.8
			} else if recentVolume < int(float64(avgVolume)*0.7) {
				model.VolumeExhaustion = 0.5
			}
		}
	}

	// 3. Whale回流信号
	for _, ai := range state.AIs {
		if ai.Type == "Whale" && ai.HasSold && ai.Cash > 50000 {
			model.WhaleReentrySignals++
		}
	}

	// 贝叶斯更新
	prior := 0.30
	likelihoodRatios := []float64{}

	if model.OversoldIndicator > 0.7 {
		likelihoodRatios = append(likelihoodRatios, 3.0)
	}
	if model.VolumeExhaustion > 0.6 {
		likelihoodRatios = append(likelihoodRatios, 2.5)
	}
	if model.WhaleReentrySignals >= 2 {
		likelihoodRatios = append(likelihoodRatios, 2.0)
	}
	if state.ConsecutiveFallDays >= 3 {
		likelihoodRatios = append(likelihoodRatios, 1.5) // 均值回归
	}

	posterior := prior
	for _, lr := range likelihoodRatios {
		posterior = (lr * posterior) / (lr*posterior + (1 - posterior))
	}

	model.RecoveryProbability = posterior
	model.ContinuedDeclineProbability = math.Max(0, 0.8-posterior) // 剩余为横盘概率

	return model
}

// 预测AI行为
func predictAIBehavior(ai *AI, state *GameState) *AIBehaviorModel {
	model := &AIBehaviorModel{
		TraderID:   ai.ID,
		TraderType: ai.Type,
	}

	if ai.HasSold {
		// 已出货，不再预测
		model.ProbHold = 1.0
		return model
	}

	profitRatio := (state.Price - ai.Cost) / ai.Cost

	// 根据AI类型预测
	switch ai.SubType {
	case "刺客": // Whale - 追求高利润快进快出
		if profitRatio > ai.TargetProfit*0.8 {
			model.ProbNextSell = 0.85
		} else if state.CurrentEvent.FearModifier > 2.0 {
			model.ProbNextSell = 0.80
		} else {
			model.ProbHold = 0.70
		}
		model.ProfitThreshold = ai.TargetProfit

	case "打板": // Whale - 激进追涨
		if state.IsMonsterStock {
			model.ProbHold = 0.80
		} else if profitRatio > 0.5 {
			model.ProbNextSell = 0.75
		} else {
			model.ProbHold = 0.60
		}

	case "网格": // Quant - 高频网格交易
		if profitRatio > 0.03 {
			model.ProbNextSell = 0.60
		} else if profitRatio < -0.03 {
			model.ProbNextBuy = 0.70
		} else {
			model.ProbHold = 0.70
		}

	case "新韭": // Retail - 追涨杀跌
		if state.IsMonsterStock {
			model.ProbHold = 0.80
		} else if profitRatio < -0.05 && state.ConsecutiveFallDays >= 2 {
			model.ProbNextSell = 0.70 // 恐慌割肉
		} else {
			model.ProbHold = 0.60
		}

	case "国家队": // Institution - 稳定市场
		if state.CrashWarningLevel >= 3 {
			model.ProbNextBuy = 0.80 // 救市
		} else {
			model.ProbHold = 0.90
		}

	default:
		model.ProbHold = 0.70
	}

	// 归一化
	total := model.ProbNextSell + model.ProbNextBuy + model.ProbHold
	if total > 0 {
		model.ProbNextSell /= total
		model.ProbNextBuy /= total
		model.ProbHold /= total
	}

	return model
}

// 辅助函数：将Session转换为整数（用于缓存key）
func sessionToInt(session string) int {
	switch session {
	case "集合竞价":
		return 0
	case "早盘":
		return 1
	case "午盘":
		return 2
	case "尾盘":
		return 3
	default:
		return 0
	}
}

// 辅助函数：计算总回合数
func calculateTurnNumber(day int, session string) int {
	// 每天4个session（集合竞价、早盘、午盘、尾盘）
	baseTurns := (day - 1) * 4
	sessionTurn := sessionToInt(session)
	return baseTurns + sessionTurn
}

// 生成高级分析（整合反身性+贝叶斯）
func generateAdvancedAnalysis(state *GameState, histCache *HistoricalStateCache) AdvancedAnalysis {
	// 检查缓存（同一回合内复用分析结果）
	currentTurn := state.Day*10 + sessionToInt(state.Session)
	if GlobalCache.IsValid && GlobalCache.LastUpdateTurn == currentTurn {
		return GlobalCache.LastAnalysis
	}

	analysis := AdvancedAnalysis{}

	// 1. 基础建议（保持原有逻辑）
	analysis.BasicAdvice = generateStrategyAdvice(state)

	// 2. 反身性分析
	analysis.ReflexivityMetrics = calculateReflexivityMetrics(state, histCache)
	analysis.ReflexivitySignals = generateReflexivitySignals(analysis.ReflexivityMetrics, state)

	// 3. 贝叶斯分析
	analysis.BayesianAnalysis = runBayesianAnalysis(state, histCache, analysis.ReflexivityMetrics)

	// 4. 综合风险评分 (0-100)
	riskScore := 0.0
	riskScore += analysis.BayesianAnalysis.CrashProbability.Posterior * 40                      // 崩盘概率权重40%
	riskScore += analysis.ReflexivityMetrics.PanicSellIntensity * 3                             // 恐慌强度权重30%
	riskScore += (1.0 - analysis.BayesianAnalysis.ReversalProbability.RecoveryProbability) * 30 // 反转概率权重30%

	if riskScore > 100 {
		riskScore = 100
	}
	analysis.OverallRiskScore = riskScore

	// 5. 提炼核心洞察（最多3条）
	insights := []string{}

	// 洞察1: 最紧迫的风险
	if analysis.BayesianAnalysis.CrashProbability.Posterior > 0.6 {
		insights = append(insights, fmt.Sprintf("崩盘概率高达%.0f%%，强烈建议立即离场",
			analysis.BayesianAnalysis.CrashProbability.Posterior*100))
	} else if len(analysis.ReflexivitySignals) > 0 {
		signal := analysis.ReflexivitySignals[0]
		insights = append(insights, fmt.Sprintf("%s信号出现，市场进入反身性循环", signal.Type))
	}

	// 洞察2: 最优时机
	if analysis.BayesianAnalysis.BestExitConfidence > 0.4 {
		insights = append(insights, fmt.Sprintf("最佳离场窗口: 第%d天 (置信度%.0f%%)",
			analysis.BayesianAnalysis.BestExitDay,
			analysis.BayesianAnalysis.BestExitConfidence*100))
	}

	// 洞察3: AI行为预警
	whaleSellCount := 0
	for _, pred := range analysis.BayesianAnalysis.AIBehaviorPrediction {
		if pred.ProbNextSell > 0.6 {
			whaleSellCount++
		}
	}
	if whaleSellCount >= 2 {
		insights = append(insights, fmt.Sprintf("预测下回合将有%d个大资金出货", whaleSellCount))
	} else if analysis.BayesianAnalysis.ReversalProbability.RecoveryProbability > 0.6 {
		insights = append(insights, fmt.Sprintf("价格反转概率%.0f%%，存在抄底机会",
			analysis.BayesianAnalysis.ReversalProbability.RecoveryProbability*100))
	}

	// 限制最多3条
	if len(insights) > 3 {
		insights = insights[:3]
	}
	analysis.KeyInsights = insights

	// 更新缓存
	GlobalCache.LastAnalysis = analysis
	GlobalCache.LastUpdateTurn = currentTurn
	GlobalCache.IsValid = true

	return analysis
}

// 渲染反身性趋势图 (ASCII 艺术)
func renderReflexivityTrend(metrics ReflexivityMetrics) string {
	if len(metrics.RecentPriceChanges) < 3 {
		return "数据不足"
	}

	// 使用最近的数据点（最多10个）
	data := metrics.RecentPriceChanges
	if len(data) > 10 {
		data = data[len(data)-10:]
	}

	// 计算相对变化率（归一化到0-8范围）
	min, max := data[0], data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// 避免除零
	if max-min < 0.01 {
		return "价格波动过小"
	}

	// 映射到 0-8 的柱状图高度
	bars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	var trend string
	for _, price := range data {
		normalized := (price - min) / (max - min)
		index := int(normalized * 7)
		if index > 7 {
			index = 7
		}
		trend += bars[index]
	}

	return trend
}

// 教学提示函数
func showEducationTip(signalType string) {
	if !GlobalSettings.ShowEducation {
		return
	}

	if SeenSignals[signalType] {
		return // 已经看过了
	}

	SeenSignals[signalType] = true
	GlobalStats.ReflexivityPatternsSeen[signalType]++

	fmt.Printf("\n%s┌────────────────── 📚 反身性小课堂 ──────────────────┐%s\n", Yellow, Reset)

	switch signalType {
	case "恐慌踩踏":
		fmt.Printf("%s│%s  【恐慌踩踏】是典型的自我强化下跌循环：\n", Yellow, Reset)
		fmt.Printf("%s│%s  大资金出逃 → 价格下跌 → 散户恐慌 → 更多人卖出 → 加速下跌\n", Yellow, Reset)
		fmt.Printf("%s│%s  🎯 应对策略：提前识别信号，在踩踏开始前离场\n", Yellow, Reset)
	case "FOMO狂热":
		fmt.Printf("%s│%s  【FOMO狂热】是情绪推动的泡沫循环：\n", Yellow, Reset)
		fmt.Printf("%s│%s  价格上涨 → FOMO情绪 → 散户追涨 → 价格进一步上涨 → 泡沫积聚\n", Yellow, Reset)
		fmt.Printf("%s│%s  🎯 应对策略：保持理性，在极度狂热时逐步减仓\n", Yellow, Reset)
	case "大资金出逃":
		fmt.Printf("%s│%s  【大资金出逃】意味着聪明钱已经开始撤退：\n", Yellow, Reset)
		fmt.Printf("%s│%s  游资形成羊群效应，主力资金有序撤离\n", Yellow, Reset)
		fmt.Printf("%s│%s  🎯 应对策略：跟随大资金的脚步，不要做最后的接盘侠\n", Yellow, Reset)
	case "散户接盘":
		fmt.Printf("%s│%s  【散户接盘】是典型的反身性陷阱：\n", Yellow, Reset)
		fmt.Printf("%s│%s  散户高位追涨接盘，而大资金正在出货\n", Yellow, Reset)
		fmt.Printf("%s│%s  🎯 应对策略：避免在价格高位追涨，警惕大资金动向\n", Yellow, Reset)
	case "价格泡沫":
		fmt.Printf("%s│%s  【价格泡沫】表示价格已严重脱离基本面：\n", Yellow, Reset)
		fmt.Printf("%s│%s  市场情绪推动价格远离合理估值\n", Yellow, Reset)
		fmt.Printf("%s│%s  🎯 应对策略：泡沫必然破裂，及时止盈\n", Yellow, Reset)
	}

	fmt.Printf("%s└──────────────────────────────────────────────────────┘%s\n", Yellow, Reset)
	fmt.Println("\n按 Enter 继续...")
	waitEnter(bufio.NewReader(os.Stdin))
}

// 显示高级量化分析
func displayAdvancedAnalysis(analysis AdvancedAnalysis) {
	if !GlobalSettings.ShowAdvancedAnalysis {
		return // 用户关闭了高级分析
	}

	fmt.Printf("\n%s┌─────────── 🧠 高级量化分析 (反身性+贝叶斯) ───────────┐%s\n", Cyan, Reset)

	// ===== 反身性信号 =====
	if len(analysis.ReflexivitySignals) > 0 {
		fmt.Printf("%s│%s  【反身性信号】\n", Cyan, Reset)
		for _, signal := range analysis.ReflexivitySignals {
			color := Red
			icon := "⚠️"
			if signal.IsBullish {
				color = Green
				icon = "✅"
			}
			fmt.Printf("%s│%s    %s %s%s%s (强度: %.1f/10)\n",
				Cyan, Reset, icon, color, signal.Type, Reset, signal.Strength)
			fmt.Printf("%s│%s       └─ %s\n", Cyan, Reset, signal.Description)

			// 显示教学提示（首次出现时）
			showEducationTip(signal.Type)
		}
	} else {
		fmt.Printf("%s│%s  【反身性信号】%s 暂无明显反身性效应%s\n", Cyan, Reset, Green, Reset)
	}

	// ===== 价格趋势可视化 =====
	trend := renderReflexivityTrend(analysis.ReflexivityMetrics)
	fmt.Printf("%s│%s  【价格走势】 %s\n", Cyan, Reset, trend)

	// ===== 贝叶斯概率分析 =====
	fmt.Printf("%s│%s  【贝叶斯概率推演】\n", Cyan, Reset)
	bayes := analysis.BayesianAnalysis

	// 崩盘概率
	crashProb := bayes.CrashProbability.Posterior * 100
	crashColor := Green
	crashLevel := "低风险"
	if crashProb > 60 {
		crashColor = Red
		crashLevel = "极高风险"
	} else if crashProb > 40 {
		crashColor = Red
		crashLevel = "高风险"
	} else if crashProb > 25 {
		crashColor = Yellow
		crashLevel = "中等风险"
	}

	fmt.Printf("%s│%s    崩盘概率: %s%.1f%%%s (%s)\n",
		Cyan, Reset, crashColor, crashProb, Reset, crashLevel)

	// 最佳离场窗口
	fmt.Printf("%s│%s    最佳离场窗口: %s第%d天%s (置信度: %.0f%%)\n",
		Cyan, Reset, Yellow, bayes.BestExitDay, Reset, bayes.BestExitConfidence*100)

	// 价格反转概率
	recoveryProb := bayes.ReversalProbability.RecoveryProbability * 100
	declineProb := bayes.ReversalProbability.ContinuedDeclineProbability * 100

	fmt.Printf("%s│%s    价格反转概率: %s%.1f%%%s  |  继续下跌: %s%.1f%%%s\n",
		Cyan, Reset, Green, recoveryProb, Reset, Red, declineProb, Reset)

	// AI行为预测
	if len(bayes.AIBehaviorPrediction) > 0 {
		whaleNextSellCount := 0
		for id, pred := range bayes.AIBehaviorPrediction {
			if pred.ProbNextSell > 0.6 && pred.TraderType == "Whale" {
				whaleNextSellCount++
			}
			// 显示详细AI预测（仅显示概率>0.6的）
			if pred.ProbNextSell > 0.6 || pred.ProbNextBuy > 0.6 {
				action := "卖出"
				prob := pred.ProbNextSell * 100
				actionColor := Red
				if pred.ProbNextBuy > pred.ProbNextSell {
					action = "买入"
					prob = pred.ProbNextBuy * 100
					actionColor = Green
				}

				// 简化ID显示
				shortID := id
				if len(id) > 8 {
					shortID = id[:8]
				}

				fmt.Printf("%s│%s       • %s: %s%s%s概率%.0f%%\n",
					Cyan, Reset, shortID, actionColor, action, Reset, prob)
			}
		}

		if whaleNextSellCount > 0 {
			fmt.Printf("%s│%s    ⚠️  预测下回合出货的大资金: %s%d 个%s\n",
				Cyan, Reset, Red, whaleNextSellCount, Reset)
		}
	}

	// ===== 核心洞察 =====
	if len(analysis.KeyInsights) > 0 {
		fmt.Printf("%s│%s  【核心洞察】\n", Cyan, Reset)
		for i, insight := range analysis.KeyInsights {
			fmt.Printf("%s│%s    %s%d. %s%s\n", Cyan, Reset, Yellow, i+1, insight, Reset)
		}
	}

	// ===== 综合风险评分 =====
	riskColor := Green
	if analysis.OverallRiskScore > 70 {
		riskColor = Red
	} else if analysis.OverallRiskScore > 40 {
		riskColor = Yellow
	}

	fmt.Printf("%s│%s  【综合风险评分】%s%.0f/100%s\n",
		Cyan, Reset, riskColor, analysis.OverallRiskScore, Reset)

	fmt.Printf("%s└───────────────────────────────────────────────────────┘%s\n", Cyan, Reset)
}

// 排名评级系统
type PlayerRank struct {
	Grade       string // S, A, B, C, D, F
	Title       string // 称号
	Score       int    // 总分
	ProfitScore int    // 收益分
	TimingScore int    // 时机分
	RiskScore   int    // 风控分
	Comment     string // 评语
}

func calculatePlayerRank(state *GameState) PlayerRank {
	rank := PlayerRank{}

	// 计算最终净资产和收益率 (Gross - MarginDebt)
	finalAsset := float64(state.PlayerShares)*state.Price + state.PlayerCash - state.MarginDebt
	finalProfit := (finalAsset - state.InitialAsset) / state.InitialAsset

	// 1. 收益分 (0-40分)
	profitScore := 0
	if finalProfit > 0.3 {
		profitScore = 40
	} else if finalProfit > 0.25 {
		profitScore = 35
	} else if finalProfit > 0.2 {
		profitScore = 30
	} else if finalProfit > 0.15 {
		profitScore = 25
	} else if finalProfit > 0.1 {
		profitScore = 20
	} else if finalProfit > 0.05 {
		profitScore = 15
	} else if finalProfit > 0 {
		profitScore = 10
	} else if finalProfit > -0.1 {
		profitScore = 5
	} else {
		profitScore = 0
	}
	rank.ProfitScore = profitScore

	// 2. 时机分 (0-30分) - 是否成功逃顶
	timingScore := 0
	hasSoldBefore := state.PlayerSoldDay > 0 // 是否卖出过
	isHoldingAtEnd := state.PlayerShares > 0 // 最终是否持仓

	if state.IsCrashed {
		// 崩盘时
		if hasSoldBefore && !isHoldingAtEnd {
			timingScore = 30 // 完美逃顶，空仓躲过崩盘
		} else if hasSoldBefore && isHoldingAtEnd {
			timingScore = 15 // 卖出过但又买回来，被部分埋
		} else {
			timingScore = 0 // 从未卖出，全程持仓被埋
		}
	} else {
		// 没崩盘
		if hasSoldBefore {
			// 卖出过
			if state.PlayerSoldDay <= 5 {
				timingScore = 10 // 太早卖出
			} else if state.PlayerSoldDay <= 10 {
				timingScore = 20 // 中期卖出，不错
			} else {
				timingScore = 25 // 后期卖出，很好
			}
		} else {
			// 从未卖出，格局到底
			if finalProfit > 0.2 {
				timingScore = 25 // 格局到底赚麻了
			} else if finalProfit > 0.1 {
				timingScore = 15
			} else {
				timingScore = 5
			}
		}
	}
	rank.TimingScore = timingScore

	// 3. 风控分 (0-30分) - 风险控制能力
	riskScore := 0
	if state.IsCrashed {
		if hasSoldBefore && !isHoldingAtEnd {
			riskScore = 30 // 完美风控，空仓躲过崩盘
		} else if hasSoldBefore && isHoldingAtEnd {
			riskScore = 10 // 卖出过但又买回来，风控不够坚决
		} else {
			riskScore = 0 // 风控失败，全程持仓
		}
	} else {
		if hasSoldBefore {
			// 虽然没崩盘，但主动止盈也不错
			if finalProfit > 0.1 {
				riskScore = 25
			} else if finalProfit > 0 {
				riskScore = 20
			} else {
				riskScore = 10 // 止损了
			}
		} else {
			// 格局到底
			if finalProfit > 0.2 {
				riskScore = 28 // 大胆且成功
			} else if finalProfit > 0.1 {
				riskScore = 22
			} else if finalProfit > 0 {
				riskScore = 15
			} else {
				riskScore = 5 // 该跑不跑
			}
		}
	}
	rank.RiskScore = riskScore

	// 总分
	rank.Score = profitScore + timingScore + riskScore

	// 评级
	if rank.Score >= 85 {
		rank.Grade = "S"
		rank.Title = "🏆 股神"
		rank.Comment = "完美操作！你对市场的理解已臻化境，精准逃顶，收益最大化。巴菲特看了都要点赞！"
	} else if rank.Score >= 70 {
		rank.Grade = "A"
		rank.Title = "💎 高手"
		rank.Comment = "优秀表现！你成功把握了市场节奏，在风险可控的情况下获得了可观收益。"
	} else if rank.Score >= 55 {
		rank.Grade = "B"
		rank.Title = "📈 老手"
		rank.Comment = "不错的操作。虽然还有提升空间，但你展现了一定的市场敏感度和决策能力。"
	} else if rank.Score >= 40 {
		rank.Grade = "C"
		rank.Title = "🔰 新手"
		rank.Comment = "及格水平。你对市场有基本认知，但在时机把握和风险控制上还需要更多历练。"
	} else if rank.Score >= 25 {
		rank.Grade = "D"
		rank.Title = "🌿 韭菜"
		rank.Comment = "需要改进。你的操作过于冲动或保守，没有很好地平衡收益与风险。多复盘，多思考。"
	} else {
		rank.Grade = "F"
		rank.Title = "💀 血本无归"
		rank.Comment = "惨痛教训。你被市场狠狠上了一课。记住：贪婪和恐惧是投资最大的敌人。"
	}

	return rank
}

// 计算零和博弈的买卖压力（纯零和机制）
// 直接使用AI的真实卖出数据，反映真实供需关系
func calculateZeroSumPressure(state *GameState, todaySellPressure int, todayWhaleSell int, todayRetailSell int) {
	// 卖压就是AI实际卖出的股数
	state.SellPressure = todaySellPressure
	state.WhaleSelling = todayWhaleSell
	state.RetailSelling = todayRetailSell

	// 零和博弈：没有外部买家，买压=0
	// 只有玩家的买入（当空仓时买入）算作买压
	// 注意：这里不计算假想的买压，因为零和博弈中，
	// 价格下跌本身就会吸引新的买家（玩家抄底）
	state.BuyPressure = 0
	state.WhaleBuying = 0
	state.RetailBuying = 0

	// 如果玩家空仓，且价格跌得多，可以假设有买意愿
	// 这反映了零和博弈中"有人卖必有人买"的规律
	if state.PlayerShares == 0 && state.Price < state.LastPrice*0.9 {
		// 价格跌超10%，假设有抄底意愿
		state.BuyPressure = todaySellPressure / 3 // 买压为卖压的1/3（供大于求）
	}
}

// 辅助函数：计算动态流动性
func calculateDynamicLiquidity(state *GameState) float64 {
	// 基础流动性由事件情绪决定
	baseLiquidity := 0.15 * state.CurrentEvent.Sentiment // 6%-22.5%

	// 如果已有大量AI逃跑，流动性急剧下降
	totalShares := 0
	for _, ai := range state.AIs {
		totalShares += ai.Shares
	}
	// 玩家如果持仓，也计入总持仓
	totalShares += state.PlayerShares

	if totalShares > 0 {
		escapedRatio := float64(state.TotalEscapedAIShares) / float64(totalShares+state.TotalEscapedAIShares)
		if escapedRatio > 0.15 {
			baseLiquidity *= 0.7 // 逃跑>15%，流动性下降30%
		}
		if escapedRatio > 0.3 {
			baseLiquidity *= 0.5 // 逃跑>30%，流动性再下降50%
		}
	}

	// 价格连续下跌时，流动性枯竭
	if state.ConsecutiveFallDays >= 2 {
		baseLiquidity *= 0.8
	}

	return max(baseLiquidity, 0.03) // 最低3%流动性
}

// 辅助函数：min/max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func processTurn(state *GameState) {
	fmt.Printf("\n"+Cyan+"  [ 第%d天 %s 撮合清算中... ]"+Reset+"\n", state.Day, state.Session)
	time.Sleep(400 * time.Millisecond)

	// 1. 清空上回合盘口数据
	state.TotalBuyDemandCash = 0
	state.TotalSellSupplyShares = 0
	state.WhaleBuying, state.WhaleSelling = 0, 0
	state.RetailBuying, state.RetailSelling = 0, 0
	state.FakeBuyPressure, state.FakeSellPressure = 0, 0

	priceChangePct := (state.Price - state.LastPrice) / state.LastPrice

	// 妖股逻辑更新：连涨3天判定为妖股
	if priceChangePct > 0.05 {
		state.ConsecutiveGrowthDays++
	} else if priceChangePct <= 0 {
		state.ConsecutiveGrowthDays = 0
		state.IsMonsterStock = false
	}
	if state.ConsecutiveGrowthDays >= 3 {
		state.IsMonsterStock = true
	}

	// 连续下跌与崩盘预警
	if priceChangePct < -0.05 {
		state.ConsecutiveFallDays++
	} else if priceChangePct > 0 {
		state.ConsecutiveFallDays = 0
	}
	if state.ConsecutiveFallDays >= 2 {
		state.CrashWarningLevel = max_int(state.CrashWarningLevel, 3)
	}

	// 2. AI 制定挂单决策
	for _, ai := range state.AIs {
		ai.OrderType = ""
		ai.OrderShares = 0
		ai.OrderCash = 0
		ai.StatusFlag = ""

		profitRatio := 0.0
		if ai.Cost > 0 {
			profitRatio = state.Price / ai.Cost
		}
		eventFear := state.CurrentEvent.FearModifier

		// === 特殊机制：庄家配资爆仓 ===
		if ai.MarginDebt > 0 {
			netAsset := float64(ai.Shares)*state.Price + ai.Cash - ai.MarginDebt
			if netAsset <= 0 {
				ai.StatusFlag = "ForcedLiquidation"
				ai.OrderType = "Sell"
				ai.OrderShares = ai.Shares
				ai.Cash = 0
				continue
			}
		}

		// === 特殊机制：国家队救市 ===
		if ai.SubType == "国家队" {
			if state.CrashWarningLevel >= 4 && state.ConsecutiveFallDays >= 2 {
				ai.StatusFlag = "Bailout"
				ai.OrderType = "Buy"
				ai.OrderCash = ai.Cash
			}
			continue
		}

		// === 妖股共识机制 ===
		activeFearBasis := ai.FearBasis
		activeTargetProfit := ai.TargetProfit
		if state.IsMonsterStock && (ai.SubType == "新韭" || ai.SubType == "打板") {
			activeFearBasis = 0.0
			activeTargetProfit = 100.0
		}

		if ai.Shares == 0 {
			reenter := false
			investRatio := 1.0
			switch ai.Type {
			case "Whale":
				if ai.SubType == "刺客" {
					if state.NextEvent.Sentiment > 1.3 && rand.Float64() < 0.80 {
						reenter = true
					} else if state.Price < ai.Cost*0.60 && rand.Float64() < 0.80 {
						reenter = true
					}
				} else { // 机构长线
					if state.Price < ai.Cost*0.50 {
						reenter = true
					}
				}
			case "Quant":
				if ai.SubType == "网格" {
					if state.Price < ai.Cost*0.97 && rand.Float64() < 0.8 {
						reenter = true
						investRatio = 0.2
						ai.StatusFlag = "GridTrading"
					}
				} else if ai.SubType == "打板" {
					if state.ConsecutiveFallDays >= 2 && priceChangePct > 0.05 && rand.Float64() < 0.75 {
						reenter = true
						investRatio = 0.5
					} else if state.IsMonsterStock {
						reenter = true
						investRatio = 1.0
					}
				}
			default: // Retail
				if (state.CurrentEvent.Sentiment > 1.1 || priceChangePct > 0.08) && rand.Float64() < 0.50 {
					reenter = true
				} else if state.Price < ai.Cost*0.50 && rand.Float64() < 0.05 {
					reenter = true
				}
				if state.IsMonsterStock && ai.SubType == "新韭" {
					reenter = true
				}
			}

			if reenter && ai.Cash > 1000 {
				ai.OrderType = "Buy"
				ai.OrderCash = ai.Cash * investRatio
				if ai.Type == "Whale" && state.Session == "早盘" && rand.Float64() < 0.3 {
					ai.StatusFlag = "Spoofing"
					state.FakeSellPressure += int(ai.OrderCash/state.Price) * 3
				}
			}
		} else {
			shouldSell := false
			sellRatio := 0.0
			if ai.Type == "Whale" {
				if ai.SubType == "刺客" {
					if profitRatio > activeTargetProfit {
						shouldSell = true
						sellRatio = 1.0
					} else if profitRatio > 1.1 && rand.Float64() < 0.3 {
						shouldSell = true
						sellRatio = 0.4
						ai.StatusFlag = "Shakeout"
					} else if eventFear > 2.0 {
						shouldSell = true
						sellRatio = 1.0
					}
				} else { // 机构长线
					if profitRatio > activeTargetProfit {
						shouldSell = true
						sellRatio = 0.5
					}
				}
			} else if ai.Type == "Quant" {
				if ai.SubType == "网格" {
					if profitRatio > 1.03 {
						shouldSell = true
						sellRatio = 0.5
						ai.StatusFlag = "GridTrading"
					} else if profitRatio < 0.97 && ai.Cash > 1000 {
						shouldSell = false
						ai.OrderType = "Buy"
						ai.OrderCash = ai.Cash * 0.5
						ai.StatusFlag = "GridTrading"
					}
				} else {
					if profitRatio > activeTargetProfit || rand.Float64() < 0.05 {
						shouldSell = true
						sellRatio = 1.0
					}
				}
			} else { // Retail
				panicProb := activeFearBasis * eventFear * 0.05
				if ai.SubType == "新韭" {
					if profitRatio > activeTargetProfit && rand.Float64() < 0.6 {
						shouldSell = true
						sellRatio = 0.5
					} else if profitRatio < 0.95 && priceChangePct < -0.05 && rand.Float64() < 0.7 {
						shouldSell = true
						sellRatio = 1.0
					} else if rand.Float64() < panicProb {
						shouldSell = true
						sellRatio = 1.0
					}
				} else { // 老散
					if profitRatio > activeTargetProfit && rand.Float64() < 0.8 {
						shouldSell = true
						sellRatio = 1.0
					}
				}
			}
			if shouldSell {
				ai.OrderType = "Sell"
				ai.OrderShares = int(float64(ai.Shares) * sellRatio)
				if ai.OrderShares == 0 {
					ai.OrderShares = ai.Shares
				}
			}
		}
	}

	// 3. 汇总全场订单
	totalSellShares := state.PlayerOrderShares
	totalBuyCash := state.PlayerOrderCash
	for _, ai := range state.AIs {
		if ai.OrderType == "Sell" {
			totalSellShares += ai.OrderShares
		}
		if ai.OrderType == "Buy" {
			totalBuyCash += ai.OrderCash
		}
	}
	state.TotalBuyDemandCash = totalBuyCash
	state.TotalSellSupplyShares = totalSellShares
	totalBuyDemandShares := int(totalBuyCash / state.Price)
	state.TotalBuyDemandShares = totalBuyDemandShares

	if totalBuyDemandShares == 0 && totalSellShares == 0 {
		state.Price = state.Price * (1.0 + (rand.Float64()*0.01 - 0.005))
		state.AddChronicle()
		advanceTime(state)
		return
	}

	// 4. 定价引擎
	demandRatio := 1.0
	if totalSellShares > 0 {
		demandRatio = float64(totalBuyDemandShares) / float64(totalSellShares)
	} else {
		demandRatio = 5.0
	}

	priceModifier := 0.0
	if demandRatio > 2.0 {
		priceModifier = 0.10
	} else if demandRatio > 1.2 {
		priceModifier = 0.05
	} else if demandRatio > 0.8 {
		priceModifier = (rand.Float64()*0.04 - 0.02)
	} else if demandRatio > 0.5 {
		priceModifier = -0.05
	} else {
		priceModifier = -0.10
	}

	state.LastPrice = state.Price
	state.Price = state.Price * (1.0 + priceModifier)
	if state.Price < 0.1 {
		state.Price = 0.1
	}

	// 5. 撮合
	var buyProration, sellProration float64 = 1.0, 1.0
	realBuyExpectedShares := int(totalBuyCash / state.Price)
	if realBuyExpectedShares > 0 && totalSellShares > 0 {
		if realBuyExpectedShares > totalSellShares {
			sellProration = 1.0
			buyProration = float64(totalSellShares) / float64(realBuyExpectedShares)
		} else {
			buyProration = 1.0
			sellProration = float64(realBuyExpectedShares) / float64(totalSellShares)
		}
	} else if realBuyExpectedShares <= 0 {
		sellProration = 0.0
	} else if totalSellShares <= 0 {
		buyProration = 0.0
	}

	// 6. 账户清算 (Fixing the Bug here)
	whaleStatusSnapshot := "沉静"
	for _, ai := range state.AIs {
		if ai.Type == "Whale" && ai.StatusFlag != "" {
			whaleStatusSnapshot = ai.StatusFlag
			break
		}
	}
	marketContext := "正常"
	if state.IsMonsterStock {
		marketContext = "妖股狂热"
	} else if state.CrashWarningLevel >= 3 {
		marketContext = "极端恐慌"
	}

	if state.PlayerOrderType == "Sell" {
		actualSell := int(float64(state.PlayerOrderShares) * sellProration)
		if actualSell > 0 {
			revenue := float64(actualSell) * state.Price
			state.PlayerCash += revenue
			state.PlayerShares -= actualSell // FIXED: Deduct shares
			state.PlayerAvailableShares += (state.PlayerOrderShares - actualSell)
			state.LastActionMessage = fmt.Sprintf(" | 撮合成功: 卖出 %d 股", actualSell)
			state.TradePoints = append(state.TradePoints, TradePoint{
				Day: state.Day, Session: state.Session, Action: "Sell", Price: state.Price, Shares: actualSell, WhaleStatus: whaleStatusSnapshot, MarketContext: marketContext,
			})
		} else {
			state.PlayerAvailableShares += state.PlayerOrderShares
			state.LastActionMessage = "[系统退单] 撮合失败：无人接盘"
		}
	} else if state.PlayerOrderType == "Buy" {
		expectedBuy := int(state.PlayerOrderCash / state.Price)
		actualBuy := int(float64(expectedBuy) * buyProration)
		if actualBuy > 0 {
			cost := float64(actualBuy) * state.Price
			state.PlayerShares += actualBuy
			state.PlayerFrozenShares += actualBuy
			state.PlayerCash += (state.PlayerOrderCash - cost)
			totalCostBefore := float64(state.PlayerShares-actualBuy) * state.PlayerAvgCost
			state.PlayerAvgCost = (totalCostBefore + cost) / float64(state.PlayerShares)
			state.LastActionMessage = fmt.Sprintf("[交割单] 抢到 %d 股", actualBuy)
			state.TradePoints = append(state.TradePoints, TradePoint{
				Day: state.Day, Session: state.Session, Action: "Buy", Price: state.Price, Shares: actualBuy, WhaleStatus: whaleStatusSnapshot, MarketContext: marketContext,
			})
		} else {
			state.PlayerCash += state.PlayerOrderCash
			state.LastActionMessage = "[系统退单] 没买到"
		}
	}
	state.PlayerOrderType = ""
	state.PlayerOrderShares = 0
	state.PlayerOrderCash = 0

	for _, ai := range state.AIs {
		if ai.OrderType == "Sell" {
			actualSell := int(float64(ai.OrderShares) * sellProration)
			ai.Shares -= actualSell
			ai.Cash += float64(actualSell) * state.Price
			if ai.Type == "Whale" {
				state.WhaleSelling += actualSell
				if actualSell > 1000 {
					state.AddLog(fmt.Sprintf("%s 🐋 游资砸盘: %s 抛售 %d 股%s", Red, ai.Name, actualSell, Reset))
				}
			}
			if ai.Type == "Retail" {
				state.RetailSelling += actualSell
			}
			if ai.StatusFlag == "ForcedLiquidation" {
				state.SpeakLog(fmt.Sprintf("%s 💥 爆仓强平: %s 资金链断裂被强制清场%s", Red, ai.Name, Reset))
			}
		} else if ai.OrderType == "Buy" {
			actualBuy := int(float64(int(ai.OrderCash/state.Price)) * buyProration)
			cost := float64(actualBuy) * state.Price
			if actualBuy > 0 {
				oldTotal := float64(ai.Shares) * ai.Cost
				ai.Shares += actualBuy
				ai.Cost = (oldTotal + cost) / float64(ai.Shares)
				ai.Cash -= cost
			if ai.Type == "Whale" && actualBuy > 1000 {
				state.SpeakLog(fmt.Sprintf("%s 🐋 游资进场: %s 抢筹 %d 股%s", Green, ai.Name, actualBuy, Reset))
			}
			if ai.StatusFlag == "Bailout" {
				state.SpeakLog(fmt.Sprintf("%s 🛡️ 国家队救市: %s 开启无限额护盘%s", Red, ai.Name, Reset))
			}
			}
			if ai.Type == "Whale" {
				state.WhaleBuying += actualBuy
			}
			if ai.Type == "Retail" {
				state.RetailBuying += actualBuy
			}
		}
		if ai.Shares == 0 {
			ai.HasSold = true
		} else {
			ai.HasSold = false
		}
	}
	// 价格异动日志
	if math.Abs(priceModifier) > 0.08 {
		color := Green
		if priceModifier < 0 {
			color = Red
		}
		state.AddLog(fmt.Sprintf("%s 📊 价格剧震: 本回合波幅达到 %.1f%%%s", color, priceModifier*100, Reset))
	}
	state.BuyPressure = int(totalBuyCash / state.Price)
	state.SellPressure = totalSellShares
	state.TotalSellSupplyShares += state.FakeSellPressure
	state.TotalBuyDemandShares += state.FakeBuyPressure
	state.PriceHistory = append(state.PriceHistory, state.Price)

	// 更新日内高低价
	if state.Price > state.DayHigh || state.Session == "早盘" {
		state.DayHigh = state.Price
	}
	if state.Price < state.DayLow || state.Session == "早盘" || state.DayLow == 0 {
		state.DayLow = state.Price
	}
	// 记录成交量 (撮合成功的总量)
	turnVolume := int(float64(realBuyExpectedShares) * buyProration)
	state.VolumeHistory = append(state.VolumeHistory, turnVolume)

	state.AddChronicle()
	state.IntelUsedThisTurn = false // 重置情报使用状态
	advanceTime(state)
}

func advanceTime(state *GameState) {
	if state.Session == "早盘" {
		state.Session = "盘中上午"
	} else if state.Session == "盘中上午" {
		state.Session = "盘中下午"
	} else if state.Session == "盘中下午" {
		state.Session = "尾盘"
	} else {
		state.Day++
		state.Session = "早盘"
		state.PlayerAvailableShares += state.PlayerFrozenShares
		state.PlayerFrozenShares = 0
		state.CurrentEvent = state.NextEvent
		state.NextEvent = selectNextEventByMarkov(state.CurrentEvent, Events)
		state.AddLog(fmt.Sprintf("%s 💡 市场风向变化: %s%s%s", Purple, Yellow, state.CurrentEvent.Title, Reset))

		// 生成今日运势
		fortunes := []string{
			"宜：格局，忌：核按钮",
			"宜：空仓观望，忌：满仓梭哈",
			"宜：建仓，忌：贪婪",
			"宜：止损，忌：死扛",
			"宜：吃肉，忌：关灯面",
		}
		state.DailyFortune = fortunes[rand.Intn(len(fortunes))]

		// 30% 概率产生市场传闻（暗示明天的事件）
		if rand.Float64() < 0.3 {
			rumors := map[EventCategory]string{
				ExtremeOptimistic: "传闻：政策面可能有重磅利好正在路上...",
				MildOptimistic:    "传闻：某行业大资金正在悄悄建仓...",
				Neutral:           "传闻：明天可能是个无聊的平盘日...",
				MildPessimistic:   "传闻：市场情绪似乎在悄悄转冷...",
				ExtremePanic:      "传闻：小心，今晚可能要出重大利空雷...",
			}
			state.AddLog(fmt.Sprintf("%s 🕵️ 小道消息: %s%s", Cyan, rumors[state.NextEvent.Category], Reset))
		}
	}

	// 杠杆爆仓检测
	if state.MarginDebt > 0 {
		netAsset := float64(state.PlayerShares)*state.Price + state.PlayerCash - state.MarginDebt
		if netAsset <= 0 {
			state.IsMarginCalled = true
			state.IsGameOver = true
			state.IsCrashed = true
			fmt.Printf("\n" + Red + ">>> 💥【强制平仓】你已资不抵债！配资公司强平了你的所有爆仓股票！\n" + Reset)
			time.Sleep(2 * time.Second)
			return
		}
	}

	if state.Day > state.MaxDays {
		state.IsGameOver = true
	}
}

func min_int(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max_int(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// 渲染增强版K线图表
func renderEnhancedKLine(state *GameState) {
	if len(state.PriceHistory) < 5 || len(state.VolumeHistory) < 1 {
		return
	}

	// 取最近40个数据点
	displayPoints := 40
	if len(state.PriceHistory) < displayPoints {
		displayPoints = len(state.PriceHistory)
	}
	if len(state.VolumeHistory) < displayPoints {
		displayPoints = len(state.VolumeHistory)
	}
	history := state.PriceHistory[len(state.PriceHistory)-displayPoints:]
	volumeHistory := state.VolumeHistory[len(state.VolumeHistory)-displayPoints:]

	// 计算价格范围
	minPrice := history[0]
	maxPrice := history[0]
	for _, p := range history {
		if p < minPrice {
			minPrice = p
		}
		if p > maxPrice {
			maxPrice = p
		}
	}
	priceRange := maxPrice - minPrice
	if priceRange == 0 {
		priceRange = 1
	}

	// 图表高度
	chartHeight := 12

	fmt.Printf("\n%s┌─────────────────── 📈 增强K线图 (最近%d点) ─────────────────┐%s\n",
		Cyan, displayPoints, Reset)

	// 渲染价格K线（从上到下）
	for row := chartHeight; row >= 0; row-- {
		// 左侧价格标签
		priceAtRow := minPrice + (float64(row)/float64(chartHeight))*priceRange
		fmt.Printf("%s│%s %s$%6.2f%s │", Cyan, Reset, Yellow, priceAtRow, Reset)

		// 绘制K线
		for i, price := range history {
			normalized := (price - minPrice) / priceRange
			barHeight := int(normalized * float64(chartHeight))

			// 判断涨跌
			isRising := true
			if i > 0 {
				isRising = price >= history[i-1]
			}

			// 当前行是否有K线
			var char string
			if barHeight == row {
				if isRising {
					char = Green + "█" + Reset
				} else {
					char = Red + "█" + Reset
				}
			} else if barHeight > row {
				if isRising {
					char = Green + "│" + Reset
				} else {
					char = Red + "│" + Reset
				}
			} else {
				char = " "
			}

			fmt.Print(char)
		}

		// 右侧标注
		if row == chartHeight {
			fmt.Printf(" %s%.2f%s", Red, maxPrice, Reset)
		} else if row == 0 {
			fmt.Printf(" %s%.2f%s", Green, minPrice, Reset)
		} else if row == chartHeight/2 {
			avgPrice := (maxPrice + minPrice) / 2
			fmt.Printf(" %s%.2f%s", Yellow, avgPrice, Reset)
		}

		fmt.Println()
	}

	// 底部边框
	fmt.Printf("%s│%s ────────┼", Cyan, Reset)
	for range history {
		fmt.Print("─")
	}
	fmt.Println()

	// 成交量柱状图
	fmt.Printf("%s│%s %s成交量%s  │", Cyan, Reset, Yellow, Reset)

	maxVolume := volumeHistory[0]
	for _, v := range volumeHistory {
		if v > maxVolume {
			maxVolume = v
		}
	}

	volumeBars := []rune(" ▁▂▃▄▅▆▇█")
	for _, vol := range volumeHistory {
		normalized := float64(vol) / float64(maxVolume)
		idx := int(normalized * float64(len(volumeBars)-1))
		if idx >= len(volumeBars) {
			idx = len(volumeBars) - 1
		}
		fmt.Printf("%s%c%s", Blue, volumeBars[idx], Reset)
	}
	fmt.Println()

	// 统计信息
	fmt.Printf("%s│%s  %s涨%s: %sMA5=%.2f%s  %sMA10=%.2f%s  %s当前=%.2f%s\n",
		Cyan, Reset, Green, Reset,
		Yellow, calculateMA(history, 5), Reset,
		Yellow, calculateMA(history, 10), Reset,
		Yellow, state.Price, Reset)

	fmt.Printf("%s└───────────────────────────────────────────────────────────────┘%s\n", Cyan, Reset)
}

// 计算移动平均线
func calculateMA(prices []float64, period int) float64 {
	if len(prices) < period {
		period = len(prices)
	}
	if period == 0 {
		return 0
	}

	sum := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		sum += prices[i]
	}
	return sum / float64(period)
}

// 渲染 K线趋势图 (Sparkline) - 保留原版用于简单显示
func renderSparkline(history []float64) string {
	if len(history) == 0 {
		return ""
	}
	bars := []rune(" ▂▃▄▅▆▇█")
	min := history[0]
	max := history[0]
	for _, v := range history {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	if max == min {
		return strings.Repeat("▄", len(history))
	}

	var sb strings.Builder
	for _, v := range history {
		idx := int((v - min) / (max - min) * float64(len(bars)-1))
		sb.WriteRune(bars[idx])
	}
	return sb.String()
}

// stripAnsi 去除 ANSI 转义码，用于计算显示宽度
func stripAnsi(s string) string {
	var out strings.Builder
	inEsc := false
	for _, r := range s {
		if r == '\033' {
			inEsc = true
			continue
		}
		if inEsc {
			if r == 'm' {
				inEsc = false
			}
			continue
		}
		out.WriteRune(r)
	}
	return out.String()
}

// runeWidth 粗略计算字符串显示宽度（中文=2，其余=1）
func runeWidth(s string) int {
	w := 0
	for _, r := range s {
		if r > 0x7F {
			w += 2
		} else {
			w++
		}
	}
	return w
}

// 渲染筹码分布图 (Cost Distribution / Volume by Price)
func renderCostDistribution(state *GameState) {
	type Holder struct {
		cost   float64
		shares int
	}
	holders := []Holder{}
	if state.PlayerShares > 0 {
		holders = append(holders, Holder{state.PlayerAvgCost, state.PlayerShares})
	}
	for _, ai := range state.AIs {
		if !ai.HasSold && ai.Shares > 0 {
			holders = append(holders, Holder{ai.Cost, ai.Shares})
		}
	}
	if len(holders) == 0 {
		fmt.Println("  (暂无持仓数据)")
		return
	}

	minP, maxP := state.Price, state.Price
	for _, h := range holders {
		if h.cost < minP {
			minP = h.cost
		}
		if h.cost > maxP {
			maxP = h.cost
		}
	}
	minP = minP * 0.95
	maxP = maxP * 1.05
	if maxP <= minP {
		maxP = minP + 1.0
	}

	numBuckets := 9
	bucketSize := (maxP - minP) / float64(numBuckets)
	buckets := make([]float64, numBuckets)
	for _, h := range holders {
		idx := int((h.cost - minP) / bucketSize)
		if idx < 0 {
			idx = 0
		}
		if idx >= numBuckets {
			idx = numBuckets - 1
		}
		buckets[idx] += float64(h.shares)
	}
	maxVal := 0.0
	for _, v := range buckets {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		return
	}

	barMaxWidth := 18
	fmt.Printf("  %s筹码分布图 (持仓成本 vs 仓量)%s\n", Cyan, Reset)
	for i := numBuckets - 1; i >= 0; i-- {
		price := minP + float64(i)*bucketSize + bucketSize/2
		barLen := int(buckets[i] / maxVal * float64(barMaxWidth))
		bar := strings.Repeat("█", barLen)

		isCurrentPriceBucket := price >= state.Price-bucketSize && price <= state.Price+bucketSize
		barColor := Green
		marker := ""
		if isCurrentPriceBucket {
			barColor = Yellow
			marker = Yellow + " ◄ 现价" + Reset
		} else if price > state.Price {
			barColor = Red // 上方套牢盘=阻力
		}
		fmt.Printf("  %s$%5.2f%s │%s%-*s%s│%s\n",
			barColor, price, Reset,
			barColor, barMaxWidth, bar, Reset,
			marker)
	}
}

// renderEventCard 渲染带边框的事件卡牌（animate=true 时有翻牌动画）
func renderEventCard(event Event, animate bool) {
	if animate {
		fmt.Print("\033[H\033[2J")
		fmt.Printf("\n\n  %s╭────────────────────────────────────╮%s\n", Yellow, Reset)
		fmt.Printf("  %s│        📰  翻牌中...               │%s\n", Yellow, Reset)
		fmt.Printf("  %s╰────────────────────────────────────╯%s\n", Yellow, Reset)
		time.Sleep(600 * time.Millisecond)
	}

	sentimentLabel := "😐 中性"
	sentimentColor := Yellow
	switch {
	case event.Sentiment > 1.3:
		sentimentLabel = "🚀 强烈乐观"
		sentimentColor = Red
	case event.Sentiment > 1.1:
		sentimentLabel = "😊 温和乐观"
		sentimentColor = Green
	case event.Sentiment < 0.7:
		sentimentLabel = "☠️  极度恐慌"
		sentimentColor = Red
	case event.Sentiment < 0.9:
		sentimentLabel = "😨 温和悲观"
		sentimentColor = Yellow
	}

	fearBar := ""
	fearLevel := int(event.FearModifier / 3.5 * 10)
	if fearLevel > 10 {
		fearLevel = 10
	}
	fearBar = strings.Repeat("▓", fearLevel) + strings.Repeat("░", 10-fearLevel)

	fmt.Printf("  %s╭──────────────────────────────────────────╮%s\n", Purple, Reset)
	fmt.Printf("  %s│%s  %-40s%s%s│%s\n", Purple, Yellow, event.Title, Reset, Purple, Reset)
	fmt.Printf("  %s├──────────────────────────────────────────┤%s\n", Purple, Reset)
	fmt.Printf("  %s│%s  情绪: %-30s%s%s│%s\n", Purple, sentimentColor, sentimentLabel, Reset, Purple, Reset)
	fmt.Printf("  %s│%s  恐慌: [%s] %.1fx%s%-15s%s%s│%s\n", Purple, sentimentColor, fearBar, event.FearModifier, Reset, "", Purple, Reset, Reset)
	fmt.Printf("  %s├──────────────────────────────────────────┤%s\n", Purple, Reset)

	// 描述文字自动换行（每行18个汉字宽度）
	desc := []rune(event.Desc)
	lineWidth := 20 // rune 数
	for len(desc) > 0 {
		end := lineWidth
		if end > len(desc) {
			end = len(desc)
		}
		chunk := string(desc[:end])
		desc = desc[end:]
		fmt.Printf("  %s│%s  %-40s%s%s│%s\n", Purple, Reset, chunk, Reset, Purple, Reset)
	}
	fmt.Printf("  %s╰──────────────────────────────────────────╯%s\n", Purple, Reset)
}

// 渲染关键警报（闪烁效果）
func renderCriticalAlerts(state *GameState) {
	const Blink = "\033[5m"     // ANSI闪烁代码
	const Bold = "\033[1m"      // 粗体
	const BlinkOff = "\033[25m" // 关闭闪烁

	alerts := []string{}

	// 1. 崩盘预警
	if state.CrashWarningLevel >= 4 {
		alerts = append(alerts, fmt.Sprintf("%s%s%s🚨🚨 市场崩盘警报！立即离场！ 🚨🚨%s%s",
			Bold, Blink, Red, BlinkOff, Reset))
	} else if state.CrashWarningLevel >= 3 {
		alerts = append(alerts, fmt.Sprintf("%s%s%s⚠️  高度崩盘风险，建议减仓%s%s",
			Bold, Blink, Red, BlinkOff, Reset))
	}

	// 2. 爆仓风险
	if state.MarginDebt > 0 {
		netAsset := float64(state.PlayerShares)*state.Price + state.PlayerCash
		if netAsset <= state.MarginDebt*1.1 {
			alerts = append(alerts, fmt.Sprintf("%s%s%s💥💥 爆仓警告！距离强平仅%.0f%% 💥💥%s%s",
				Bold, Blink, Red, (netAsset/state.MarginDebt-1)*100, BlinkOff, Reset))
		} else if netAsset <= state.MarginDebt*1.3 {
			alerts = append(alerts, fmt.Sprintf("%s%s%s⚡ 接近爆仓线，警惕风险！%s%s",
				Bold, Blink, Yellow, BlinkOff, Reset))
		}
	}

	// 3. 残局模式倒计时
	if state.IsEndgameMode && state.EndgameScenario != nil {
		currentTurn := calculateTurnNumberHelper(state.Day, state.Session)
		elapsedTurns := currentTurn - state.EndgameStartTurn
		remainingTurns := state.EndgameScenario.TimeLimit - elapsedTurns

		if remainingTurns <= 3 && remainingTurns > 0 {
			alerts = append(alerts, fmt.Sprintf("%s%s%s⏰ 残局倒计时：仅剩%d回合！%s%s",
				Bold, Blink, Red, remainingTurns, BlinkOff, Reset))
		}
	}

	// 4. 大资金集体出逃
	whaleEscaped := 0
	for _, ai := range state.AIs {
		if ai.Type == "Whale" && ai.HasSold {
			whaleEscaped++
		}
	}
	totalWhales := 0
	for _, ai := range state.AIs {
		if ai.Type == "Whale" {
			totalWhales++
		}
	}
	if totalWhales > 0 && float64(whaleEscaped)/float64(totalWhales) >= 0.6 {
		alerts = append(alerts, fmt.Sprintf("%s%s%s🐋 大资金集体出逃（%d/%d已离场）%s%s",
			Bold, Blink, Red, whaleEscaped, totalWhales, BlinkOff, Reset))
	}

	// 5. 极端波动
	if len(state.PriceHistory) >= 2 {
		lastPrice := state.PriceHistory[len(state.PriceHistory)-2]
		change := math.Abs((state.Price - lastPrice) / lastPrice)
		if change >= 0.15 {
			direction := "暴涨"
			if state.Price < lastPrice {
				direction = "暴跌"
			}
			alerts = append(alerts, fmt.Sprintf("%s%s%s⚡ 单回合%s%.0f%%，极端波动！%s%s",
				Bold, Blink, Yellow, direction, change*100, BlinkOff, Reset))
		}
	}

	// 6. 连续跌停风险
	if state.ConsecutiveFallDays >= 3 {
		alerts = append(alerts, fmt.Sprintf("%s%s%s📉 已连续下跌%d天，跌停风险！%s%s",
			Bold, Blink, Red, state.ConsecutiveFallDays, BlinkOff, Reset))
	}

	// 显示警报
	if len(alerts) > 0 {
		fmt.Printf("\n%s%s╔═══════════════════ ⚠️  紧急警报 ⚠️  ═══════════════════╗%s%s\n",
			Bold, Red, Reset, Reset)
		for _, alert := range alerts {
			fmt.Printf("%s║%s  %s\n", Red, Reset, alert)
		}
		fmt.Printf("%s╚═══════════════════════════════════════════════════════════╝%s\n", Red, Reset)
	}
}

// 渲染动态风险热力图
func renderRiskHeatmap(state *GameState) {
	// 计算各维度风险值 (0-100)

	// 1. 崩盘风险
	crashRisk := float64(state.CrashWarningLevel) * 25.0
	if crashRisk > 100 {
		crashRisk = 100
	}

	// 2. AI逃跑风险
	escapedCount := 0
	whaleEscaped := 0
	for _, ai := range state.AIs {
		if ai.HasSold {
			escapedCount++
			if ai.Type == "Whale" {
				whaleEscaped++
			}
		}
	}
	aiExodusRisk := (float64(escapedCount) / float64(len(state.AIs))) * 100
	if whaleEscaped >= 3 {
		aiExodusRisk = math.Min(aiExodusRisk+30, 100)
	}

	// 3. 价格波动风险
	priceVolatility := 0.0
	if len(state.PriceHistory) >= 5 {
		recentPrices := state.PriceHistory[len(state.PriceHistory)-5:]
		avgPrice := 0.0
		for _, p := range recentPrices {
			avgPrice += p
		}
		avgPrice /= float64(len(recentPrices))

		variance := 0.0
		for _, p := range recentPrices {
			variance += math.Pow(p-avgPrice, 2)
		}
		stdDev := math.Sqrt(variance / float64(len(recentPrices)))
		priceVolatility = (stdDev / avgPrice) * 1000 // 放大到0-100范围
		if priceVolatility > 100 {
			priceVolatility = 100
		}
	}

	// 4. 杠杆风险
	leverageRisk := 0.0
	if state.MarginDebt > 0 {
		netAsset := float64(state.PlayerShares)*state.Price + state.PlayerCash
		leverageRatio := state.MarginDebt / math.Max(netAsset, 1)
		leverageRisk = leverageRatio * 50 // 2倍杠杆=100风险
		if leverageRisk > 100 {
			leverageRisk = 100
		}
	}

	// 5. 流动性风险
	liquidityRisk := 0.0
	if state.ConsecutiveFallDays >= 3 {
		liquidityRisk = 60.0
	} else if state.ConsecutiveFallDays >= 2 {
		liquidityRisk = 40.0
	}
	// 加上AI逃跑带来的流动性枯竭
	if aiExodusRisk > 50 {
		liquidityRisk += (aiExodusRisk - 50) * 0.8
	}
	if liquidityRisk > 100 {
		liquidityRisk = 100
	}

	// 6. 情绪风险 (FOMO/恐慌)
	sentimentRisk := 0.0
	if state.CurrentEvent.Sentiment > 1.4 {
		sentimentRisk = 80.0 // 极度乐观=高风险
	} else if state.CurrentEvent.Sentiment > 1.2 {
		sentimentRisk = 50.0
	} else if state.CurrentEvent.FearModifier > 2.5 {
		sentimentRisk = 90.0 // 极度恐慌=高风险
	} else if state.CurrentEvent.FearModifier > 2.0 {
		sentimentRisk = 60.0
	}

	// 综合风险评分
	totalRisk := (crashRisk*0.25 + aiExodusRisk*0.20 + priceVolatility*0.15 +
		leverageRisk*0.20 + liquidityRisk*0.10 + sentimentRisk*0.10)

	// 渲染热力图
	fmt.Printf("\n%s╔═══════════════════ 🔥 动态风险热力图 ═══════════════════╗%s\n", Red, Reset)

	// 渲染函数：根据风险值返回渐变色和条形图
	renderRiskBar := func(label string, risk float64) {
		barWidth := 30
		filled := int(risk / 100 * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}

		// 渐变色：绿->黄->橙->红
		var barColor string
		var riskLabel string
		if risk < 25 {
			barColor = Green
			riskLabel = "安全"
		} else if risk < 50 {
			barColor = Yellow
			riskLabel = "警惕"
		} else if risk < 75 {
			barColor = "\033[38;5;208m" // 橙色
			riskLabel = "危险"
		} else {
			barColor = Red
			riskLabel = "极危"
		}

		bar := barColor + strings.Repeat("█", filled) + Reset +
			strings.Repeat("░", barWidth-filled)

		fmt.Printf("%s║%s  %-10s [%s] %s%3.0f%% %s%s\n",
			Red, Reset, label, bar, barColor, risk, riskLabel, Reset)
	}

	// 显示各项风险
	renderRiskBar("崩盘风险", crashRisk)
	renderRiskBar("AI出逃", aiExodusRisk)
	renderRiskBar("价格波动", priceVolatility)
	renderRiskBar("杠杆风险", leverageRisk)
	renderRiskBar("流动性", liquidityRisk)
	renderRiskBar("情绪风险", sentimentRisk)

	fmt.Printf("%s║%s  %s\n", Red, Reset, strings.Repeat("─", 56))

	// 综合评估
	overallColor := Green
	overallLabel := "✅ 相对安全"
	overallIcon := "🟢"
	if totalRisk >= 75 {
		overallColor = Red
		overallLabel = "🚨 极度危险"
		overallIcon = "🔴"
	} else if totalRisk >= 50 {
		overallColor = "\033[38;5;208m"
		overallLabel = "⚠️  高度警惕"
		overallIcon = "🟠"
	} else if totalRisk >= 25 {
		overallColor = Yellow
		overallLabel = "⚡ 谨慎操作"
		overallIcon = "🟡"
	}

	fmt.Printf("%s║%s  %s 综合风险: %s%.0f/100 %s%s\n",
		Red, Reset, overallIcon, overallColor, totalRisk, overallLabel, Reset)

	// 关键建议
	if totalRisk >= 75 && state.PlayerShares > 0 {
		fmt.Printf("%s║%s  %s💥 建议: 立即减仓或清仓离场！%s\n", Red, Reset, Red, Reset)
	} else if leverageRisk > 60 && state.MarginDebt > 0 {
		fmt.Printf("%s║%s  %s⚠️  建议: 杠杆过高，警惕爆仓风险！%s\n", Red, Reset, Yellow, Reset)
	} else if aiExodusRisk > 70 {
		fmt.Printf("%s║%s  %s⚠️  建议: 大量AI已逃离，流动性危机！%s\n", Red, Reset, Yellow, Reset)
	} else if totalRisk < 30 && state.PlayerShares == 0 {
		fmt.Printf("%s║%s  %s💡 提示: 当前风险较低，可考虑入场%s\n", Red, Reset, Green, Reset)
	}

	fmt.Printf("%s╚═══════════════════════════════════════════════════════════╝%s\n", Red, Reset)
}

// 渲染AI资金流向雷达
func renderAIMoneyFlowRadar(state *GameState) {
	// 统计各类AI的资金流向
	whaleInflow := 0.0
	whaleOutflow := 0.0
	quantInflow := 0.0
	quantOutflow := 0.0
	retailInflow := 0.0
	retailOutflow := 0.0

	activeWhales := 0
	activeQuants := 0
	activeRetails := 0

	for _, ai := range state.AIs {
		if ai.HasSold || ai.Shares == 0 {
			// 已出货的AI计入资金流出
			if ai.Type == "Whale" {
				whaleOutflow += ai.Cash
			} else if ai.Type == "Quant" {
				quantOutflow += ai.Cash
			} else {
				retailOutflow += ai.Cash
			}
		} else {
			// 持仓AI的市值算流入
			marketValue := float64(ai.Shares) * state.Price
			if ai.Type == "Whale" {
				whaleInflow += marketValue
				activeWhales++
			} else if ai.Type == "Quant" {
				quantInflow += marketValue
				activeQuants++
			} else {
				retailInflow += marketValue
				activeRetails++
			}
		}
	}

	// 计算净流向
	whaleNet := whaleInflow - whaleOutflow
	quantNet := quantInflow - quantOutflow
	retailNet := retailInflow - retailOutflow

	// 雷达图渲染
	fmt.Printf("\n%s╔═══════════════════ 📡 AI资金流向雷达 ═══════════════════╗%s\n", Purple, Reset)
	fmt.Printf("%s║%s", Purple, Reset)

	// 第一行：Whale状态
	whaleColor := Green
	whaleArrow := "↑"
	if whaleNet < 0 {
		whaleColor = Red
		whaleArrow = "↓"
	}
	fmt.Printf("  %s🐋 大资金%s: %s%s $%.0fk%s (%d个活跃)",
		Purple, Reset, whaleColor, whaleArrow, math.Abs(whaleNet)/1000, Reset, activeWhales)

	// 雷达可视化符号
	whalePower := math.Min(math.Abs(whaleNet)/50000, 5) // 最多5个箭头
	whaleRadar := strings.Repeat(whaleArrow, int(whalePower))
	if whaleRadar == "" {
		whaleRadar = "─"
	}
	fmt.Printf(" [%s%s%s]\n", whaleColor, whaleRadar, Reset)

	fmt.Printf("%s║%s", Purple, Reset)

	// 第二行：Quant状态
	quantColor := Green
	quantArrow := "↑"
	if quantNet < 0 {
		quantColor = Red
		quantArrow = "↓"
	}
	fmt.Printf("  %s🤖 量化队%s: %s%s $%.0fk%s (%d个活跃)",
		Blue, Reset, quantColor, quantArrow, math.Abs(quantNet)/1000, Reset, activeQuants)

	quantPower := math.Min(math.Abs(quantNet)/30000, 5)
	quantRadar := strings.Repeat(quantArrow, int(quantPower))
	if quantRadar == "" {
		quantRadar = "─"
	}
	fmt.Printf(" [%s%s%s]\n", quantColor, quantRadar, Reset)

	fmt.Printf("%s║%s", Purple, Reset)

	// 第三行：Retail状态
	retailColor := Green
	retailArrow := "↑"
	if retailNet < 0 {
		retailColor = Red
		retailArrow = "↓"
	}
	fmt.Printf("  %s🥬 散户群%s: %s%s $%.0fk%s (%d个活跃)",
		Green, Reset, retailColor, retailArrow, math.Abs(retailNet)/1000, Reset, activeRetails)

	retailPower := math.Min(math.Abs(retailNet)/20000, 5)
	retailRadar := strings.Repeat(retailArrow, int(retailPower))
	if retailRadar == "" {
		retailRadar = "─"
	}
	fmt.Printf(" [%s%s%s]\n", retailColor, retailRadar, Reset)

	// 第四行：市场情绪总结
	totalNet := whaleNet + quantNet + retailNet
	marketSentiment := "平衡"
	sentimentColor := Yellow
	sentimentIcon := "⚖️"

	if totalNet > 50000 {
		marketSentiment = "强势流入"
		sentimentColor = Green
		sentimentIcon = "🚀"
	} else if totalNet > 20000 {
		marketSentiment = "温和流入"
		sentimentColor = Green
		sentimentIcon = "📈"
	} else if totalNet < -50000 {
		marketSentiment = "恐慌出逃"
		sentimentColor = Red
		sentimentIcon = "💥"
	} else if totalNet < -20000 {
		marketSentiment = "资金撤离"
		sentimentColor = Red
		sentimentIcon = "📉"
	}

	fmt.Printf("%s║%s  %s 总体态势: %s%s%s (净流向: %s$%.0fk%s)\n",
		Purple, Reset, sentimentIcon, sentimentColor, marketSentiment, Reset,
		sentimentColor, totalNet/1000, Reset)

	// 关键预警
	if whaleNet < -100000 && activeWhales <= 2 {
		fmt.Printf("%s║%s  %s⚠️  警告: 大资金大规模撤离，市场流动性危机！%s\n", Purple, Reset, Red, Reset)
	} else if whaleNet > 100000 && retailNet < -50000 {
		fmt.Printf("%s║%s  %s💡 注意: 大资金入场而散户逃离，可能是主力建仓%s\n", Purple, Reset, Yellow, Reset)
	} else if retailNet > 80000 && whaleNet < -50000 {
		fmt.Printf("%s║%s  %s⚠️  警告: 散户疯狂追涨而大资金出货，典型出货信号！%s\n", Purple, Reset, Red, Reset)
	}

	fmt.Printf("%s╚═══════════════════════════════════════════════════════════╝%s\n", Purple, Reset)
}

// 渲染残局模式专属HUD
func renderEndgameHUD(state *GameState) {
	if state.EndgameScenario == nil {
		return
	}

	scenario := state.EndgameScenario

	// 难度颜色
	diffColor := Green
	diffIcon := "⭐"
	switch scenario.Difficulty {
	case "简单":
		diffColor = Green
		diffIcon = "⭐"
	case "中等":
		diffColor = Yellow
		diffIcon = "⭐⭐"
	case "困难":
		diffColor = Red
		diffIcon = "⭐⭐⭐"
	case "地狱":
		diffColor = Purple
		diffIcon = "🔥🔥🔥"
	}

	// 计算进度
	currentTurn := calculateTurnNumberHelper(state.Day, state.Session)
	elapsedTurns := currentTurn - state.EndgameStartTurn
	totalTurns := scenario.TimeLimit
	progressPercent := float64(elapsedTurns) / float64(totalTurns) * 100
	if progressPercent > 100 {
		progressPercent = 100
	}

	// 进度条渲染
	barWidth := 30
	filledWidth := int(progressPercent / 100 * float64(barWidth))
	if filledWidth > barWidth {
		filledWidth = barWidth
	}

	progressColor := Green
	if progressPercent > 75 {
		progressColor = Red
	} else if progressPercent > 50 {
		progressColor = Yellow
	}

	progressBar := progressColor + strings.Repeat("█", filledWidth) + Reset +
		strings.Repeat("░", barWidth-filledWidth)

	// 盈利计算
	netAsset := float64(state.PlayerShares)*state.Price + state.PlayerCash - state.MarginDebt
	currentProfit := ((netAsset - state.InitialAsset) / state.InitialAsset) * 100
	targetProfit := scenario.TargetProfit * 100

	profitColor := Red
	profitStatus := "未达标"
	if currentProfit >= targetProfit {
		profitColor = Green
		profitStatus = "✓ 已达标"
	}

	// 渲染HUD
	fmt.Printf("\n%s╔═══════════════════════ 🎯 残局挑战 ═══════════════════════╗%s\n", Purple, Reset)
	fmt.Printf("%s║%s  关卡: %s%s%s (难度: %s%s%s)\n",
		Purple, Reset, Cyan, scenario.Name, Reset, diffColor, diffIcon, Reset)
	fmt.Printf("%s║%s  进度: [%s] %d/%d 回合 (%.0f%%)\n",
		Purple, Reset, progressBar, elapsedTurns, totalTurns, progressPercent)
	fmt.Printf("%s║%s  目标: 盈利 %s%.1f%%%s  |  当前: %s%.1f%% %s%s\n",
		Purple, Reset, Yellow, targetProfit, Reset, profitColor, currentProfit, profitStatus, Reset)

	// 教学要点（显示前2个）
	if len(scenario.TeachingPoints) > 0 {
		fmt.Printf("%s║%s  要点: ", Purple, Reset)
		for i := 0; i < len(scenario.TeachingPoints) && i < 2; i++ {
			if i > 0 {
				fmt.Printf(" • ")
			}
			fmt.Printf("%s%s%s", Yellow, scenario.TeachingPoints[i], Reset)
		}
		fmt.Println()
	}

	// 时间警告
	if progressPercent > 80 {
		remainingTurns := totalTurns - elapsedTurns
		fmt.Printf("%s║%s  %s⚠️  警告: 仅剩 %d 回合！%s\n",
			Purple, Reset, Red, remainingTurns, Reset)
	}

	fmt.Printf("%s╚═══════════════════════════════════════════════════════════╝%s\n", Purple, Reset)
}

// 渲染进度条（用于多空对比）
func renderProgressBar(label string, val1, val2 int, color1, color2 string) string {
	total := val1 + val2
	if total == 0 {
		return label + " [----------] 平衡"
	}
	width := 20
	p1 := int(float64(val1) / float64(total) * float64(width))
	p2 := width - p1
	bar := color1 + strings.Repeat("█", p1) + color2 + strings.Repeat("█", p2) + Reset
	return fmt.Sprintf("%s [%s] %d:%d", label, bar, val1, val2)
}

func renderFrame(state *GameState, histCache *HistoricalStateCache) {
	// 恢复为仅重置光标或添加适量空行，配合分隔符使用
	fmt.Println("\n" + strings.Repeat("━", 70))

	// ── 环境氛围渲染 ──
	borderColor := Cyan
	borderChar := "═"
	marketMood := "平稳"
	if state.IsMonsterStock {
		borderColor = Purple
		borderChar = "🐉"
		marketMood = "狂热"
	} else if state.CrashWarningLevel >= 3 {
		borderColor = Red
		borderChar = "🚨"
		marketMood = "高危"
	}

	// ── 顶部 HUD ──
	sessionColor := Purple
	if state.Session == "尾盘" {
		sessionColor = Blue
	}
	change := ((state.Price - state.LastPrice) / state.LastPrice) * 100
	priceColor := Red
	arrow := "▲"
	if change < 0 {
		priceColor = Green
		arrow = "▼"
	}

	netAsset := float64(state.PlayerShares)*state.Price + state.PlayerCash - state.MarginDebt
	assetProfit := ((netAsset - state.InitialAsset) / state.InitialAsset) * 100
	profitColor := Green
	if assetProfit < 0 {
		profitColor = Red
	}

	data := loadAchievementData()
	title := getTraderTitle(data.Level)

	fmt.Printf("%s╔%s╗%s\n", borderColor, strings.Repeat(borderChar, 31), Reset)
	fmt.Printf("%s║%s  操盘手: %s%s (Lv.%d)%s  │  情报点: %d  │  %s %s%d天-%s%s%s (%s)\n",
		borderColor, Reset, Cyan, title, data.Level, Reset, data.IntelPoints, borderColor, Yellow, state.Day, sessionColor, state.Session, Reset, marketMood)
	fmt.Printf("%s║%s  现价: %s$%.2f %s (%+.2f%%)%s  │  日内: %s$%.2f - $%.2f%s\n",
		borderColor, Reset, priceColor, state.Price, arrow, change, Reset, Yellow, state.DayLow, state.DayHigh, Reset)
	// 计算换手率（安全访问VolumeHistory）
	turnoverRate := 0.0
	if len(state.VolumeHistory) > 0 && state.TotalMarketShares > 0 {
		turnoverRate = float64(state.VolumeHistory[len(state.VolumeHistory)-1]) / float64(state.TotalMarketShares) * 100
	}
	fmt.Printf("%s║%s  资产: %s$%.0f %s(%+.1f%%)%s  │  日内换手: %.1f%%  │  VIX恐慌: %d/100\n",
		borderColor, Reset, profitColor, netAsset, profitColor, assetProfit, Reset, turnoverRate, state.CrashWarningLevel*25)
	fmt.Printf("%s╚%s╝%s\n", borderColor, strings.Repeat(borderChar, 31), Reset)

	// ── 残局模式HUD ──
	if state.IsEndgameMode {
		renderEndgameHUD(state)
	}

	// ── 关键警报（闪烁提醒）──
	renderCriticalAlerts(state)

	// ── 市场心理仪表盘 ──
	if state.DailyFortune != "" {
		fmt.Printf("  %s🔮 操盘黄历: %s%s%s\n", Cyan, Yellow, state.DailyFortune, Reset)
	}
	sentimentBar := renderProgressBar("  🤝 全服博弈(多:空)", state.TotalBuyDemandShares, state.TotalSellSupplyShares, Green, Red)
	fmt.Println(sentimentBar)

	// ── 实时快讯 (Ticker) ──
	if len(state.MarketLogs) > 0 {
		lastLog := state.MarketLogs[len(state.MarketLogs)-1]
		fmt.Printf("  %s📢 实时快讯: %s%s\n", Yellow, lastLog, Reset)
	}

	// ── 操作反馈 ──
	if state.LastActionMessage != "" {
		fmt.Printf("%s┌──────────────────────────────────────────────────────────────┐%s\n", Green, Reset)
		fmt.Printf("%s│%s  %s\n", Green, Reset, state.LastActionMessage)
		fmt.Printf("%s└──────────────────────────────────────────────────────────────┘%s\n", Green, Reset)
	}

	// ── 事件卡牌 ──
	renderEventCard(state.CurrentEvent, false)

	// ── 核心监控区 ──
	fmt.Printf("\n%s┌────────────────────── 📊 核心监控 ──────────────────────────┐%s\n", Cyan, Reset)

	recentDays := 35
	if len(state.PriceHistory) < recentDays {
		recentDays = len(state.PriceHistory)
	}
	trendStr := renderSparkline(state.PriceHistory[len(state.PriceHistory)-recentDays:])
	fmt.Printf("%s│%s  实时K线: [%s%s%s]\n", Cyan, Reset, Cyan, trendStr, Reset)

	if state.PlayerShares == 0 {
		fmt.Printf("%s│%s  持仓: %s【空仓中】%s", Cyan, Reset, Blue, Reset)
	} else {
		frozenInfo := ""
		if state.PlayerFrozenShares > 0 {
			frozenInfo = fmt.Sprintf("%s (可用:%d / 冻结:%d T+1)%s", Yellow, state.PlayerAvailableShares, state.PlayerFrozenShares, Reset)
		}
		fmt.Printf("%s│%s  持仓: %s%d 股%s (均价:$%.2f)%s", Cyan, Reset, Red, state.PlayerShares, Reset, state.PlayerAvgCost, frozenInfo)
	}
	fmt.Printf("  %s现金: $%.2f%s\n", Yellow, state.PlayerCash, Reset)

	if state.MarginDebt > 0 {
		fmt.Printf("%s│%s  %s⚠️  杠杆风险: 负债$%.2f  (随时面临爆仓强平)%s\n", Cyan, Reset, Red, state.MarginDebt, Reset)
	}

	// 盘口动态 L2 Detail
	fmt.Printf("%s│%s  盘口: 买盘需求%d / 卖盘抛压%d  │  净流向: %d\n", Cyan, Reset, state.TotalBuyDemandShares, state.TotalSellSupplyShares, state.TotalBuyDemandShares-state.TotalSellSupplyShares)

	fmt.Printf("%s│%s  资金流: ", Cyan, Reset)
	if state.WhaleSelling > 0 {
		fmt.Printf("%s🐋砸盘%d %s", Red, state.WhaleSelling, Reset)
	}
	if state.WhaleBuying > 0 {
		fmt.Printf("%s🐋抢筹%d %s", Green, state.WhaleBuying, Reset)
	}
	if state.RetailSelling > 0 {
		fmt.Printf("%s🥬割肉%d %s", Red, state.RetailSelling, Reset)
	}
	if state.RetailBuying > 0 {
		fmt.Printf("%s追高%d %s", Green, state.RetailBuying, Reset)
	}
	fmt.Println()
	fmt.Printf("%s└──────────────────────────────────────────────────────────────┘%s\n", Cyan, Reset)

	// ── 辅助分析 ──
	fmt.Printf("\n%s┌────────────────────── 🧠 辅助分析 ──────────────────────────┐%s\n", Cyan, Reset)
	renderCostDistribution(state)

	if state.PlayerShares > 0 {
		advice := generateStrategyAdvice(state)
		actionColor := Green
		if advice.Action == "SELL" {
			actionColor = Red
		}
		riskColor := Green
		if advice.RiskLevel == "高风险" {
			riskColor = Yellow
		} else if advice.RiskLevel == "极高风险" {
			riskColor = Red
		}
		fmt.Printf("  %s💡 建议: %s%s %s%s  (风险:%s%s%s)\n", Cyan, actionColor, advice.Confidence, advice.Action, Reset, riskColor, advice.RiskLevel, Reset)
		fmt.Printf("    原因: %s\n", advice.Reason)
	}
	fmt.Printf("%s└──────────────────────────────────────────────────────────────┘%s\n", Cyan, Reset)

	// ── 增强K线图表 ──
	if len(state.PriceHistory) >= 10 {
		renderEnhancedKLine(state)
	}

	// 显示高级分析（仅当历史数据足够且玩家持有股票时）
	if state.PlayerShares > 0 && len(histCache.States) >= 3 {
		advancedAnalysis := generateAdvancedAnalysis(state, histCache)
		displayAdvancedAnalysis(advancedAnalysis)
	}

	// ── 动态风险热力图 ──
	renderRiskHeatmap(state)

	// ── 实时动态 Feed ──
	fmt.Printf("\n%s┌────────────────────── 📰 市场 Feed ──────────────────────────┐%s\n", Yellow, Reset)
	displayLogs := state.MarketLogs
	if len(displayLogs) > 5 {
		displayLogs = displayLogs[len(displayLogs)-5:]
	}
	for _, log := range displayLogs {
		fmt.Printf("  %s\n", log)
	}
	fmt.Printf("%s└──────────────────────────────────────────────────────────────┘%s\n", Yellow, Reset)

	// ── AI资金流向雷达 ──
	renderAIMoneyFlowRadar(state)

	// ── AI 列表 ──
	fmt.Printf("\n%s┌────────────────────── 🤖 AI 实时仓位 ──────────────────────┐%s\n", Cyan, Reset)
	sortedAIs := make([]*AI, len(state.AIs))
	copy(sortedAIs, state.AIs)
	sort.Slice(sortedAIs, func(i, j int) bool { return sortedAIs[i].Shares > sortedAIs[j].Shares })

	for _, ai := range sortedAIs {
		aiProfit := 0.0
		if ai.Cost > 0 {
			aiProfit = ((state.Price - ai.Cost) / ai.Cost) * 100
		}
		statusStr := fmt.Sprintf("%s%+.0f%%%s", Red, aiProfit, Reset)
		if ai.HasSold || ai.Shares == 0 {
			statusStr = "\033[90m空仓\033[0m"
		}

		// 恢复并美化特殊状态标签
		statusFlagStr := ""
		if ai.StatusFlag == "Spoofing" {
			statusFlagStr = Purple + "[诱多中] " + Reset
		} else if ai.StatusFlag == "GridTrading" {
			statusFlagStr = Blue + "[网格中] " + Reset
		} else if ai.StatusFlag == "Bailout" {
			statusFlagStr = Red + "[砸锅卖铁救市] " + Reset
		} else if ai.StatusFlag == "ForcedLiquidation" {
			statusFlagStr = Yellow + "[爆仓清场] " + Reset
		} else if ai.StatusFlag == "Shakeout" {
			statusFlagStr = Cyan + "[震仓洗盘] " + Reset
		}

		nameColor := Cyan
		if ai.Type == "Whale" {
			nameColor = Purple
		} else if ai.Type == "Quant" {
			nameColor = Blue
		} else {
			nameColor = Green
		}
		fmt.Printf("  %s[%-8s] %-12s%s │ %5d股 │ %-8s │ %s%s%s%s\n",
			nameColor, ai.SubType, ai.Name, Reset, ai.Shares, statusStr, statusFlagStr, Gray, ai.LastOpinion, Reset)
	}
	fmt.Printf("%s└──────────────────────────────────────────────────────────────┘%s\n", Cyan, Reset)
}

// 渲染战后复盘分析
func renderTradeRecap(state *GameState) {
	if len(state.PriceHistory) == 0 {
		return
	}

	fmt.Printf("\n" + Cyan + "📊 【上帝视角：全场博弈复盘】" + Reset + "\n")

	// 1. 渲染 K 线标尺
	history := state.PriceHistory
	bars := []rune(" ▂▃▄▅▆▇█")
	minP, maxP := history[0], history[0]
	for _, v := range history {
		if v < minP {
			minP = v
		}
		if v > maxP {
			maxP = v
		}
	}

	fmt.Printf("  走势: ")
	for _, v := range history {
		idx := 0
		if maxP > minP {
			idx = int((v - minP) / (maxP - minP) * float64(len(bars)-1))
		}
		fmt.Printf("%c", bars[idx])
	}
	fmt.Println()

	// 2. 渲染买卖点对齐 (B/S)
	trades := make(map[int]string)
	for _, tp := range state.TradePoints {
		sIdx := 0
		switch tp.Session {
		case "早盘":
			sIdx = 0
		case "盘中上午":
			sIdx = 1
		case "盘中下午":
			sIdx = 2
		case "尾盘":
			sIdx = 3
		}
		// Index logic: day 1 start is 0.
		// Day 1 Morning result is index 1.
		idx := (tp.Day-1)*4 + sIdx + 1
		if idx < len(history) {
			char := ""
			if tp.Action == "Buy" {
				char = Green + "B" + Reset
			} else {
				char = Red + "S" + Reset
			}
			trades[idx] = char
		}
	}

	fmt.Printf("  操作: ")
	for i := range history {
		if char, ok := trades[i]; ok {
			fmt.Print(char)
		} else {
			fmt.Print(" ")
		}
	}
	fmt.Println("  (B=买入, S=卖出)")

	// 3. 详细诊断建议
	if len(state.TradePoints) > 0 {
		fmt.Printf("\n" + Yellow + "💡 【操盘手诊断报告】" + Reset + "\n")
		for _, tp := range state.TradePoints {
			icon := "🟢"
			if tp.Action == "Sell" {
				icon = "🔴"
			}

			diag := "这一步操作策略极其稳健。"
			if tp.Action == "Buy" {
				if tp.WhaleStatus == "Spoofing" {
					diag = Red + "警报！你被游资的假买盘/假新闻勾引了，典型的诱多接盘。" + Reset
				} else if tp.MarketContext == "妖股狂热" {
					diag = Purple + "击鼓传花：你参与了末路狂欢，勇气可嘉，但也要小心被埋。" + Reset
				} else if tp.WhaleStatus == "Bailout" {
					diag = Green + "神操作！你竟然跟国家队同一时间抄底，你是内幕人士吗？" + Reset
				}
			} else { // Sell
				if tp.WhaleStatus == "Spoofing" {
					diag = Red + "心理防线崩溃！游资用一笔假抛单就把你吓得割肉离场了。" + Reset
				} else if tp.MarketContext == "极端恐慌" {
					diag = Yellow + "散户本能：你在最黑暗的时刻选择了逃跑，却没看到国家队正在跌停板捡筹码。" + Reset
				}
			}

			fmt.Printf("  [%d天 %s] %s %s %d股 @$%.2f -> %s\n",
				tp.Day, tp.Session, icon, tp.Action, tp.Shares, tp.Price, diag)
		}
	}
}

// 渲染操盘编年史
func renderChronicle(state *GameState) {
	fmt.Print("\033[H\033[2J") // 清屏
	fmt.Println(Purple + "╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║             操盘编年史 - 历史的每一刻                ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝" + Reset)

	fmt.Printf("\n%s  %-8s %-12s %-10s %-8s %-15s %-10s%s\n",
		Cyan, "时间", "价格", "涨跌", "玩家", "庄家动向", "市场情绪", Reset)
	fmt.Println(Cyan + "  " + strings.Repeat("━", 75) + Reset)

	// 从最新到最旧显示（倒序遍历）
	for i := len(state.Chronicle) - 1; i >= 0; i-- {
		e := state.Chronicle[i]
		color := Reset
		if e.Change > 0.03 {
			color = Red
		} else if e.Change < -0.03 {
			color = Green
		}

		playerColor := Gray
		if e.PlayerAction == "买入" || e.PlayerAction == "持仓" {
			playerColor = Red
		}

		whaleColor := Reset
		if e.WhaleAction == "拉升" || e.WhaleAction == "吸筹" {
			whaleColor = Red
		} else if e.WhaleAction == "砸盘" || e.WhaleAction == "洗盘" {
			whaleColor = Green
		}

		moodColor := Reset
		if e.MarketMood == "狂热" {
			moodColor = Purple
		} else if e.MarketMood == "恐慌" {
			moodColor = Green
		}

		timeStr := fmt.Sprintf("D%d-%s", e.Day, e.Session)
		fmt.Printf("  %-10s %s$%7.2f %s%7.1f%% %s%-6s %s%-12s %s%-10s%s\n",
			timeStr,
			Reset, e.Price,
			color, e.Change*100,
			playerColor, e.PlayerAction,
			whaleColor, e.WhaleAction,
			moodColor, e.MarketMood, Reset)

		// 打印重要事件（如果是早盘）
		if e.Session == "早盘" {
			fmt.Printf("    %s[事件] %s%s\n", Gray, e.Event, Reset)
		}
	}

	fmt.Println(Cyan + "\n  " + strings.Repeat("━", 75) + Reset)
	fmt.Print("\n按回车键返回结算页面...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}

// 操盘手人格诊断
func diagnoseTrader(state *GameState, rank PlayerRank) (string, string) {
	// 维度1：风险偏好 (Risk Appetite)
	riskAppetite := "稳健型"
	if state.MarginDebt > 0 {
		riskAppetite = "激进型"
	}
	if state.IsMarginCalled {
		riskAppetite = "赌徒型"
	}

	// 维度2：持仓耐心 (Patience)
	tradingCount := 0
	for _, e := range state.Chronicle {
		if e.PlayerAction == "买入" || e.PlayerAction == "卖出" {
			tradingCount++
		}
	}
	patience := "波段持仓"
	if tradingCount > 10 {
		patience = "频繁短线"
	} else if tradingCount <= 2 {
		patience = "长线格局"
	}

	// 维度3：执行力 (Execution)
	execution := "随性交易"
	if rank.RiskScore > 25 {
		execution = "铁律执行"
	} else if rank.RiskScore < 10 {
		execution = "犹豫不决"
	}

	// 判定人格
	archetype := "市场观察者"
	description := "你对市场有基本的参与感，但尚未形成坚定的交易系统。"

	if riskAppetite == "赌徒型" {
		archetype = "【亡命之徒】"
		description = "你极度迷恋杠杆，试图在一次波动中改变命运。建议：学会尊重市场，远离高利贷。"
	} else if patience == "频繁短线" && rank.ProfitScore < 15 {
		archetype = "【手续费贡献者】"
		description = "你频繁进出，试图抓住每一个波动，却在摩擦成本中损耗了利润。建议：减少操作，等待大趋势。"
	} else if patience == "长线格局" && rank.ProfitScore > 30 {
		archetype = "【冷酷的狙击手】"
		description = "你拥有极佳的耐心，能精准捕捉主升浪并格局到底。建议：保持节奏，你是天生的作手。"
	} else if execution == "铁律执行" && state.IsCrashed && state.PlayerShares == 0 {
		archetype = "【纪律捍卫者】"
		description = "你对风险极度敏感，在危险来临前果断离场。建议：你的风控能力是你在修罗场生存的基石。"
	} else if rank.Grade == "S" {
		archetype = "【天选之子】"
		description = "本局操作堪称教科书级别。无论行情如何，你总能站在赢家的一边。"
	}

	return archetype, description
}

// 生成游戏后复盘报告
func generatePostGameReport(state *GameState, histCache *HistoricalStateCache) {
	// 文件名：包含时间戳和主题
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("stock_game_report_%s_%s.txt", CurrentTheme.Name, timestamp)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf(Red+"无法创建报告文件: %v"+Reset+"\n", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// ===== 报告标题 =====
	writer.WriteString("╔═══════════════════════════════════════════════════════════════╗\n")
	writer.WriteString("║              妖股搏杀 - 复盘报告                             ║\n")
	writer.WriteString("╚═══════════════════════════════════════════════════════════════╝\n\n")

	writer.WriteString(fmt.Sprintf("生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	writer.WriteString(fmt.Sprintf("游戏主题: %s\n", CurrentTheme.Name))
	writer.WriteString(fmt.Sprintf("游戏时长: %d 天 (从 Day 1 到 Day %d)\n", state.Day, state.Day))
	writer.WriteString(fmt.Sprintf("最终股价: $%.2f\n", state.Price))
	writer.WriteString(fmt.Sprintf("是否崩盘: %v\n\n", state.IsCrashed))

	// ===== 玩家表现 =====
	finalAsset := float64(state.PlayerShares)*state.Price + state.PlayerCash - state.MarginDebt
	profitRate := (finalAsset - 100000) / 100000 * 100

	writer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	writer.WriteString("【玩家表现】\n")
	writer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	writer.WriteString(fmt.Sprintf("初始资金: $100,000.00\n"))
	writer.WriteString(fmt.Sprintf("最终资产: $%.2f\n", finalAsset))
	writer.WriteString(fmt.Sprintf("收益率: %+.2f%%\n", profitRate))
	writer.WriteString(fmt.Sprintf("最终持仓: %d 股\n", state.PlayerShares))
	writer.WriteString(fmt.Sprintf("现金余额: $%.2f\n", state.PlayerCash))
	if state.MarginDebt > 0 {
		writer.WriteString(fmt.Sprintf("配资欠款: $%.2f\n", state.MarginDebt))
	}
	writer.WriteString("\n")

	// ===== 反身性模式分析 =====
	writer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	writer.WriteString("【反身性模式统计】\n")
	writer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	if len(GlobalStats.ReflexivityPatternsSeen) > 0 {
		writer.WriteString("本局游戏中出现的反身性模式:\n\n")
		for pattern, count := range GlobalStats.ReflexivityPatternsSeen {
			writer.WriteString(fmt.Sprintf("  • %s: 出现 %d 次\n", pattern, count))
		}
	} else {
		writer.WriteString("本局未检测到明显的反身性模式。\n")
	}
	writer.WriteString("\n")

	// ===== 贝叶斯预测准确性 =====
	writer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	writer.WriteString("【贝叶斯预测准确性】\n")
	writer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// 崩盘预测准确性
	if len(histCache.States) > 0 {
		// 简化版：检查最后一次高级分析的崩盘概率预测
		if GlobalCache.IsValid {
			predictedCrash := GlobalCache.LastAnalysis.BayesianAnalysis.CrashProbability.Posterior > 0.5
			actualCrash := state.IsCrashed

			writer.WriteString(fmt.Sprintf("最后预测崩盘概率: %.1f%%\n", GlobalCache.LastAnalysis.BayesianAnalysis.CrashProbability.Posterior*100))
			writer.WriteString(fmt.Sprintf("实际结果: %s\n", map[bool]string{true: "崩盘", false: "未崩盘"}[actualCrash]))

			if predictedCrash == actualCrash {
				writer.WriteString("✅ 崩盘预测命中！\n")
			} else {
				writer.WriteString("❌ 崩盘预测未命中。\n")
			}
		} else {
			writer.WriteString("无预测数据记录。\n")
		}
	}
	writer.WriteString("\n")

	// ===== 最佳离场时机对比 =====
	writer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	writer.WriteString("【离场时机对比】\n")
	writer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	if GlobalCache.IsValid && state.PlayerSoldDay > 0 {
		suggestedDay := GlobalCache.LastAnalysis.BayesianAnalysis.BestExitDay
		actualDay := state.PlayerSoldDay
		timingError := actualDay - suggestedDay

		writer.WriteString(fmt.Sprintf("AI建议离场日: 第 %d 天\n", suggestedDay))
		writer.WriteString(fmt.Sprintf("实际离场日: 第 %d 天 (%s)\n", actualDay, state.PlayerSoldSession))
		writer.WriteString(fmt.Sprintf("时机偏差: %+d 天\n", timingError))

		if timingError == 0 {
			writer.WriteString("✨ 完美离场！\n")
		} else if timingError > 0 {
			writer.WriteString("⚠️ 离场偏晚，可能错过最佳时机。\n")
		} else {
			writer.WriteString("⚠️ 离场偏早，可能放弃了部分利润。\n")
		}
	} else if state.PlayerShares > 0 {
		writer.WriteString("玩家未离场，全程持仓到底。\n")
	} else {
		writer.WriteString("无离场记录。\n")
	}
	writer.WriteString("\n")

	// ===== 操作日志摘要 =====
	writer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	writer.WriteString("【操作日志摘要】\n")
	writer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	buyCount := 0
	sellCount := 0
	for _, entry := range state.Chronicle {
		if entry.PlayerAction == "买入" {
			buyCount++
		} else if entry.PlayerAction == "卖出" {
			sellCount++
		}
	}

	writer.WriteString(fmt.Sprintf("总操作次数: %d 次\n", len(state.Chronicle)))
	writer.WriteString(fmt.Sprintf("买入次数: %d 次\n", buyCount))
	writer.WriteString(fmt.Sprintf("卖出次数: %d 次\n", sellCount))
	writer.WriteString("\n")

	// ===== 历史价格走势 =====
	if len(histCache.States) > 0 {
		writer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		writer.WriteString("【历史价格走势】(最近20个时段，从新到旧)\n")
		writer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		writer.WriteString(fmt.Sprintf("%-6s %-10s %-10s %-15s\n", "Day", "Session", "Price", "Player Action"))
		writer.WriteString("─────────────────────────────────────────────────\n")
		// 从最新到最旧显示（倒序遍历）
		for i := len(histCache.States) - 1; i >= 0; i-- {
			snap := histCache.States[i]
			writer.WriteString(fmt.Sprintf("%-6d %-10s $%-9.2f %-15s\n",
				snap.Day, snap.Session, snap.Price, snap.PlayerAction))
		}
		writer.WriteString("\n")
	}

	// ===== 结语 =====
	writer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	writer.WriteString("【复盘总结】\n")
	writer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	writer.WriteString("\n市场是一个零和博弈的修罗场。\n")
	writer.WriteString("每一次交易决策都是对市场理解的检验。\n")
	writer.WriteString("反身性理论告诉我们：认知影响现实，现实反过来影响认知。\n")
	writer.WriteString("贝叶斯概率提醒我们：用证据更新信念，在不确定性中寻找确定性。\n\n")
	writer.WriteString("继续磨练，在实战中成长。\n\n")
	writer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	writer.WriteString("报告结束\n")

	writer.Flush()

	fmt.Printf(Green+"\n📊 复盘报告已生成: %s"+Reset+"\n", filename)
}

func renderGameOver(state *GameState, reader *bufio.Reader) {
	fmt.Println("\n" + Yellow + "==================== 终局审判 ====================" + Reset)
	fmt.Printf(Cyan+"主题模式: %s\n"+Reset, CurrentTheme.Name)

	// 计算最终资产 (市值 + 现金 - 负债)
	finalAsset := float64(state.PlayerShares)*state.Price + state.PlayerCash - state.MarginDebt
	hasSoldBefore := state.PlayerSoldDay > 0
	isHoldingAtEnd := state.PlayerShares > 0

	if state.IsMarginCalled {
		fmt.Println(Red + "【🪦 爆仓出局】你因过度使用配资杠杆，账户净资产已归零！" + Reset)
		fmt.Println(Red + "配资公司为了止损，已暴力强平了你的持仓碎股。你已倾家荡产。" + Reset)
		fmt.Printf("\n☠️ 杠杆是魔鬼：最终净资产：$%.2f (欠款已被配资公司截留)\n", finalAsset)
	} else if state.IsCrashed {
		fmt.Println(Green + "【惨剧发生】盘口雪崩！获利盘踩踏狂涌，承接力瞬间贯穿！" + Reset)
		fmt.Println(Green + "所有人还没来得及撤单，天地斩连续跌停已经关头！" + Reset)

		if hasSoldBefore && !isHoldingAtEnd {
			fmt.Printf("\n🎉 奇迹生还！你成功逃顶，空仓躲过了崩盘！\n")
			fmt.Printf("最终安全现金：$%.2f\n", state.PlayerCash)
		} else if hasSoldBefore && isHoldingAtEnd {
			fmt.Printf("\n😰 功亏一篑！你卖出过但又买了回来，部分被埋...\n")
			fmt.Printf("最终资产：$%.2f (权益合计)\n", finalAsset)
		} else {
			fmt.Println("\n☠️ 贪婪的坟墓！你在犹豫要不要多吃一个点的时候，被直接活埋。")
			fmt.Printf("账户被强平，资产腰斩，惨烈收场：$%.2f\n", finalAsset)
		}
	} else {
		fmt.Println("【平缓落地】主力极其克制，居然拖到了最后也没闪崩。")
		fmt.Printf("你的最终净资产：$%.2f\n", finalAsset)
		if len(state.TradeHistory) > 0 {
			fmt.Printf("(交易了%d次)\n", len(state.TradeHistory))
		}
	}
	fmt.Println(Yellow + "==================================================" + Reset)
	renderTradeRecap(state)
	// 零和财富掠夺报告
	fmt.Println("\n" + Red + "💀 【修罗场·财富掠夺排行榜】 (零和博弈本质)" + Reset)
	fmt.Printf("%-15s %-12s %-12s %-10s\n", "角色", "初始资产", "最终资产", "财富转移")
	fmt.Println(strings.Repeat("-", 60))

	// 将玩家和 AI 统一排序
	type WealthEntry struct {
		Name     string
		Initial  float64
		Final    float64
		Transfer float64
	}
	ranking := []WealthEntry{
		{"你 (Player)", state.InitialAsset, finalAsset, finalAsset - state.InitialAsset},
	}
	for _, ai := range state.AIs {
		finalAI := ai.Cash + float64(ai.Shares)*state.Price - ai.MarginDebt
		ranking = append(ranking, WealthEntry{
			ai.Name, state.InitialAIAssets[ai.ID], finalAI, finalAI - state.InitialAIAssets[ai.ID],
		})
	}
	sort.Slice(ranking, func(i, j int) bool { return ranking[i].Transfer > ranking[j].Transfer })

	for _, e := range ranking {
		color := Reset
		title := "  "
		if e.Transfer > 0 {
			color = Red
			title = "🏆" // 掠食者
		} else if e.Transfer < 0 {
			color = Green
			title = "📉" // 猎物
		}
		fmt.Printf("%s%-15s %10.0f %10.0f %s%10.0f %s %s\n",
			Reset, e.Name, e.Initial, e.Final, color, e.Transfer, title, Reset)
	}
	fmt.Println(strings.Repeat("-", 60))

	rank := calculatePlayerRank(state)
	finalProfit := (finalAsset - state.InitialAsset) / state.InitialAsset

	// 根据评级显示不同颜色
	gradeColor := Reset
	switch rank.Grade {
	case "S":
		gradeColor = Purple
	case "A":
		gradeColor = Blue
	case "B":
		gradeColor = Cyan
	case "C":
		gradeColor = Yellow
	case "D":
		gradeColor = Red
	case "F":
		gradeColor = Red
	}

	fmt.Println("\n" + Purple + "╔════════════════════════════════════════════════════╗" + Reset)
	fmt.Println(Purple + "║            " + gradeColor + rank.Title + Purple + "           ║" + Reset)
	fmt.Println(Purple + "╚════════════════════════════════════════════════════╝" + Reset)

	fmt.Printf("\n  综合评级: %s%s 级%s   总分: %s%d/100%s\n", gradeColor, rank.Grade, Reset, gradeColor, rank.Score, Reset)
	fmt.Println("\n  ━━━━━━━━━━━━━ 详细评分 ━━━━━━━━━━━━━")
	fmt.Printf("    💰 收益分:  %d/40 分  (最终收益 %+.1f%%)\n", rank.ProfitScore, finalProfit*100)
	fmt.Printf("    ⏱️  时机分:  %d/30 分  ", rank.TimingScore)
	if hasSoldBefore {
		if !isHoldingAtEnd {
			fmt.Printf("(第%d天 %s 离场)\n", state.PlayerSoldDay, state.PlayerSoldSession)
		} else {
			fmt.Printf("(第%d天 %s 曾卖出，后又买回)\n", state.PlayerSoldDay, state.PlayerSoldSession)
		}
	} else {
		fmt.Printf("(格局到底)\n")
	}
	fmt.Printf("    🛡️  风控分:  %d/30 分  ", rank.RiskScore)
	if state.IsCrashed {
		if hasSoldBefore && !isHoldingAtEnd {
			fmt.Printf("(成功逃顶)\n")
		} else {
			fmt.Printf("(被埋)\n")
		}
	} else {
		fmt.Printf("(安全落地)\n")
	}

	fmt.Println("\n  ━━━━━━━━━━━━━ 评语 ━━━━━━━━━━━━━")
	fmt.Printf("  %s\n", rank.Comment)

	// 操盘手人格诊断
	archetype, diagDesc := diagnoseTrader(state, rank)
	fmt.Println("\n  ━━━━━━━━━━━━━ 🧠 操盘手人格诊断 ━━━━━━━━━━━━━")
	fmt.Printf("    人格画像: %s%s%s\n", Purple, archetype, Reset)
	fmt.Printf("    专业建议: %s\n", diagDesc)

	fmt.Println("\n" + Cyan + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + Reset)

	// 保存战绩到历史记录
	record := GameRecord{
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
		Theme:       CurrentTheme.Name,
		FinalProfit: finalProfit,
		Grade:       rank.Grade,
		Score:       rank.Score,
		IsCrashed:   state.IsCrashed,
		PlayerSold:  hasSoldBefore && !isHoldingAtEnd, // 卖出后未再买回
		SoldDay:     state.PlayerSoldDay,
		SoldSession: state.PlayerSoldSession,
		ProfitScore: rank.ProfitScore,
		TimingScore: rank.TimingScore,
		RiskScore:   rank.RiskScore,
	}

	err := saveGameRecord(record)
	if err != nil {
		fmt.Printf(Red+"\n⚠️ 保存战绩失败: %v\n"+Reset, err)
	} else {
		fmt.Println(Green + "\n✅ 战绩已保存到历史记录" + Reset)
	}

	// 检测并显示新成就
	newAchievements := checkAchievements(state, rank)
	if len(newAchievements) > 0 {
		fmt.Println("\n" + Gold + "🎊 【成就解锁】 🎊" + Reset)
		for _, ach := range newAchievements {
			fmt.Printf("  %s⭐ %s%s: %s\n", Gold, ach.Name, Reset, ach.Description)
		}
		fmt.Println(Cyan + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + Reset)
	}

	for {
		fmt.Println("\n" + Yellow + "请选择后续操作：" + Reset)
		fmt.Println(" [1] 查看《操盘编年史》(全周期深度复盘)")
		fmt.Println(" [2] 结束并返回主界面")
		fmt.Print(Green + "\n请输入编号: " + Reset)

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "1" {
			renderChronicle(state)
			// 看完编年史后重新显示结算总结，方便对比
			fmt.Print("\033[H\033[2J")
			fmt.Println(Cyan + "--- 已返回结算总结 ---" + Reset)
			fmt.Printf("最终收益: %+.1f%% | 评级: %s%s%s\n", finalProfit*100, gradeColor, rank.Grade, Reset)
			continue
		} else {
			break
		}
	}
}
