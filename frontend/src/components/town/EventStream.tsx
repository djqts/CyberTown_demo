import { useRef, useEffect, useState } from "react";
import { useTownStore } from "../../store/town-store";
import { focusGodotNpc, focusGodotLocation } from "../../lib/godot-bridge";
import { getNpcPortrait } from "../../lib/utils";
import gsap from "gsap";
import { Zap, MessageCircle, MapPin, Heart, AlertTriangle, Activity, ChevronDown, ChevronUp } from "lucide-react";

type EventStyle = {
  icon: React.ReactNode;
  accent: string;
  bg: string;
  border: string;
  label: string;
};

const STYLES: Record<string, EventStyle> = {
  "town.news.generated": {
    icon: <Zap size={13} />, accent: "text-amber-400", bg: "bg-amber-400/5", border: "border-amber-400/20", label: "新闻",
  },
  "story.event.triggered": {
    icon: <Zap size={13} />, accent: "text-amber-400", bg: "bg-amber-400/5", border: "border-amber-400/20", label: "故事",
  },
  "npc.interaction.generated": {
    icon: <MessageCircle size={13} />, accent: "text-violet-400", bg: "bg-violet-400/5", border: "border-violet-400/20", label: "互动",
  },
  "npc.mood.changed": {
    icon: <Heart size={13} />, accent: "text-rose-400", bg: "bg-rose-400/5", border: "border-rose-400/20", label: "情绪",
  },
  "npc.moved": {
    icon: <MapPin size={13} />, accent: "text-sky-400", bg: "bg-sky-400/5", border: "border-sky-400/20", label: "移动",
  },
  "npc.replied": {
    icon: <MessageCircle size={13} />, accent: "text-emerald-400", bg: "bg-emerald-400/5", border: "border-emerald-400/20", label: "对话",
  },
  "npc.activity.generated": {
    icon: <Activity size={13} />, accent: "text-[#B8C2CC]", bg: "bg-white/[0.02]", border: "border-white/5", label: "活动",
  },
  "npc.idle.action": {
    icon: <Activity size={13} />, accent: "text-[#B8C2CC]/60", bg: "bg-white/[0.01]", border: "border-white/5", label: "休闲",
  },
  "npc.goal.changed": {
    icon: <Activity size={13} />, accent: "text-cyan-400", bg: "bg-cyan-400/5", border: "border-cyan-400/20", label: "目标",
  },
  "godot.offline": {
    icon: <AlertTriangle size={13} />, accent: "text-red-400", bg: "bg-red-400/5", border: "border-red-400/20", label: "离线",
  },
};

const FALLBACK_STYLE: EventStyle = {
  icon: <Activity size={13} />, accent: "text-[#B8C2CC]/50", bg: "bg-white/[0.01]", border: "border-white/5", label: "系统",
};

export function EventStream() {
  const events = useTownStore((s) => s.events);
  const locations = useTownStore((s) => s.locations);
  const listRef = useRef<HTMLDivElement>(null);
  const prevLen = useRef(0);
  const [expanded, setExpanded] = useState<number>(0);

  useEffect(() => {
    if (!listRef.current || events.length === 0) return;
    const newCount = events.length - prevLen.current;
    prevLen.current = events.length;
    if (newCount <= 0) return;

    const items = listRef.current.children;
    const fresh = Array.from(items).slice(0, Math.min(newCount, items.length));
    gsap.fromTo(fresh,
      { opacity: 0, y: -12, scale: 0.96 },
      { opacity: 1, y: 0, scale: 1, duration: 0.35, stagger: 0.06, ease: "power3.out" }
    );
  }, [events.length]);

  const selectNpc = useTownStore((s) => s.setSelectedNpc);

  function handleClick(evt: { type: string; data: Record<string, unknown> }, index: number) {
    if (evt.type === "npc.interaction.generated") {
      setExpanded(expanded === index ? 0 : index);
      return;
    }
    const npcId = evt.data.npc_id as number | undefined;
    const locId = evt.data.location_id as number | undefined;
    if (evt.type === "npc.replied" && npcId) {
      selectNpc(npcId, (evt.data as { location_id?: number }).location_id ?? 0);
      window.dispatchEvent(new CustomEvent("town.openChat"));
      return;
    }
    if (npcId) focusGodotNpc(npcId);
    else if (locId) focusGodotLocation(locId);
  }

  function titleFor(evt: { type: string; data: Record<string, unknown> }): string {
    const d = evt.data;
    switch (evt.type) {
      case "npc.activity.generated":
      case "npc.idle.action":
        return `${d.npc_name ?? ""} ${d.action ?? ""}`;
      case "npc.mood.changed":
        return `${d.npc_name ?? ""} 情绪 ${d.old_mood ?? ""} → ${d.new_mood ?? ""}`;
      case "npc.goal.changed":
        return `${d.npc_name ?? ""} 目标变为 「${d.new_goal ?? ""}」`;
      case "npc.moved": {
        const locName = locations.find((l) => l.name === d.to_location)?.name ?? d.to_location;
        return `${d.npc_name ?? ""} 前往 ${locName}`;
      }
      case "story.event.triggered":
        return `⚡ ${d.story_title ?? ""}`;
      case "npc.interaction.generated": {
        const a = d.npc_a as { name?: string; role?: string } | undefined;
        const b = d.npc_b as { name?: string; role?: string } | undefined;
        return `${a?.name ?? ""} 与 ${b?.name ?? ""} 交谈`;
      }
      case "npc.replied":
        return `${d.npc_name ?? ""}：${(d.content as string)?.slice(0, 80) ?? ""}`;
      case "town.news.generated":
        return `📰 ${(d.message as string) ?? (d.story_title as string) ?? ""}`;
      case "npc.relationship.changed": {
        const delta = d.delta as number ?? 0;
        return `${d.from_npc_name ?? ""} → ${d.to_npc_name ?? ""} 关系${delta > 0 ? "+" : ""}${delta}`;
      }
      case "godot.offline":
        return "后端不可用，切换到离线预览模式";
      default:
        return evt.type;
    }
  }

  return (
    <div className="h-full overflow-auto">
      <div ref={listRef} className="p-3 space-y-2 min-h-[200px]">
        {events.length === 0 && (
          <p className="text-center text-[13px] text-[#B8C2CC]/30 py-16">等待小镇苏醒...</p>
        )}
        {events.slice(0, 100).map((evt, i) => {
          const style = STYLES[evt.type] ?? FALLBACK_STYLE;
          const time = new Date(evt.createdAt).toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit", second: "2-digit" });
          const isExpanded = expanded === i;
          const isInteraction = evt.type === "npc.interaction.generated";

          return (
            <button
              key={`${evt.createdAt}-${i}`}
              onClick={() => handleClick(evt, i)}
              className={`w-full text-left rounded-xl border ${style.border} ${style.bg} p-3 transition-all duration-200 hover:scale-[1.01] hover:shadow-lg hover:shadow-black/20`}
            >
              {/* Header */}
              <div className="flex items-start gap-2.5">
                <div className={`mt-0.5 shrink-0 ${style.accent}`}>{style.icon}</div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <span className={`text-[10px] font-medium px-1.5 py-0.5 rounded-md ${style.accent} ${style.bg} border ${style.border}`}>
                      {style.label}
                    </span>
                    <span className="text-[10px] text-[#B8C2CC]/30 font-mono">{time}</span>
                    {isInteraction && (
                      <span className="ml-auto text-[#B8C2CC]/40">
                        {isExpanded ? <ChevronUp size={12} /> : <ChevronDown size={12} />}
                      </span>
                    )}
                  </div>
                  <p className={`text-[13px] leading-relaxed ${style.accent}`}>{titleFor(evt)}</p>
                </div>
              </div>

              {/* Expanded dialogue detail */}
              {isExpanded && isInteraction && (
                <div className="mt-3 pt-3 border-t border-violet-400/10 space-y-2">
                  {((evt.data.dialogue as Array<{ speaker: string; speech: string; action: string; emotion: string }>) ?? []).map((line, di) => {
                    const isA = (evt.data.npc_a as { name?: string })?.name === line.speaker;
                    const npc = isA ? (evt.data.npc_a as { name?: string; role?: string }) : (evt.data.npc_b as { name?: string; role?: string });
                    const portraitUrl = npc?.role ? getNpcPortrait(npc.role) : "";
                    return (
                      <div key={di} className={`flex gap-2 ${isA ? "" : "flex-row-reverse"}`}>
                        {portraitUrl ? (
                          <img src={portraitUrl} alt={line.speaker}
                            className="w-7 h-7 rounded-full object-cover border border-white/10 shrink-0 mt-0.5" />
                        ) : (
                          <div className="w-7 h-7 rounded-full bg-white/5 flex items-center justify-center text-[9px] text-[#B8C2CC] shrink-0 mt-0.5">
                            {line.speaker[0]}
                          </div>
                        )}
                        <div className={`max-w-[85%] ${isA ? "" : "text-right"}`}>
                          <div className={`rounded-xl px-2.5 py-1.5 text-[12px] leading-relaxed ${
                            isA ? "bg-violet-400/10 text-[#F3F5F7] rounded-tl-sm" : "bg-white/5 text-[#F3F5F7] rounded-tr-sm"
                          }`}>
                            <p>{line.speech}</p>
                          </div>
                          {line.action && (
                            <p className="text-[10px] text-[#B8C2CC]/40 mt-0.5 italic">* {line.action}</p>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
}
