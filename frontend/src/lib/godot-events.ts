import type { GodotEvent } from "./town-types";

type Handler = (event: GodotEvent) => void;

const eventTypes: GodotEvent["type"][] = [
  "godot.map.ready",
  "godot.npc.selected",
  "godot.location.selected",
  "godot.ws.message",
  "godot.ws.connected",
  "godot.ws.disconnected",
  "godot.offline",
  "godot.validation.failed",
];

export function subscribeGodotEvents(handler: Handler) {
  window.onGodotEvent = handler;

  const listeners = eventTypes.map((type) => {
    const listener = (event: Event) => {
      handler((event as CustomEvent<GodotEvent>).detail);
    };
    window.addEventListener(type, listener);
    return { type, listener };
  });

  return () => {
    if (window.onGodotEvent === handler) {
      window.onGodotEvent = undefined;
    }
    listeners.forEach(({ type, listener }) => {
      window.removeEventListener(type, listener);
    });
  };
}
