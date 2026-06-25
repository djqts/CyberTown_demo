package seed

import (
	"context"
	"time"

	qdrant "github.com/qdrant/go-client/qdrant"

	"backend/internal/logger"
	"backend/internal/memory"
)

var worldFacts = []string{
	// 小镇概况
	"晨曦镇坐落在森林与湖泊之间的山谷中，有17座建筑和15位居民。广场钟楼是小镇的地标，钟声每天准时响起。",
	"小镇的日常生活围绕广场展开：早晨咖啡馆飘香，午间集市热闹，傍晚酒馆灯火通明。",
	"每月第一个周末是广场集市日，全镇居民摆摊交易、演奏音乐、交换故事。",

	// 广场区
	"埃德蒙是晨曦镇的老镇长，慈祥有威严。他在市政厅办公，每天上午会在广场散步，和每一位居民打招呼。",
	"奥托是小镇唯一的铁匠，在铁匠铺打铁为生。他还自愿维护广场钟楼，每天早上第一件事就是检查齿轮。",
	"钟楼的钟声是小镇的心跳。奥托对它有近乎偏执的责任感——如果钟声不准，他宁可熬夜也要修好。",

	// 左区商业街
	"莉娜的咖啡馆是小镇的社交中心。她的拿铁和可颂是镇上有名的早餐搭配。她善于倾听，知道每个人的喜好。",
	"皮埃尔是面包师，每天凌晨五点开始揉面。他的可颂外酥里嫩，和莉娜合伙供应早餐。两人是多年的商业伙伴。",
	"菲奥娜在花店里种满了春天的花。她和杰克合作种植花卉，常为咖啡馆和图书馆送装饰花束。",
	"托马斯是木匠，手艺精湛。他负责维护全镇的木质建筑，常和奥托合作——铁件配木件。",

	// 公共服务
	"克莱尔是小镇唯一的医生，在诊所工作。她温和专业，关心每一个居民的健康，尤其担心埃德蒙镇长的身体。",
	"索菲亚是学校老师，温柔耐心。她独自负责镇上所有孩子的教育，和图书管理员艾琳合作推荐课外读物。",
	"艾琳管理着小镇图书馆。她安静温柔，知识渊博，能记住每一本书的位置和借阅记录。",
	"米娅是一个八岁的小女孩，活泼好奇。她每天下午帮大人跑腿送信，最崇拜的人是冒险者薇拉。",

	// 湖区与农田
	"杰克是农夫，在远左农田区种植蔬菜。他和菲奥娜合作种花，向皮埃尔供应面粉，和沃尔特交换食材。",
	"沃尔特是老渔夫，每天清晨在湖边垂钓。他的渔获供应给玛莎的酒馆。他话不多，但每句话都有分量。",

	// 右区与森林
	"玛莎是酒馆老板，豪爽直率。她的酒馆是小镇夜间社交的中心，奥托和沃尔特常来下棋，卢卡斯在此演奏。",
	"卢卡斯是小镇的音乐家，浪漫而略带忧郁。他每天都在公园凉亭练琴，傍晚在玛莎的酒馆演奏。",
	"薇拉是住在森林边缘的冒险者，独立神秘。她见过外面的世界，但选择回到小镇。她暗中守护着这片宁静。",

	// 小镇关系
	"莉娜和菲奥娜是在商业街上结识的好友，每天早上互道早安。菲奥娜常送花给咖啡馆装饰。",
	"奥托和玛莎是认识多年的老朋友。玛莎了解奥托不爱说话的性格，总是默默给他倒好麦芽酒。",
	"米娅和薇拉之间有一种特别的友谊。薇拉给米娅讲森林里的冒险故事，米娅则分享学校里学的新字。",
	"埃德蒙把米娅当成自己的孙女儿。他常在广场上给米娅糖果，关心她的学习和成长。",
}

// SeedWorldKnowledge 将世界知识向量化后写入 Qdrant（幂等，collection 不存在时跳过）。
func SeedWorldKnowledge(client *qdrant.Client, appLog *logger.AppLogger) {
	ctx := context.Background()

	exists, err := client.CollectionExists(ctx, "world_knowledge")
	if err != nil || !exists {
		appLog.Warn("world_knowledge collection 不存在，跳过世界知识初始化", "err", err)
		return
	}

	// 检查是否已有数据
	result, err := client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: "world_knowledge",
		Query:          qdrant.NewQuery(make([]float32, memory.EmbedDim)...),
		Limit:          qdrant.PtrOf(uint64(1)),
	})
	if err == nil && len(result) > 0 {
		appLog.Info("世界知识已存在，跳过")
		return
	}

	appLog.Info("开始导入世界知识", "count", len(worldFacts))
	for i, fact := range worldFacts {
		vec := memory.Embed(fact)
		pointID := qdrant.NewIDNum(uint64(i + 1))

		_, err := client.Upsert(ctx, &qdrant.UpsertPoints{
			CollectionName: "world_knowledge",
			Points: []*qdrant.PointStruct{
				{
					Id: pointID,
					Payload: map[string]*qdrant.Value{
						"content": {Kind: &qdrant.Value_StringValue{StringValue: fact}},
					},
					Vectors: qdrant.NewVectors(vec...),
				},
			},
		})
		if err != nil {
			appLog.Error(err, "导入世界知识失败", "index", i)
			return
		}
		time.Sleep(50 * time.Millisecond) // 避免 Qdrant 过载
	}
	appLog.Info("世界知识导入完成", "count", len(worldFacts))
}
