"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import { MessageSquare, X, Send, Bot, User, Loader2, TrendingUp, TrendingDown, Minus, Newspaper, Lightbulb, BarChart3, AlertTriangle, ExternalLink } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { ProactiveSuggestions } from "@/features/chat/components/ProactiveSuggestions";

// --- Types ---

interface ChatMessage {
  role: "user" | "assistant";
  content: string;
  structured?: StructuredResponse | null;
  isTyping?: boolean;
  isTimeout?: boolean;
}

interface PriceData {
  symbol: string;
  price: number;
  change: number;
  changePercent: number;
  volume?: number;
  source?: string;
}

interface AnalysisData {
  signal: string;
  confidence: number;
  indicators?: Record<string, string>;
  summary?: string;
  trendAssessment?: string;
  bullishCount?: number;
  bearishCount?: number;
}

interface NewsArticle {
  title: string;
  url?: string;
  source?: string;
  snippet?: string;
}

interface AdviceData {
  recommendation: string;
  reasoning: string;
  action?: string;
  riskAssessment?: string;
}

interface StructuredResponse {
  data?: PriceData[];
  analysis?: AnalysisData;
  news?: NewsArticle[];
  advice?: AdviceData;
  symbols?: string[];
  confidence?: number;
  citations?: CitationLink[];
  tokenUsage?: { used: number; budget: number };
}

interface CitationLink {
  label: string;
  url: string;
}

interface HistoryEntry {
  role: "user" | "assistant";
  content: string;
}

// --- Symbol detection ---

const SYMBOL_REGEX = /\b([A-Z]{3,4})\b/g;

function detectSymbols(text: string): string[] {
  const matches = text.match(SYMBOL_REGEX);
  if (!matches) return [];
  const common = new Set(["THE", "AND", "FOR", "ARE", "BUT", "NOT", "YOU", "ALL", "CAN", "HER", "WAS", "ONE", "OUR", "OUT", "HAS", "HIS", "HOW", "ITS", "MAY", "NEW", "NOW", "OLD", "SEE", "WAY", "WHO", "DID", "GET", "HIM", "LET", "SAY", "SHE", "TOO", "USE"]);
  return [...new Set(matches.filter(m => !common.has(m)))];
}

function highlightSymbols(text: string): React.ReactNode[] {
  const parts = text.split(SYMBOL_REGEX);
  const result: React.ReactNode[] = [];
  for (let i = 0; i < parts.length; i++) {
    const part = parts[i];
    if (i % 2 === 1 && /^[A-Z]{3,4}$/.test(part)) {
      result.push(
        <span key={i} className="bg-indigo-500/20 text-indigo-300 px-1 rounded font-mono text-xs font-semibold">
          {part}
        </span>
      );
    } else {
      result.push(part);
    }
  }
  return result;
}

// --- Structured section renderers ---

function PriceCards({ data }: { data: PriceData[] }) {
  return (
    <div className="space-y-1.5 mt-2">
      <div className="flex items-center gap-1.5 text-xs text-zinc-400 font-medium uppercase tracking-wide">
        <BarChart3 size={12} /> Market Data
      </div>
      {data.map((d, i) => {
        const isUp = d.change > 0;
        const isDown = d.change < 0;
        return (
          <div key={i} className="flex items-center justify-between bg-zinc-900/60 border border-zinc-700/40 rounded-lg px-3 py-2">
            <span className="font-mono text-sm font-semibold text-white">{d.symbol}</span>
            <div className="text-right">
              <div className="text-sm font-medium text-white">{d.price?.toLocaleString()}</div>
              <div className={`text-xs flex items-center gap-0.5 ${isUp ? "text-green-400" : isDown ? "text-red-400" : "text-zinc-400"}`}>
                {isUp ? <TrendingUp size={10} /> : isDown ? <TrendingDown size={10} /> : <Minus size={10} />}
                {d.change > 0 ? "+" : ""}{d.change?.toFixed(2)} ({d.changePercent > 0 ? "+" : ""}{d.changePercent?.toFixed(2)}%)
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}

function AnalysisSection({ analysis }: { analysis: AnalysisData }) {
  const signalColor = analysis.signal?.toLowerCase().includes("bullish")
    ? "text-green-400"
    : analysis.signal?.toLowerCase().includes("bearish")
      ? "text-red-400"
      : "text-yellow-400";

  return (
    <div className="space-y-1.5 mt-2">
      <div className="flex items-center gap-1.5 text-xs text-zinc-400 font-medium uppercase tracking-wide">
        <TrendingUp size={12} /> Analysis
      </div>
      <div className="bg-zinc-900/60 border border-zinc-700/40 rounded-lg px-3 py-2 space-y-1.5">
        <div className="flex items-center justify-between">
          <span className={`text-sm font-semibold ${signalColor}`}>{analysis.signal || "N/A"}</span>
          {analysis.confidence != null && (
            <span className="text-xs bg-zinc-800 px-2 py-0.5 rounded-full text-zinc-300">
              Confidence: {analysis.confidence}%
            </span>
          )}
        </div>
        {analysis.summary && <p className="text-xs text-zinc-300 leading-relaxed">{analysis.summary}</p>}
        {analysis.trendAssessment && <p className="text-xs text-zinc-400">{analysis.trendAssessment}</p>}
        {(analysis.bullishCount != null || analysis.bearishCount != null) && (
          <div className="flex gap-3 text-xs">
            {analysis.bullishCount != null && <span className="text-green-400">▲ {analysis.bullishCount} bullish</span>}
            {analysis.bearishCount != null && <span className="text-red-400">▼ {analysis.bearishCount} bearish</span>}
          </div>
        )}
      </div>
    </div>
  );
}

function NewsSection({ articles }: { articles: NewsArticle[] }) {
  return (
    <div className="space-y-1.5 mt-2">
      <div className="flex items-center gap-1.5 text-xs text-zinc-400 font-medium uppercase tracking-wide">
        <Newspaper size={12} /> News
      </div>
      <div className="space-y-1">
        {articles.slice(0, 5).map((a, i) => (
          <div key={i} className="bg-zinc-900/60 border border-zinc-700/40 rounded-lg px-3 py-1.5">
            {a.url ? (
              <a href={a.url} target="_blank" rel="noopener noreferrer" className="text-xs text-indigo-300 hover:text-indigo-200 hover:underline leading-snug block">
                {a.title}
              </a>
            ) : (
              <span className="text-xs text-zinc-300 leading-snug block">{a.title}</span>
            )}
            {a.source && <span className="text-[10px] text-zinc-500">{a.source}</span>}
          </div>
        ))}
      </div>
    </div>
  );
}

function AdviceSection({ advice }: { advice: AdviceData }) {
  return (
    <div className="space-y-1.5 mt-2">
      <div className="flex items-center gap-1.5 text-xs text-zinc-400 font-medium uppercase tracking-wide">
        <Lightbulb size={12} /> Advice
      </div>
      <div className="bg-zinc-900/60 border border-zinc-700/40 rounded-lg px-3 py-2 space-y-1">
        {advice.action && (
          <span className={`text-xs font-semibold px-2 py-0.5 rounded-full ${advice.action === "buy" ? "bg-green-500/20 text-green-400" : advice.action === "sell" ? "bg-red-500/20 text-red-400" : "bg-yellow-500/20 text-yellow-400"}`}>
            {advice.action.toUpperCase()}
          </span>
        )}
        <p className="text-sm text-zinc-200 leading-relaxed">{advice.recommendation}</p>
        {advice.reasoning && <p className="text-xs text-zinc-400 leading-relaxed">{advice.reasoning}</p>}
        {advice.riskAssessment && (
          <p className="text-xs text-zinc-500">Risk: {advice.riskAssessment}</p>
        )}
      </div>
    </div>
  );
}

// --- Typing indicator ---

function TypingIndicator() {
  return (
    <div className="flex gap-1 items-center px-1 py-0.5">
      {[0, 1, 2].map(i => (
        <motion.div
          key={i}
          className="w-1.5 h-1.5 bg-indigo-400 rounded-full"
          animate={{ opacity: [0.3, 1, 0.3] }}
          transition={{ duration: 1, repeat: Infinity, delay: i * 0.2 }}
        />
      ))}
    </div>
  );
}

// --- Citation links ---

function CitationsSection({ citations }: { citations: CitationLink[] }) {
  return (
    <div className="space-y-1 mt-2">
      <div className="flex items-center gap-1.5 text-xs text-zinc-400 font-medium uppercase tracking-wide">
        <ExternalLink size={12} /> Sources
      </div>
      <div className="flex flex-wrap gap-1.5">
        {citations.map((c, i) => (
          <a
            key={i}
            href={c.url}
            target="_blank"
            rel="noopener noreferrer"
            className="text-[11px] text-indigo-300 hover:text-indigo-200 bg-indigo-500/10 hover:bg-indigo-500/20 px-2 py-0.5 rounded-full transition inline-flex items-center gap-1"
          >
            <ExternalLink size={9} />
            {c.label}
          </a>
        ))}
      </div>
    </div>
  );
}

// --- Token budget warning ---

function TokenBudgetWarning({ used, budget }: { used: number; budget: number }) {
  const pct = budget > 0 ? (used / budget) * 100 : 0;
  if (pct < 80) return null;
  return (
    <div className="flex items-center gap-1.5 text-xs text-yellow-400 bg-yellow-500/10 border border-yellow-500/20 rounded-lg px-2.5 py-1.5 mx-3 mb-1">
      <AlertTriangle size={12} />
      <span>Token usage: {Math.round(pct)}% — approaching limit</span>
    </div>
  );
}


// --- Parse backend response ---

function parseResponse(data: Record<string, unknown>): { content: string; structured: StructuredResponse | null } {
  // If the backend returns structured sections, parse them
  if (data.data || data.analysis || data.news || data.advice) {
    const structured: StructuredResponse = {};

    if (data.data && Array.isArray(data.data)) {
      structured.data = data.data as PriceData[];
    }
    if (data.analysis && typeof data.analysis === "object") {
      structured.analysis = data.analysis as AnalysisData;
    }
    if (data.news && Array.isArray(data.news)) {
      structured.news = data.news as NewsArticle[];
    }
    if (data.advice && typeof data.advice === "object") {
      structured.advice = data.advice as AdviceData;
    }
    if (data.symbols && Array.isArray(data.symbols)) {
      structured.symbols = data.symbols as string[];
    }
    if (typeof data.confidence === "number") {
      structured.confidence = data.confidence;
    }
    if (data.citations && Array.isArray(data.citations)) {
      structured.citations = data.citations as CitationLink[];
    }
    if (data.tokenUsage && typeof data.tokenUsage === "object") {
      structured.tokenUsage = data.tokenUsage as { used: number; budget: number };
    }

    const summary = typeof data.reply === "string" ? data.reply : typeof data.summary === "string" ? data.summary : "";
    return { content: summary, structured };
  }

  // Fallback: plain reply
  const content = typeof data.reply === "string" ? data.reply : typeof data.error === "string" ? data.error : "No response.";
  return { content, structured: null };
}

// --- Main component ---

const MAX_HISTORY = 10;
const TIMEOUT_MS = 45_000;

export function ChatWidget() {
  const [isOpen, setIsOpen] = useState(false);
  const [messages, setMessages] = useState<ChatMessage[]>([
    { role: "assistant", content: "Hi! I'm your AI financial advisor. How can I help you analyze the market today?" },
  ]);
  const [input, setInput] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [tokenUsage, setTokenUsage] = useState<{ used: number; budget: number } | null>(null);
  const chatEndRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Auto-scroll to latest message
  const scrollToBottom = useCallback(() => {
    chatEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, []);

  useEffect(() => {
    scrollToBottom();
  }, [messages, scrollToBottom]);

  // Build conversation history (last 10 messages)
  const buildHistory = useCallback((): HistoryEntry[] => {
    return messages
      .filter(m => !m.isTyping)
      .slice(-MAX_HISTORY)
      .map(m => ({ role: m.role, content: m.content }));
  }, [messages]);

  const handleSend = async () => {
    const trimmed = input.trim();
    if (!trimmed || isLoading) return;

    const userMessage = trimmed;
    setInput("");
    setIsLoading(true);

    // Add user message
    setMessages(prev => [...prev, { role: "user", content: userMessage }]);

    // Add typing indicator
    setMessages(prev => [...prev, { role: "assistant", content: "", isTyping: true }]);

    // Detect symbols
    const symbols = detectSymbols(userMessage);
    const primarySymbol = symbols[0] || "";

    // Read AI config from localStorage
    let provider = "bedrock";
    let model = "anthropic.claude-3-sonnet-20240229-v1:0";
    let apiKey = "";
    let awsAccessKey = "";
    let awsSecretKey = "";
    let awsRegion = "";

    try {
      const saved = localStorage.getItem("myfi_ai_config");
      if (saved) {
        const parsed = JSON.parse(saved);
        if (parsed.provider) provider = parsed.provider;
        if (parsed.model) model = parsed.model;
        if (parsed.apiKey) apiKey = parsed.apiKey;
        if (parsed.awsAccessKey) awsAccessKey = parsed.awsAccessKey;
        if (parsed.awsSecretKey) awsSecretKey = parsed.awsSecretKey;
        if (parsed.awsRegion) awsRegion = parsed.awsRegion;
      }
    } catch {
      // ignore parse errors
    }

    const history = buildHistory();

    // Setup abort controller for 45s timeout
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), TIMEOUT_MS);
    let timedOut = false;

    try {
      const response = await fetch("http://localhost:8080/api/chat", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        signal: controller.signal,
        body: JSON.stringify({
          message: userMessage,
          symbol: primarySymbol,
          history,
          provider,
          model,
          apiKey,
          awsAccessKey,
          awsSecretKey,
          awsRegion,
        }),
      });

      clearTimeout(timeoutId);
      const data = await response.json();
      const { content, structured } = parseResponse(data);

      if (structured?.tokenUsage) {
        setTokenUsage(structured.tokenUsage);
      }

      setMessages(prev => {
        const updated = [...prev];
        // Replace typing indicator
        const typingIdx = updated.findLastIndex(m => m.isTyping);
        if (typingIdx >= 0) {
          updated[typingIdx] = { role: "assistant", content, structured };
        } else {
          updated.push({ role: "assistant", content, structured });
        }
        return updated;
      });
    } catch (err) {
      clearTimeout(timeoutId);

      if (err instanceof DOMException && err.name === "AbortError") {
        timedOut = true;
      }

      setMessages(prev => {
        const updated = [...prev];
        const typingIdx = updated.findLastIndex(m => m.isTyping);
        const errorMsg: ChatMessage = timedOut
          ? { role: "assistant", content: "Taking longer than expected... The AI pipeline timed out after 45 seconds. Please try a simpler query or try again later.", isTimeout: true }
          : { role: "assistant", content: "Sorry, I couldn't reach the backend server. Please check your connection." };

        if (typingIdx >= 0) {
          updated[typingIdx] = errorMsg;
        } else {
          updated.push(errorMsg);
        }
        return updated;
      });
    } finally {
      setIsLoading(false);
    }
  };

  // Handle Enter to send, Shift+Enter for newline
  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  // Auto-resize textarea
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
      textareaRef.current.style.height = Math.min(textareaRef.current.scrollHeight, 96) + "px";
    }
  }, [input]);

  return (
    <>
      {/* Floating toggle button */}
      <button
        onClick={() => setIsOpen(true)}
        className="fixed bottom-6 right-6 w-14 h-14 bg-gradient-to-r from-indigo-500 to-purple-600 rounded-full shadow-xl shadow-indigo-500/20 flex items-center justify-center text-white hover:scale-105 transition-transform z-50"
        aria-label="Open chat"
      >
        <MessageSquare size={24} />
      </button>

      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ opacity: 0, y: 20, scale: 0.95 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: 20, scale: 0.95 }}
            transition={{ duration: 0.2 }}
            className="fixed bottom-24 right-6 w-96 h-[32rem] bg-zinc-900 border border-zinc-700/50 rounded-2xl shadow-2xl flex flex-col z-50 overflow-hidden backdrop-blur-xl"
          >
            {/* Header */}
            <div className="px-5 py-4 bg-zinc-800/80 border-b border-zinc-700/50 flex justify-between items-center backdrop-blur-sm">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-gradient-to-tr from-indigo-500 to-purple-500 flex items-center justify-center">
                  <Bot size={18} className="text-white" />
                </div>
                <div>
                  <h3 className="font-semibold text-white">EziStock AI Advisor</h3>
                  <p className="text-xs text-green-400">Online</p>
                </div>
              </div>
              <button onClick={() => setIsOpen(false)} className="text-zinc-400 hover:text-white transition p-1 rounded-md hover:bg-zinc-700" aria-label="Close chat">
                <X size={20} />
              </button>
            </div>

            {/* Chat messages area */}
            <div className="flex-1 overflow-y-auto p-4 space-y-3">
              {/* Proactive suggestions when no user messages yet */}
              {messages.length <= 1 && !isLoading && (
                <ProactiveSuggestions onSelect={(text) => { setInput(text); }} />
              )}
              <AnimatePresence initial={false}>
                {messages.map((msg, i) => (
                  <motion.div
                    key={i}
                    initial={{ opacity: 0, y: 8 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.2 }}
                    className={`flex gap-2.5 ${msg.role === "user" ? "flex-row-reverse" : ""}`}
                  >
                    <div className={`w-7 h-7 rounded-full flex-shrink-0 flex items-center justify-center mt-0.5 ${msg.role === "user" ? "bg-zinc-800" : "bg-indigo-500"}`}>
                      {msg.role === "user" ? <User size={14} className="text-zinc-400" /> : <Bot size={14} className="text-white" />}
                    </div>
                    <div className={`max-w-[80%] ${msg.role === "user" ? "text-right" : ""}`}>
                      {msg.isTyping ? (
                        <div className="bg-zinc-800 rounded-2xl rounded-tl-sm px-3 py-2.5 border border-zinc-700/50">
                          <TypingIndicator />
                        </div>
                      ) : (
                        <>
                          <div className={`p-3 rounded-2xl text-sm leading-relaxed ${msg.role === "user" ? "bg-indigo-600 text-white rounded-tr-sm" : "bg-zinc-800 text-zinc-200 rounded-tl-sm border border-zinc-700/50"} ${msg.isTimeout ? "border-yellow-500/30" : ""}`}>
                            {msg.role === "user" ? highlightSymbols(msg.content) : msg.content}
                          </div>
                          {/* Structured response sections */}
                          {msg.structured && (
                            <div className="mt-1.5 space-y-1">
                              {msg.structured.data && msg.structured.data.length > 0 && (
                                <PriceCards data={msg.structured.data} />
                              )}
                              {msg.structured.analysis && (
                                <AnalysisSection analysis={msg.structured.analysis} />
                              )}
                              {msg.structured.news && msg.structured.news.length > 0 && (
                                <NewsSection articles={msg.structured.news} />
                              )}
                              {msg.structured.advice && (
                                <AdviceSection advice={msg.structured.advice} />
                              )}
                              {msg.structured.citations && msg.structured.citations.length > 0 && (
                                <CitationsSection citations={msg.structured.citations} />
                              )}
                            </div>
                          )}
                        </>
                      )}
                    </div>
                  </motion.div>
                ))}
              </AnimatePresence>
              <div ref={chatEndRef} />
            </div>

            {/* Token budget warning */}
            {tokenUsage && <TokenBudgetWarning used={tokenUsage.used} budget={tokenUsage.budget} />}

            {/* Input area */}
            <div className="p-3 bg-zinc-800/50 border-t border-zinc-700/50">
              <div className="flex items-end gap-2">
                <textarea
                  ref={textareaRef}
                  value={input}
                  onChange={e => setInput(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder="Ask about markets, stocks..."
                  rows={1}
                  className="flex-1 bg-zinc-900 border border-zinc-700 rounded-xl py-2.5 px-4 text-sm text-white focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all placeholder-zinc-500 resize-none max-h-24 scrollbar-thin"
                />
                <button
                  onClick={handleSend}
                  disabled={!input.trim() || isLoading}
                  className="w-10 h-10 bg-indigo-600 rounded-xl flex items-center justify-center text-white disabled:opacity-50 disabled:bg-zinc-700 transition flex-shrink-0"
                  aria-label="Send message"
                >
                  {isLoading ? <Loader2 size={18} className="animate-spin" /> : <Send size={18} />}
                </button>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  );
}
