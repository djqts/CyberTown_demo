import { useState, useEffect } from "react";
import { useTownStore } from "../store/town-store";
import { TopBar } from "../components/layout/TopBar";
import { GodotMapEmbed } from "../components/godot/GodotMapEmbed";
import { NpcList } from "../components/town/NpcList";
import { EventStream } from "../components/town/EventStream";
import { ChatPanel } from "../components/town/ChatPanel";
import { NpcDetailPanel } from "../components/town/NpcDetailPanel";
import { MapPin, Newspaper, MessageCircle, X, ArrowLeft } from "lucide-react";

export default function TownApp() {
  const selectedNpcId = useTownStore((s) => s.selectedNpcId);
  const npcCount = useTownStore((s) => s.npcs.length);
  const clearSelection = useTownStore((s) => s.clearSelection);
  const [panel, setPanel] = useState<"closed" | "npcs" | "events">("closed");
  const [chatOpen, setChatOpen] = useState(false);

  useEffect(() => {
    function openChat() { setChatOpen(true); }
    window.addEventListener("town.openChat", openChat);
    return () => window.removeEventListener("town.openChat", openChat);
  }, []);

  // When NPC selected (from map click or list click), auto-open NPC panel
  useEffect(() => {
    if (selectedNpcId != null) {
      setPanel("npcs");
    }
  }, [selectedNpcId]);

  return (
    <div className="flex h-screen flex-col bg-[#0f1923] text-[#F3F5F7]">
      <TopBar />

      {/* Map fills entire remaining space — index.html pattern */}
      <div className="flex-1 min-h-0 relative">
        <GodotMapEmbed />

        {/* Floating toggle buttons — top-right of map */}
        <div className="absolute top-3 right-3 z-30 flex gap-2">
          <button
            onClick={() => setPanel(panel === "npcs" ? "closed" : "npcs")}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[12px] font-medium backdrop-blur-sm transition-all ${
              panel === "npcs"
                ? "bg-[#7CCB8A]/20 text-[#7CCB8A] border border-[#7CCB8A]/30"
                : "bg-[#0f1923]/80 text-[#B8C2CC] border border-white/10 hover:border-white/20"
            }`}
          >
            <MapPin size={13} /> 居民 · {npcCount}
          </button>
          <button
            onClick={() => setPanel(panel === "events" ? "closed" : "events")}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[12px] font-medium backdrop-blur-sm transition-all ${
              panel === "events"
                ? "bg-[#7CCB8A]/20 text-[#7CCB8A] border border-[#7CCB8A]/30"
                : "bg-[#0f1923]/80 text-[#B8C2CC] border border-white/10 hover:border-white/20"
            }`}
          >
            <Newspaper size={13} /> 事件
          </button>
          {selectedNpcId && (
            <button
              onClick={() => setChatOpen(!chatOpen)}
              className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[12px] font-medium backdrop-blur-sm transition-all ${
                chatOpen
                  ? "bg-[#7CCB8A]/20 text-[#7CCB8A] border border-[#7CCB8A]/30"
                  : "bg-[#0f1923]/80 text-[#B8C2CC] border border-white/10 hover:border-white/20"
              }`}
            >
              <MessageCircle size={13} /> 对话
            </button>
          )}
        </div>

        {/* Floating right panel — slides over map */}
        {panel !== "closed" && (
          <div className="absolute top-3 right-3 bottom-3 w-[320px] z-20 rounded-xl border border-[#1e2d3d] bg-[#111b26]/95 backdrop-blur-md shadow-2xl flex flex-col overflow-hidden"
               style={{ top: "48px" }}>
            <div className="flex items-center justify-between px-3 py-2 border-b border-[#1e2d3d] shrink-0">
              <div className="flex items-center gap-2">
                {panel === "npcs" && selectedNpcId && (
                  <button onClick={clearSelection} className="text-[#B8C2CC]/50 hover:text-[#7CCB8A] transition-colors" title="返回列表">
                    <ArrowLeft size={14} />
                  </button>
                )}
                <span className="text-[12px] font-medium text-[#B8C2CC]">
                  {panel === "npcs" ? `居民 · ${npcCount}` : "实时事件"}
                </span>
              </div>
              <button onClick={() => { setPanel("closed"); clearSelection(); }} className="text-[#B8C2CC]/40 hover:text-[#B8C2CC] transition-colors">
                <X size={14} />
              </button>
            </div>
            <div className="flex-1 min-h-0 overflow-y-auto">
              {panel === "npcs" ? (
                selectedNpcId ? <NpcDetailPanel /> : <NpcList />
              ) : (
                <EventStream />
              )}
            </div>
          </div>
        )}

        {/* Floating chat — bottom overlay. Always mounted so NPC replies are captured. */}
        <div className={`absolute bottom-3 left-3 right-3 z-20 h-[200px] rounded-xl border border-[#1e2d3d] bg-[#0d1520]/95 backdrop-blur-md shadow-2xl overflow-hidden transition-all duration-200 ${chatOpen && selectedNpcId ? '' : 'hidden'}`}>
          <div className="absolute top-2 right-3 z-10">
            <button onClick={() => setChatOpen(false)} className="text-[#B8C2CC]/40 hover:text-[#B8C2CC] transition-colors">
              <X size={14} />
            </button>
          </div>
          <ChatPanel />
        </div>
      </div>
    </div>
  );
}
