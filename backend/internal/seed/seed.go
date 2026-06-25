package seed

import (
	"backend/internal/logger"
	"backend/internal/model"

	"gorm.io/gorm"
)

func SeedDemo(db *gorm.DB, appLog *logger.AppLogger) {
	var count int64
	db.Model(&model.Town{}).Where("name = ?", "晨曦镇").Count(&count)
	if count > 0 {
		appLog.Info("种子数据已存在，跳过")
		return
	}

	appLog.Info("开始导入种子数据（15 NPC / 17 地点）")

	err := db.Transaction(func(tx *gorm.DB) error {
		town := model.Town{Name: "晨曦镇", CurrentDay: 0, CurrentMinute: 360}
		if err := tx.Create(&town).Error; err != nil {
			return err
		}

		locations := []model.Location{
			{Name: "广场", Longitude: "0", Latitude: "0"},
			{Name: "咖啡馆", Longitude: "0.5", Latitude: "0.3"},
			{Name: "钟楼", Longitude: "-0.3", Latitude: "0.8"},
			{Name: "市政厅", Longitude: "0", Latitude: "0.2"},
			{Name: "图书馆", Longitude: "0.2", Latitude: "-0.1"},
			{Name: "花店", Longitude: "0.4", Latitude: "0.2"},
			{Name: "铁匠铺", Longitude: "0.6", Latitude: "0.6"},
			{Name: "诊所", Longitude: "-0.1", Latitude: "-0.2"},
			{Name: "农舍", Longitude: "-0.8", Latitude: "0.1"},
			{Name: "钓鱼小屋", Longitude: "-0.2", Latitude: "0.9"},
			{Name: "学校", Longitude: "-0.1", Latitude: "-0.3"},
			{Name: "面包店", Longitude: "0.5", Latitude: "0.1"},
			{Name: "酒馆", Longitude: "0.7", Latitude: "0.5"},
			{Name: "公园凉亭", Longitude: "0.2", Latitude: "-0.5"},
			{Name: "手工工坊", Longitude: "-0.1", Latitude: "0.1"},
			{Name: "住宅区", Longitude: "0.1", Latitude: "-0.7"},
			{Name: "森林营地", Longitude: "0.9", Latitude: "0.4"},
		}
		for i := range locations {
			locations[i].TownID = int64(town.ID)
		}
		if err := tx.Create(&locations).Error; err != nil {
			return err
		}

		locMap := make(map[string]uint, len(locations))
		for _, l := range locations {
			locMap[l.Name] = l.ID
		}

		npcs := []model.NPC{
			{Name: "埃德蒙", Role: "镇长", Gender: "男", AgeGroup: "老年", Personality: "慈祥有威严，关心每一位镇民", Catchphrase: "小镇虽小，每个人都很重要。", Appearance: "灰白头发，整洁胡须，圆框眼镜，棕色西装", Mood: "content", Energy: 70, CurrentGoal: "巡视小镇", TownID: town.ID, LocationID: locMap["市政厅"], Status: "idle"},
			{Name: "莉娜", Role: "咖啡馆主", Gender: "女", AgeGroup: "青年", Personality: "温暖开朗，善于倾听，小镇情报中心", Catchphrase: "来杯咖啡吗？今天的故事配咖啡正好。", Appearance: "棕色发髻，绿色围裙配米色衬衫", Mood: "cheerful", Energy: 90, CurrentGoal: "准备开店", TownID: town.ID, LocationID: locMap["咖啡馆"], Status: "idle"},
			{Name: "艾琳", Role: "图书管理员", Gender: "女", AgeGroup: "中年", Personality: "安静温柔，知识渊博，观察力敏锐", Catchphrase: "每本书里都藏着一个世界。", Appearance: "长发披肩，圆框眼镜，素色长裙", Mood: "calm", Energy: 75, CurrentGoal: "整理新到书籍", TownID: town.ID, LocationID: locMap["图书馆"], Status: "idle"},
			{Name: "菲奥娜", Role: "花店店主", Gender: "女", AgeGroup: "青年", Personality: "活泼阳光，热爱自然，浪漫主义", Catchphrase: "每一朵花都有它的花语。", Appearance: "金色短发，碎花围裙，笑容灿烂", Mood: "happy", Energy: 85, CurrentGoal: "布置春季花卉", TownID: town.ID, LocationID: locMap["花店"], Status: "idle"},
			{Name: "奥托", Role: "铁匠", Gender: "男", AgeGroup: "中年", Personality: "固执但善良，对钟楼有偏执的责任感", Catchphrase: "好的手艺需要时间，就像好的钟表一样。", Appearance: "深棕短发，浓密胡须，皮围裙，持锤", Mood: "focused", Energy: 80, CurrentGoal: "检查钟楼", TownID: town.ID, LocationID: locMap["钟楼"], Status: "idle"},
			{Name: "克莱尔", Role: "医生", Gender: "女", AgeGroup: "中年", Personality: "温和专业，细心负责，像小镇的母亲", Catchphrase: "预防永远比治疗重要。", Appearance: "盘发，白大褂，听诊器挂在颈间", Mood: "composed", Energy: 70, CurrentGoal: "整理病历", TownID: town.ID, LocationID: locMap["诊所"], Status: "idle"},
			{Name: "杰克", Role: "农夫", Gender: "男", AgeGroup: "青年", Personality: "朴实勤劳，乐观豁达，和自然关系亲密", Catchphrase: "土地不会骗人，你付出多少它就回报多少。", Appearance: "健壮身材，草帽，粗布衣", Mood: "content", Energy: 75, CurrentGoal: "照料春苗", TownID: town.ID, LocationID: locMap["农舍"], Status: "idle"},
			{Name: "沃尔特", Role: "渔夫", Gender: "男", AgeGroup: "老年", Personality: "沉稳耐心，话不多但有智慧", Catchphrase: "钓鱼教会我一件事：好东西值得等待。", Appearance: "花白胡须，旧渔帽，手持钓竿", Mood: "peaceful", Energy: 65, CurrentGoal: "湖边垂钓", TownID: town.ID, LocationID: locMap["钓鱼小屋"], Status: "idle"},
			{Name: "索菲亚", Role: "教师", Gender: "女", AgeGroup: "青年", Personality: "温柔耐心，充满教育热情，喜欢孩子", Catchphrase: "每个孩子都是一颗等待发芽的种子。", Appearance: "淡蓝连衣裙，书本常抱怀中", Mood: "warm", Energy: 80, CurrentGoal: "准备今日课程", TownID: town.ID, LocationID: locMap["学校"], Status: "idle"},
			{Name: "皮埃尔", Role: "面包师", Gender: "男", AgeGroup: "中年", Personality: "乐观开朗，热爱烘焙，有点小骄傲", Catchphrase: "新鲜出炉的幸福，谁要来一块？", Appearance: "微胖身材，白色厨师帽，面粉沾在围裙上", Mood: "jolly", Energy: 85, CurrentGoal: "烤今日第一批面包", TownID: town.ID, LocationID: locMap["面包店"], Status: "idle"},
			{Name: "玛莎", Role: "酒馆老板", Gender: "女", AgeGroup: "中年", Personality: "豪爽直率，见多识广，夜晚的守护者", Catchphrase: "无论白天多糟糕，一杯好酒和一个好听众总能解决问题。", Appearance: "红发挽髻，琥珀耳环，暖色长裙", Mood: "friendly", Energy: 70, CurrentGoal: "准备午餐食材", TownID: town.ID, LocationID: locMap["酒馆"], Status: "idle"},
			{Name: "卢卡斯", Role: "音乐家", Gender: "男", AgeGroup: "青年", Personality: "浪漫自由，略带忧郁，艺术气质浓厚", Catchphrase: "音乐是心灵的语言，不需要翻译。", Appearance: "长发束后，随身携带乐器，神情梦幻", Mood: "dreamy", Energy: 65, CurrentGoal: "寻找新旋律灵感", TownID: town.ID, LocationID: locMap["公园凉亭"], Status: "idle"},
			{Name: "托马斯", Role: "木匠", Gender: "男", AgeGroup: "中年", Personality: "稳重可靠，手艺精湛，寡言实干", Catchphrase: "好木头会说话，你得学会听。", Appearance: "粗壮手臂，帆布围裙，腰间挂着卷尺", Mood: "steady", Energy: 80, CurrentGoal: "修复广场长椅", TownID: town.ID, LocationID: locMap["手工工坊"], Status: "idle"},
			{Name: "米娅", Role: "小女孩", Gender: "女", AgeGroup: "儿童", Personality: "活泼好奇，天真可爱，喜欢帮大人跑腿", Catchphrase: "我能帮忙！我已经不是小孩子了！", Appearance: "浅棕双马尾，春黄色连衣裙，抱玩具兔", Mood: "playful", Energy: 95, CurrentGoal: "上学/送信", TownID: town.ID, LocationID: locMap["住宅区"], Status: "idle"},
			{Name: "薇拉", Role: "冒险者", Gender: "女", AgeGroup: "青年", Personality: "独立神秘，自信潇洒，见过外面的世界", Catchphrase: "森林外面还有更大的世界，但这里才是家。", Appearance: "紫色短发，单眼眼罩，深绿斗篷配皮甲", Mood: "confident", Energy: 75, CurrentGoal: "探索森林边缘", TownID: town.ID, LocationID: locMap["森林营地"], Status: "idle"},
		}
		if err := tx.Create(&npcs).Error; err != nil {
			return err
		}

		schedules := []model.NPCSchedule{
			// 埃德蒙
			{NPCID: npcs[0].ID, StartTime: "08:00", EndTime: "10:00", LocationID: locMap["市政厅"], Action: "办公"},
			{NPCID: npcs[0].ID, StartTime: "10:00", EndTime: "12:00", LocationID: locMap["广场"], Action: "散步巡视"},
			{NPCID: npcs[0].ID, StartTime: "12:00", EndTime: "14:00", LocationID: locMap["咖啡馆"], Action: "午餐"},
			{NPCID: npcs[0].ID, StartTime: "14:00", EndTime: "17:00", LocationID: locMap["市政厅"], Action: "处理公务"},
			{NPCID: npcs[0].ID, StartTime: "17:00", EndTime: "18:00", LocationID: locMap["广场"], Action: "傍晚散步"},
			// 莉娜
			{NPCID: npcs[1].ID, StartTime: "07:30", EndTime: "12:00", LocationID: locMap["咖啡馆"], Action: "开店营业"},
			{NPCID: npcs[1].ID, StartTime: "12:00", EndTime: "14:00", LocationID: locMap["咖啡馆"], Action: "午间忙时"},
			{NPCID: npcs[1].ID, StartTime: "14:00", EndTime: "15:00", LocationID: locMap["广场"], Action: "休息散步"},
			{NPCID: npcs[1].ID, StartTime: "15:00", EndTime: "18:00", LocationID: locMap["咖啡馆"], Action: "下午营业"},
			{NPCID: npcs[1].ID, StartTime: "18:00", EndTime: "19:00", LocationID: locMap["广场"], Action: "关店后散步"},
			// 艾琳
			{NPCID: npcs[2].ID, StartTime: "08:30", EndTime: "12:00", LocationID: locMap["图书馆"], Action: "开馆整理"},
			{NPCID: npcs[2].ID, StartTime: "12:00", EndTime: "14:00", LocationID: locMap["公园凉亭"], Action: "午休阅读"},
			{NPCID: npcs[2].ID, StartTime: "14:00", EndTime: "17:30", LocationID: locMap["图书馆"], Action: "协助借阅"},
			{NPCID: npcs[2].ID, StartTime: "17:30", EndTime: "18:00", LocationID: locMap["广场"], Action: "闭馆后散步"},
			// 菲奥娜
			{NPCID: npcs[3].ID, StartTime: "08:00", EndTime: "10:00", LocationID: locMap["花店"], Action: "浇花整理"},
			{NPCID: npcs[3].ID, StartTime: "10:00", EndTime: "12:00", LocationID: locMap["花店"], Action: "制作花束"},
			{NPCID: npcs[3].ID, StartTime: "12:00", EndTime: "14:00", LocationID: locMap["广场"], Action: "送花到广场"},
			{NPCID: npcs[3].ID, StartTime: "14:00", EndTime: "17:00", LocationID: locMap["花店"], Action: "照料温室"},
			{NPCID: npcs[3].ID, StartTime: "17:00", EndTime: "18:00", LocationID: locMap["咖啡馆"], Action: "与莉娜闲聊"},
			// 奥托
			{NPCID: npcs[4].ID, StartTime: "07:00", EndTime: "08:00", LocationID: locMap["钟楼"], Action: "检查钟楼"},
			{NPCID: npcs[4].ID, StartTime: "08:00", EndTime: "12:00", LocationID: locMap["铁匠铺"], Action: "开铺锻造"},
			{NPCID: npcs[4].ID, StartTime: "12:00", EndTime: "13:00", LocationID: locMap["酒馆"], Action: "午休"},
			{NPCID: npcs[4].ID, StartTime: "14:00", EndTime: "17:00", LocationID: locMap["铁匠铺"], Action: "下午锻造"},
			{NPCID: npcs[4].ID, StartTime: "17:00", EndTime: "18:00", LocationID: locMap["钟楼"], Action: "复查钟楼"},
			// 克莱尔
			{NPCID: npcs[5].ID, StartTime: "08:00", EndTime: "10:00", LocationID: locMap["诊所"], Action: "开诊"},
			{NPCID: npcs[5].ID, StartTime: "10:00", EndTime: "12:00", LocationID: locMap["广场"], Action: "巡诊"},
			{NPCID: npcs[5].ID, StartTime: "12:00", EndTime: "14:00", LocationID: locMap["诊所"], Action: "整理病历"},
			{NPCID: npcs[5].ID, StartTime: "14:00", EndTime: "17:00", LocationID: locMap["诊所"], Action: "接诊"},
			{NPCID: npcs[5].ID, StartTime: "17:00", EndTime: "18:00", LocationID: locMap["住宅区"], Action: "回家"},
			// 杰克
			{NPCID: npcs[6].ID, StartTime: "06:00", EndTime: "10:00", LocationID: locMap["农舍"], Action: "田间劳作"},
			{NPCID: npcs[6].ID, StartTime: "10:00", EndTime: "12:00", LocationID: locMap["面包店"], Action: "送蔬菜"},
			{NPCID: npcs[6].ID, StartTime: "12:00", EndTime: "14:00", LocationID: locMap["农舍"], Action: "田边午餐"},
			{NPCID: npcs[6].ID, StartTime: "14:00", EndTime: "17:00", LocationID: locMap["农舍"], Action: "继续农活"},
			{NPCID: npcs[6].ID, StartTime: "17:00", EndTime: "18:00", LocationID: locMap["酒馆"], Action: "偶尔去酒馆"},
			// 沃尔特
			{NPCID: npcs[7].ID, StartTime: "06:00", EndTime: "10:00", LocationID: locMap["钓鱼小屋"], Action: "湖边钓鱼"},
			{NPCID: npcs[7].ID, StartTime: "10:00", EndTime: "12:00", LocationID: locMap["钓鱼小屋"], Action: "整理渔获"},
			{NPCID: npcs[7].ID, StartTime: "12:00", EndTime: "14:00", LocationID: locMap["酒馆"], Action: "送鱼到酒馆"},
			{NPCID: npcs[7].ID, StartTime: "14:00", EndTime: "17:00", LocationID: locMap["钓鱼小屋"], Action: "修补渔网"},
			{NPCID: npcs[7].ID, StartTime: "17:00", EndTime: "18:00", LocationID: locMap["钓鱼小屋"], Action: "傍晚垂钓"},
			// 索菲亚
			{NPCID: npcs[8].ID, StartTime: "08:00", EndTime: "12:00", LocationID: locMap["学校"], Action: "上课"},
			{NPCID: npcs[8].ID, StartTime: "12:00", EndTime: "14:00", LocationID: locMap["学校"], Action: "午休备课"},
			{NPCID: npcs[8].ID, StartTime: "14:00", EndTime: "15:30", LocationID: locMap["学校"], Action: "下午课"},
			{NPCID: npcs[8].ID, StartTime: "15:30", EndTime: "17:00", LocationID: locMap["学校"], Action: "批改作业"},
			{NPCID: npcs[8].ID, StartTime: "17:00", EndTime: "18:00", LocationID: locMap["图书馆"], Action: "与艾琳交流"},
			// 皮埃尔
			{NPCID: npcs[9].ID, StartTime: "05:00", EndTime: "08:00", LocationID: locMap["面包店"], Action: "开始烘焙"},
			{NPCID: npcs[9].ID, StartTime: "08:00", EndTime: "12:00", LocationID: locMap["面包店"], Action: "开门营业"},
			{NPCID: npcs[9].ID, StartTime: "12:00", EndTime: "14:00", LocationID: locMap["面包店"], Action: "第二批出炉"},
			{NPCID: npcs[9].ID, StartTime: "14:00", EndTime: "17:00", LocationID: locMap["面包店"], Action: "研究新配方"},
			{NPCID: npcs[9].ID, StartTime: "17:00", EndTime: "18:00", LocationID: locMap["咖啡馆"], Action: "给莉娜送面包"},
			// 玛莎
			{NPCID: npcs[10].ID, StartTime: "10:00", EndTime: "12:00", LocationID: locMap["酒馆"], Action: "准备食材"},
			{NPCID: npcs[10].ID, StartTime: "12:00", EndTime: "14:00", LocationID: locMap["酒馆"], Action: "午餐营业"},
			{NPCID: npcs[10].ID, StartTime: "14:00", EndTime: "17:00", LocationID: locMap["酒馆"], Action: "休息准备"},
			{NPCID: npcs[10].ID, StartTime: "17:00", EndTime: "22:00", LocationID: locMap["酒馆"], Action: "晚市营业"},
			// 卢卡斯
			{NPCID: npcs[11].ID, StartTime: "09:00", EndTime: "11:00", LocationID: locMap["公园凉亭"], Action: "练琴"},
			{NPCID: npcs[11].ID, StartTime: "11:00", EndTime: "12:00", LocationID: locMap["广场"], Action: "散步获取灵感"},
			{NPCID: npcs[11].ID, StartTime: "12:00", EndTime: "14:00", LocationID: locMap["公园凉亭"], Action: "午休"},
			{NPCID: npcs[11].ID, StartTime: "14:00", EndTime: "17:00", LocationID: locMap["公园凉亭"], Action: "创作"},
			{NPCID: npcs[11].ID, StartTime: "18:00", EndTime: "21:00", LocationID: locMap["酒馆"], Action: "晚间演奏"},
			// 托马斯
			{NPCID: npcs[12].ID, StartTime: "07:30", EndTime: "09:00", LocationID: locMap["手工工坊"], Action: "开工"},
			{NPCID: npcs[12].ID, StartTime: "09:00", EndTime: "12:00", LocationID: locMap["广场"], Action: "外出修缮"},
			{NPCID: npcs[12].ID, StartTime: "12:00", EndTime: "14:00", LocationID: locMap["手工工坊"], Action: "午休"},
			{NPCID: npcs[12].ID, StartTime: "14:00", EndTime: "17:30", LocationID: locMap["手工工坊"], Action: "制作家具"},
			{NPCID: npcs[12].ID, StartTime: "17:30", EndTime: "18:00", LocationID: locMap["铁匠铺"], Action: "与奥托议事"},
			// 米娅
			{NPCID: npcs[13].ID, StartTime: "08:00", EndTime: "12:00", LocationID: locMap["学校"], Action: "上学"},
			{NPCID: npcs[13].ID, StartTime: "12:00", EndTime: "14:00", LocationID: locMap["住宅区"], Action: "回家午餐"},
			{NPCID: npcs[13].ID, StartTime: "14:00", EndTime: "15:00", LocationID: locMap["广场"], Action: "分发信件"},
			{NPCID: npcs[13].ID, StartTime: "15:00", EndTime: "17:00", LocationID: locMap["公园凉亭"], Action: "玩耍"},
			{NPCID: npcs[13].ID, StartTime: "17:00", EndTime: "18:00", LocationID: locMap["住宅区"], Action: "回家"},
			// 薇拉
			{NPCID: npcs[14].ID, StartTime: "07:00", EndTime: "10:00", LocationID: locMap["森林营地"], Action: "探索森林"},
			{NPCID: npcs[14].ID, StartTime: "10:00", EndTime: "12:00", LocationID: locMap["铁匠铺"], Action: "修理装备"},
			{NPCID: npcs[14].ID, StartTime: "12:00", EndTime: "14:00", LocationID: locMap["森林营地"], Action: "午休"},
			{NPCID: npcs[14].ID, StartTime: "14:00", EndTime: "17:00", LocationID: locMap["森林营地"], Action: "探索森林"},
			{NPCID: npcs[14].ID, StartTime: "17:00", EndTime: "20:00", LocationID: locMap["酒馆"], Action: "分享冒险故事"},
		}
		if err := tx.Create(&schedules).Error; err != nil {
			return err
		}

		relationships := []model.NPCRelationship{
			{NPCID: npcs[1].ID, TargetNPCID: npcs[4].ID, Affinity: 65, Trust: 60, Tag: "友好照顾"},
			{NPCID: npcs[4].ID, TargetNPCID: npcs[1].ID, Affinity: 55, Trust: 50, Tag: "感激"},
			{NPCID: npcs[1].ID, TargetNPCID: npcs[3].ID, Affinity: 75, Trust: 70, Tag: "商业街好友"},
			{NPCID: npcs[3].ID, TargetNPCID: npcs[1].ID, Affinity: 75, Trust: 70, Tag: "商业街好友"},
			{NPCID: npcs[4].ID, TargetNPCID: npcs[10].ID, Affinity: 80, Trust: 75, Tag: "多年老友"},
			{NPCID: npcs[10].ID, TargetNPCID: npcs[4].ID, Affinity: 80, Trust: 75, Tag: "多年老友"},
			{NPCID: npcs[13].ID, TargetNPCID: npcs[0].ID, Affinity: 70, Trust: 75, Tag: "like_family"},
			{NPCID: npcs[0].ID, TargetNPCID: npcs[13].ID, Affinity: 80, Trust: 85, Tag: "like_family"},
			{NPCID: npcs[13].ID, TargetNPCID: npcs[14].ID, Affinity: 65, Trust: 55, Tag: "崇拜"},
			{NPCID: npcs[14].ID, TargetNPCID: npcs[13].ID, Affinity: 60, Trust: 60, Tag: "保护欲"},
			{NPCID: npcs[6].ID, TargetNPCID: npcs[9].ID, Affinity: 60, Trust: 55, Tag: "商业合作"},
			{NPCID: npcs[4].ID, TargetNPCID: npcs[12].ID, Affinity: 60, Trust: 65, Tag: "职业合作"},
			{NPCID: npcs[12].ID, TargetNPCID: npcs[4].ID, Affinity: 60, Trust: 65, Tag: "职业合作"},
		}
		if err := tx.Create(&relationships).Error; err != nil {
			return err
		}

		storyEvents := []model.StoryEvent{
			{TownID: town.ID, Title: "钟楼维护日", Description: "奥托进行每月一次的钟楼大检修，钟声将暂停一个上午", Status: "ready",
				TriggerCondition: `{"time_range":"morning","probability":0.12,"cooldown_hours":12}`,
				Effects:          `{"npc_effects":[{"npc_name":"奥托","mood":"focused","goal":"全面检修钟楼"},{"npc_name":"埃德蒙","mood":"content","goal":"通知居民钟楼维护"},{"npc_name":"托马斯","mood":"steady","goal":"协助奥托检修"}]}`},
			{TownID: town.ID, Title: "咖啡馆季节限定", Description: "莉娜根据季节推出新饮品，邀请全镇品鉴。皮埃尔为此特制了搭配甜点", Status: "ready",
				TriggerCondition: `{"time_range":"morning","probability":0.10,"cooldown_hours":16}`,
				Effects:          `{"npc_effects":[{"npc_name":"莉娜","mood":"excited","goal":"推出季节限定饮品"},{"npc_name":"皮埃尔","mood":"excited","goal":"特制搭配甜点"},{"npc_name":"菲奥娜","mood":"happy","goal":"装饰咖啡馆"}]}`},
			{TownID: town.ID, Title: "米娅的冒险", Description: "米娅帮埃德蒙送信时，在薇拉的帮助下经历了一场小冒险", Status: "ready",
				TriggerCondition: `{"time_range":"afternoon","probability":0.10,"cooldown_hours":12}`,
				Effects:          `{"npc_effects":[{"npc_name":"米娅","mood":"excited","goal":"送信冒险"},{"npc_name":"薇拉","mood":"confident","goal":"保护米娅"},{"npc_name":"埃德蒙","mood":"content","goal":"等待信件"}]}`},
			{TownID: town.ID, Title: "森林的发现", Description: "薇拉在森林探索中发现了一些不寻常的痕迹，引起了奥托和玛莎的兴趣", Status: "ready",
				TriggerCondition: `{"time_range":"evening","probability":0.08,"cooldown_hours":24}`,
				Effects:          `{"npc_effects":[{"npc_name":"薇拉","mood":"curious","goal":"调查森林发现"},{"npc_name":"奥托","mood":"curious","goal":"协助调查"},{"npc_name":"玛莎","mood":"curious","goal":"在酒馆组织讨论"}]}`},
			{TownID: town.ID, Title: "广场集市日", Description: "每月一次的广场集市。菲奥娜布置花展，杰克带来新鲜蔬菜，卢卡斯现场演奏", Status: "ready",
				TriggerCondition: `{"time_range":"morning","trigger_day":"market_day","probability":1.0}`,
				Effects:          `{"npc_effects":[{"npc_name":"菲奥娜","mood":"excited","goal":"布置花展"},{"npc_name":"皮埃尔","mood":"jolly","goal":"准备面包摊位"},{"npc_name":"杰克","mood":"happy","goal":"摆出新鲜蔬菜"},{"npc_name":"卢卡斯","mood":"inspired","goal":"集市演奏"},{"npc_name":"沃尔特","mood":"content","goal":"卖鲜鱼"}]}`},
			{TownID: town.ID, Title: "晚间演奏会", Description: "卢卡斯在酒馆举办晚间演奏。玛莎准备了特别的菜单，全镇居民聚集聆听", Status: "ready",
				TriggerCondition: `{"time_range":"evening","probability":0.12,"cooldown_hours":16}`,
				Effects:          `{"npc_effects":[{"npc_name":"卢卡斯","mood":"inspired","goal":"准备演奏曲目"},{"npc_name":"玛莎","mood":"excited","goal":"准备特别菜单"},{"npc_name":"艾琳","mood":"calm","goal":"为演奏会写介绍词"}]}`},
			{TownID: town.ID, Title: "湖边垂钓日", Description: "沃尔特邀请杰克和奥托去湖边钓鱼。这是男人的休憩时光", Status: "ready",
				TriggerCondition: `{"time_range":"morning","probability":0.08,"cooldown_hours":20}`,
				Effects:          `{"npc_effects":[{"npc_name":"沃尔特","mood":"peaceful","goal":"教年轻人钓鱼技巧"},{"npc_name":"杰克","mood":"content","goal":"学习钓鱼"},{"npc_name":"奥托","mood":"content","goal":"休息放松"}]}`},
			{TownID: town.ID, Title: "花展筹备", Description: "菲奥娜正在为即将到来的花展做准备。她需要杰克的帮助搬运新到的花盆", Status: "ready",
				TriggerCondition: `{"time_range":"morning","probability":0.10,"cooldown_hours":18}`,
				Effects:          `{"npc_effects":[{"npc_name":"菲奥娜","mood":"excited","goal":"筹备花展"},{"npc_name":"杰克","mood":"happy","goal":"帮忙搬运花盆"},{"npc_name":"索菲亚","mood":"warm","goal":"带学生参观花展"}]}`},
		}
		if err := tx.Create(&storyEvents).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		appLog.Error(err, "种子数据导入失败")
		return
	}

	appLog.Info("种子数据导入完成",
		"town", "晨曦镇",
		"locations", 17,
		"npcs", 15,
		"schedules", 75,
		"relationships", 13,
		"story_events", 5,
	)
}
