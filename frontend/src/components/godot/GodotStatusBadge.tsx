import { Badge } from "../ui/Badge";
import { useTownStore } from "../../store/town-store";
import { Wifi, WifiOff } from "lucide-react";

export function GodotStatusBadge() {
  const online = useTownStore((s) => s.online);
  const wsConnected = useTownStore((s) => s.wsConnected);

  if (!online) {
    return (
      <Badge variant="warning" className="gap-1">
        <WifiOff size={12} />
        离线预览
      </Badge>
    );
  }

  return (
    <Badge variant={wsConnected ? "accent" : "warning"} className="gap-1">
      <Wifi size={12} />
      {wsConnected ? "在线" : "连接中"}
    </Badge>
  );
}
