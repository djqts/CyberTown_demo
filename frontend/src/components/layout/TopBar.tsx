import { useTownStore } from "../../store/town-store";
import { GodotStatusBadge } from "../godot/GodotStatusBadge";
import { Clock, Cloud, Users, Play, Loader2 } from "lucide-react";
import { useState, useRef, useEffect } from "react";

export function TopBar() {
  const townState = useTownStore((s) => s.townState);
  const npcs = useTownStore((s) => s.npcs);
  const online = useTownStore((s) => s.online);
  const wsConnected = useTownStore((s) => s.wsConnected);
  const [demoRunning, setDemoRunning] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout>>();

  useEffect(() => {
    function onDone() { setDemoRunning(false); if (timerRef.current) clearTimeout(timerRef.current); }
    window.addEventListener("demo.completed", onDone);
    return () => window.removeEventListener("demo.completed", onDone);
  }, []);

  async function triggerDemo() {
    if (demoRunning) return;
    setDemoRunning(true);
    try { await fetch("http://localhost:8080/api/demo/trigger", { method: "POST" }); } catch { /* ignore */ }
    timerRef.current = setTimeout(() => setDemoRunning(false), 300_000); // 5min safety
  }

  return (
    <div className="flex h-12 items-center gap-3 border-b border-[#1e2d3d] bg-[#111b26] px-4 shrink-0">
      <span className="text-[15px] font-semibold text-[#F3F5F7] tracking-tight">
        🌸 {townState?.name ?? "晨曦镇"}
      </span>
      <div className="h-5 w-px bg-[#1e2d3d]" />
      {townState && (
        <>
          <span className="text-[12px] text-[#B8C2CC] flex items-center gap-1.5">
            <Clock size={13} />
            Day {townState.day} · {townState.time}
          </span>
          <span className="text-[12px] text-[#B8C2CC] flex items-center gap-1.5">
            <Cloud size={13} />
            {townState.season} · {townState.weather}
          </span>
        </>
      )}
      <div className="flex-1" />
      <span className="text-[12px] text-[#B8C2CC]/70 flex items-center gap-1.5">
        <Users size={13} />
        {npcs.length} NPC{online ? "" : " (离线)"}
      </span>
      <button
        onClick={triggerDemo}
        disabled={demoRunning}
        className={`flex items-center gap-1 px-2.5 py-1 rounded-md text-[11px] font-medium transition-all border ${
          demoRunning
            ? "bg-amber-400/10 text-amber-400 border-amber-400/20 cursor-wait"
            : "bg-[#7CCB8A]/15 text-[#7CCB8A] hover:bg-[#7CCB8A]/25 border-[#7CCB8A]/20"
        }`}
        title={demoRunning ? "演示运行中..." : "触发演示脚本"}
      >
        {demoRunning ? <Loader2 size={11} className="animate-spin" /> : <Play size={11} />}
        {demoRunning ? "运行中" : "演示"}
      </button>
      <GodotStatusBadge />
    </div>
  );
}
