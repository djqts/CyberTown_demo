package interaction

import "math/rand"

type DialogueTemplate struct {
	Initiator        string `json:"initiator"`
	Responder        string `json:"responder"`
	InitiatorEmotion string `json:"initiator_emotion"`
	ResponderEmotion string `json:"responder_emotion"`
}

var TemplateDialoguePool = map[string][]DialogueTemplate{
	"友好照顾": {
		{Initiator: "来，这是我刚为你准备的。", Responder: "谢谢你，总是这么贴心。", InitiatorEmotion: "caring", ResponderEmotion: "grateful"},
		{Initiator: "你今天看起来很累，没事吧？", Responder: "还好，就是有点忙。谢谢关心。", InitiatorEmotion: "caring", ResponderEmotion: "touched"},
	},
	"商业街好友": {
		{Initiator: "今天的客人多吗？", Responder: "还不错，你呢？", InitiatorEmotion: "friendly", ResponderEmotion: "friendly"},
		{Initiator: "你听说了吗？镇上新来的那个旅人。", Responder: "听说了，好像很有意思。", InitiatorEmotion: "curious", ResponderEmotion: "curious"},
	},
	"多年老友": {
		{Initiator: "好久没一起喝酒了。", Responder: "是啊，这阵子太忙了。", InitiatorEmotion: "warm", ResponderEmotion: "content"},
		{Initiator: "还记得那年春天的事吗？", Responder: "哈哈，怎么可能忘。", InitiatorEmotion: "nostalgic", ResponderEmotion: "amused"},
	},
	"职业合作": {
		{Initiator: "你那边最近进展怎么样？", Responder: "一切顺利，多亏你的帮助。", InitiatorEmotion: "professional", ResponderEmotion: "grateful"},
	},
	"商业合作": {
		{Initiator: "这批材料质量很好。", Responder: "那就好，我一直很重视品质。", InitiatorEmotion: "satisfied", ResponderEmotion: "proud"},
	},
	"like_family": {
		{Initiator: "今天在学校学了什么？", Responder: "索菲亚老师教我们认了好多新字！", InitiatorEmotion: "warm", ResponderEmotion: "excited"},
	},
	"崇拜": {
		{Initiator: "薇拉姐姐，再讲一个冒险故事吧！", Responder: "好，不过这次不能太长。", InitiatorEmotion: "excited", ResponderEmotion: "amused"},
	},
	"default": {
		{Initiator: "今天天气真不错。", Responder: "是啊，适合出去走走。", InitiatorEmotion: "content", ResponderEmotion: "content"},
		{Initiator: "最近镇上有什么新鲜事吗？", Responder: "好像没什么特别的，一切照旧。", InitiatorEmotion: "curious", ResponderEmotion: "neutral"},
	},
}

func GetDialogueTemplate(tag string) *DialogueTemplate {
	if templates, ok := TemplateDialoguePool[tag]; ok && len(templates) > 0 {
		return &templates[rand.Intn(len(templates))]
	}
	d := TemplateDialoguePool["default"]
	return &d[rand.Intn(len(d))]
}
