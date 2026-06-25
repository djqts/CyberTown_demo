import { Card, CardContent } from "../ui/Card";
import { Tabs } from "../ui/Tabs";
import { ScrollArea } from "../ui/ScrollArea";
import { LocationList } from "../town/LocationList";
import { NpcList } from "../town/NpcList";
import { ChatPanel } from "../town/ChatPanel";
import { EventStream } from "../town/EventStream";
import { NpcDetailPanel } from "../town/NpcDetailPanel";
import { LocationDetailPanel } from "../town/LocationDetailPanel";
import { TownStateBar } from "../town/TownStateBar";
import { useTownStore } from "../../store/town-store";
import { useState } from "react";

export function SidePanel() {
  const selectedNpcId = useTownStore((s) => s.selectedNpcId);
  const selectedLocationId = useTownStore((s) => s.selectedLocationId);
  const [tab, setTab] = useState("town");

  const tabs = [
    { key: "town", label: "小镇" },
    { key: "npc", label: selectedNpcId ? "NPC ✓" : "NPC" },
    { key: "location", label: selectedLocationId ? "地点 ✓" : "地点" },
    { key: "chat", label: "聊天" },
    { key: "events", label: "事件" },
  ];

  return (
    <div className="flex h-full flex-col">
      <Card className="flex-1 rounded-none border-0">
        <CardContent className="p-0">
          <Tabs tabs={tabs} active={tab} onChange={setTab}>
            <ScrollArea className="h-[calc(100vh-11rem)]">
              {tab === "town" && <TownStateBar />}
              {tab === "town" && <LocationList />}
              {tab === "npc" && (selectedNpcId ? <NpcDetailPanel /> : <NpcList />)}
              {tab === "location" && (selectedLocationId ? <LocationDetailPanel /> : <LocationList />)}
              {tab === "chat" && <ChatPanel />}
              {tab === "events" && <EventStream />}
            </ScrollArea>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
}
