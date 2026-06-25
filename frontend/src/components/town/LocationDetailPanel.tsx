import { useTownStore } from "../../store/town-store";
import { focusGodotLocation, focusGodotNpc } from "../../lib/godot-bridge";
import { moodLabel } from "../../lib/utils";
import { Button } from "../ui/Button";
import { Badge } from "../ui/Badge";
import { MapPin, Users } from "lucide-react";

export function LocationDetailPanel() {
  const locations = useTownStore((s) => s.locations);
  const npcs = useTownStore((s) => s.npcs);
  const selectedLocationId = useTownStore((s) => s.selectedLocationId);

  const loc = locations.find((l) => l.id === selectedLocationId);
  const npcsHere = npcs.filter((n) => n.location_id === loc?.id);

  if (!loc) return null;

  return (
    <div className="space-y-3 p-4">
      <div className="flex items-center gap-2">
        <h3 className="text-base font-semibold text-[#F3F5F7]">{loc.name}</h3>
        <Badge variant="default">{loc.district ?? ""}</Badge>
      </div>

      <Button size="sm" variant="outline" className="gap-1" onClick={() => focusGodotLocation(loc.id)}>
        <MapPin size={13} /> 聚焦地图
      </Button>

      {npcsHere.length > 0 && (
        <div>
          <p className="mb-1.5 text-[13px] text-[#B8C2CC] flex items-center gap-1">
            <Users size={13} /> 在场 NPC ({npcsHere.length})
          </p>
          <div className="space-y-1">
            {npcsHere.map((n) => (
              <Button
                key={n.id}
                variant="ghost"
                size="sm"
                className="w-full justify-start gap-2 h-auto py-1"
                onClick={() => focusGodotNpc(n.id)}
              >
                <span
                  className="inline-block h-4 w-4 shrink-0 rounded-full border border-white/20"
                  style={{ backgroundColor: n.portrait_color ?? "#888" }}
                />
                <span className="text-[13px] text-[#F3F5F7]">{n.name}</span>
                <span className="text-[12px] text-[#B8C2CC]">{n.role}</span>
                <span className="flex-1" />
                <span className="text-[11px] text-[#B8C2CC]/70">{moodLabel(n.mood)}</span>
              </Button>
            ))}
          </div>
        </div>
      )}

      {npcsHere.length === 0 && (
        <p className="text-[13px] text-[#B8C2CC]/60">暂无 NPC</p>
      )}
    </div>
  );
}
