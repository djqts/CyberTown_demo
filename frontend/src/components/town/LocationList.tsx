import { useTownStore } from "../../store/town-store";
import { focusGodotLocation } from "../../lib/godot-bridge";
import { Button } from "../ui/Button";
import { MapPin } from "lucide-react";

const DISTRICT_ORDER = ["central", "left_commercial", "right", "farmland", "lake", "park", "residential", "forest"];
const DISTRICT_LABELS: Record<string, string> = {
  central: "中央广场区",
  left_commercial: "左区商业街",
  right: "右区",
  farmland: "远左农田区",
  lake: "湖区",
  park: "公园区",
  residential: "住宅区",
  forest: "森林边缘",
};

export function LocationList() {
  const locations = useTownStore((s) => s.locations);
  const npcs = useTownStore((s) => s.npcs);
  const selectedLocationId = useTownStore((s) => s.selectedLocationId);

  const npcCountByLoc: Record<number, number> = {};
  for (const n of npcs) {
    npcCountByLoc[n.location_id] = (npcCountByLoc[n.location_id] ?? 0) + 1;
  }

  const byDistrict: Record<string, typeof locations> = {};
  for (const loc of locations) {
    const d = loc.district ?? "";
    if (!byDistrict[d]) byDistrict[d] = [];
    byDistrict[d].push(loc);
  }

  return (
    <div className="space-y-3 p-4">
      {DISTRICT_ORDER.map((d) => {
        const locs = byDistrict[d];
        if (!locs?.length) return null;
        return (
          <div key={d}>
            <div className="mb-1.5 text-[11px] font-medium uppercase tracking-wider text-[#B8C2CC]/60">
              {DISTRICT_LABELS[d] ?? d}
            </div>
            <div className="flex flex-wrap gap-1.5">
              {locs.map((loc) => {
                const count = npcCountByLoc[loc.id] ?? 0;
                const isSelected = selectedLocationId === loc.id;
                return (
                  <Button
                    key={loc.id}
                    variant={isSelected ? "default" : "ghost"}
                    size="sm"
                    onClick={() => focusGodotLocation(loc.id)}
                    className="gap-1"
                  >
                    <MapPin size={12} />
                    {loc.name}
                    {count > 0 && (
                      <span className="ml-0.5 rounded-full bg-[#7CCB8A]/20 px-1.5 text-[11px] text-[#7CCB8A]">
                        {count}
                      </span>
                    )}
                  </Button>
                );
              })}
            </div>
          </div>
        );
      })}
    </div>
  );
}
