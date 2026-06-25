package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"backend/internal/event"
)

type demoHandler struct {
	pub              *event.Publisher
	seq              atomic.Int64
	directBroadcast  func(eventType string, data map[string]any) // bypasses RabbitMQ
	locNameToID      map[string]uint                             // location name → ID lookup
}

func newDemoHandler(pub *event.Publisher) *demoHandler { return &demoHandler{pub: pub} }

func (h *demoHandler) SetDirectBroadcast(fn func(eventType string, data map[string]any)) {
	h.directBroadcast = fn
}

func (h *demoHandler) Trigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	go h.run()
}

func (h *demoHandler) emit(evtType string, payload any) {
	b, _ := json.Marshal(payload)
	seq := h.seq.Add(1)
	h.pub.Publish(context.TODO(), &event.Event{
		EventID:   fmt.Sprintf("demo-%d-%d", time.Now().UnixNano(), seq),
		EventType: evtType, TownID: 1, ActorType: event.ActorTypeSystem, ActorID: "demo",
		Payload: b, CreatedAt: time.Now(),
	})
	// Brief pause between emits prevents RabbitMQ channel buffer overrun
	// when multiple events are emitted in rapid succession within the same batch.
	time.Sleep(10 * time.Millisecond)
}

func (h *demoHandler) news(msg string) {
	h.emit(event.EventTypeTownNewsGenerated, map[string]any{"story_title": "晨曦日报", "message": msg})
}
func (h *demoHandler) mood(id uint, name, from, to string) {
	h.emit(event.EventTypeNPCMoodChanged, map[string]any{"npc_id": id, "npc_name": name, "old_mood": from, "new_mood": to})
}
func (h *demoHandler) goal(id uint, name, from, to, reason string) {
	h.emit(event.EventTypeNPCGoalChanged, map[string]any{"npc_id": id, "npc_name": name, "old_goal": from, "new_goal": to, "reason": reason})
}
func (h *demoHandler) move(id uint, name, from, to string) {
	fromID := h.locNameToID[from]
	toID := h.locNameToID[to]
	payload := map[string]any{
		"npc_id": id, "npc_name": name,
		"from_location": from, "to_location": to,
		"from_location_id": fromID, "to_location_id": toID,
	}
	if h.directBroadcast != nil {
		h.directBroadcast(event.EventTypeNPCMoved, payload)
	}
	h.emit(event.EventTypeNPCMoved, payload)
}
func (h *demoHandler) act(id uint, name, action, mood string) {
	h.emit(event.EventTypeNPCActivityGenerated, map[string]any{"npc_id": id, "npc_name": name, "action": action, "mood": mood})
}
func (h *demoHandler) chat(aID uint, aN, aR string, bID uint, bN, bR string, lines []map[string]string, mc map[uint]string, rel int) {
	h.emit(event.EventTypeNPCInteractionGenerated, map[string]any{
		"npc_a": map[string]any{"id": aID, "name": aN, "role": aR}, "npc_b": map[string]any{"id": bID, "name": bN, "role": bR},
		"dialogue": lines, "mood_changes": mc,
		"rel_deltas": []map[string]any{{"from_npc_id": aID, "to_npc_id": bID, "delta": rel, "reason": "conversation"}},
	})
}

func (h *demoHandler) run() {
	// ═══════════════════════════════════════════
	// ACT 1: MORNING — 小镇日常 (0-60s)
	// ═══════════════════════════════════════════
	h.news("📰 晨曦初现。春日的阳光洒在屋顶上，小镇开始了新的一天。")
	time.Sleep(3 * time.Second)

	h.goal(35, "莉娜", "休息", "准备开店", "新的一天")
	h.goal(38, "奥托", "休息", "检查钟楼", "新的一天")
	h.goal(43, "皮埃尔", "休息", "烤面包", "新的一天")
	h.goal(37, "菲奥娜", "休息", "浇花", "新的一天")
	time.Sleep(3 * time.Second)

	h.move(35, "莉娜", "住宅区", "咖啡馆")
	h.move(43, "皮埃尔", "住宅区", "面包店")
	h.move(38, "奥托", "住宅区", "钟楼")
	time.Sleep(5 * time.Second)

	h.act(35, "莉娜", "推开咖啡馆的木门，开始煮今天的第一壶咖啡", "cheerful")
	h.act(43, "皮埃尔", "揉着面团，烤箱已经预热好了，第一批可颂马上出炉", "jolly")
	h.act(38, "奥托", "准时来到钟楼下，仰头检查齿轮的运转——这是每天雷打不动的仪式", "focused")
	time.Sleep(5 * time.Second)

	h.mood(35, "莉娜", "resting", "cheerful")
	h.mood(38, "奥托", "resting", "focused")
	h.mood(43, "皮埃尔", "resting", "jolly")
	time.Sleep(3 * time.Second)

	// ═══════════════════════════════════════════
	// ACT 2: INCIDENT — 钟楼需要大检修 (60-140s)
	// ═══════════════════════════════════════════
	h.news("📰 奥托宣布今天要进行每月一次的钟楼大检修。钟声将暂停一个上午，埃德蒙镇长通知了全镇居民。")
	h.emit(event.EventTypeStoryEventTriggered, json.RawMessage(`{"story_title":"钟楼维护日","npc_id":38,"npc_name":"奥托","old_mood":"focused","new_mood":"focused","old_goal":"检查钟楼","new_goal":"全面检修钟楼"}`))
	time.Sleep(3 * time.Second)

	h.goal(38, "奥托", "检查钟楼", "全面检修钟楼", "每月维护日")
	h.goal(34, "埃德蒙", "办公", "通知居民钟楼维护", "钟楼维护日")
	h.goal(46, "托马斯", "修缮桌椅", "协助奥托检修", "钟楼维护日")
	time.Sleep(5 * time.Second)

	h.move(34, "埃德蒙", "市政厅", "钟楼")
	h.move(46, "托马斯", "手工工坊", "钟楼")
	time.Sleep(5 * time.Second)

	h.act(38, "奥托", "打开钟楼的齿轮箱，仔细检查每一个齿轮的磨损情况。这是个大工程", "focused")
	h.act(46, "托马斯", "拿出工具箱里的精密工具，和奥托一起拆下需要更换的齿轮", "steady")
	h.act(34, "埃德蒙", "站在钟楼下向围观的居民解释维护计划，让大家不必担心", "content")
	time.Sleep(8 * time.Second)

	// 莉娜送咖啡
	h.move(35, "莉娜", "咖啡馆", "钟楼")
	time.Sleep(5 * time.Second)

	h.chat(35, "莉娜", "咖啡馆主", 38, "奥托", "铁匠",
		[]map[string]string{
			{"speaker": "莉娜", "speech": "奥托！听说你今天要大修钟楼，我给你和托马斯带了咖啡。", "action": "端着两杯热咖啡走来", "emotion": "caring"},
			{"speaker": "奥托", "speech": "你总是这么贴心。主齿轮确实磨损了，得换新的。", "action": "擦了擦手上的油污，接过咖啡", "emotion": "focused"},
			{"speaker": "莉娜", "speech": "有什么需要尽管说。皮埃尔还特制了新的可颂，说是给维修队加油的。", "action": "微笑着指了指面包店的方向", "emotion": "cheerful"},
		}, map[uint]string{35: "cheerful", 38: "focused"}, 2,
	)
	time.Sleep(8 * time.Second)

	// 皮埃尔也来帮忙
	h.move(43, "皮埃尔", "面包店", "钟楼")
	h.act(43, "皮埃尔", "端着一盘刚出炉的可颂来到钟楼下，分给围观的居民", "jolly")
	time.Sleep(5 * time.Second)

	// 菲奥娜送花装饰
	h.move(37, "菲奥娜", "花店", "钟楼")
	h.act(37, "菲奥娜", "抱着一束雏菊来到钟楼，插在临时的工作台上，给维修队带来一点春天的色彩", "happy")
	time.Sleep(8 * time.Second)

	// ═══════════════════════════════════════════
	// ACT 3: RESOLUTION — 钟楼修好，小镇庆祝 (140-250s)
	// ═══════════════════════════════════════════
	h.news("📰 经过一上午的努力，奥托和托马斯成功更换了钟楼的齿轮。钟声重新响起，比之前更加清亮！")
	time.Sleep(3 * time.Second)

	h.goal(38, "奥托", "全面检修钟楼", "休息放松", "维护完成")
	h.mood(38, "奥托", "focused", "content")
	h.mood(46, "托马斯", "steady", "content")
	time.Sleep(3 * time.Second)

	h.act(38, "奥托", "站在钟楼下，听着重新响起的钟声，嘴角露出一丝满意的微笑", "content")
	h.act(46, "托马斯", "收拾好工具，和奥托握了握手。合作愉快", "content")
	time.Sleep(5 * time.Second)

	// 居民们返回日常
	h.move(35, "莉娜", "钟楼", "咖啡馆")
	h.move(43, "皮埃尔", "钟楼", "面包店")
	h.move(37, "菲奥娜", "钟楼", "花店")
	h.move(34, "埃德蒙", "钟楼", "市政厅")
	time.Sleep(8 * time.Second)

	// 晚间：酒馆庆祝
	h.news("📰 傍晚时分，居民们聚集在玛莎的酒馆。卢卡斯即兴演奏了一曲，庆祝钟楼维护顺利完成。小镇恢复了平日的节奏。")
	time.Sleep(3 * time.Second)

	h.move(38, "奥托", "钟楼", "酒馆")
	h.move(46, "托马斯", "钟楼", "酒馆")
	h.move(34, "埃德蒙", "市政厅", "酒馆")
	h.move(45, "卢卡斯", "公园凉亭", "酒馆")
	time.Sleep(8 * time.Second)

	h.act(44, "玛莎", "点亮酒馆的灯笼，给每个人的杯子倒满麦芽酒。今晚由她请客", "friendly")
	h.act(45, "卢卡斯", "抱着琴找到角落的位置，即兴弹起了一首欢快的庆祝曲", "inspired")
	time.Sleep(8 * time.Second)

	h.chat(38, "奥托", "铁匠", 44, "玛莎", "酒馆老板",
		[]map[string]string{
			{"speaker": "玛莎", "speech": "奥托！今天你是英雄。全镇都听到钟声了，比之前还好听！", "action": "倒了一大杯麦芽酒推到他面前", "emotion": "friendly"},
			{"speaker": "奥托", "speech": "多亏了托马斯帮忙。齿轮换了新的，至少能用三年。", "action": "难得露出放松的笑容", "emotion": "content"},
		}, map[uint]string{38: "content", 44: "friendly"}, 1,
	)
	time.Sleep(8 * time.Second)

	// ═══════════════════════════════════════════
	// EPILOGUE — 小镇入眠 (250-270s)
	// ═══════════════════════════════════════════
	h.move(38, "奥托", "酒馆", "住宅区")
	h.move(34, "埃德蒙", "酒馆", "住宅区")
	h.move(46, "托马斯", "酒馆", "住宅区")
	h.move(45, "卢卡斯", "酒馆", "公园凉亭")
	time.Sleep(5 * time.Second)

	h.act(35, "莉娜", "关上咖啡馆的门，在窗边坐下翻看今天的账本。今天又是充实的一天", "content")
	h.act(38, "奥托", "回到家门口，回头望了一眼钟楼的方向。钟声在夜色中回荡，一切安好", "content")
	time.Sleep(5 * time.Second)

	// Restore all mood
	h.mood(35, "莉娜", "cheerful", "cheerful")
	h.mood(38, "奥托", "content", "content")
	h.mood(37, "菲奥娜", "happy", "happy")
	h.mood(43, "皮埃尔", "jolly", "jolly")
	h.mood(34, "埃德蒙", "content", "content")
	h.mood(46, "托马斯", "content", "steady")
	h.mood(45, "卢卡斯", "inspired", "dreamy")
	h.mood(44, "玛莎", "friendly", "friendly")
	time.Sleep(3 * time.Second)

	h.news("📰 夜色渐深，小镇恢复了宁静。钟声依旧准时，咖啡馆飘着香气，酒馆里的故事还在继续。明天又是新的一天。这就是晨曦镇——一个普通但温暖的日子。")

	// Notify frontend that demo is done
	b, _ := json.Marshal(map[string]string{"status": "done"})
	h.pub.Publish(context.TODO(), &event.Event{
		EventID: "demo-done", EventType: "demo.completed", TownID: 1,
		ActorType: event.ActorTypeSystem, ActorID: "demo",
		Payload: b, CreatedAt: time.Now(),
	})
}
