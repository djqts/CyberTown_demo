import { create } from "zustand";
import type { Location, Npc, TownState } from "../lib/town-types";

type TownEvent = { type: string; data: Record<string, unknown>; createdAt: number };
const EVENTS_STORAGE_KEY = "town_events";
const MAX_EVENTS = 500;
const MAX_EVENT_AGE_MS = 2 * 60 * 60 * 1000; // 2 hours

function loadEvents(): TownEvent[] {
  try {
    const raw = localStorage.getItem(EVENTS_STORAGE_KEY);
    if (!raw) return [];
    const all = JSON.parse(raw) as TownEvent[];
    const cutoff = Date.now() - MAX_EVENT_AGE_MS;
    return all.filter((e) => e.createdAt > cutoff);
  } catch { return []; }
}

function saveEvents(events: TownEvent[]) {
  try { localStorage.setItem(EVENTS_STORAGE_KEY, JSON.stringify(events)); } catch { /* ignore */ }
}

// Track last event for dedup within a short window
let _lastEventKey = "";
let _lastEventTime = 0;
const DEDUP_WINDOW_MS = 500;

function isDuplicate(type: string, data: Record<string, unknown>): boolean {
  const key = type + "|" + JSON.stringify(data);
  const now = Date.now();
  if (key === _lastEventKey && now - _lastEventTime < DEDUP_WINDOW_MS) {
    return true;
  }
  _lastEventKey = key;
  _lastEventTime = now;
  return false;
}

type TownStore = {
  ready: boolean;
  online: boolean;
  wsConnected: boolean;
  townState: TownState | null;
  locations: Location[];
  npcs: Npc[];
  selectedNpcId: number | null;
  selectedLocationId: number | null;
  events: TownEvent[];
  setReadyData(data: { online: boolean; town_state: TownState; locations: Location[]; npcs: Npc[] }): void;
  setSelectedNpc(npcId: number, locationId: number): void;
  setSelectedLocation(locationId: number): void;
  clearSelection(): void;
  setOnline(online: boolean): void;
  setWsConnected(connected: boolean): void;
  pushEvent(type: string, data: Record<string, unknown>): void;
  updateNpcLocation(npcId: number, toLocation: string, locations: Location[]): void;
  updateNpcMood(npcId: number, newMood: string): void;
  updateNpcGoal(npcId: number, newGoal: string): void;
};

export const useTownStore = create<TownStore>((set) => ({
  ready: false,
  online: false,
  wsConnected: false,
  townState: null,
  locations: [],
  npcs: [],
  selectedNpcId: null,
  selectedLocationId: null,
  events: loadEvents(),
  setReadyData: (data) =>
    set({
      ready: true,
      online: data.online,
      townState: data.town_state,
      locations: data.locations,
      npcs: data.npcs,
    }),
  setSelectedNpc: (npcId, locationId) => set({ selectedNpcId: npcId, selectedLocationId: locationId }),
  setSelectedLocation: (locationId) => set({ selectedLocationId: locationId }),
  clearSelection: () => set({ selectedNpcId: null, selectedLocationId: null }),
  setOnline: (online) => set({ online }),
  setWsConnected: (connected) => set({ wsConnected: connected }),
  pushEvent: (type, data) =>
    set((state) => {
      if (isDuplicate(type, data)) return { events: state.events };
      const events = [{ type, data, createdAt: Date.now() }, ...state.events].slice(0, MAX_EVENTS);
      saveEvents(events);
      return { events };
    }),
  updateNpcLocation: (npcId, toLocation, locations) =>
    set((state) => {
      const loc = locations.find((l) => l.name === toLocation);
      return {
        npcs: state.npcs.map((n) =>
          n.id === npcId && loc ? { ...n, location_id: loc.id } : n
        ),
      };
    }),
  updateNpcMood: (npcId, newMood) =>
    set((state) => ({
      npcs: state.npcs.map((n) => (n.id === npcId ? { ...n, mood: newMood } : n)),
    })),
  updateNpcGoal: (npcId, newGoal) =>
    set((state) => ({
      npcs: state.npcs.map((n) => (n.id === npcId ? { ...n, current_goal: newGoal } : n)),
    })),
}));
