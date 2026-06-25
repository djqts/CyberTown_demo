import { useEffect, useState } from "react";
import { useTownStore } from "../../store/town-store";
import { focusGodotLocation } from "../../lib/godot-bridge";
import { moodLabel, getNpcPortrait } from "../../lib/utils";
import { getNpcDetail } from "../../lib/api";
import { Badge } from "../ui/Badge";
import { Button } from "../ui/Button";
import { MapPin, Zap, Target, MessageCircle, ArrowLeft } from "lucide-react";

type NpcDetail = Record<string, unknown>;

export function NpcDetailPanel() {
  const npcs = useTownStore((s) => s.npcs);
  const locations = useTownStore((s) => s.locations);
  const selectedNpcId = useTownStore((s) => s.selectedNpcId);
  const [detail, setDetail] = useState<NpcDetail | null>(null);

  const clearSelection = useTownStore((s) => s.clearSelection);
  const npc = npcs.find((n) => n.id === selectedNpcId);
  const loc = locations.find((l) => l.id === npc?.location_id);

  useEffect(() => {
    if (selectedNpcId == null) return;
    let cancelled = false;
    getNpcDetail(selectedNpcId)
      .then((d) => { if (!cancelled) setDetail(d as NpcDetail); })
      .catch(() => { /* use basic npc data */ });
    return () => { cancelled = true; };
  }, [selectedNpcId]);

  if (!npc) return null;

  const name = (detail?.name ?? npc.name) as string;
  const role = (detail?.role ?? npc.role) as string;
  const gender = (detail?.gender ?? npc.gender) as string;
  const ageGroup = (detail?.age_group ?? npc.age_group) as string;
  const mood = (detail?.mood ?? npc.mood) as string;
  const energy = (detail?.energy ?? npc.energy) as number;
  const goal = (detail?.current_goal ?? npc.current_goal) as string;
  const personality = detail?.personality as string | undefined;
  const catchphrase = detail?.catchphrase as string | undefined;
  const appearance = detail?.appearance as string | undefined;
  const nameFirst = (name)[0] ?? "?";

  return (
    <div className="space-y-3 p-4">
      <button onClick={clearSelection} className="flex items-center gap-1 text-[12px] text-[#B8C2CC]/50 hover:text-[#7CCB8A] transition-colors">
        <ArrowLeft size={13} /> 返回居民列表
      </button>
      <div className="flex items-start gap-3">
        {getNpcPortrait(npc.role) ? (
          <img src={getNpcPortrait(npc.role)} alt={name as string}
            className="h-16 w-16 rounded-full object-cover shadow-lg border-2 border-white/10 shrink-0" />
        ) : (
          <div
            className="flex h-16 w-16 shrink-0 items-center justify-center rounded-full text-lg font-bold text-white"
            style={{ background: `linear-gradient(135deg, ${npc.portrait_color ?? "#888"}, ${npc.portrait_color ?? "#888"}88)` }}
          >{nameFirst}</div>
        )}
        <div>
          <h3 className="text-base font-semibold text-[#F3F5F7]">{name}</h3>
          <p className="text-[13px] text-[#B8C2CC]">
            {role} · {gender} · {ageGroup}
          </p>
        </div>
      </div>

      <div className="flex flex-wrap gap-1.5">
        <Badge variant="accent" className="gap-1">
          <Zap size={11} /> {moodLabel(mood)}
        </Badge>
        <Badge variant="default" className="gap-1">
          精力 {energy}%
        </Badge>
        {loc && (
          <Button variant="ghost" size="sm" className="gap-1 h-6 text-[12px]" onClick={() => focusGodotLocation(loc.id)}>
            <MapPin size={11} /> {loc.name}
          </Button>
        )}
      </div>

      {goal && (
        <p className="flex items-start gap-1.5 text-[13px] text-[#B8C2CC]">
          <Target size={13} className="mt-0.5 shrink-0" />
          {goal}
        </p>
      )}

      {personality && (
        <p className="text-[13px] text-[#F3F5F7]">{personality}</p>
      )}
      {catchphrase && (
        <p className="text-[13px] italic text-[#B8C2CC] flex items-start gap-1.5">
          <MessageCircle size={13} className="mt-0.5 shrink-0" />
          &ldquo;{catchphrase}&rdquo;
        </p>
      )}
      {appearance && (
        <p className="text-[12px] text-[#B8C2CC]/70">{appearance}</p>
      )}

      {detail?.relationships && (detail.relationships as Array<{npc_name: string; npc_id: number; affinity: number; tag: string}>).length > 0 && (
        <div className="mt-3 pt-3 border-t border-[#1e2d3d]">
          <p className="text-[11px] font-medium text-[#B8C2CC]/50 mb-2">人际关系</p>
          <div className="space-y-1">
            {(detail.relationships as Array<{npc_name: string; npc_id: number; affinity: number; tag: string}>).map((rel) => (
              <div key={rel.npc_id} className="flex items-center gap-2 text-[12px]">
                <div className="flex-1 text-[#F3F5F7]">{rel.npc_name}</div>
                <span className="text-[10px] px-1.5 py-0.5 rounded bg-white/5 text-[#B8C2CC]">{rel.tag}</span>
                <div className="w-16 h-1.5 rounded-full bg-white/5">
                  <div className="h-full rounded-full bg-[#7CCB8A]/60" style={{width: `${rel.affinity}%`}} />
                </div>
                <span className="text-[10px] text-[#B8C2CC]/50 w-6 text-right">{rel.affinity}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
