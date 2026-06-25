/// <reference types="vite/client" />

import type { GodotEvent } from "./lib/town-types";

declare global {
  interface Window {
    CYBERTOWN_CONFIG?: {
      httpBaseUrl: string;
      wsUrl: string;
      userToken: string;
    };
    cybertownGodot?: {
      focusLocation(locationId: number): void;
      focusNpc(npcId: number): void;
      highlightLocation(locationId: number, duration?: number): void;
      highlightNpc(npcId: number, duration?: number): void;
      selectNpc(npcId: number): void;
      selectLocation(locationId: number): void;
      getLocationData(): string;
      getNpcData(): string;
      getTownState(): string;
      getNpcsAtLocation(locationId: number): string;
      getOnlineStatus(): boolean;
      getWsConnected(): boolean;
    };
    onGodotEvent?: (event: GodotEvent) => void;
    __townWs?: WebSocket;
    __sendTownMessage?: (npcId: number, content: string) => void;
    __onChatReply?: (data: { npc_id: number; npc_name: string; content: string }) => void;
    __diag?: {
      entries: { time: number; type: string; data: Record<string, unknown> }[];
      record(type: string, data: Record<string, unknown>): void;
      report(): { total: number; byType: Record<string, number>; movedCount: number; moved: Record<string, unknown>[] };
      clear(): void;
    };
    __godotStart?: (canvas: HTMLCanvasElement) => Promise<void>;
    __godotStarted?: boolean;
    __godotReady?: boolean;
    Engine?: new (config: Record<string, unknown>) => { startGame(opts: Record<string, unknown>): Promise<void> };
  }
}

export {};
