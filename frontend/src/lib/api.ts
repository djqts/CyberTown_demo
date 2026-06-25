import type { TownState, Npc, Location } from "./town-types";

const baseUrl = import.meta.env.VITE_TOWN_HTTP_BASE_URL ?? "http://localhost:8080";

export async function getTownState(): Promise<TownState> {
  const res = await fetch(`${baseUrl}/api/town/state`);
  if (!res.ok) throw new Error(`GET /api/town/state failed`);
  return res.json();
}

export async function getNpcs(): Promise<Npc[]> {
  const res = await fetch(`${baseUrl}/api/npcs`);
  if (!res.ok) throw new Error(`GET /api/npcs failed`);
  return res.json();
}

export async function getNpcDetail(npcId: number) {
  const res = await fetch(`${baseUrl}/api/npcs/${npcId}`);
  if (!res.ok) throw new Error(`GET /api/npcs/${npcId} failed`);
  return res.json();
}

export async function getMapData(): Promise<{
  locations: Location[];
  districts: Array<{ id: string; name: string; x: number; y: number; width: number; height: number; color: string }>;
  roads: Array<{ from: number; to: number; name: string }>;
}> {
  const res = await fetch(`${baseUrl}/api/map`);
  if (!res.ok) {
    // Fallback: return embedded map data if API not yet implemented
    return FALLBACK_MAP_DATA;
  }
  return res.json();
}

// Embedded fallback when /api/map is not yet implemented on backend
const FALLBACK_MAP_LOCATIONS: Location[] = [
  { id: 1, name: "广场", icon: "🏛", district: "central", size: "large", x: 480, y: 350 },
  { id: 2, name: "咖啡馆", icon: "☕", district: "left_commercial", size: "medium", x: 180, y: 280 },
  { id: 3, name: "钟楼", icon: "🕐", district: "central", size: "large", x: 520, y: 120 },
  { id: 4, name: "市政厅", icon: "🏛", district: "central", size: "medium", x: 440, y: 300 },
  { id: 5, name: "图书馆", icon: "📚", district: "central", size: "medium", x: 520, y: 440 },
  { id: 6, name: "花店", icon: "🌸", district: "left_commercial", size: "small", x: 120, y: 330 },
  { id: 7, name: "铁匠铺", icon: "🔨", district: "right", size: "medium", x: 720, y: 280 },
  { id: 8, name: "诊所", icon: "🏥", district: "left_commercial", size: "medium", x: 180, y: 420 },
  { id: 9, name: "农舍", icon: "🌾", district: "farmland", size: "medium", x: 60, y: 180 },
  { id: 10, name: "钓鱼小屋", icon: "🎣", district: "lake", size: "small", x: 420, y: 60 },
  { id: 11, name: "学校", icon: "🏫", district: "left_commercial", size: "medium", x: 260, y: 440 },
  { id: 12, name: "面包店", icon: "🍞", district: "left_commercial", size: "small", x: 220, y: 250 },
  { id: 13, name: "酒馆", icon: "🍺", district: "right", size: "medium", x: 760, y: 380 },
  { id: 14, name: "公园凉亭", icon: "🎵", district: "park", size: "medium", x: 420, y: 600 },
  { id: 15, name: "手工工坊", icon: "🪚", district: "left_commercial", size: "small", x: 280, y: 350 },
  { id: 16, name: "住宅区", icon: "🏠", district: "residential", size: "medium", x: 200, y: 650 },
  { id: 17, name: "森林营地", icon: "🏕", district: "forest", size: "medium", x: 880, y: 580 },
];

const FALLBACK_MAP_DATA = {
  locations: FALLBACK_MAP_LOCATIONS,
  roads: [] as Array<{ from: number; to: number }>,
  districts: [
    { id: "central", name: "中央广场区", x: 350, y: 250, width: 300, height: 200, color: "#fffde7" },
    { id: "left_commercial", name: "左区商业街", x: 100, y: 150, width: 220, height: 350, color: "#f3e5f5" },
    { id: "right", name: "右区", x: 680, y: 150, width: 250, height: 350, color: "#e3f2fd" },
    { id: "farmland", name: "远左农田区", x: 0, y: 50, width: 180, height: 200, color: "#e8f5e9" },
    { id: "lake", name: "湖区", x: 350, y: 0, width: 200, height: 100, color: "#e0f7fa" },
    { id: "park", name: "公园区", x: 300, y: 550, width: 280, height: 150, color: "#c8e6c9" },
    { id: "residential", name: "住宅区", x: 150, y: 600, width: 150, height: 130, color: "#fff3e0" },
    { id: "forest", name: "森林边缘", x: 800, y: 520, width: 200, height: 200, color: "#dcedc8" },
  ],
};

export function getFallbackLocations(): Location[] {
  return FALLBACK_MAP_LOCATIONS;
}
