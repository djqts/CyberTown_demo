import { useState, useRef, useEffect, useCallback } from "react";
import { useTownStore } from "../../store/town-store";
import { Send, Loader2, MessageCircle } from "lucide-react";
import { getNpcPortrait } from "../../lib/utils";
import gsap from "gsap";

type Message = {
  id: number;
  sender: string;
  content: string;
  isUser: boolean;
  timestamp: number;
};

const STORAGE_KEY = (npcId: number) => `chat_history_${npcId}`;
const MAX_MESSAGES = 200;

function loadMessages(npcId: number): Message[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY(npcId));
    return raw ? (JSON.parse(raw) as Message[]) : [];
  } catch {
    return [];
  }
}

function saveMessages(npcId: number, msgs: Message[]) {
  try {
    localStorage.setItem(STORAGE_KEY(npcId), JSON.stringify(msgs.slice(-MAX_MESSAGES)));
  } catch { /* storage full */ }
}

export function ChatPanel() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const [sending, setSending] = useState(false);
  const selectedNpcId = useTownStore((s) => s.selectedNpcId);
  const selectedNpc = useTownStore((s) => s.npcs.find((n) => n.id === s.selectedNpcId));
  const scrollRef = useRef<HTMLDivElement>(null);
  const idRef = useRef(0);
  const msgsEndRef = useRef<HTMLDivElement>(null);

  // Load messages when NPC changes
  useEffect(() => {
    if (!selectedNpcId) { setMessages([]); return; }
    const saved = loadMessages(selectedNpcId);
    idRef.current = saved.length > 0 ? Math.max(...saved.map((m) => m.id)) + 1 : 0;
    setMessages(saved);
  }, [selectedNpcId]);

  // Save messages whenever they change
  useEffect(() => {
    if (!selectedNpcId || messages.length === 0) return;
    saveMessages(selectedNpcId, messages);
  }, [messages, selectedNpcId]);

  // Register global callback for NPC replies (direct from WS, bypasses store)
  useEffect(() => {
    const currentNpcId = selectedNpcId;
    window.__onChatReply = (data: { npc_id: number; npc_name: string; content: string }) => {
      if (!currentNpcId || data.npc_id === currentNpcId) {
        setMessages((msgs) => [...msgs, {
          id: ++idRef.current, sender: data.npc_name ?? "NPC",
          content: data.content ?? "", isUser: false, timestamp: Date.now(),
        }]);
        setSending(false);
      }
    };
    return () => { delete window.__onChatReply; };
  }, [selectedNpcId]);

  useEffect(() => {
    msgsEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  useEffect(() => {
    if (!msgsEndRef.current) return;
    const parent = msgsEndRef.current.parentElement;
    if (!parent) return;
    const newMsgs = parent.querySelectorAll(".chat-msg-new");
    if (newMsgs.length > 0) {
      gsap.fromTo(newMsgs, { opacity: 0, y: 8 }, { opacity: 1, y: 0, duration: 0.25, stagger: 0.05, ease: "power2.out" });
    }
    newMsgs.forEach((el) => el.classList.remove("chat-msg-new"));
  }, [messages]);

  const sendMessage = useCallback(() => {
    const content = input.trim();
    if (!content || !selectedNpcId) return;
    window.__sendTownMessage?.(selectedNpcId, content);
    const newMsg: Message = {
      id: ++idRef.current, sender: "你", content, isUser: true, timestamp: Date.now(),
    };
    setMessages((msgs) => [...msgs, newMsg]);
    setInput("");
    setSending(true);
  }, [input, selectedNpcId]);

  if (!selectedNpcId) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 p-12 text-[#B8C2CC]/40">
        <MessageCircle size={32} strokeWidth={1} />
        <p className="text-[13px]">点击左侧 NPC 开始对话</p>
      </div>
    );
  }

  return (
    <div className="flex gap-3 h-full p-3">
      {/* Messages area */}
      <div className="flex-1 flex flex-col min-w-0">
        <div className="flex items-center gap-2 mb-2 shrink-0">
          {selectedNpc?.role && getNpcPortrait(selectedNpc.role) ? (
            <img src={getNpcPortrait(selectedNpc.role)} alt={selectedNpc.name}
              className="w-7 h-7 rounded-full object-cover border border-white/10 shrink-0" />
          ) : (
            <div className="w-7 h-7 rounded-full flex items-center justify-center text-[9px] font-bold text-white shrink-0"
              style={{ background: `linear-gradient(135deg, ${selectedNpc?.portrait_color ?? "#888"}, ${selectedNpc?.portrait_color ?? "#888"}99)` }}>
              {selectedNpc?.name?.[0] ?? "?"}
            </div>
          )}
          <span className="text-[13px] font-medium text-[#F3F5F7]">{selectedNpc?.name ?? "NPC"}</span>
        </div>
        <div ref={scrollRef} className="flex-1 overflow-auto min-h-0">
          <div className="space-y-1.5">
            {messages.map((msg) => (
              <div key={msg.id} className={`chat-msg-new flex ${msg.isUser ? "justify-end" : "justify-start"}`}>
                <div className={`max-w-[75%] rounded-xl px-3 py-1.5 text-[12px] leading-relaxed ${
                  msg.isUser
                    ? "bg-[#7CCB8A]/15 text-[#E8F5E9] rounded-br-md"
                    : "bg-[#1e2d3d]/50 text-[#F3F5F7] rounded-bl-md"
                }`}>
                  <div className="text-[10px] text-[#B8C2CC]/50 mb-0.5">{msg.sender}</div>
                  {msg.content}
                </div>
              </div>
            ))}
            <div ref={msgsEndRef} />
          </div>
        </div>
      </div>

      {/* Input area */}
      <div className="w-[280px] shrink-0 flex flex-col justify-end gap-2">
        <input
          placeholder={`和 ${selectedNpc?.name ?? "NPC"} 说点什么...`}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && sendMessage()}
          className="w-full h-9 rounded-lg border border-[#1e2d3d] bg-[#0a1018] px-3 text-[12px] text-[#F3F5F7] placeholder:text-[#B8C2CC]/30 outline-none focus:border-[#7CCB8A]/40 transition-colors"
        />
        <button
          onClick={sendMessage}
          disabled={sending || !input.trim()}
          className="w-full h-9 rounded-lg bg-[#7CCB8A] hover:bg-[#6ab878] disabled:opacity-30 transition-all flex items-center justify-center gap-1.5 text-[12px] font-medium text-[#0f1923]"
        >
          {sending ? <Loader2 size={13} className="animate-spin" /> : <Send size={13} />}
          发送
        </button>
      </div>
    </div>
  );
}
