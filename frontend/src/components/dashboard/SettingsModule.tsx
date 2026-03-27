
"use client";

import { useState, useEffect } from "react";
import {
  Bot, Key, Cloud, Cpu, ChevronRight, Info, CheckCircle,
  AlertCircle, RefreshCw, Save, Zap, Globe, Eye, EyeOff
} from "lucide-react";

export interface AIConfig {
  provider: "openai" | "bedrock" | "anthropic" | "google" | "qwen";
  model: string;
  apiKey: string;
  awsAccessKey?: string;
  awsSecretKey?: string;
  awsRegion?: string;
}

const DEFAULT_CONFIG: AIConfig = {
  provider: "bedrock",
  model: "anthropic.claude-3-sonnet-20240229-v1:0",
  apiKey: "",
  awsAccessKey: "",
  awsSecretKey: "",
  awsRegion: "us-east-1"
};

const PROVIDERS = [
  {
    id: "bedrock",
    name: "AWS Bedrock",
    icon: Cloud,
    color: "from-orange-500 to-amber-500",
    desc: "Run models from Anthropic, Meta, Amazon via AWS cloud.",
    badge: "Enterprise"
  },
  {
    id: "openai",
    name: "OpenAI",
    icon: Zap,
    color: "from-green-500 to-emerald-500",
    desc: "Access GPT-4o, GPT-4 Turbo and other OpenAI models.",
    badge: "Popular"
  },
  {
    id: "anthropic",
    name: "Anthropic",
    icon: Bot,
    color: "from-violet-500 to-purple-500",
    desc: "Claude 3 family: Sonnet, Haiku, Opus.",
    badge: null
  },
  {
    id: "google",
    name: "Google Gemini",
    icon: Globe,
    color: "from-blue-500 to-cyan-500",
    desc: "Gemini 1.5 Pro and Flash — strong multimodal capability.",
    badge: null
  },
  {
    id: "qwen",
    name: "Qwen (Alibaba)",
    icon: Cpu,
    color: "from-rose-500 to-pink-500",
    desc: "Qwen-Turbo, Plus and Max via DashScope API.",
    badge: "New"
  }
];

function formatModelName(id: string): string {
  const names: Record<string, string> = {
    "amazon.nova-lite-v2:0": "Nova 2 Lite",
    "amazon.nova-premier-v1:0": "Nova Premier",
    "amazon.nova-pro-v1:0": "Nova Pro",
    "amazon.nova-lite-v1:0": "Nova Lite",
    "amazon.nova-micro-v1:0": "Nova Micro",
    "amazon.titan-text-premier-v1:0": "Titan Text Premier",
    "amazon.titan-text-express-v1": "Titan Text Express",
    "amazon.titan-text-lite-v1": "Titan Text Lite",
    // Claude 4.x — Verified IDs only
    "anthropic.claude-opus-4-6-v1": "Claude Opus 4.6",
    "anthropic.claude-sonnet-4-6": "Claude Sonnet 4.6",
    "anthropic.claude-haiku-4-5-20251001-v1:0": "Claude Haiku 4.5",
    // Claude 3.7
    "anthropic.claude-3-7-sonnet-20250219-v1:0": "Claude 3.7 Sonnet",
    // Claude 3.5
    "anthropic.claude-3-5-sonnet-20240620-v1:0": "Claude 3.5 Sonnet",
    "anthropic.claude-3-5-haiku-20241022-v1:0": "Claude 3.5 Haiku",
    "anthropic.claude-3-opus-20240229-v1:0": "Claude 3 Opus",
    "anthropic.claude-3-sonnet-20240229-v1:0": "Claude 3 Sonnet",
    "anthropic.claude-3-haiku-20240307-v1:0": "Claude 3 Haiku",
    "anthropic.claude-v2:1": "Claude 2.1",
    "anthropic.claude-instant-v1": "Claude Instant",
    "meta.llama4-maverick-17b-instruct-v1:0": "Llama 4 Maverick 17B",
    "meta.llama4-scout-17b-instruct-v1:0": "Llama 4 Scout 17B",
    "meta.llama3-3-70b-instruct-v1:0": "Llama 3.3 70B",
    "meta.llama3-2-90b-instruct-v2:0": "Llama 3.2 90B",
    "meta.llama3-2-11b-instruct-v2:0": "Llama 3.2 11B",
    "meta.llama3-2-3b-instruct-v2:0": "Llama 3.2 3B",
    "meta.llama3-2-1b-instruct-v2:0": "Llama 3.2 1B",
    "meta.llama3-1-405b-instruct-v1:0": "Llama 3.1 405B",
    "meta.llama3-1-70b-instruct-v1:0": "Llama 3.1 70B",
    "meta.llama3-1-8b-instruct-v1:0": "Llama 3.1 8B",
    "meta.llama3-70b-instruct-v1:0": "Llama 3 70B",
    "meta.llama3-8b-instruct-v1:0": "Llama 3 8B",
    // DeepSeek
    "deepseek.r1-v1:0": "DeepSeek R1",
    "deepseek.deepseek-v3-1-v1:0": "DeepSeek V3.1",
    // Mistral
    "mistral.mistral-large-2402-v1:0": "Mistral Large (Feb '24)",
    "mistral.mistral-large-2407-v1:0": "Mistral Large (Jul '24)",
    "mistral.mistral-small-2402-v1:0": "Mistral Small",
    "mistral.mixtral-8x7b-instruct-v0:1": "Mixtral 8x7B",
    "mistral.mistral-7b-instruct-v0:2": "Mistral 7B",
    "cohere.command-r-plus-v1:0": "Command R+",
    "cohere.command-r-v1:0": "Command R",
    "cohere.command-text-v14": "Command Text",
    "cohere.command-light-text-v14": "Command Light",
    "ai21.jamba-1-5-large-v1:0": "Jamba 1.5 Large",
    "ai21.jamba-1-5-mini-v1:0": "Jamba 1.5 Mini",
    "ai21.jamba-instruct-v1:0": "Jamba Instruct",
    "gpt-4o": "GPT-4o",
    "gpt-4-turbo": "GPT-4 Turbo",
    "gpt-3.5-turbo": "GPT-3.5 Turbo",
    "gemini-1.5-pro": "Gemini 1.5 Pro",
    "gemini-1.5-flash": "Gemini 1.5 Flash",
    "qwen-turbo": "Qwen Turbo",
    "qwen-plus": "Qwen Plus",
    "qwen-max": "Qwen Max",
  };
  return names[id] ?? id;
}

const BEDROCK_GROUPS: { label: string; prefix: string; extra?: string }[] = [
  { label: "Amazon Nova",          prefix: "amazon.nova" },
  { label: "Amazon Titan",         prefix: "amazon.titan" },
  { label: "Anthropic Claude 4.x", prefix: "anthropic.claude-opus-4", extra: "anthropic.claude-sonnet-4" },
  { label: "Anthropic Claude 3.7", prefix: "anthropic.claude-3-7" },
  { label: "Anthropic Claude 3.5", prefix: "anthropic.claude-3-5" },
  { label: "Anthropic Claude 3",   prefix: "anthropic.claude-3-" },
  { label: "Anthropic Claude 2",   prefix: "anthropic.claude-v2", extra: "anthropic.claude-instant" },
  { label: "DeepSeek",             prefix: "deepseek." },
  { label: "Meta Llama 4",         prefix: "meta.llama4" },
  { label: "Meta Llama 3.3",       prefix: "meta.llama3-3" },
  { label: "Meta Llama 3.2",       prefix: "meta.llama3-2" },
  { label: "Meta Llama 3.1",       prefix: "meta.llama3-1" },
  { label: "Meta Llama 3",         prefix: "meta.llama3-" },
  { label: "Mistral",              prefix: "mistral." },
  { label: "Cohere",               prefix: "cohere." },
  { label: "AI21 Jamba",           prefix: "ai21." },
];

export function SettingsModule() {
  const [config, setConfig] = useState<AIConfig>(DEFAULT_CONFIG);
  const [availableModels, setAvailableModels] = useState<string[]>([]);
  const [isFetching, setIsFetching] = useState(false);
  const [fetchStatus, setFetchStatus] = useState<"idle" | "success" | "error">("idle");
  const [saved, setSaved] = useState(false);
  const [showApiKey, setShowApiKey] = useState(false);
  const [showSecretKey, setShowSecretKey] = useState(false);

  useEffect(() => {
    const saved = localStorage.getItem("myfi_ai_config");
    if (saved) {
      try {
        const parsed = JSON.parse(saved);
        if (parsed.provider && parsed.model) {
          setConfig(parsed);
          fetchModels(parsed.provider, parsed.apiKey, parsed.awsAccessKey, parsed.awsSecretKey, parsed.awsRegion);
        }
      } catch (e) {
        console.error("Failed to parse AI config from localStorage");
      }
    }
  }, []);

  const fetchModels = async (provider: string, apiKey: string, access?: string, secret?: string, region?: string) => {
    setIsFetching(true);
    setFetchStatus("idle");
    try {
      const payload: any = { provider, apiKey };
      if (provider === "bedrock") {
        payload.awsAccessKey = access;
        payload.awsSecretKey = secret;
        payload.awsRegion = region;
      }

      const res = await fetch("http://localhost:8080/api/models", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });
      const data = await res.json();
      if (data.models && data.models.length > 0) {
        setAvailableModels(data.models);
        setFetchStatus("success");
      } else {
        setAvailableModels([]);
        setFetchStatus("error");
      }
    } catch (err) {
      console.error("Error fetching models", err);
      setAvailableModels([]);
      setFetchStatus("error");
    } finally {
      setIsFetching(false);
    }
  };

  const handleSave = () => {
    localStorage.setItem("myfi_ai_config", JSON.stringify(config));
    setSaved(true);
    setTimeout(() => setSaved(false), 2500);
  };

  const handleProviderChange = (providerId: string) => {
    const newConfig = { ...config, provider: providerId as AIConfig["provider"], model: "" };
    setConfig(newConfig);
    setAvailableModels([]);
    setFetchStatus("idle");
    fetchModels(providerId, config.apiKey, config.awsAccessKey, config.awsSecretKey, config.awsRegion);
  };

  const selectedProvider = PROVIDERS.find(p => p.id === config.provider);
  const needsApiKey = ["openai", "anthropic", "google", "qwen"].includes(config.provider);

  return (
    <div className="w-full max-w-5xl mx-auto px-4 py-8">
      {/* Page Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-white mb-1 flex items-center gap-3">
          <Bot className="text-indigo-400" size={32} /> AI Model Configuration
        </h1>
        <p className="text-zinc-400 text-sm">
          Select your preferred AI provider and model. Settings are stored locally in your browser.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* LEFT: Provider Selection */}
        <div className="lg:col-span-1">
          <h2 className="text-xs uppercase font-semibold text-zinc-500 tracking-wider mb-3">AI Provider</h2>
          <div className="flex flex-col gap-2">
            {PROVIDERS.map(p => {
              const Icon = p.icon;
              const isActive = config.provider === p.id;
              return (
                <button
                  key={p.id}
                  onClick={() => handleProviderChange(p.id)}
                  className={`flex items-center gap-3 px-4 py-3 rounded-xl border transition-all duration-200 text-left w-full group ${
                    isActive
                      ? "border-indigo-500 bg-indigo-500/10"
                      : "border-zinc-800 bg-zinc-900/60 hover:border-zinc-600 hover:bg-zinc-800/60"
                  }`}
                >
                  <div className={`w-8 h-8 rounded-lg bg-gradient-to-br ${p.color} flex items-center justify-center flex-shrink-0`}>
                    <Icon size={16} className="text-white" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className={`text-sm font-semibold ${isActive ? "text-white" : "text-zinc-300 group-hover:text-white"}`}>
                        {p.name}
                      </span>
                      {p.badge && (
                        <span className="text-[10px] px-1.5 py-0.5 rounded bg-indigo-500/20 text-indigo-300 font-bold">
                          {p.badge}
                        </span>
                      )}
                    </div>
                    <p className="text-[11px] text-zinc-500 truncate">{p.desc}</p>
                  </div>
                  {isActive && <ChevronRight size={16} className="text-indigo-400 flex-shrink-0" />}
                </button>
              );
            })}
          </div>
        </div>

        {/* RIGHT: Configuration */}
        <div className="lg:col-span-2 flex flex-col gap-5">

          {/* Provider Banner */}
          {selectedProvider && (
            <div className={`rounded-xl bg-gradient-to-r ${selectedProvider.color} p-px`}>
              <div className="bg-zinc-900 rounded-xl p-5 flex items-start gap-4">
                <div className={`w-12 h-12 rounded-xl bg-gradient-to-br ${selectedProvider.color} flex items-center justify-center flex-shrink-0 shadow-lg`}>
                  <selectedProvider.icon size={22} className="text-white" />
                </div>
                <div>
                  <h3 className="text-lg font-bold text-white">{selectedProvider.name}</h3>
                  <p className="text-sm text-zinc-400 mt-1">{selectedProvider.desc}</p>
                </div>
              </div>
            </div>
          )}

          {/* AWS Bedrock Credentials */}
          {config.provider === "bedrock" && (
            <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-4">
              <div className="flex items-center gap-2 mb-1">
                <Key size={16} className="text-orange-400" />
                <h3 className="text-sm font-semibold text-white">AWS Credentials</h3>
              </div>
              <div>
                <label className="text-xs text-zinc-400 uppercase tracking-wider font-semibold block mb-1.5">Access Key ID</label>
                <input
                  type="text"
                  value={config.awsAccessKey || ""}
                  onChange={(e) => setConfig({ ...config, awsAccessKey: e.target.value })}
                  placeholder="AKIAIOSFODNN7EXAMPLE"
                  className="w-full bg-black border border-zinc-700 text-white text-sm rounded-lg focus:ring-2 focus:ring-orange-500/50 focus:border-orange-500 outline-none p-2.5 font-mono placeholder-zinc-600"
                />
              </div>
              <div>
                <label className="text-xs text-zinc-400 uppercase tracking-wider font-semibold block mb-1.5">Secret Access Key</label>
                <div className="relative">
                  <input
                    type={showSecretKey ? "text" : "password"}
                    value={config.awsSecretKey || ""}
                    onChange={(e) => setConfig({ ...config, awsSecretKey: e.target.value })}
                    placeholder="wJalrXUtnFEMI..."
                    className="w-full bg-black border border-zinc-700 text-white text-sm rounded-lg focus:ring-2 focus:ring-orange-500/50 focus:border-orange-500 outline-none p-2.5 pr-10 font-mono placeholder-zinc-600"
                  />
                  <button onClick={() => setShowSecretKey(!showSecretKey)} className="absolute right-3 top-2.5 text-zinc-500 hover:text-white transition">
                    {showSecretKey ? <EyeOff size={16} /> : <Eye size={16} />}
                  </button>
                </div>
              </div>
              <div>
                <label className="text-xs text-zinc-400 uppercase tracking-wider font-semibold block mb-1.5">Region</label>
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={config.awsRegion || ""}
                    onChange={(e) => setConfig({ ...config, awsRegion: e.target.value })}
                    placeholder="us-east-1"
                    className="flex-1 bg-black border border-zinc-700 text-white text-sm rounded-lg focus:ring-2 focus:ring-orange-500/50 focus:border-orange-500 outline-none p-2.5 font-mono placeholder-zinc-600"
                  />
                  <button
                    onClick={() => fetchModels(config.provider, config.apiKey, config.awsAccessKey, config.awsSecretKey, config.awsRegion)}
                    disabled={isFetching}
                    className="flex items-center gap-2 px-4 py-2 bg-orange-500 hover:bg-orange-400 disabled:opacity-60 text-white text-sm font-semibold rounded-lg transition shadow-md"
                  >
                    <RefreshCw size={14} className={isFetching ? "animate-spin" : ""} />
                    {isFetching ? "Loading..." : "Fetch Models"}
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* API Key for other providers */}
          {needsApiKey && (
            <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-4">
              <div className="flex items-center gap-2">
                <Key size={16} className="text-indigo-400" />
                <h3 className="text-sm font-semibold text-white">API Key</h3>
              </div>
              <div>
                <label className="text-xs text-zinc-400 uppercase tracking-wider font-semibold block mb-1.5">Secret Key</label>
                <div className="relative">
                  <input
                    type={showApiKey ? "text" : "password"}
                    value={config.apiKey || ""}
                    onChange={(e) => setConfig({ ...config, apiKey: e.target.value })}
                    placeholder={`Enter your ${selectedProvider?.name} API Key...`}
                    className="w-full bg-black border border-zinc-700 text-white text-sm rounded-lg focus:ring-2 focus:ring-indigo-500/50 focus:border-indigo-500 outline-none p-2.5 pr-10 font-mono placeholder-zinc-600"
                  />
                  <button onClick={() => setShowApiKey(!showApiKey)} className="absolute right-3 top-2.5 text-zinc-500 hover:text-white transition">
                    {showApiKey ? <EyeOff size={16} /> : <Eye size={16} />}
                  </button>
                </div>
                <p className="text-xs text-zinc-500 mt-1.5 flex items-center gap-1.5">
                  <Info size={11} /> Stored locally in your browser. Never sent anywhere except the backend proxy.
                </p>
              </div>
              <button
                onClick={() => fetchModels(config.provider, config.apiKey)}
                disabled={isFetching}
                className="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-60 text-white text-sm font-semibold rounded-lg transition shadow-md"
              >
                <RefreshCw size={14} className={isFetching ? "animate-spin" : ""} />
                {isFetching ? "Fetching Models..." : "Fetch Available Models"}
              </button>
            </div>
          )}

          {/* Model Selection */}
          <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-2">
                <Cpu size={16} className="text-indigo-400" />
                <h3 className="text-sm font-semibold text-white">Model Selection</h3>
              </div>
              {fetchStatus === "success" && (
                <span className="text-xs flex items-center gap-1 text-green-400">
                  <CheckCircle size={12} /> {availableModels.length} models
                </span>
              )}
              {fetchStatus === "error" && (
                <span className="text-xs flex items-center gap-1 text-red-400">
                  <AlertCircle size={12} /> Failed to load
                </span>
              )}
            </div>

            {availableModels.length > 0 ? (
              config.provider === "bedrock" ? (
                // Grouped optgroup dropdown for Bedrock
                <select
                  value={config.model}
                  onChange={(e) => setConfig({ ...config, model: e.target.value })}
                  className="w-full bg-black border border-zinc-700 text-white text-sm rounded-lg focus:ring-2 focus:ring-indigo-500/50 focus:border-indigo-500 outline-none p-3"
                >
                  <option value="">— Select a model —</option>
                  {BEDROCK_GROUPS.map(group => {
                    const groupModels = availableModels.filter(m =>
                      m.startsWith(group.prefix) ||
                      (group.extra ? m.startsWith(group.extra) : false)
                    );
                    if (groupModels.length === 0) return null;
                    return (
                      <optgroup key={group.label} label={group.label}>
                        {groupModels.map(m => (
                          <option key={m} value={m}>{formatModelName(m)}</option>
                        ))}
                      </optgroup>
                    );
                  })}
                </select>
              ) : (
                // Simple dropdown for other providers
                <select
                  value={config.model}
                  onChange={(e) => setConfig({ ...config, model: e.target.value })}
                  className="w-full bg-black border border-zinc-700 text-white text-sm rounded-lg focus:ring-2 focus:ring-indigo-500/50 focus:border-indigo-500 outline-none p-3"
                >
                  <option value="">— Select a model —</option>
                  {availableModels.map(m => (
                    <option key={m} value={m}>{formatModelName(m)}</option>
                  ))}
                </select>
              )
            ) : (
              <div className="border border-dashed border-zinc-700 rounded-xl p-6 text-center">
                <Cpu size={28} className="text-zinc-700 mx-auto mb-2" />
                <p className="text-sm text-zinc-500">
                  {isFetching ? "Fetching available models..." : config.provider === "bedrock" ? "Enter AWS credentials and click Fetch Models." : "Enter your API key and click Fetch Available Models."}
                </p>
              </div>
            )}
          </div>

          {/* Save Button */}
          <button
            onClick={handleSave}
            className={`flex items-center justify-center gap-2 w-full py-3.5 rounded-xl text-sm font-bold shadow-lg transition-all duration-200 ${
              saved
                ? "bg-green-600 text-white shadow-green-600/20"
                : "bg-indigo-600 hover:bg-indigo-500 text-white shadow-indigo-600/20"
            }`}
          >
            {saved ? (
              <><CheckCircle size={18} /> Settings Saved!</>
            ) : (
              <><Save size={18} /> Save Configuration</>
            )}
          </button>

          {config.model && (
            <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl px-4 py-3 flex items-center gap-3">
              <CheckCircle size={16} className="text-green-400 flex-shrink-0" />
              <div className="text-sm text-zinc-300">
                Active: <span className="text-white font-semibold">{selectedProvider?.name}</span>
                {" · "}<span className="text-indigo-300">{formatModelName(config.model)}</span>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
