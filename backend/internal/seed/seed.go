package seed

import (
	"backend/internal/logger"
	"backend/internal/model"

	"gorm.io/gorm"
)

// SeedDemo 初始化最小演示数据。若已存在则跳过。
func SeedDemo(db *gorm.DB, appLog *logger.AppLogger) {
	var count int64
	db.Model(&model.Town{}).Where("name = ?", "晨曦镇").Count(&count)
	if count > 0 {
		appLog.Info("种子数据已存在，跳过")
		return
	}

	appLog.Info("开始导入种子数据")

	err := db.Transaction(func(tx *gorm.DB) error {
		town := model.Town{
			Name:          "晨曦镇",
			CurrentDay:    0,
			CurrentMinute: 0,
		}
		if err := tx.Create(&town).Error; err != nil {
			return err
		}

		locations := []model.Location{
			{Name: "广场", Longitude: "0", Latitude: "0"},
			{Name: "咖啡馆", Longitude: "0.5", Latitude: "0.3"},
			{Name: "钟楼", Longitude: "-0.3", Latitude: "0.8"},
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
			{
				TownID:      town.ID,
				LocationID:  locMap["咖啡馆"],
				Name:        "莉娜",
				Role:        "咖啡师",
				Personality: "热情开朗，喜欢和每位客人闲聊",
				Status:      "idle",
				CurrentGoal: "为客人冲泡今日特调",
			},
			{
				TownID:      town.ID,
				LocationID:  locMap["钟楼"],
				Name:        "奥托",
				Role:        "钟表匠",
				Personality: "沉默寡言，对机械有极致追求",
				Status:      "idle",
				CurrentGoal: "校准钟楼齿轮",
			},
			{
				TownID:      town.ID,
				LocationID:  locMap["广场"],
				Name:        "米娅",
				Role:        "邮差",
				Personality: "好奇心旺盛，喜欢打探镇上新鲜事",
				Status:      "idle",
				CurrentGoal: "派送今日信件",
			},
		}
		if err := tx.Create(&npcs).Error; err != nil {
			return err
		}

		schedules := []model.NPCSchedule{
			// 莉娜: 上午咖啡馆工作, 下午广场休息
			{NPCID: npcs[0].ID, StartTime: "08:00", EndTime: "12:00", LocationID: locMap["咖啡馆"], Action: "冲泡咖啡"},
			{NPCID: npcs[0].ID, StartTime: "14:00", EndTime: "16:00", LocationID: locMap["广场"], Action: "午后散步"},
			// 奥托: 上午钟楼修表, 晚间咖啡馆喝茶
			{NPCID: npcs[1].ID, StartTime: "06:00", EndTime: "12:00", LocationID: locMap["钟楼"], Action: "修理钟表"},
			{NPCID: npcs[1].ID, StartTime: "14:00", EndTime: "18:00", LocationID: locMap["咖啡馆"], Action: "品尝下午茶"},
			// 米娅: 邮差送信路线 广场→咖啡馆→钟楼
			{NPCID: npcs[2].ID, StartTime: "08:00", EndTime: "09:00", LocationID: locMap["广场"], Action: "整理信件"},
			{NPCID: npcs[2].ID, StartTime: "10:00", EndTime: "11:00", LocationID: locMap["咖啡馆"], Action: "投递信件"},
			{NPCID: npcs[2].ID, StartTime: "13:00", EndTime: "14:00", LocationID: locMap["钟楼"], Action: "投递信件"},
		}
		if err := tx.Create(&schedules).Error; err != nil {
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
		"locations", 3,
		"npcs", 3,
		"schedules", 7,
	)
}
