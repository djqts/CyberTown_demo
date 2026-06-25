export function setupGodotConfig() {
  window.CYBERTOWN_CONFIG = {
    httpBaseUrl: import.meta.env.VITE_TOWN_HTTP_BASE_URL ?? "http://localhost:8080",
    wsUrl: import.meta.env.VITE_TOWN_WS_URL ?? "ws://localhost:8080/ws",
    userToken: import.meta.env.VITE_TOWN_USER_TOKEN ?? "web-user",
  };
}

export function focusGodotLocation(locationId: number) {
  window.cybertownGodot?.focusLocation(locationId);
}

export function focusGodotNpc(npcId: number) {
  if (window.cybertownGodot) {
    window.cybertownGodot.selectNpc(npcId);
    window.cybertownGodot.focusNpc(npcId);
  }
}
