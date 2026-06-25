import { clsx, type ClassValue } from "clsx";

export function cn(...inputs: ClassValue[]) {
  return clsx(inputs);
}

export const MOOD_MAP: Record<string, string> = {
  cheerful: "愉快", happy: "开心", excited: "兴奋", inspired: "受启发",
  content: "满足", calm: "平静", focused: "专注", composed: "镇定",
  peaceful: "安宁", warm: "温和", jolly: "快活", friendly: "友好",
  dreamy: "梦幻", steady: "稳定", playful: "顽皮", confident: "自信",
  anxious: "焦虑", worried: "担心", sad: "悲伤", angry: "生气", tired: "疲惫",
  curious: "好奇", caring: "关心", cautious: "谨慎", alert: "警觉", helpful: "乐于助人",
  grateful: "感激", touched: "感动", nostalgic: "怀旧", amused: "被逗乐",
  professional: "专业", satisfied: "满意", proud: "自豪", neutral: "平静",
};

export function moodLabel(mood: string): string {
  return MOOD_MAP[mood] ?? mood;
}

export const DISTRICTS: Record<string, string> = {
  central: "中央广场区", left_commercial: "左区商业街", right: "右区",
  farmland: "远左农田区", lake: "湖区", park: "公园区",
  residential: "住宅区", forest: "森林边缘",
};

export const NPC_PORTRAITS: Record<string, string> = {
  "镇长": "/npc/mayor.png",
  "咖啡馆主": "/npc/cafe.png",
  "图书管理员": "/npc/librarian.png",
  "花店店主": "/npc/florist.png",
  "铁匠": "/npc/blacksmith.png",
  "医生": "/npc/doctor.png",
  "农夫": "/npc/farmer.png",
  "渔夫": "/npc/fisher.png",
  "教师": "/npc/teacher.png",
  "面包师": "/npc/baker.png",
  "酒馆老板": "/npc/bartender.png",
  "音乐家": "/npc/musician.png",
  "木匠": "/npc/carpenter.png",
  "小女孩": "/npc/girl.png",
  "冒险者": "/npc/adventurer.png",
};

export function getNpcPortrait(role: string): string {
  return NPC_PORTRAITS[role] ?? "";
}
