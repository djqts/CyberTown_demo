import { useEffect, useRef, useState, useCallback } from "react";
import { useTownStore } from "../../store/town-store";
import { getTownState, getNpcs, getFallbackLocations } from "../../lib/api";
import { TownMap } from "../town/TownMap";

const WS_URL = import.meta.env.VITE_TOWN_WS_URL ?? "ws://localhost:8080/ws";

export function GodotMapEmbed() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [mode, setMode] = useState<"loading" | "godot" | "fallback">("loading");
  const [statusText, setStatusText] = useState("加载中...");
  const setReadyData = useTownStore((s) => s.setReadyData);
  const setSelectedNpc = useTownStore((s) => s.setSelectedNpc);
  const setSelectedLocation = useTownStore((s) => s.setSelectedLocation);
  const setOnline = useTownStore((s) => s.setOnline);
  const setWsConnected = useTownStore((s) => s.setWsConnected);
  const pushEvent = useTownStore((s) => s.pushEvent);

  // Initialize Godot and set up event listener
  useEffect(() => {
    let cleanup: (() => void) | undefined;

    function handleGodotEvent(e: Event | CustomEvent | Record<string, unknown>) {
      // Godot's onGodotEvent passes {type, data}, dispatchEvent passes CustomEvent with {detail: {type, data}}
      const customDetail = (e as CustomEvent).detail as Record<string, unknown> | undefined;
      const event: Record<string, unknown> | undefined =
        customDetail?.type ? customDetail :           // dispatchEvent path
        (e as Record<string, unknown>).type ? (e as Record<string, unknown>) :  // onGodotEvent raw object
        undefined;
      if (!event?.type) return;

      (window as any).__diag?.record("godot." + (event.type as string), (event.data ?? {}) as Record<string, unknown>);

      switch (event.type) {
        case "godot.map.ready":
          setReadyData(event.data);
          setMode("godot");
          break;
        case "godot.npc.selected": {
          const d = event.data as Record<string, unknown>;
          setSelectedNpc(d.npc_id as number, (d.location_id as number) ?? 0);
          break;
        }
        case "godot.location.selected": {
          const locId = (event.data as Record<string, unknown>).location_id as number;
          setSelectedLocation(locId);
          // Auto-select first NPC at this location for immediate detail view
          const npcsHere = useTownStore.getState().npcs.filter((n) => n.location_id === locId);
          if (npcsHere.length > 0) {
            setSelectedNpc(npcsHere[0].id, locId);
          }
          break;
        }
        case "godot.ws.connected": setWsConnected(true); break;
        case "godot.ws.disconnected": setWsConnected(false); break;
        case "godot.offline":
          setOnline(false);
          pushEvent(event.type, event.data);
          break;
        case "godot.ws.message": {
          const d = event.data?.data ?? {};
          pushEvent(event.data?.type ?? "", d as Record<string, unknown>);
          break;
        }
      }
    }

    const eventTypes = [
      "godot.map.ready", "godot.npc.selected", "godot.location.selected",
      "godot.ws.connected", "godot.ws.disconnected", "godot.offline",
      "godot.ws.message", "godot.validation.failed",
    ];

    eventTypes.forEach((t) => window.addEventListener(t, handleGodotEvent));
    window.onGodotEvent = handleGodotEvent;
    cleanup = () => {
      eventTypes.forEach((t) => window.removeEventListener(t, handleGodotEvent));
      delete window.onGodotEvent;
    };

    // Set up Godot config for the engine
    if (!window.CYBERTOWN_CONFIG) {
      window.CYBERTOWN_CONFIG = {
        httpBaseUrl: "http://localhost:8080",
        wsUrl: "ws://localhost:8080/ws",
        userToken: "web-user",
      };
    }

    // Centralized WS — used by ChatPanel to send messages (index.html pattern)
    let ws: WebSocket | null = null;
    function connectWs() {
      ws = new WebSocket(WS_URL + "?user_token=web-user");
      ws.onopen = () => { setWsConnected(true); setOnline(true); };
      ws.onmessage = (e) => {
        try {
          const m = JSON.parse(e.data); console.log("[WS-raw]", m.type);
          if (m.type && (m.type.startsWith("npc.") || m.type.startsWith("story.") || m.type === "demo.completed")) {
            console.log("[WS-broadcast]", m.type, JSON.stringify(m.data).slice(0, 200));
          }
          (window as any).__diag?.record(m.type ?? "unknown", m.data ?? {});
          pushEvent(m.type ?? "", m.data ?? {});
          if (m.type === "npc.replied" && window.__onChatReply) {
            window.__onChatReply(m.data as any);
          }
          if (m.type === "demo.completed") {
            window.dispatchEvent(new CustomEvent("demo.completed"));
          }
        } catch (err) { console.error("[WS] parse error:", err); }
      };
      ws.onclose = () => { setWsConnected(false); setTimeout(connectWs, 3000); };
      ws.onerror = () => ws?.close();
    }
    connectWs();

    window.__sendTownMessage = function(npcId: number, content: string) {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: "user.message", data: { npc_id: npcId, content, user_token: "web-user" } }));
      }
    };

    // Start Godot engine on canvas
    const canvas = canvasRef.current;
    if (canvas && window.__godotStart) {
      canvas.id = "canvas";
      window.__godotStart(canvas);
    }

    // Poll for bridge + fallback timer
    let attempts = 0;
    const iv = setInterval(() => {
      attempts++;
      if (window.cybertownGodot) {
        clearInterval(iv);
        setStatusText("");
        return;
      }
      if (attempts === 6) setStatusText("等待 Godot 初始化...");
      if (attempts >= 30) {
        clearInterval(iv);
        setMode("fallback");
        startFallback();
      }
    }, 1000);

    return () => {
      clearInterval(iv);
      cleanup?.();
    };
  }, []);

  async function startFallback() {
    try {
      const [townState, npcs] = await Promise.all([getTownState(), getNpcs()]);
      setReadyData({ online: true, town_state: townState, locations: getFallbackLocations(), npcs });
      setOnline(true);
      setStatusText("");
    } catch {
      setStatusText("无法连接后端");
    }
  }

  return (
    <div id="cybertown-shell" className="relative w-full h-full overflow-hidden bg-[#18202A]"
         >
      {/* Godot loading overlay */}
      <div id="godot-status" className="absolute inset-0 z-20 flex flex-col items-center justify-center bg-[#242424]"
           style={{ visibility: mode === "loading" && !statusText.includes("SVG") ? "visible" : "hidden" }}>
        <progress id="godot-status-progress" className="w-1/2" />
        <div id="godot-status-notice" className="hidden mt-4 mx-8 p-4 rounded-lg bg-[#5b3943] border border-[#9b3943] text-[#e0e0e0] text-center" />
      </div>

      {/* Status indicator */}
      {statusText && (
        <div className="pointer-events-none absolute right-3 top-3 z-10 inline-flex min-h-8 max-w-[min(360px,calc(100%-24px))] items-center rounded-lg border border-white/15 bg-[#18202A]/80 px-3 text-[13px] leading-tight text-[#F3F5F7] shadow-xl">
          {statusText}
        </div>
      )}

      {mode === "fallback" ? (
        <TownMap />
      ) : (
        <canvas ref={canvasRef} id="canvas" className="block w-full h-full outline-none" />
      )}
    </div>
  );
}
