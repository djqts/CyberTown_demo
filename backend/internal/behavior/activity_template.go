package behavior

import "math/rand"

// RoleActivityPool maps NPC roles to their in-role activities (when at their home building).
var RoleActivityPool = map[string][]string{
	"镇长": {
		"翻阅桌上的文件，批注着今天的公务",
		"站在市政厅门口，向路过的镇民微笑点头",
		"拿出怀表看了看时间，又到了巡视小镇的时候了",
	},
	"咖啡馆主": {
		"擦拭手中的咖啡杯，让它们闪闪发光",
		"煮上一壶新咖啡，整个店铺都弥漫着香气",
		"在吧台后调配新品，偶尔尝一口味道",
	},
	"图书管理员": {
		"轻轻拂去书架上的灰尘，把归还的书放回原位",
		"安静地翻着书页，偶尔在笔记本上记录心得",
		"调整阅读区的灯光，让每一束光都落在恰到好处的位置",
	},
	"花店店主": {
		"给花架上的盆栽逐一浇水，每一株都充满生机",
		"修剪花枝，精心搭配成一束漂亮的花艺",
		"在花店门口摆放今天推荐的鲜花，吸引路人驻足",
	},
	"铁匠": {
		"拉着风箱让炉火烧得更旺，准备锻造新的铁器",
		"举起铁锤有节奏地敲打烧红的铁块，火花四溅",
		"把刚打造好的铁件浸入冷水，呲的一声冒出白烟",
	},
	"医生": {
		"整理医疗器械，把每件工具摆放得井井有条",
		"在病历本上写下今天的巡诊记录",
		"在诊所门口给草药浇水，检查长势",
	},
	"农夫": {
		"弯腰检查作物的生长情况，拔掉几株杂草",
		"擦了擦额头的汗水，继续田间劳作",
		"把刚摘的蔬菜放进竹篮，准备送到镇上",
	},
	"渔夫": {
		"安静地坐在湖边，盯着水面上的浮标一动不动",
		"修补破损的渔网，手法熟练",
		"整理鱼篓里的收获，挑出最大的几条",
	},
	"教师": {
		"在黑板上工整地写下今天的课程内容",
		"批改学生的作业，偶尔露出欣慰的微笑",
		"在校园的花坛里浇花，等待学生们到来",
	},
	"面包师": {
		"用力揉着面团，手法熟练又有节奏",
		"从烤炉中取出金黄的面包，用手指轻弹听声音",
		"把刚出炉的可颂整齐摆好，趁热试吃了一个",
	},
	"酒馆老板": {
		"擦拭酒杯，把它们整齐排列在吧台后",
		"检查酒架上的库存，在笔记本上记录需要补充的品类",
		"整理桌椅，为晚上的客人铺好桌布",
	},
	"音乐家": {
		"拨动琴弦，一段悠扬的旋律飘散开来",
		"在笔记本上飞快地写下几个音符，灵感来了",
		"闭眼感受风声，寻找旋律的灵感",
	},
	"木匠": {
		"仔细测量木料的尺寸，用铅笔做好标记",
		"用刨子推过木面，薄薄的木花卷曲着落下",
		"检查刚做好的家具，用手指抚摸每一个接缝",
	},
	"小女孩": {
		"蹦蹦跳跳地穿过广场，书包在背后一颠一颠的",
		"抱着玩具兔，小声和它说着今天的趣事",
		"在花坛边蹲下来，好奇地看一只蝴蝶停在花瓣上",
	},
	"冒险者": {
		"磨着剑刃，动作利落又专注",
		"查看那张磨损的地图，在某个位置画了个圈",
		"眺望森林深处，目光锐利而沉着",
	},
}

// LocationActivityPool maps location names to generic activities anyone can do there.
var LocationActivityPool = map[string][]string{
	"广场": {
		"在广场上散步，享受着阳光和微风", "坐在长椅上看着来往的居民",
		"和路过的熟人打招呼，聊了几句天气", "驻足看了一会儿鸽子啄食",
	},
	"咖啡馆": {
		"点了一杯咖啡，坐在窗边慢慢品尝", "翻看吧台上的报纸，看到一条有趣的消息",
		"和邻桌的客人闲聊了几句镇上的新闻",
	},
	"钟楼": {
		"抬头看着高耸的钟楼，指针在阳光下闪着光", "在钟楼下驻足，听了听齿轮转动的声音",
		"绕着钟楼走了一圈，欣赏着石墙上的雕刻",
	},
	"市政厅": {
		"在市政厅公告栏前停下来，看看有什么新的通知", "坐在等候区的椅子上整理了一下物品",
	},
	"图书馆": {
		"在书架间穿行，随手抽出一本书翻了几页", "坐在阅读区的桌前安静地看了一会儿书",
		"找到一本感兴趣的书，靠在窗边读了起来",
	},
	"花店": {
		"凑近一束花闻了闻，露出满意的微笑", "看着花店门口五彩缤纷的花卉，心情变好了",
	},
	"铁匠铺": {
		"站在门口看铁匠干活，铁锤的声音有节奏地响着", "挑了挑展示架上的铁器，挑了件趁手的",
	},
	"诊所": {
		"在候诊区坐下来整理了一下衣物", "看了看墙上贴的健康提示，若有所思",
	},
	"农舍": {
		"站在田边看着一片绿油油的庄稼，深吸了一口新鲜空气", "帮农夫捡了几颗掉在地上的蔬菜",
	},
	"钓鱼小屋": {
		"在湖边坐下来，把脚伸进水里，感受着湖水的清凉", "看着湖面上的涟漪发呆，偶尔有条鱼跃出水面",
	},
	"学校": {
		"在教室窗外听了听孩子们朗朗的读书声", "在操场边看着孩子们嬉笑打闹，脸上挂着微笑",
	},
	"面包店": {
		"被面包的香气吸引，在门口张望了一下", "买了两个刚出炉的面包，趁热咬了一口",
	},
	"酒馆": {
		"在吧台边坐下，点了一杯喝的", "和旁边的客人聊起了今天的见闻",
		"看着墙上挂着的旧照片，想起了一些往事",
	},
	"公园凉亭": {
		"在凉亭里坐下，感受着微风拂过", "看着公园里的花草树木，心情格外平静",
		"沿着公园的小径散步，踩在落叶上发出沙沙的声响",
	},
	"手工工坊": {
		"站在工坊门口看着匠人专注地工作", "拿起一块木料端详了一下纹理",
	},
	"住宅区": {
		"在家门口的小花园里打理了一下花草", "坐在门廊前晒了会儿太阳，和邻居挥了挥手",
	},
	"森林营地": {
		"站在森林边缘，深吸了一口带着松脂香味的空气", "仔细查看树干上的标记，确认了方向",
		"蹲下来观察地上的动物足迹，看得出神",
	},
}

// GetActivity returns a contextually appropriate activity for an NPC at a given location.
func GetActivity(role string, locationName string) string {
	// 70% chance: role-appropriate if at the "right" place, else location-appropriate
	isHome := isHomeLocation(role, locationName)

	if isHome && rand.Intn(100) < 70 {
		actions := RoleActivityPool[role]
		if len(actions) > 0 {
			return actions[rand.Intn(len(actions))]
		}
	}

	// Location-based fallback
	actions := LocationActivityPool[locationName]
	if len(actions) > 0 {
		return actions[rand.Intn(len(actions))]
	}

	// Generic fallback
	return "环顾四周，安静地待了一会儿"
}

func isHomeLocation(role, location string) bool {
	home := map[string]string{
		"镇长": "市政厅", "咖啡馆主": "咖啡馆", "图书管理员": "图书馆",
		"花店店主": "花店", "铁匠": "铁匠铺", "医生": "诊所",
		"农夫": "农舍", "渔夫": "钓鱼小屋", "教师": "学校",
		"面包师": "面包店", "酒馆老板": "酒馆", "音乐家": "公园凉亭",
		"木匠": "手工工坊", "小女孩": "学校", "冒险者": "森林营地",
	}
	return home[role] == location
}

// GetRoleActions returns actions for a given role (used by BehaviorService).
func GetRoleActions(role string) []string {
	if actions, ok := RoleActivityPool[role]; ok {
		return actions
	}
	return []string{"环顾四周", "安静地站着"}
}

// MoodActionPool maps moods to expressive actions.
var MoodActionPool = map[string][]string{
	"cheerful":  {"愉快地哼着小调", "脸上挂着温暖的笑容"},
	"happy":     {"脚步轻快，心情愉悦", "忍不住笑出声来"},
	"excited":   {"眼里闪着光，充满期待", "激动地搓着手"},
	"content":   {"神情从容，一切尽在掌握", "安详地享受当下的平静"},
	"calm":      {"目光平静如水", "沉默而专注地做着手上的事"},
	"focused":   {"全神贯注，周围一切都似乎消失了", "眉头微皱，专注异常"},
	"anxious":   {"不安地踱着步子", "眉头紧锁，不时叹气"},
	"worried":   {"忧心忡忡地望向远方", "反复检查手中的东西"},
	"sad":       {"低头不语，神情落寞", "默默地擦拭着眼角"},
	"tired":     {"打着哈欠，眼皮沉重", "无力地靠在墙上"},
	"curious":   {"好奇地四处张望", "蹲下来仔细观察"},
	"caring":    {"温柔地看着对方", "轻轻地拍了拍对方的肩膀"},
	"confident": {"昂首挺胸，步伐坚定", "嘴角挂着自信的微笑"},
	"inspired":  {"眼中闪过灵感的光芒", "停下脚步，沉浸在思绪中"},
	"playful":   {"开心地转了个圈", "兴致勃勃地哼着歌"},
}

// GetMoodAction returns a mood-specific action.
func GetMoodAction(mood string) string {
	if actions, ok := MoodActionPool[mood]; ok && len(actions) > 0 {
		return actions[rand.Intn(len(actions))]
	}
	return ""
}
