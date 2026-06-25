export type Location = {
  id: number;
  name: string;
  icon?: string;
  district?: string;
  size?: "small" | "medium" | "large" | string;
  x?: number;
  y?: number;
};

export type Npc = {
  id: number;
  name: string;
  role: string;
  location_id: number;
  location_name?: string;
  mood: string;
  energy: number;
  current_goal: string;
  gender?: string;
  age_group?: string;
  portrait_color?: string;
};

export type TownState = {
  name: string;
  day?: number;
  minute?: number;
  time: string;
  season: string;
  weather: string;
};

export type GodotEvent =
  | { type: "godot.map.ready"; data: { online: boolean; town_state: TownState; locations: Location[]; npcs: Npc[] } }
  | { type: "godot.npc.selected"; data: { npc_id: number; location_id: number } }
  | { type: "godot.location.selected"; data: { location_id: number } }
  | { type: "godot.ws.message"; data: { type: string; data: Record<string, unknown> } }
  | { type: "godot.ws.connected"; data: Record<string, never> }
  | { type: "godot.ws.disconnected"; data: Record<string, never> }
  | { type: "godot.offline"; data: { reason: string } }
  | { type: "godot.validation.failed"; data: { ok: false; errors: string[]; warnings: string[] } };
