package story

type StoryTemplate struct {
	ID          string
	Title       string
	Description string
	Condition   string
	Effects     string
}

func DefaultTemplates() []StoryTemplate {
	return []StoryTemplate{
		{
			ID: "ST-001", Title: "钟楼故障",
			Description: "广场钟楼突然停摆，齿轮出现异常",
			Condition:   `{"time_range":"morning","probability":0.15,"cooldown_hours":8}`,
			Effects:     `{"npc_effects":[{"npc_name":"奥托","mood":"anxious","goal":"修好钟楼"},{"npc_name":"埃德蒙","mood":"concerned"},{"npc_name":"莉娜","mood":"caring","goal":"帮助奥托"}]}`,
		},
		{
			ID: "ST-002", Title: "咖啡馆新菜单",
			Description: "莉娜试制了一款春季限定饮品",
			Condition:   `{"time_range":"noon","probability":0.1,"cooldown_hours":12}`,
			Effects:     `{"npc_effects":[{"npc_name":"莉娜","mood":"excited","goal":"邀请镇民试喝"},{"npc_name":"皮埃尔","mood":"excited","goal":"搭配新面包"}]}`,
		},
		{
			ID: "ST-003", Title: "邮件丢失",
			Description: "米娅在送信途中发现少了一封信",
			Condition:   `{"time_range":"afternoon","probability":0.1,"cooldown_hours":12}`,
			Effects:     `{"npc_effects":[{"npc_name":"米娅","mood":"worried","goal":"寻找丢失邮件"},{"npc_name":"埃德蒙","mood":"concerned"},{"npc_name":"薇拉","mood":"helpful","goal":"追踪线索"}]}`,
		},
		{
			ID: "ST-004", Title: "神秘旅人到访",
			Description: "一位陌生的旅人出现在小镇，带来外面世界的消息",
			Condition:   `{"time_range":"any","probability":0.05,"cooldown_hours":24}`,
			Effects:     `{"npc_effects":[{"npc_name":"薇拉","mood":"curious","goal":"了解来意"},{"npc_name":"玛莎","mood":"curious","goal":"招待旅人"},{"npc_name":"埃德蒙","mood":"cautious","goal":"确保小镇安全"}]}`,
		},
		{
			ID: "ST-005", Title: "广场集市日",
			Description: "每月一次的集市日，全镇参与",
			Condition:   `{"time_range":"morning","trigger_day":"market_day","probability":1.0}`,
			Effects:     `{"npc_effects":[{"npc_name":"菲奥娜","mood":"excited","goal":"准备花展"},{"npc_name":"皮埃尔","mood":"excited","goal":"准备面包"},{"npc_name":"杰克","mood":"happy","goal":"摆出蔬菜"},{"npc_name":"卢卡斯","mood":"inspired","goal":"集市演奏"}]}`,
		},
	}
}
