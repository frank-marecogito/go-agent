// Package hello 提供智能问候语生成功能
// 支持多种场景模板和语气调节
package hello

import (
	"fmt"
	"time"
)

// SceneType 表示问候场景类型
type SceneType string

const (
	// SceneBusiness 商务场景 - 正式、专业
	SceneBusiness SceneType = "business"
	// SceneCasual 非正式场景 - 轻松、友好
	SceneCasual SceneType = "casual"
	// SceneFormal 正式场景 - 庄重、礼貌
	SceneFormal SceneType = "formal"
	// SceneCelebration 庆祝场景 - 热情、喜悦
	SceneCelebration SceneType = "celebration"
	// SceneCondolence 慰问场景 - 关切、温暖
	SceneCondolence SceneType = "condolence"
	// SceneDaily 日常场景 - 自然、亲切
	SceneDaily SceneType = "daily"
	// SceneFestival 节日场景 - 喜庆、祝福
	SceneFestival SceneType = "festival"
)

// ToneType 表示语气类型
type ToneType string

const (
	// ToneProfessional 专业语气 - 正式、严谨
	ToneProfessional ToneType = "professional"
	// ToneFriendly 友好语气 - 亲切、温暖
	ToneFriendly ToneType = "friendly"
	// ToneEnthusiastic 热情语气 - 活力、积极
	ToneEnthusiastic ToneType = "enthusiastic"
	// ToneWarm 温暖语气 - 关怀、体贴
	ToneWarm ToneType = "warm"
	// ToneHumorous 幽默语气 - 轻松、有趣
	ToneHumorous ToneType = "humorous"
)

// LengthType 表示问候语长度
type LengthType string

const (
	// LengthShort 短问候 - 简洁明了
	LengthShort LengthType = "short"
	// LengthMedium 中等长度 - 平衡
	LengthMedium LengthType = "medium"
	// LengthLong 长问候 - 详细周到
	LengthLong LengthType = "long"
)

// SceneTemplate 场景模板结构
type SceneTemplate struct {
	ID          string            // 场景 ID
	Name        string            // 场景名称
	Type        SceneType         // 场景类型
	Description string            // 场景描述
	Templates   []TemplateVariant // 模板变体
	Keywords    []string          // 相关关键词
}

// TemplateVariant 模板变体（支持不同长度和语气）
type TemplateVariant struct {
	Length    LengthType          // 长度类型
	Tone      ToneType            // 语气类型
	Patterns  []string            // 问候模式
	Examples  []string            // 示例
	Variables []TemplateVariable  // 可替换变量
}

// TemplateVariable 模板变量
type TemplateVariable struct {
	Name        string   // 变量名
	Description string   // 变量描述
	Default     string   // 默认值
	Options     []string // 可选值
}

// GetPresetScenes 获取预设场景模板（5+ 种场景）
func GetPresetScenes() []SceneTemplate {
	return []SceneTemplate{
		{
			ID:          "business_meeting",
			Name:        "商务会议",
			Type:        SceneBusiness,
			Description: "适用于商务会议、客户拜访等正式商务场合",
			Keywords:    []string{"会议", "商务", "客户", "合作", "洽谈"},
			Templates: []TemplateVariant{
				{
					Length: LengthShort,
					Tone:   ToneProfessional,
					Patterns: []string{
						"尊敬的{recipient}，您好！感谢您抽出宝贵时间。",
						"{recipient}您好，期待与您的会面。",
					},
					Examples: []string{
						"尊敬的张总，您好！感谢您抽出宝贵时间。",
						"李经理您好，期待与您的会面。",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "收件人称呼", Default: "您", Options: []string{}},
					},
				},
				{
					Length: LengthMedium,
					Tone:   ToneProfessional,
					Patterns: []string{
						"尊敬的{recipient}，您好！非常荣幸能有机会与您会面。期待我们今天的交流能够富有成效，为未来的合作奠定良好基础。",
						"{recipient}您好！感谢您百忙之中抽空参加此次会议。我们已做好充分准备，期待与您深入探讨合作事宜。",
					},
					Examples: []string{
						"尊敬的王总，您好！非常荣幸能有机会与您会面。期待我们今天的交流能够富有成效，为未来的合作奠定良好基础。",
						"赵总监您好！感谢您百忙之中抽空参加此次会议。我们已做好充分准备，期待与您深入探讨合作事宜。",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "收件人称呼", Default: "您", Options: []string{}},
					},
				},
				{
					Length: LengthLong,
					Tone:   ToneProfessional,
					Patterns: []string{
						"尊敬的{recipient}，您好！首先，请允许我代表团队向您致以诚挚的问候。非常感谢您对我们工作的支持与信任。今天的会议对我们而言意义非凡，我们已做了充分的准备，期待与您就合作细节进行深入交流。相信通过我们的共同努力，一定能够达成互利共赢的合作成果。",
					},
					Examples: []string{
						"尊敬的陈总，您好！首先，请允许我代表团队向您致以诚挚的问候。非常感谢您对我们工作的支持与信任。今天的会议对我们而言意义非凡，我们已做了充分的准备，期待与您就合作细节进行深入交流。相信通过我们的共同努力，一定能够达成互利共赢的合作成果。",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "收件人称呼", Default: "您", Options: []string{}},
					},
				},
			},
		},
		{
			ID:          "casual_friend",
			Name:        "朋友聚会",
			Type:        SceneCasual,
			Description: "适用于朋友聚会、休闲聊天等轻松场合",
			Keywords:    []string{"朋友", "聚会", "休闲", "聊天", "放松"},
			Templates: []TemplateVariant{
				{
					Length: LengthShort,
					Tone:   ToneFriendly,
					Patterns: []string{
						"嘿{recipient}！好久不见，最近怎么样？",
						"嗨{recipient}，想你了！有空聚聚吗？",
					},
					Examples: []string{
						"嘿小明！好久不见，最近怎么样？",
						"嗨小红，想你了！有空聚聚吗？",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "朋友称呼", Default: "朋友", Options: []string{}},
					},
				},
				{
					Length: LengthMedium,
					Tone:   ToneFriendly,
					Patterns: []string{
						"嗨{recipient}！最近过得怎么样？好久没联系了，挺想你的。找个时间一起出来坐坐，喝杯咖啡聊聊天吧！",
						"{recipient}，好久不见啦！看你朋友圈最近挺忙的，要注意休息哦。有空的话我们一起吃个饭吧！",
					},
					Examples: []string{
						"嗨小明！最近过得怎么样？好久没联系了，挺想你的。找个时间一起出来坐坐，喝杯咖啡聊聊天吧！",
						"小红，好久不见啦！看你朋友圈最近挺忙的，要注意休息哦。有空的话我们一起吃个饭吧！",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "朋友称呼", Default: "朋友", Options: []string{}},
					},
				},
				{
					Length: LengthLong,
					Tone:   ToneFriendly,
					Patterns: []string{
						"嘿{recipient}！真的好久没见了，算算都有大半年了吧？最近工作忙吗？生活还顺利吗？我一直记得咱们以前一起度过的那些快乐时光，真的很怀念。找个周末出来聚聚吧，叫上其他朋友，咱们好好聊聊近况，分享一下各自的故事。期待见到你！",
					},
					Examples: []string{
						"嘿小明！真的好久没见了，算算都有大半年了吧？最近工作忙吗？生活还顺利吗？我一直记得咱们以前一起度过的那些快乐时光，真的很怀念。找个周末出来聚聚吧，叫上其他朋友，咱们好好聊聊近况，分享一下各自的故事。期待见到你！",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "朋友称呼", Default: "朋友", Options: []string{}},
					},
				},
			},
		},
		{
			ID:          "formal_ceremony",
			Name:        "正式典礼",
			Type:        SceneFormal,
			Description: "适用于婚礼、毕业典礼、颁奖仪式等正式场合",
			Keywords:    []string{"典礼", "仪式", "庆典", "正式", "庄重"},
			Templates: []TemplateVariant{
				{
					Length: LengthShort,
					Tone:   ToneProfessional,
					Patterns: []string{
						"尊敬的{recipient}，谨向您致以最诚挚的祝贺。",
						"{recipient}，恭贺您迎来这重要时刻。",
					},
					Examples: []string{
						"尊敬的王先生，谨向您致以最诚挚的祝贺。",
						"李女士，恭贺您迎来这重要时刻。",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "收件人称呼", Default: "您", Options: []string{}},
					},
				},
				{
					Length: LengthMedium,
					Tone:   ToneProfessional,
					Patterns: []string{
						"尊敬的{recipient}，在这喜庆的日子里，谨向您及家人致以最热烈的祝贺。愿这美好时刻成为您人生中难忘的回忆，祝福您未来更加辉煌。",
						"{recipient}，欣闻喜讯，不胜欢欣。在此庄重时刻，向您表达我最诚挚的祝福。愿幸福与荣耀常伴您左右。",
					},
					Examples: []string{
						"尊敬的张先生，在这喜庆的日子里，谨向您及家人致以最热烈的祝贺。愿这美好时刻成为您人生中难忘的回忆，祝福您未来更加辉煌。",
						"李女士，欣闻喜讯，不胜欢欣。在此庄重时刻，向您表达我最诚挚的祝福。愿幸福与荣耀常伴您左右。",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "收件人称呼", Default: "您", Options: []string{}},
					},
				},
				{
					Length: LengthLong,
					Tone:   ToneProfessional,
					Patterns: []string{
						"尊敬的{recipient}，值此隆重而美好的时刻，我谨代表个人向您致以最崇高的敬意和最热烈的祝贺。这一重要里程碑见证了您的卓越成就与不懈努力，令人由衷钦佩。愿此庆典成为您人生新篇章的华丽开端，祝愿您在未来的道路上继续绽放光彩，收获更多幸福与成功。",
					},
					Examples: []string{
						"尊敬的陈先生，值此隆重而美好的时刻，我谨代表个人向您致以最崇高的敬意和最热烈的祝贺。这一重要里程碑见证了您的卓越成就与不懈努力，令人由衷钦佩。愿此庆典成为您人生新篇章的华丽开端，祝愿您在未来的道路上继续绽放光彩，收获更多幸福与成功。",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "收件人称呼", Default: "您", Options: []string{}},
					},
				},
			},
		},
		{
			ID:          "celebration_birthday",
			Name:        "生日庆祝",
			Type:        SceneCelebration,
			Description: "适用于生日派对、周年纪念等庆祝场合",
			Keywords:    []string{"生日", "庆祝", "周年", "派对", "祝福"},
			Templates: []TemplateVariant{
				{
					Length: LengthShort,
					Tone:   ToneEnthusiastic,
					Patterns: []string{
						"{recipient}，生日快乐！🎂 祝你今天超级开心！",
						"Happy Birthday {recipient}！愿你的一天充满惊喜和欢乐！",
					},
					Examples: []string{
						"小明，生日快乐！🎂 祝你今天超级开心！",
						"Happy Birthday 小红！愿你的一天充满惊喜和欢乐！",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "寿星称呼", Default: "你", Options: []string{}},
					},
				},
				{
					Length: LengthMedium,
					Tone:   ToneEnthusiastic,
					Patterns: []string{
						"亲爱的{recipient}，生日快乐！🎉 在这个特别的日子里，祝你身体健康、事业顺利、天天开心！愿新的一岁带给你更多美好与惊喜！",
						"{recipient}，今天是你的大日子！生日快乐！🎁 感谢有你出现在我的生命中，愿你的笑容永远灿烂如阳光！",
					},
					Examples: []string{
						"亲爱的小明，生日快乐！🎉 在这个特别的日子里，祝你身体健康、事业顺利、天天开心！愿新的一岁带给你更多美好与惊喜！",
						"小红，今天是你的大日子！生日快乐！🎁 感谢有你出现在我的生命中，愿你的笑容永远灿烂如阳光！",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "寿星称呼", Default: "你", Options: []string{}},
					},
				},
				{
					Length: LengthLong,
					Tone:   ToneEnthusiastic,
					Patterns: []string{
						"最最亲爱的{recipient}，生日快乐！🎂🎉🎁 今天是属于你的闪耀日子，整个世界都在为你庆祝！回顾过去的一年，你取得了那么多令人骄傲的成就；展望新的一岁，相信会有更多美好在等着你。愿你永远保持这份热情与活力，追逐梦想，收获幸福。今天，请尽情享受属于你的时刻，因为你是最棒的！",
					},
					Examples: []string{
						"最最爱的小明，生日快乐！🎂🎉🎁 今天是属于你的闪耀日子，整个世界都在为你庆祝！回顾过去的一年，你取得了那么多令人骄傲的成就；展望新的一岁，相信会有更多美好在等着你。愿你永远保持这份热情与活力，追逐梦想，收获幸福。今天，请尽情享受属于你的时刻，因为你是最棒的！",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "寿星称呼", Default: "你", Options: []string{}},
					},
				},
			},
		},
		{
			ID:          "condolence_sympathy",
			Name:        "慰问关怀",
			Type:        SceneCondolence,
			Description: "适用于慰问病人、悼念逝者、安慰失意者等场合",
			Keywords:    []string{"慰问", "关怀", "安慰", "支持", "陪伴"},
			Templates: []TemplateVariant{
				{
					Length: LengthShort,
					Tone:   ToneWarm,
					Patterns: []string{
						"{recipient}，请保重身体。我们都在你身边。",
						"亲爱的{recipient}，愿你早日恢复健康。",
					},
					Examples: []string{
						"张叔叔，请保重身体。我们都在你身边。",
						"亲爱的小红，愿你早日恢复健康。",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "收件人称呼", Default: "您", Options: []string{}},
					},
				},
				{
					Length: LengthMedium,
					Tone:   ToneWarm,
					Patterns: []string{
						"亲爱的{recipient}，得知你最近的情况，我非常牵挂。请相信困难只是暂时的，好好照顾自己。有任何需要，随时告诉我，我会尽力帮助你。",
						"{recipient}，在这个艰难的时刻，请记得你并不孤单。我们都在关心着你，支持着你。愿你坚强面对，早日走出阴霾。",
					},
					Examples: []string{
						"亲爱的小明，得知你最近的情况，我非常牵挂。请相信困难只是暂时的，好好照顾自己。有任何需要，随时告诉我，我会尽力帮助你。",
						"小红，在这个艰难的时刻，请记得你并不孤单。我们都在关心着你，支持着你。愿你坚强面对，早日走出阴霾。",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "收件人称呼", Default: "您", Options: []string{}},
					},
				},
				{
					Length: LengthLong,
					Tone:   ToneWarm,
					Patterns: []string{
						"亲爱的{recipient}，听闻你正在经历一段艰难时光，我的心中满是牵挂。人生路上难免会遇到风雨，但请相信，阳光总在风雨后。这段日子里，请允许自己休息，允许自己脆弱，因为坚强不是独自硬撑，而是懂得接受关爱。请记住，你从来都不是一个人——家人、朋友，还有我，都在你身边，愿意为你分担、为你守候。愿你慢慢好起来，愿温暖与希望常伴你左右。",
					},
					Examples: []string{
						"亲爱的小明，听闻你正在经历一段艰难时光，我的心中满是牵挂。人生路上难免会遇到风雨，但请相信，阳光总在风雨后。这段日子里，请允许自己休息，允许自己脆弱，因为坚强不是独自硬撑，而是懂得接受关爱。请记住，你从来都不是一个人——家人、朋友，还有我，都在你身边，愿意为你分担、为你守候。愿你慢慢好起来，愿温暖与希望常伴你左右。",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "收件人称呼", Default: "您", Options: []string{}},
					},
				},
			},
		},
		{
			ID:          "daily_greeting",
			Name:        "日常问候",
			Type:        SceneDaily,
			Description: "适用于日常打招呼、晨间问候、晚安问候等",
			Keywords:    []string{"日常", "早安", "晚安", "问候", "关心"},
			Templates: []TemplateVariant{
				{
					Length: LengthShort,
					Tone:   ToneFriendly,
					Patterns: []string{
						"早安{recipient}！今天也要加油哦！☀️",
						"晚安{recipient}，好梦！🌙",
						"嗨{recipient}，今天过得怎么样？",
					},
					Examples: []string{
						"早安小明！今天也要加油哦！☀️",
						"晚安小红，好梦！🌙",
						"嗨小李，今天过得怎么样？",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "收件人称呼", Default: "你", Options: []string{}},
					},
				},
				{
					Length: LengthMedium,
					Tone:   ToneFriendly,
					Patterns: []string{
						"早上好{recipient}！新的一天开始了，愿你今天心情愉快，工作顺利。记得吃早餐，照顾好自己哦！",
						"{recipient}，晚上好！忙碌了一天，辛苦啦。早点休息，明天又是充满活力的一天。晚安，好梦！",
					},
					Examples: []string{
						"早上好小明！新的一天开始了，愿你今天心情愉快，工作顺利。记得吃早餐，照顾好自己哦！",
						"小红，晚上好！忙碌了一天，辛苦啦。早点休息，明天又是充满活力的一天。晚安，好梦！",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "收件人称呼", Default: "你", Options: []string{}},
					},
				},
				{
					Length: LengthLong,
					Tone:   ToneFriendly,
					Patterns: []string{
						"亲爱的{recipient}，早安！清晨的阳光洒进窗台，新的一天又开始了。愿你今天带着好心情迎接每一个挑战，收获满满的成就与快乐。工作再忙也要记得适当休息，多喝水，按时吃饭。期待听到你分享今天的精彩故事！",
					},
					Examples: []string{
						"亲爱的小明，早安！清晨的阳光洒进窗台，新的一天又开始了。愿你今天带着好心情迎接每一个挑战，收获满满的成就与快乐。工作再忙也要记得适当休息，多喝水，按时吃饭。期待听到你分享今天的精彩故事！",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "收件人称呼", Default: "你", Options: []string{}},
					},
				},
			},
		},
		{
			ID:          "festival_spring",
			Name:        "春节祝福",
			Type:        SceneFestival,
			Description: "适用于春节、新年等传统节日祝福",
			Keywords:    []string{"春节", "新年", "节日", "祝福", "团圆"},
			Templates: []TemplateVariant{
				{
					Length: LengthShort,
					Tone:   ToneEnthusiastic,
					Patterns: []string{
						"{recipient}，新年快乐！🧧 恭喜发财，万事如意！",
						"春节快乐{recipient}！祝你阖家幸福，身体健康！",
					},
					Examples: []string{
						"张叔叔，新年快乐！🧧 恭喜发财，万事如意！",
						"春节快乐小红！祝你阖家幸福，身体健康！",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "收件人称呼", Default: "您", Options: []string{}},
					},
				},
				{
					Length: LengthMedium,
					Tone:   ToneEnthusiastic,
					Patterns: []string{
						"亲爱的{recipient}，新春佳节到，给您拜年啦！🎊 祝您在新的一年里身体健康、事业有成、家庭幸福、万事如意！感谢过去一年的关照，期待新的一年继续携手同行！",
						"{recipient}，春节快乐！🏮 辞旧迎新之际，愿你烦恼清零，快乐加倍；愿好运常伴，幸福安康。新的一年，我们一起加油！",
					},
					Examples: []string{
						"亲爱的小明，新春佳节到，给您拜年啦！🎊 祝您在新的一年里身体健康、事业有成、家庭幸福、万事如意！感谢过去一年的关照，期待新的一年继续携手同行！",
						"小红，春节快乐！🏮 辞旧迎新之际，愿你烦恼清零，快乐加倍；愿好运常伴，幸福安康。新的一年，我们一起加油！",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "收件人称呼", Default: "您", Options: []string{}},
					},
				},
				{
					Length: LengthLong,
					Tone:   ToneEnthusiastic,
					Patterns: []string{
						"最最亲爱的{recipient}，新春大吉！🎉🧨🏮 爆竹声中一岁除，春风送暖入屠苏。在这辞旧迎新的美好时刻，我衷心祝愿你和家人：新春快乐、身体健康、事业腾飞、财源广进、阖家幸福！感谢过去一年里你的陪伴与支持，那些共同度过的美好时光我都珍藏在心。新的一年，愿我们继续携手同行，创造更多精彩回忆。愿你每一天都充满阳光与欢笑，每一步都走得坚定而从容。春节快乐，万事胜意！",
					},
					Examples: []string{
						"最最爱的小明，新春大吉！🎉🧨🏮 爆竹声中一岁除，春风送暖入屠苏。在这辞旧迎新的美好时刻，我衷心祝愿你和家人：新春快乐、身体健康、事业腾飞、财源广进、阖家幸福！感谢过去一年里你的陪伴与支持，那些共同度过的美好时光我都珍藏在心。新的一年，愿我们继续携手同行，创造更多精彩回忆。愿你每一天都充满阳光与欢笑，每一步都走得坚定而从容。春节快乐，万事胜意！",
					},
					Variables: []TemplateVariable{
						{Name: "recipient", Description: "收件人称呼", Default: "您", Options: []string{}},
					},
				},
			},
		},
	}
}

// GetSceneByID 根据 ID 获取场景模板
func GetSceneByID(id string) (*SceneTemplate, error) {
	scenes := GetPresetScenes()
	for _, scene := range scenes {
		if scene.ID == id {
			return &scene, nil
		}
	}
	return nil, fmt.Errorf("scene not found: %s", id)
}

// GetScenesByType 根据类型获取场景模板
func GetScenesByType(sceneType SceneType) []SceneTemplate {
	scenes := GetPresetScenes()
	var result []SceneTemplate
	for _, scene := range scenes {
		if scene.Type == sceneType {
			result = append(result, scene)
		}
	}
	return result
}

// GetTimeBasedGreeting 根据时间获取合适的日常问候
func GetTimeBasedGreeting(recipient string) string {
	hour := time.Now().Hour()
	
	if hour >= 5 && hour < 12 {
		return fmt.Sprintf("早安%s！新的一天开始了，愿你今天心情愉快！☀️", recipient)
	} else if hour >= 12 && hour < 18 {
		return fmt.Sprintf("下午好%s！午后时光，记得适当休息哦！☕", recipient)
	} else if hour >= 18 && hour < 23 {
		return fmt.Sprintf("晚上好%s！忙碌了一天，辛苦啦！🌆", recipient)
	} else {
		return fmt.Sprintf("晚安%s，好梦！🌙", recipient)
	}
}
