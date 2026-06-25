import { useRef, useEffect } from "react";
import { useTownStore } from "../../store/town-store";
import { focusGodotNpc } from "../../lib/godot-bridge";
import { moodLabel, getNpcPortrait } from "../../lib/utils";
import gsap from "gsap";

const NEGATIVE = ["anxious","worried","sad","angry","tired"];
const POSITIVE = ["cheerful","happy","excited","inspired","jolly","warm","playful","confident"];

function moodBorderClass(mood: string) {
  if (POSITIVE.includes(mood)) return "border-l-[#78B860]";
  if (NEGATIVE.includes(mood)) return "border-l-[#e94560]";
  return "border-l-[#3B4B5C]";
}

export function NpcList() {
  const npcs = useTownStore((s) => s.npcs);
  const locations = useTownStore((s) => s.locations);
  const selectedNpcId = useTownStore((s) => s.selectedNpcId);
  const selectNpc = useTownStore((s) => s.setSelectedNpc);
  const locMap = new Map(locations.map((l) => [l.id, l]));
  const listRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!listRef.current) return;
    const items = listRef.current.children;
    gsap.fromTo(items,
      { opacity: 0, x: -16 },
      { opacity: 1, x: 0, duration: 0.35, stagger: 0.04, ease: "power2.out" }
    );
  }, [npcs.length]);

  return (
    <div ref={listRef} className="space-y-0.5 p-3">
      {npcs.map((npc) => {
        const isSelected = selectedNpcId === npc.id;
        return (
          <button
            key={npc.id}
            onClick={() => {
              focusGodotNpc(npc.id);
              selectNpc(npc.id, npc.location_id);
            }}
            className={`w-full flex items-center gap-2.5 rounded-lg px-3 py-2 text-left transition-all duration-200 border-l-[3px] ${
              isSelected
                ? "bg-white/10 border-l-[#7CCB8A]"
                : `bg-white/[0.02] hover:bg-white/5 ${moodBorderClass(npc.mood)}`
            }`}
          >
            <div className="relative shrink-0">
              {getNpcPortrait(npc.role) ? (
                <img src={getNpcPortrait(npc.role)} alt={npc.name}
                  className="w-10 h-10 rounded-full object-cover shadow-md border-2 border-white/10" />
              ) : (
                <div
                  className="w-10 h-10 rounded-full flex items-center justify-center text-xs font-bold text-white shadow-md"
                  style={{ background: `linear-gradient(135deg, ${npc.portrait_color ?? "#888"}, ${npc.portrait_color ?? "#888"}99)` }}
                >{npc.name[0]}</div>
              )}
              {npc.energy < 30 && (
                <span className="absolute -bottom-0.5 -right-0.5 w-3 h-3 bg-amber-400 rounded-full border border-[#18202A]" title="疲惫" />
              )}
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-1.5">
                <span className="text-[13px] font-medium text-[#F3F5F7] truncate">{npc.name}</span>
                <span className="text-[11px] text-[#B8C2CC]">{npc.role}</span>
              </div>
              {npc.current_goal && (
                <p className="text-[11px] text-[#B8C2CC]/70 truncate">{npc.current_goal}</p>
              )}
            </div>
            <div className="flex flex-col items-end gap-0.5 shrink-0">
              <span className="text-[12px]">{moodLabel(npc.mood)}</span>
              {npc.location_name && <span className="text-[10px] text-[#B8C2CC]/50">{npc.location_name}</span>}
            </div>
          </button>
        );
      })}
    </div>
  );
}
