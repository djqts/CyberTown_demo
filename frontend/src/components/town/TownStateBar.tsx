import { useTownStore } from "../../store/town-store";
import { Badge } from "../ui/Badge";
import { MapPin, Users, Clock, Cloud, Sun } from "lucide-react";

export function TownStateBar() {
  const townState = useTownStore((s) => s.townState);
  const npcs = useTownStore((s) => s.npcs);
  const locations = useTownStore((s) => s.locations);
  const online = useTownStore((s) => s.online);

  if (!townState) return null;

  return (
    <div className="space-y-3 p-4">
      <h2 className="text-lg font-semibold text-[#F3F5F7]">{townState.name}</h2>
      <div className="flex flex-wrap gap-2">
        <Badge variant="accent" className="gap-1">
          <Clock size={12} /> {townState.time}
        </Badge>
        <Badge variant="default" className="gap-1">
          <Sun size={12} /> {townState.season}
        </Badge>
        <Badge variant="default" className="gap-1">
          <Cloud size={12} /> {townState.weather}
        </Badge>
        <Badge variant={online ? "accent" : "warning"}>
          {online ? "在线" : "离线"}
        </Badge>
      </div>
      <div className="flex gap-4 text-[13px] text-[#B8C2CC]">
        <span className="flex items-center gap-1"><Users size={13} /> NPC: {npcs.length}</span>
        <span className="flex items-center gap-1"><MapPin size={13} /> 地点: {locations.length}</span>
      </div>
    </div>
  );
}
