"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import {
  Send, Paperclip, ToggleLeft, ToggleRight, TrendingUp, TrendingDown,
  Minus, Eye, Flame, Zap, ChevronRight,
} from "lucide-react";

// ---- Types ----

interface ChatMessage { role: "user" | "assistant"; content: string; isTyping?: boolean; }
interface MarketCard { label: string; value: string; change?: string; changePos?: boolean; badge?: string; }

// ---- Mock data (replace with API calls) ----

const MARKET_CARDS: MarketCard[] = [
  { label: "VN-Index",          value: "1,268.45", change: "-0.61%", changePos: false },
  { label: "P/E thị trường",    value: "12.7x",    change: "+0.2x",  changePos: true  },
  { label: "P/B thị trường",    value: "1.54x",    change: "0.00x",  changePos: true  },
  { label: "Chu kỳ thị trường", value: "Mở rộng",  badge: "62%",                      },
];

const WATCHLIST = [
  { symbol: "VHM",  price: "43,700", pct: "+5.05%" , trend: +1 },
  { symbol: "VIC",  price: "141,000", pct: "+2.97%", trend: +1 },
  { symbol: "HSG",  price: "22,850",  pct: "+2.93%", trend: +1 },
  { symbol: "FPT",  price: "132,000", pct: "+2.09%", trend: +1 },
  { symbol: "HPG",  price: "27,250",  pct: "+1.68%", trend: +1 },
  { symbol: "MSN",  price: "76,000",  pct: "+0.13%", trend: +1 },
  { symbol: "GEX",  price: "26,000",  pct: "-0.38%", trend: -1 },
];

const BREAKTHROUGHS = [
  { symbol: "BSR",  pct: "+6.50%", vol: "13M" },
  { symbol: "GEE",  pct: "+5.50%", vol: "2.3M" },
  { symbol: "PVB",  pct: "+4.96%", vol: "1.1M" },
  { symbol: "VCI",  pct: "+4.47%", vol: "4.5M" },
  { symbol: "VHM",  pct: "+4.42%", vol: "19M" },
];

const NEWS_ITEMS = [
  { time: "07:45", title: "VN-Index mở phiên giảm nhẹ, nhóm Vincgroup bứt phá", source: "Cafef" },
  { time: "07:30", title: "Khối ngoại quay lại mua ròng sau 5 phiên bán liên tiếp", source: "VnEconomy" },
  { time: "07:15", title: "NHNN bơm tiền qua kênh OMO, lãi suất liên ngân hàng giảm", source: "ThienViet" },
  { time: "06:55", title: "Dầu thô tiếp tục đà tăng sau thông tin từ OPEC+", source: "Bloomberg" },
  { time: "06:30", title: "Chứng khoán Mỹ chốt phiên: S&P 500 tăng 0.8% do số liệu việc làm", source: "Reuters" },
];

const SUGGESTIONS = [
  "Phân tích VNM hiện tại có nên mua không?",
  "Top cổ phiếu upside cao nhất hôm nay?",
  "Giải thích chu kỳ thị trường VN-Index",
  "So sánh HPG và HSG trong ngành thép",
];

// ---- Main Component ----

export default function AgentPage() {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [expertMode, setExpertMode] = useState(false);
  const [activeFilter, setActiveFilter] = useState("Việt Nam");
  const chatEndRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
      textareaRef.current.style.height = Math.min(textareaRef.current.scrollHeight, 120) + "px";
    }
  }, [input]);

  const handleSend = useCallback(async () => {
    const trimmed = input.trim();
    if (!trimmed || isLoading) return;
    setInput("");
    setIsLoading(true);
    setMessages(prev => [...prev, { role: "user", content: trimmed }]);
    setMessages(prev => [...prev, { role: "assistant", content: "", isTyping: true }]);

    try {
      const res = await fetch("http://localhost:8080/api/chat", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ message: trimmed, symbol: "", history: [] }),
      });
      const data = await res.json();
      const reply = data.reply ?? data.error ?? "Không có phản hồi.";
      setMessages(prev => {
        const updated = [...prev];
        const idx = updated.findLastIndex(m => m.isTyping);
        if (idx >= 0) updated[idx] = { role: "assistant", content: reply };
        else updated.push({ role: "assistant", content: reply });
        return updated;
      });
    } catch {
      setMessages(prev => {
        const updated = [...prev];
        const idx = updated.findLastIndex(m => m.isTyping);
        const err = { role: "assistant" as const, content: "Không thể kết nối server." };
        if (idx >= 0) updated[idx] = err; else updated.push(err);
        return updated;
      });
    } finally {
      setIsLoading(false);
    }
  }, [input, isLoading]);

  const handleKey = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); handleSend(); }
  };

  return (
    <div className="flex h-[calc(100vh-3rem)] gap-0 -mx-4 md:-mx-6 -mt-4 md:-mt-6 overflow-hidden animate-fade-in">

      {/* ── CENTER: Chat area ── */}
      <div className="flex flex-col flex-1 min-w-0 border-r border-border-theme">

        {/* Market overview cards strip */}
        <div className="px-4 py-3 border-b border-border-theme bg-surface/10 flex-shrink-0">
          <div className="flex items-center gap-1.5 mb-2">
            <span className="text-[10px] text-text-muted font-semibold uppercase tracking-widest">Thị trường</span>
            {["Việt Nam", "Tài sản số"].map(f => (
              <button key={f} onClick={() => setActiveFilter(f)}
                className={`px-2.5 py-0.5 rounded text-[10px] font-medium transition ${activeFilter === f ? "bg-indigo-600 text-white" : "text-text-muted hover:text-foreground"}`}
              >{f}</button>
            ))}
          </div>
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-2">
            {MARKET_CARDS.map(card => (
              <div key={card.label} className="bg-card-bg border border-border-theme rounded-lg px-3 py-2">
                <p className="text-[10px] text-text-muted mb-0.5">{card.label}</p>
                <p className="text-sm font-bold text-foreground tabular-nums">{card.value}</p>
                {card.change && (
                  <p className={`text-[10px] font-semibold flex items-center gap-0.5 ${card.changePos ? "text-positive" : "text-negative"}`}>
                    {card.changePos ? <TrendingUp size={9}/> : <TrendingDown size={9}/>}
                    {card.change}
                  </p>
                )}
                {card.badge && (
                  <p className="text-[10px] text-emerald-400 font-semibold">{card.badge}</p>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Chat messages */}
        <div className="flex-1 overflow-y-auto px-4 py-4 space-y-3 custom-scrollbar">
          {/* News feed when no messages */}
          {messages.length === 0 && (
            <div className="space-y-4">
              {/* Suggestions */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                {SUGGESTIONS.map((s, i) => (
                  <button key={i} onClick={() => setInput(s)}
                    className="text-left px-4 py-3 bg-card-bg border border-border-theme rounded-xl text-xs text-foreground hover:border-indigo-500/40 hover:bg-surface transition group">
                    <span className="flex items-center gap-2">
                      <Zap size={12} className="text-indigo-400 flex-shrink-0" />
                      {s}
                      <ChevronRight size={12} className="ml-auto text-text-muted group-hover:text-indigo-400 transition" />
                    </span>
                  </button>
                ))}
              </div>

              {/* News feed */}
              <div>
                <p className="text-[10px] font-semibold text-text-muted uppercase tracking-widest mb-2 flex items-center gap-1">
                  <Flame size={10} className="text-orange-400" /> Tin tức hôm nay
                </p>
                <div className="space-y-1">
                  {NEWS_ITEMS.map((item, i) => (
                    <div key={i} className="flex items-start gap-3 px-3 py-2 rounded-lg hover:bg-surface transition cursor-pointer">
                      <span className="text-[10px] text-text-muted pt-0.5 w-10 flex-shrink-0 tabular-nums">{item.time}</span>
                      <div className="flex-1 min-w-0">
                        <p className="text-xs font-medium text-foreground hover:text-indigo-400 transition line-clamp-1">{item.title}</p>
                        <p className="text-[10px] text-text-muted">{item.source}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* Actual chat messages */}
          {messages.map((msg, i) => (
            <div key={i} className={`flex gap-2.5 ${msg.role === "user" ? "flex-row-reverse" : ""}`}>
              <div className={`w-6 h-6 rounded-full flex-shrink-0 flex items-center justify-center mt-0.5 text-[10px] font-bold ${msg.role === "user" ? "bg-indigo-600 text-white" : "bg-gradient-to-br from-indigo-500 to-purple-600 text-white"}`}>
                {msg.role === "user" ? "U" : "AI"}
              </div>
              <div className={`max-w-[75%] ${msg.role === "user" ? "text-right" : ""}`}>
                {msg.isTyping ? (
                  <div className="bg-card-bg border border-border-theme rounded-xl px-3 py-2.5 flex gap-1 items-center">
                    {[0,1,2].map(i => <div key={i} className="w-1.5 h-1.5 bg-indigo-400 rounded-full animate-bounce" style={{ animationDelay: `${i*0.15}s` }} />)}
                  </div>
                ) : (
                  <div className={`px-3 py-2.5 rounded-xl text-xs leading-relaxed ${msg.role === "user" ? "bg-indigo-600 text-white rounded-tr-sm" : "bg-card-bg border border-border-theme text-foreground rounded-tl-sm"}`}>
                    {msg.content}
                  </div>
                )}
              </div>
            </div>
          ))}
          <div ref={chatEndRef} />
        </div>

        {/* Chat input */}
        <div className="p-3 border-t border-border-theme bg-card-bg flex-shrink-0">
          <div className="flex items-end gap-2">
            <button className="p-2 text-text-muted hover:text-foreground hover:bg-surface rounded-lg transition flex-shrink-0" title="Đính kèm file">
              <Paperclip size={16} />
            </button>
            <textarea
              ref={textareaRef}
              value={input}
              onChange={e => setInput(e.target.value)}
              onKeyDown={handleKey}
              placeholder="Hỏi MyFi AI về thị trường, cổ phiếu..."
              rows={1}
              className="flex-1 bg-surface border border-border-theme rounded-xl px-3 py-2 text-xs text-foreground placeholder-text-muted focus:outline-none focus:border-accent/50 resize-none max-h-28 custom-scrollbar transition"
            />
            <button
              onClick={() => setExpertMode(v => !v)}
              className="flex items-center gap-1 text-[10px] font-medium text-text-muted hover:text-foreground transition flex-shrink-0"
              title="Expert mode"
            >
              {expertMode ? <ToggleRight size={16} className="text-indigo-400" /> : <ToggleLeft size={16} />}
              <span className="hidden sm:inline">Expert</span>
            </button>
            <button
              onClick={handleSend}
              disabled={!input.trim() || isLoading}
              className="p-2 bg-indigo-600 hover:bg-indigo-500 rounded-xl text-white disabled:opacity-40 transition flex-shrink-0"
            >
              <Send size={15} />
            </button>
          </div>
          <p className="text-[10px] text-text-muted mt-1.5 text-center">MyFi AI có thể mắc sai sót. Hãy xác minh thông tin trước khi đưa ra quyết định đầu tư.</p>
        </div>
      </div>

      {/* ── RIGHT SIDEBAR ── */}
      <div className="hidden xl:flex flex-col w-72 flex-shrink-0 overflow-y-auto custom-scrollbar bg-card-bg">

        {/* Watchlist */}
        <div className="border-b border-border-theme">
          <div className="px-4 py-3 flex items-center justify-between">
            <h3 className="text-xs font-bold text-foreground">Danh mục theo dõi</h3>
            <Eye size={13} className="text-text-muted" />
          </div>
          <div className="divide-y divide-border-theme">
            {WATCHLIST.map(s => (
              <div key={s.symbol} className="flex items-center justify-between px-4 py-2 hover:bg-surface transition cursor-pointer">
                <span className="text-xs font-semibold text-foreground">{s.symbol}</span>
                <div className="text-right">
                  <p className="text-xs font-medium text-foreground tabular-nums">{s.price}</p>
                  <p className={`text-[10px] font-bold ${s.trend > 0 ? "text-positive" : "text-negative"}`}>{s.pct}</p>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Top Breakthroughs */}
        <div>
          <div className="px-4 py-3 flex items-center gap-1.5 border-b border-border-theme">
            <Flame size={12} className="text-orange-400" />
            <h3 className="text-xs font-bold text-foreground">Top vượt đỉnh hôm nay</h3>
          </div>
          <div className="divide-y divide-border-theme">
            {BREAKTHROUGHS.map((b, i) => (
              <div key={b.symbol} className="flex items-center gap-3 px-4 py-2 hover:bg-surface transition cursor-pointer">
                <span className="text-[10px] text-text-muted w-4 text-right">{i+1}</span>
                <span className="text-xs font-bold text-foreground flex-1">{b.symbol}</span>
                <div className="text-right">
                  <p className="text-xs font-bold text-positive">{b.pct}</p>
                  <p className="text-[10px] text-text-muted">{b.vol}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
