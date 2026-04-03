package main

import (
	"bufio"
	"encoding/json"
	"fmt"
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
	Timestamp     string  `json:"timestamp"`      // 游戏时间
	Theme         string  `json:"theme"`          // 主题模式
	FinalProfit   float64 `json:"final_profit"`   // 最终收益率
	Grade         string  `json:"grade"`          // 评级
	Score         int     `json:"score"`          // 总分
	IsCrashed     bool    `json:"is_crashed"`     // 是否崩盘
	PlayerSold    bool    `json:"player_sold"`    // 玩家是否卖出
	SoldDay       int     `json:"sold_day"`       // 卖出日期(如果卖出)
	SoldSession   string  `json:"sold_session"`   // 卖出时段(如果卖出)
	ProfitScore   int     `json:"profit_score"`   // 收益分
	TimingScore   int     `json:"timing_score"`   // 时机分
	RiskScore     int     `json:"risk_score"`     // 风控分
}

type TradePoint struct {
	Day           int
	Session       string
	Action        string    // "Buy" / "Sell"
	Price         float64
	Shares        int
	WhaleStatus   string    // 当时游资的状态 (如 "出货", "洗盘", "Spoofing")
	MarketContext string    // 当时大盘的状态 (如 "妖股狂热", "高位震荡")
}

type RecordHistory struct {
	Records []GameRecord `json:"records"`
}

const recordFilePath = "game_records.json"

type AI struct {
	ID              string
	Type            string  // "Whale", "Quant", "Retail", "Institution"
	SubType         string  // "刺客", "打板", "新韭", "老散", "国家队"
	Name            string
	Shares          int
	Cost            float64
	Cash            float64 // 新增：可用于买入的现金
	TargetProfit    float64
	FearBasis       float64
	HasSold         bool
	LastOpinion     string
	HasInsiderInfo  bool    // 是否有内幕消息（提前知道下一个事件）
	
	// 当前回合的订单
	OrderType       string  // "Buy", "Sell", ""
	OrderShares     int     // 计划卖出几股 / 计划买入几股 (买入时根据Cash算)
	OrderCash       float64 // 计划投入多少钱买
	MarginDebt      float64 // 新增：游资自身场外配资
	StatusFlag      string  // 新增：记录当前行为状态 "Spoofing", "ForcedLiquidation", "GridTrading" 等
}

// 自动交易策略
type AutoStrategy struct {
	Name           string  // 策略名称
	StopProfit     float64 // 止盈线（例如0.2表示+20%止盈）
	StopLoss       float64 // 止损线（例如-0.1表示-10%止损）
	RebuyThreshold float64 // 回买阈值（价格跌到多少时考虑买回，例如0.9表示跌10%）
	EnableRebuy    bool    // 是否启用回买
}

type GameState struct {
	Day                  int
	Session              string // "早盘" 或 "尾盘"
	MaxDays              int
	GameMode             string  // "fixed" 固定回合 或 "auto" 自动结束
	Price                float64
	LastPrice            float64
	PlayerShares         int
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
	IsGameOver           bool
	IsCrashed            bool
	CurrentEvent         Event
	NextEvent            Event    // 下一个事件（信息不对称）
	AIs                  []*AI
	TotalActiveShare     int
	CrashWarningLevel    int // 崩盘预警等级 0=正常, 1=预警, 2=危险, 3=高危, 4=极限
	ConsecutiveFallDays  int // 连续下跌天数
	TotalEscapedAIShares int // 已逃跑的AI持股总数
	TradeHistory         []string // 交易历史记录
	AutoTradeStrategy    *AutoStrategy // 自动交易策略（nil表示手动）
	TotalMarketShares    int     // 流通盘总股数 (如 100,000)
    ConsecutiveGrowthDays int    // 新增：连涨天数用于触发妖股模型
    IsMonsterStock        bool   // 新增：妖股狂热状态
	
	// 博弈数据（零和撮合）
	TotalBuyDemandCash   float64 // 本回合全场涌入的总买单金额 
	TotalBuyDemandShares int     // 本回合全场需求多少股 (预估)
	TotalSellSupplyShares int    // 本回合全场砸出多少股
	FakeSellPressure      int    // 新增：游资假抛单
	FakeBuyPressure       int    // 新增：托单

	BuyPressure          int     // 显示用的买压（原逻辑兼容）
	SellPressure         int     // 显示用的卖压（原逻辑兼容）
	WhaleBuying          int     
	WhaleSelling         int     
	RetailBuying         int     
	RetailSelling        int     
	LastActionMessage    string  // 上一回合操作反馈（显示在仪表盘中）
	
	// 玩家在输入控制台提交的暂存订单
	PlayerOrderType      string  // "Buy", "Sell", ""
	PlayerOrderShares    int
	PlayerOrderCash      float64
	
	TradePoints          []TradePoint // 上帝视角复盘记录点
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
		reader.ReadString('\n')
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
	crashedEscaped := 0  // 崩盘时成功逃顶的次数
	totalCrashed := 0    // 崩盘总次数

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

	// 显示最近10场战绩
	fmt.Println("\n" + Cyan + "📜 最近战绩 (最多显示10场)" + Reset)
	fmt.Println(Cyan + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + Reset)

	displayCount := min(10, len(history.Records))
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
	reader.ReadString('\n')
}

func selectGameMode(reader *bufio.Reader) (string, int, *AutoStrategy) {
	fmt.Print("\033[H\033[2J") // 清屏
	fmt.Println(Purple + "╔══════════════════════════════════════════════════════╗")
	fmt.Println("║       妖股搏杀 - 游戏模式选择                       ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝" + Reset)

	fmt.Println("\n" + Cyan + "请选择游戏模式：" + Reset)
	fmt.Println()
	fmt.Println(Yellow + "[1] 手动模式 - 自动结束" + Reset)
	fmt.Println("    经典模式，最多15天（30回合），崩盘自动结束")
	fmt.Println()
	fmt.Println(Yellow + "[2] 手动模式 - 60回合固定" + Reset)
	fmt.Println("    固定60回合，不会中途崩盘，适合长期博弈")
	fmt.Println()
	fmt.Println(Yellow + "[3] 自动策略模拟 - 0-1博弈测试" + Reset)
	fmt.Println("    AI自动执行止盈止损策略，测试策略效果")
	fmt.Println()

	fmt.Print(Green + "请输入编号 (1-3): " + Reset)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	choice := 1
	if len(input) > 0 {
		fmt.Sscanf(input, "%d", &choice)
	}

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

	fmt.Println("\n" + Cyan + "预设策略：" + Reset)
	fmt.Println()
	fmt.Println(Yellow + "[1] 激进策略" + Reset + " - 止盈+15%, 止损-5%, 不回买")
	fmt.Println(Yellow + "[2] 稳健策略" + Reset + " - 止盈+25%, 止损-8%, 不回买")
	fmt.Println(Yellow + "[3] 波段策略" + Reset + " - 止盈+20%, 止损-10%, 跌15%回买")
	fmt.Println(Yellow + "[4] 格局策略" + Reset + " - 止盈+40%, 止损-15%, 跌20%回买")
	fmt.Println()

	fmt.Print(Green + "请输入编号 (1-4): " + Reset)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	choice := 1
	if len(input) > 0 {
		fmt.Sscanf(input, "%d", &choice)
	}

	strategies := []AutoStrategy{
		{"激进策略", 0.15, -0.05, 0.85, false},
		{"稳健策略", 0.25, -0.08, 0.9, false},
		{"波段策略", 0.20, -0.10, 0.85, true},
		{"格局策略", 0.40, -0.15, 0.80, true},
	}

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
	reader.ReadString('\n')

	return &strategy
}

func selectTheme(reader *bufio.Reader) *ThemeMode {
	fmt.Print("\033[H\033[2J") // 清屏
	fmt.Println(Purple + "╔══════════════════════════════════════════════════════╗")
	fmt.Println("║       妖股搏杀 - 主题模式选择                       ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝" + Reset)

	fmt.Println("\n" + Cyan + "请选择你想体验的主题模式：" + Reset)
	fmt.Println()

	for i, theme := range AllThemes {
		fmt.Printf("%s[%d] %s%s\n", Yellow, i+1, theme.Name, Reset)
		fmt.Printf("    %s\n", theme.Description)
		fmt.Printf("    难度: %s\n\n", theme.Difficulty)
	}

	fmt.Print(Green + "请输入编号 (1-" + fmt.Sprintf("%d", len(AllThemes)) + "): " + Reset)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	choice := 1 // 默认经典模式
	if len(input) > 0 {
		fmt.Sscanf(input, "%d", &choice)
	}

	if choice < 1 || choice > len(AllThemes) {
		choice = 1
	}

	selectedTheme := &AllThemes[choice-1]

	fmt.Printf("\n"+Cyan+"你选择了: %s%s%s\n"+Reset, Yellow, selectedTheme.Name, Cyan)
	fmt.Printf("%s\n", selectedTheme.Description)
	fmt.Print("\n按回车键确认开始...")
	reader.ReadString('\n')

	return selectedTheme
}

func main() {
	rand.Seed(time.Now().UnixNano())
	reader := bufio.NewReader(os.Stdin)

	// 选择游戏模式
	gameMode, maxDays, autoStrategy := selectGameMode(reader)

	// 选择主题模式
	CurrentTheme = selectTheme(reader)
	Events = CurrentTheme.Events

	// 询问是否查看历史战绩
	fmt.Print("\033[H\033[2J") // 清屏
	fmt.Println(Cyan + "是否查看历史战绩？" + Reset)
	fmt.Print(Yellow + "[1] 查看历史战绩  [2] 直接开始游戏: " + Reset)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "1" {
		showRecordHistory(reader)
	}

	state := initGame(gameMode, maxDays, autoStrategy)

	fmt.Print("\033[H\033[2J") // 清屏
	fmt.Println(Yellow + "=====================================================")
	fmt.Printf("       妖股搏杀 - %s\n", CurrentTheme.Name)
	fmt.Println("       地狱修罗难度：双时段对决")
	fmt.Println("=====================================================" + Reset)
	fmt.Println(Red + "【警告】市场流动性已极度匮乏！AI 极其恐慌脆弱！" + Reset)
	fmt.Println("规则更新：")
	fmt.Println("1. 每天分为【早盘】、【盘中上午】、【盘中下午】和【尾盘】四次决策大循环！")
	fmt.Println("2. 【盘中】时间你作为散户无法看盘交易，但网格量化会在盘中暗算火拼多次！")
	fmt.Println("3. 各类庄家人设不同，盘口博弈完全零和，且增加了国家队、假单、资金链等复杂逻辑。")
	fmt.Println("4. 💡 智能策略助手会给你实时建议。")
	fmt.Print("\n按回车键开始第一天早盘的修罗场...")
	reader.ReadString('\n')

	// 翻牌动画标记：第一天早盘不做动画，之后每次新的早盘做
	prevSession := ""
	for !state.IsGameOver {
		generateOpinions(state)
		// 新的一天早盘时，做翻牌动画
		if state.Session == "早盘" && state.Day > 1 && prevSession == "尾盘" {
			renderEventCard(state.CurrentEvent, true)
		}
		prevSession = state.Session
		renderFrame(state)

		if state.Session == "盘中上午" || state.Session == "盘中下午" {
			state.PlayerOrderType = ""
			state.PlayerOrderShares = 0
			state.PlayerOrderCash = 0
			
			fmt.Printf("\n" + Purple + "  🕰️ 【%s时段】你正在上班无法看盘，机构与量化正在暗流涌动火拼中..." + Reset + "\n", state.Session)
			time.Sleep(1500 * time.Millisecond)
			processTurn(state)
			continue
		}

		// 自动策略模式 vs 手动模式
		if state.AutoTradeStrategy != nil {
			// 自动执行策略
			executeAutoStrategy(state)
			time.Sleep(500 * time.Millisecond) // 暂停让玩家看到
		} else {
			// 手动模式 - 显示综合操作面板
			fmt.Printf("\n" + Yellow + "【交易指令台】" + Reset + "\n")
			fmt.Printf(" [1] 观望 (Hold)    ")
			if state.PlayerAvailableShares > 0 {
				fmt.Printf("[2] 减仓1/3    [3] 卖出一半    [4] 核按钮清仓    ")
			} else if state.PlayerFrozenShares > 0 {
				fmt.Printf(Red + "[无法卖出] 筹码已被T+1冻结！" + Reset + "    ")
			}
			
			canBuyShares := int(state.PlayerCash / state.Price)
			if canBuyShares > 0 {
				fmt.Printf("\n [5] 建仓1/3    [6] 半仓买入    [7] 满仓梭哈    ")
			}
			if state.MarginDebt == 0 && (state.PlayerCash > 0 || state.PlayerShares > 0) {
				fmt.Printf(Purple + "[8] 🎲配资3倍杠杆满仓！" + Reset)
			}
			fmt.Printf("\n" + Green + "请输入指令号: " + Reset)

			var input string
			for {
				raw, _ := reader.ReadString('\n')
				input = strings.TrimSpace(raw)
				if input != "" {
					break
				}
				fmt.Print(Yellow + "  ⚠️  未输入，请输入编号后回车: " + Reset)
			}

			switch input {
			case "2", "3", "4":
				if state.PlayerAvailableShares > 0 {
					var sellShares int
					if input == "2" { sellShares = state.PlayerAvailableShares / 3 } else if input == "3" { sellShares = state.PlayerAvailableShares / 2 } else { sellShares = state.PlayerAvailableShares }
					if sellShares == 0 { sellShares = state.PlayerAvailableShares }
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
					if input == "5" { buyShares = canBuyShares / 3 } else if input == "6" { buyShares = canBuyShares / 2 } else { buyShares = canBuyShares }
					if buyShares == 0 { buyShares = canBuyShares }
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

		processTurn(state)
	}

	renderGameOver(state)
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
		Day:               1,
		Session:           "早盘",
		MaxDays:           maxDays,
		GameMode:          gameMode,
		Price:             initialPrice,
		LastPrice:         initialPrice,
		PlayerShares:      initialShares,
		PlayerAvailableShares: initialShares, // 初始即可全卖
		PlayerFrozenShares:    0,
		PlayerCash:        0.0, // 初始全仓股票，无现金
		PlayerAvgCost:     initialPrice,
		MarginDebt:        0.0, // 无融资负债
		IsMarginCalled:    false,
		InitialAsset:      float64(initialShares) * initialPrice, // 记录初始总资产
		AIs:               make([]*AI, 0),
		CurrentEvent:      startEvent,
		NextEvent:         selectNextEventByMarkov(startEvent, Events), // 使用马尔科夫链预生成下一个事件
		TradeHistory:      []string{},
		PriceHistory:      []float64{initialPrice}, // 初始化一条K线基础记录
		AutoTradeStrategy: autoStrategy,
		TotalMarketShares: 100000, // 全服只有 10 万股，零和博弈
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
	createAIsAdvanced(state, "Retail", "新韭", "发财梦新韭菜", 2,  8000,  80000.0, 10.8, 1.15, 2.8) 
	createAIsAdvanced(state, "Retail", "老散", "装死死扛老散", 3,  9000,  15000.0, 13.0, 1.05, 0.2) // 强迫症装死

	// 隐藏国家队 (0 股，500万备用金准备托底)
	createAIsAdvanced(state, "Institution", "国家队", "平准托底基金", 1, 0, 5000000.0, 10.0, 1.05, 0.0)

	// Whale有内幕消息优势（提前知道下一个事件），刺客有场外配资负债
	for _, ai := range state.AIs {
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
		if count > 1 { name = fmt.Sprintf("%s-%d", namePrefix, i+1) }

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
	sellScore := 0  // 卖出倾向分数
	holdScore := 0  // 持有倾向分数

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

// 排名评级系统
type PlayerRank struct {
	Grade       string  // S, A, B, C, D, F
	Title       string  // 称号
	Score       int     // 总分
	ProfitScore int     // 收益分
	TimingScore int     // 时机分
	RiskScore   int     // 风控分
	Comment     string  // 评语
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
	hasSoldBefore := state.PlayerSoldDay > 0  // 是否卖出过
	isHoldingAtEnd := state.PlayerShares > 0   // 最终是否持仓

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
		escapedRatio := float64(state.TotalEscapedAIShares) / float64(totalShares + state.TotalEscapedAIShares)
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
	if a < b { return a }
	return b
}

func max(a, b float64) float64 {
	if a > b { return a }
	return b
}

func processTurn(state *GameState) {
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
		if ai.Cost > 0 { profitRatio = state.Price / ai.Cost }
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
					if state.Price < ai.Cost*0.50 { reenter = true }
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
					state.FakeSellPressure += int(ai.OrderCash / state.Price) * 3 
				}
			}
		} else {
			shouldSell := false
			sellRatio := 0.0
			if ai.Type == "Whale" {
				if ai.SubType == "刺客" {
					if profitRatio > activeTargetProfit {
						shouldSell = true; sellRatio = 1.0 
					} else if profitRatio > 1.1 && rand.Float64() < 0.3 {
						shouldSell = true; sellRatio = 0.4 ; ai.StatusFlag = "Shakeout"
					} else if eventFear > 2.0 {
						shouldSell = true; sellRatio = 1.0 
					}
				} else { // 机构长线
					if profitRatio > activeTargetProfit {
						shouldSell = true; sellRatio = 0.5 
					}
				}
			} else if ai.Type == "Quant" {
				if ai.SubType == "网格" {
					if profitRatio > 1.03 {
						shouldSell = true; sellRatio = 0.5; ai.StatusFlag = "GridTrading"
					} else if profitRatio < 0.97 && ai.Cash > 1000 {
						shouldSell = false; ai.OrderType = "Buy"; ai.OrderCash = ai.Cash * 0.5; ai.StatusFlag = "GridTrading"
					}
				} else { 
					if profitRatio > activeTargetProfit || rand.Float64() < 0.05 {
						shouldSell = true; sellRatio = 1.0
					}
				}
			} else { // Retail
				panicProb := activeFearBasis * eventFear * 0.05
				if ai.SubType == "新韭" {
					if profitRatio > activeTargetProfit && rand.Float64() < 0.6 {
						shouldSell = true; sellRatio = 0.5
					} else if profitRatio < 0.95 && priceChangePct < -0.05 && rand.Float64() < 0.7 {
						shouldSell = true; sellRatio = 1.0 
					} else if rand.Float64() < panicProb {
						shouldSell = true; sellRatio = 1.0 
					}
				} else { // 老散
					if profitRatio > activeTargetProfit && rand.Float64() < 0.8 {
						shouldSell = true; sellRatio = 1.0 
					}
				}
			}
			if shouldSell {
				ai.OrderType = "Sell"
				ai.OrderShares = int(float64(ai.Shares) * sellRatio)
				if ai.OrderShares == 0 { ai.OrderShares = ai.Shares }
			}
		}
	}

	// 3. 汇总全场订单
	totalSellShares := state.PlayerOrderShares
	totalBuyCash := state.PlayerOrderCash
	for _, ai := range state.AIs {
		if ai.OrderType == "Sell" { totalSellShares += ai.OrderShares }
		if ai.OrderType == "Buy" { totalBuyCash += ai.OrderCash }
	}
	state.TotalBuyDemandCash = totalBuyCash
	state.TotalSellSupplyShares = totalSellShares
	totalBuyDemandShares := int(totalBuyCash / state.Price)
	state.TotalBuyDemandShares = totalBuyDemandShares
	
	if totalBuyDemandShares == 0 && totalSellShares == 0 {
		state.Price = state.Price * (1.0 + (rand.Float64()*0.01 - 0.005))
		advanceTime(state)
		return
	}

	// 4. 定价引擎
	demandRatio := 1.0
	if totalSellShares > 0 {
		demandRatio = float64(totalBuyDemandShares) / float64(totalSellShares)
	} else { demandRatio = 5.0 }

	priceModifier := 0.0
	if demandRatio > 2.0 { priceModifier = 0.10 } else if demandRatio > 1.2 { priceModifier = 0.05 } else if demandRatio > 0.8 { priceModifier = (rand.Float64()*0.04 - 0.02) } else if demandRatio > 0.5 { priceModifier = -0.05 } else { priceModifier = -0.10 }
	
	state.LastPrice = state.Price
	state.Price = state.Price * (1.0 + priceModifier)
	if state.Price < 0.1 { state.Price = 0.1 }

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
	} else if realBuyExpectedShares <= 0 { sellProration = 0.0 } else if totalSellShares <= 0 { buyProration = 0.0 }

	// 6. 账户清算 (Fixing the Bug here)
    whaleStatusSnapshot := "沉静"
    for _, ai := range state.AIs {
        if ai.Type == "Whale" && ai.StatusFlag != "" { 
            whaleStatusSnapshot = ai.StatusFlag 
            break
        }
    }
    marketContext := "正常"
    if state.IsMonsterStock { marketContext = "妖股狂热" } else if state.CrashWarningLevel >= 3 { marketContext = "极端恐慌" }

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
			if ai.Type == "Whale" { state.WhaleSelling += actualSell }
			if ai.Type == "Retail" { state.RetailSelling += actualSell }
		} else if ai.OrderType == "Buy" {
			actualBuy := int(float64(int(ai.OrderCash / state.Price)) * buyProration)
			cost := float64(actualBuy) * state.Price
			if actualBuy > 0 {
				oldTotal := float64(ai.Shares) * ai.Cost
				ai.Shares += actualBuy
				ai.Cost = (oldTotal + cost) / float64(ai.Shares)
				ai.Cash -= cost
			}
			if ai.Type == "Whale" { state.WhaleBuying += actualBuy }
			if ai.Type == "Retail" { state.RetailBuying += actualBuy }
		}
		if ai.Shares == 0 { ai.HasSold = true } else { ai.HasSold = false }
	}
	state.BuyPressure = int(totalBuyCash / state.Price)
	state.SellPressure = totalSellShares
	state.TotalSellSupplyShares += state.FakeSellPressure
	state.TotalBuyDemandShares += state.FakeBuyPressure
	state.PriceHistory = append(state.PriceHistory, state.Price)
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
	if a < b { return a }
	return b
}
func max_int(a, b int) int {
	if a > b { return a }
	return b
}

// 渲染 K线趋势图 (Sparkline)
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
		if h.cost < minP { minP = h.cost }
		if h.cost > maxP { maxP = h.cost }
	}
	minP = minP * 0.95
	maxP = maxP * 1.05
	if maxP <= minP { maxP = minP + 1.0 }

	numBuckets := 9
	bucketSize := (maxP - minP) / float64(numBuckets)
	buckets := make([]float64, numBuckets)
	for _, h := range holders {
		idx := int((h.cost - minP) / bucketSize)
		if idx < 0 { idx = 0 }
		if idx >= numBuckets { idx = numBuckets - 1 }
		buckets[idx] += float64(h.shares)
	}
	maxVal := 0.0
	for _, v := range buckets {
		if v > maxVal { maxVal = v }
	}
	if maxVal == 0 { return }

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
	if fearLevel > 10 { fearLevel = 10 }
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
		if end > len(desc) { end = len(desc) }
		chunk := string(desc[:end])
		desc = desc[end:]
		fmt.Printf("  %s│%s  %-40s%s%s│%s\n", Purple, Reset, chunk, Reset, Purple, Reset)
	}
	fmt.Printf("  %s╰──────────────────────────────────────────╯%s\n", Purple, Reset)
}

func renderFrame(state *GameState) {
	fmt.Print("\033[H\033[2J\033[3J") // 深度清屏（含滚动缓冲区）

	// ── 顶部标题栏 ──
	sessionColor := Purple
	if state.Session == "尾盘" { sessionColor = Blue }
	change := ((state.Price - state.LastPrice) / state.LastPrice) * 100
	priceColor := Red
	arrow := "↑"
	if change < 0 { priceColor = Green; arrow = "↓" }

	netAsset := float64(state.PlayerShares)*state.Price + state.PlayerCash - state.MarginDebt
	assetProfit := ((netAsset - state.InitialAsset) / state.InitialAsset) * 100
	profitColor := Green
	if assetProfit < 0 { profitColor = Red }

	fmt.Printf("%s╔══════════════════════════════════════════════════════════════╗%s\n", Yellow, Reset)
	fmt.Printf("%s║%s  第 %s%2d%s 天 [%s%s%s]  │  %s$%.2f %s %s(%+.2f%%)%s  │  净资产: %s$%.0f %s(%+.1f%%)%s\n",
		Yellow, Reset,
		Yellow, state.Day, Reset,
		sessionColor, state.Session, Reset,
		priceColor, state.Price, arrow, priceColor, change, Reset,
		profitColor, netAsset, profitColor, assetProfit, Reset)
	fmt.Printf("%s╚══════════════════════════════════════════════════════════════╝%s\n", Yellow, Reset)

	// ── 操作反馈栏 ──
	if state.LastActionMessage != "" {
		fmt.Printf("%s┌─ 上回合操作反馈 ─────────────────────────────────────────────┐%s\n", Green, Reset)
		fmt.Printf("%s│%s  %s\n", Green, Reset, state.LastActionMessage)
		fmt.Printf("%s└──────────────────────────────────────────────────────────────┘%s\n", Green, Reset)
	}

	// ── 事件卡牌区 ──
	fmt.Printf("\n%s【%s情报】%s\n", Purple, state.Session, Reset)
	renderEventCard(state.CurrentEvent, false)

	// ── 崩盘预警 ──
	if state.CrashWarningLevel >= 2 {
		warnColor := Yellow
		if state.CrashWarningLevel >= 3 { warnColor = Red }
		warningLabels := []string{"", "", "⚠️  危险", "🚨 高危", "🚨 极限黑天鹅"}
		fmt.Printf("\n%s%s 【崩盘预警 %d/4】%s 流动性极度萎缩！随时可能踩踏！%s\n",
			warnColor, warningLabels[state.CrashWarningLevel], state.CrashWarningLevel, warningLabels[state.CrashWarningLevel], Reset)
	}
	if state.ConsecutiveFallDays >= 2 {
		fmt.Printf("%s🔴 连续下跌 %d 时段，警惕多杀多踩踏！%s\n", Red, state.ConsecutiveFallDays, Reset)
	}
	if state.IsMonsterStock {
		fmt.Printf("%s🐉 【妖股狂热共识已触发】全服放弃风控，正开启无底线追高击鼓传花！%s\n", Purple, Reset)
	}

	// ── 中区：大盘与你 ──
	fmt.Printf("\n%s┌────────────────────── 📊 大盘与你 ──────────────────────────┐%s\n", Cyan, Reset)

	recentDays := 40
	if len(state.PriceHistory) < recentDays { recentDays = len(state.PriceHistory) }
	trendStr := renderSparkline(state.PriceHistory[len(state.PriceHistory)-recentDays:])
	fmt.Printf("%s│%s  K线: [%s%s%s]\n", Cyan, Reset, Cyan, trendStr, Reset)

	if state.PlayerShares == 0 {
		fmt.Printf("%s│%s  持仓: %s【空仓中】%s\n", Cyan, Reset, Blue, Reset)
	} else {
		frozenInfo := ""
		if state.PlayerFrozenShares > 0 {
			frozenInfo = fmt.Sprintf("%s  (可用:%d / 冻结:%d T+1)%s", Yellow, state.PlayerAvailableShares, state.PlayerFrozenShares, Reset)
		}
		fmt.Printf("%s│%s  持仓: %s%d 股%s  均价:$%.2f%s\n", Cyan, Reset, Red, state.PlayerShares, Reset, state.PlayerAvgCost, frozenInfo)
	}
	fmt.Printf("%s│%s  现金: $%.2f", Cyan, Reset, state.PlayerCash)
	if state.MarginDebt > 0 {
		fmt.Printf("  %s负债:$%.2f ⚡爆仓预警%s", Red, state.MarginDebt, Reset)
	}
	fmt.Println()

	// 真实现价盘口对抗
	demandShares := state.TotalBuyDemandShares
	supplyShares := state.TotalSellSupplyShares
	
	pressureStr := fmt.Sprintf("  🚀全服买盘承接: %s%d股%s   🆚   🧨全服抛压: %s%d股%s", 
		Green, demandShares, Reset,
		Red, supplyShares, Reset,
	)
	
	fmt.Printf("%s│%s\n%s│%s%s\n", Cyan, Reset, Cyan, Reset, pressureStr)
	fmt.Printf("%s│%s  追踪:", Cyan, Reset)
	if state.WhaleSelling > 0 { fmt.Printf("  %s🐋砸盘%d股%s", Red, state.WhaleSelling, Reset) }
	if state.WhaleBuying > 0 { fmt.Printf("  %s🐋抢盘%d股%s", Green, state.WhaleBuying, Reset) }
	if state.RetailSelling > 0 { fmt.Printf("  %s🥬割肉%d股%s", Red, state.RetailSelling, Reset) }
	if state.RetailBuying > 0 { fmt.Printf("  %s🥬追高%d股%s", Green, state.RetailBuying, Reset) }
	fmt.Println()
	fmt.Printf("%s└──────────────────────────────────────────────────────────────┘%s\n", Cyan, Reset)

	// ── 右区：筹码分布图 ──
	fmt.Printf("\n%s┌────────────────────── 🎯 盘口博弈 ───────────────────────────┐%s\n", Cyan, Reset)
	renderCostDistribution(state)

	// 智能策略助手
	if state.PlayerShares > 0 {
		advice := generateStrategyAdvice(state)
		actionColor := Green
		if advice.Action == "SELL" { actionColor = Red }
		riskColor := Green
		if advice.RiskLevel == "高风险" { riskColor = Yellow } else if advice.RiskLevel == "极高风险" { riskColor = Red }

		fmt.Printf("  %s────────────────── 💡 策略助手 ──────────────────%s\n", Cyan, Reset)
		fmt.Printf("  风险: %s%-6s%s  收益: %s%+.1f%%%s  建议: %s%s %s%s\n",
			riskColor, advice.RiskLevel, Reset,
			profitColor, advice.ProfitRatio*100, Reset,
			actionColor, advice.Confidence, advice.Action, Reset)
		fmt.Printf("  原因: %s\n", advice.Reason)
	}
	fmt.Printf("%s└──────────────────────────────────────────────────────────────┘%s\n", Cyan, Reset)

	// ── AI 图鉴 ──
	fmt.Printf("\n%s┌────────────────────── 🤖 盘口AI图鉴 ────────────────────────┐%s\n", Cyan, Reset)
	sortedAIs := make([]*AI, len(state.AIs))
	copy(sortedAIs, state.AIs)
	sort.Slice(sortedAIs, func(i, j int) bool { return sortedAIs[i].Shares > sortedAIs[j].Shares })

	for _, ai := range sortedAIs {
		aiProfit := 0.0
		if ai.Cost > 0 {
			aiProfit = ((state.Price - ai.Cost) / ai.Cost) * 100
		}
		statusStr := fmt.Sprintf("%s%+.0f%%%s", Red, aiProfit, Reset)
		if ai.HasSold || ai.Shares == 0 { statusStr = "\033[90m空仓伺机\033[0m" }

		nameColor := Cyan
		if ai.Type == "Whale" { nameColor = Purple } else if ai.Type == "Quant" { nameColor = Blue } else { nameColor = Green }

		opinionColor := Reset
		if ai.Shares == 0 { opinionColor = "\033[90m" }

		statusFlagStr := ""
		if ai.StatusFlag == "Spoofing" { statusFlagStr = " " + Purple + "[挂假单吓人]" + Reset }
		if ai.StatusFlag == "GridTrading" { statusFlagStr = " " + Blue + "[网格挂单]" + Reset }
		if ai.StatusFlag == "Bailout" { statusFlagStr = " " + Red + "[砸锅卖铁救市]" + Reset }
		if ai.StatusFlag == "ForcedLiquidation" { statusFlagStr = " " + Yellow + "[资不抵债强平!]" + Reset }

		fmt.Printf("  %s[%s] %-12s%s│ %5d股 │ %s │ %s%s%s%s\n",
			nameColor, ai.SubType, ai.Name, Reset,
			ai.Shares, statusStr,
			opinionColor, ai.LastOpinion, Reset, statusFlagStr)
	}
	fmt.Printf("%s└──────────────────────────────────────────────────────────────┘%s\n", Cyan, Reset)
}


// 渲染战后复盘分析
func renderTradeRecap(state *GameState) {
	if len(state.PriceHistory) == 0 { return }

	fmt.Printf("\n" + Cyan + "📊 【上帝视角：全场博弈复盘】" + Reset + "\n")
	
	// 1. 渲染 K 线标尺
	history := state.PriceHistory
	bars := []rune(" ▂▃▄▅▆▇█")
	minP, maxP := history[0], history[0]
	for _, v := range history {
		if v < minP { minP = v }
		if v > maxP { maxP = v }
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
		case "早盘": sIdx = 0
		case "盘中上午": sIdx = 1
		case "盘中下午": sIdx = 2
		case "尾盘": sIdx = 3
		}
		// Index logic: day 1 start is 0. 
		// Day 1 Morning result is index 1.
		idx := (tp.Day-1)*4 + sIdx + 1
		if idx < len(history) {
			char := ""
			if tp.Action == "Buy" { char = Green + "B" + Reset } else { char = Red + "S" + Reset }
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
			if tp.Action == "Sell" { icon = "🔴" }
			
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
func renderGameOver(state *GameState) {
	fmt.Println("\n" + Yellow + "==================== 终局审判 ====================" + Reset)
	fmt.Printf(Cyan + "主题模式: %s\n" + Reset, CurrentTheme.Name)

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
	fmt.Println("\n" + Cyan + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + Reset)

	// 保存战绩到历史记录
	record := GameRecord{
		Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
		Theme:         CurrentTheme.Name,
		FinalProfit:   finalProfit,
		Grade:         rank.Grade,
		Score:         rank.Score,
		IsCrashed:     state.IsCrashed,
		PlayerSold:    hasSoldBefore && !isHoldingAtEnd, // 卖出后未再买回
		SoldDay:       state.PlayerSoldDay,
		SoldSession:   state.PlayerSoldSession,
		ProfitScore:   rank.ProfitScore,
		TimingScore:   rank.TimingScore,
		RiskScore:     rank.RiskScore,
	}

	err := saveGameRecord(record)
	if err != nil {
		fmt.Printf(Red+"\n⚠️ 保存战绩失败: %v\n"+Reset, err)
	} else {
		fmt.Println(Green + "\n✅ 战绩已保存到历史记录" + Reset)
	}
}
