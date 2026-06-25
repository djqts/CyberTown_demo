import { useTownStore } from "../../store/town-store";
import { DISTRICTS } from "../../lib/utils";
import type { Npc } from "../../lib/town-types";

const MAP_W = 1000;
const MAP_H = 800;

const DISTRICT_ORDER = ["central", "left_commercial", "right", "farmland", "lake", "park", "residential", "forest"];

const DISTRICT_COLORS: Record<string, string> = {
  central: "#fffde7", left_commercial: "#f3e5f5", right: "#e3f2fd",
  farmland: "#e8f5e9", lake: "#e0f7fa", park: "#c8e6c9",
  residential: "#fff3e0", forest: "#dcedc8",
};

function getNpcOffset(index: number, total: number) {
  const r = 22;
  const angle = (index / Math.max(total, 1)) * Math.PI * 2 - Math.PI / 2;
  return { dx: Math.cos(angle) * r, dy: Math.sin(angle) * r };
}

export function TownMap() {
  const locations = useTownStore((s) => s.locations);
  const npcs = useTownStore((s) => s.npcs);
  const selectedNpcId = useTownStore((s) => s.selectedNpcId);
  const selectedLocationId = useTownStore((s) => s.selectedLocationId);
  const selectNpc = useTownStore((s) => s.setSelectedNpc);
  const selectLoc = useTownStore((s) => s.setSelectedLocation);

  const npcByLoc: Record<number, Npc[]> = {};
  for (const n of npcs) {
    if (!npcByLoc[n.location_id]) npcByLoc[n.location_id] = [];
    npcByLoc[n.location_id].push(n);
  }

  if (locations.length === 0) {
    return (
      <div className="flex items-center justify-center h-full text-[#B8C2CC]/60 text-[13px]">
        地图数据加载中...
      </div>
    );
  }

  return (
    <svg viewBox={`0 0 ${MAP_W} ${MAP_H}`} className="w-full h-full" style={{ background: "#e8f5e9" }}>
      {/* Districts */}
      {DISTRICT_ORDER.map((d) => {
        const locs = locations.filter((l) => (l.district ?? "") === d);
        if (!locs.length) return null;
        const xs = locs.map((l) => l.x ?? 0);
        const ys = locs.map((l) => l.y ?? 0);
        const minX = Math.min(...xs) - 60;
        const minY = Math.min(...ys) - 50;
        const maxX = Math.max(...xs) + 60;
        const maxY = Math.max(...ys) + 50;
        return (
          <g key={d}>
            <rect
              x={minX} y={minY} width={maxX - minX} height={maxY - minY}
              rx="16" fill={DISTRICT_COLORS[d] ?? "#f0f0f0"} stroke="#ccc" strokeWidth="1"
              opacity="0.6"
            />
            <text x={(minX + maxX) / 2} y={minY - 8} textAnchor="middle" fontSize="12" fill="#888" fontWeight="bold">
              {DISTRICTS[d] ?? d}
            </text>
          </g>
        );
      })}

      {/* Roads */}
      {locations.map((a) =>
        locations.filter((b) => b.id > a.id).map((b) => {
          const dist = Math.hypot((a.x ?? 0) - (b.x ?? 0), (a.y ?? 0) - (b.y ?? 0));
          if (dist > 250) return null;
          return (
            <line key={`${a.id}-${b.id}`} x1={a.x} y1={a.y} x2={b.x} y2={b.y}
              stroke="#ccc" strokeWidth="2" strokeDasharray="6,4" />
          );
        })
      )}

      {/* Location icons */}
      {locations.map((loc) => {
        const isSelected = selectedLocationId === loc.id;
        const size = loc.size === "large" ? 36 : loc.size === "small" ? 22 : 28;
        return (
          <g key={`loc-${loc.id}`}
            className="cursor-pointer transition-transform hover:scale-110"
            onClick={() => selectLoc(loc.id)}
          >
            {isSelected && (
              <circle cx={loc.x} cy={loc.y} r={size / 2 + 6} fill="none" stroke="#7CCB8A" strokeWidth="2" />
            )}
            <rect
              x={(loc.x ?? 0) - size / 2 - 4} y={(loc.y ?? 0) - size / 2 - 4}
              width={size + 8} height={size + 8} rx="6"
              fill="white" stroke={isSelected ? "#7CCB8A" : "#ccc"} strokeWidth="1.5"
            />
            <text x={loc.x} y={(loc.y ?? 0) + size / 6} textAnchor="middle" fontSize={size * 0.6}>
              {loc.icon ?? "📍"}
            </text>
            <text x={loc.x} y={(loc.y ?? 0) + size / 2 + 12} textAnchor="middle" fontSize="10" fill="#555">
              {loc.name}
            </text>
          </g>
        );
      })}

      {/* NPC dots */}
      {Object.entries(npcByLoc).map(([locId, npcList]) => {
        const loc = locations.find((l) => l.id === parseInt(locId));
        if (!loc) return null;
        return npcList.map((npc, i) => {
          const { dx, dy } = getNpcOffset(i, npcList.length);
          const isSelected = selectedNpcId === npc.id;
          return (
            <g key={`npc-${npc.id}`}
              className="cursor-pointer transition-transform hover:scale-125"
              onClick={() => selectNpc(npc.id, npc.location_id)}
            >
              <circle
                cx={(loc.x ?? 0) + dx} cy={(loc.y ?? 0) + dy} r={isSelected ? 12 : 9}
                fill={npc.portrait_color ?? "#888"} stroke="#fff" strokeWidth="2"
              />
              <text
                x={(loc.x ?? 0) + dx} y={(loc.y ?? 0) + dy + 4}
                textAnchor="middle" fontSize="9" fill="#fff" fontWeight="bold"
              >
                {npc.name[0]}
              </text>
            </g>
          );
        });
      })}
    </svg>
  );
}
