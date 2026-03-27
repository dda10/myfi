"use client";

import { useState, useEffect } from "react";
import { Settings, X } from "lucide-react";

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

export function SettingsModal() {
  const [isOpen, setIsOpen] = useState(false);
  const [config, setConfig] = useState<AIConfig>(DEFAULT_CONFIG);
  const [availableModels, setAvailableModels] = useState<string[]>([]);
  const [isFetching, setIsFetching] = useState(false);

  useEffect(() => {
    // Load existing settings from localStorage
    const saved = localStorage.getItem("myfi_ai_config");
    if (saved) {
      try {
        const parsed = JSON.parse(saved);
        if (parsed.provider && parsed.model) {
          setConfig(parsed);
          fetchModels(parsed.provider, parsed.apiKey);
        }
      } catch (e) {
        console.error("Failed to parse AI config from localStorage");
      }
    }
  }, []);

  const fetchModels = async (provider: string, apiKey: string, access?: string, secret?: string, region?: string) => {
    setIsFetching(true);
    try {
      const payload: any = { provider, apiKey };
      if (provider === 'bedrock') {
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
      } else {
        setAvailableModels([]);
      }
    } catch (err) {
      console.error("Error fetching models", err);
      setAvailableModels([]);
    } finally {
      setIsFetching(false);
    }
  };

  const handleSave = () => {
    localStorage.setItem("myfi_ai_config", JSON.stringify(config));
    setIsOpen(false);
  };

  const handleProviderChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const provider = e.target.value as AIConfig["provider"];
    const newConfig = {
      ...config,
      provider,
      model: "" // Reset model when provider changes to force selection from new list
    };
    setConfig(newConfig);
    fetchModels(provider, config.apiKey, config.awsAccessKey, config.awsSecretKey, config.awsRegion);
  };

  const formatModelName = (id: string) => {
    // Return friendly name without confusing provider prefixes and versions
    if (id.includes('claude-3-sonnet-20240229')) return "Claude 3 Sonnet";
    if (id.includes('claude-3-haiku-20240307')) return "Claude 3 Haiku";
    if (id.includes('claude-3-opus-20240229')) return "Claude 3 Opus";
    if (id.includes('claude-3-5-sonnet')) return "Claude 3.5 Sonnet";
    if (id.includes('llama3-70b')) return "Llama 3 70B";
    if (id.includes('llama3-8b')) return "Llama 3 8B";
    if (id.includes('titan')) return "Amazon Titan";
    if (id.includes('gpt-4o')) return "GPT-4o";
    if (id.includes('gpt-4-turbo')) return "GPT-4 Turbo";
    if (id.includes('gemini-1.5-pro')) return "Gemini 1.5 Pro";
    if (id.includes('gemini-1.5-flash')) return "Gemini 1.5 Flash";
    if (id.includes('qwen-turbo')) return "Qwen Turbo";
    if (id.includes('qwen-plus')) return "Qwen Plus";
    if (id.includes('qwen-max')) return "Qwen Max";
    
    // Default fallback cleaning
    return id.replace('anthropic.', '').replace('meta.', '').replace('-v1:0', '');
  };

  return (
    <>
      <button 
        onClick={() => setIsOpen(true)}
        className="p-2 text-zinc-400 hover:text-white transition rounded-full hover:bg-zinc-800"
      >
        <Settings size={20} />
      </button>

      {isOpen && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/60 backdrop-blur-sm">
          <div className="w-full max-w-md bg-zinc-900 border border-zinc-800 rounded-xl shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
            <div className="px-6 py-4 border-b border-zinc-800 flex justify-between items-center bg-zinc-800/50">
              <h2 className="text-lg font-semibold text-white">AI Configuration</h2>
              <button onClick={() => setIsOpen(false)} className="text-zinc-400 hover:text-white transition p-1 rounded hover:bg-zinc-700">
                <X size={20} />
              </button>
            </div>
            
            <div className="p-6 space-y-5">
              <div className="space-y-2">
                <label className="text-sm font-medium text-zinc-300">AI Provider</label>
                <select 
                  value={config.provider}
                  onChange={handleProviderChange}
                  className="w-full bg-zinc-950 border border-zinc-700 text-white text-sm rounded-lg focus:ring-indigo-500 focus:border-indigo-500 block p-2.5"
                >
                  <option value="bedrock">AWS Bedrock</option>
                  <option value="openai">OpenAI</option>
                  <option value="anthropic">Anthropic</option>
                  <option value="google">Google Gemini</option>
                  <option value="qwen">Qwen (Alibaba)</option>
                </select>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-zinc-300">Model ID</label>
                <select 
                  value={config.model}
                  onChange={(e) => setConfig({ ...config, model: e.target.value })}
                  className="w-full bg-zinc-950 border border-zinc-700 text-white text-sm rounded-lg focus:ring-indigo-500 focus:border-indigo-500 block p-2.5"
                  disabled={isFetching || availableModels.length === 0}
                >
                  {isFetching ? (
                    <option value="">Fetching models...</option>
                  ) : availableModels.length > 0 ? (
                    availableModels.map(m => (
                      <option key={m} value={m}>{formatModelName(m)}</option>
                    ))
                  ) : (
                    <option value="">No models available. Enter API key.</option>
                  )}
                </select>
              </div>

              {['openai', 'anthropic', 'google', 'qwen'].includes(config.provider) && (
                <div className="space-y-2">
                  <label className="text-sm font-medium text-zinc-300">API Key</label>
                  <div className="flex gap-2">
                    <input 
                      type="password"
                      value={config.apiKey || ""}
                      onChange={(e) => setConfig({ ...config, apiKey: e.target.value })}
                      placeholder="Enter API Key here..."
                      className="w-full bg-zinc-950 border border-zinc-700 text-white text-sm rounded-lg focus:ring-indigo-500 focus:border-indigo-500 block p-2.5"
                    />
                    <button 
                      onClick={() => fetchModels(config.provider, config.apiKey)}
                      className="px-4 py-2 bg-zinc-800 text-white text-sm rounded-lg hover:bg-zinc-700 transition flex-shrink-0 border border-zinc-700"
                    >
                      {isFetching ? "..." : "Fetch"}
                    </button>
                  </div>
                  <p className="text-xs text-zinc-500 mt-1">
                    Your API key is stored locally in your browser and used to fetch your available models dynamically.
                  </p>
                </div>
              )}
              {config.provider === 'bedrock' && (
                <div className="space-y-3 bg-black/20 p-4 rounded-xl border border-zinc-800">
                  <div className="space-y-1">
                    <label className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">AWS Access Key ID</label>
                    <input 
                      type="text"
                      value={config.awsAccessKey || ""}
                      onChange={(e) => setConfig({ ...config, awsAccessKey: e.target.value })}
                      placeholder="AKIAIOSFODNN7EXAMPLE"
                      className="w-full bg-zinc-950 border border-zinc-700 text-white text-sm rounded-lg focus:ring-indigo-500 focus:border-indigo-500 block p-2"
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">AWS Secret Access Key</label>
                    <input 
                      type="password"
                      value={config.awsSecretKey || ""}
                      onChange={(e) => setConfig({ ...config, awsSecretKey: e.target.value })}
                      placeholder="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
                      className="w-full bg-zinc-950 border border-zinc-700 text-white text-sm rounded-lg focus:ring-indigo-500 focus:border-indigo-500 block p-2"
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">AWS Region</label>
                    <div className="flex gap-2">
                      <input 
                        type="text"
                        value={config.awsRegion || ""}
                        onChange={(e) => setConfig({ ...config, awsRegion: e.target.value })}
                        placeholder="us-east-1"
                        className="w-full bg-zinc-950 border border-zinc-700 text-white text-sm rounded-lg focus:ring-indigo-500 focus:border-indigo-500 block p-2"
                      />
                      <button 
                        onClick={() => fetchModels(config.provider, config.apiKey, config.awsAccessKey, config.awsSecretKey, config.awsRegion)}
                        className="px-4 py-2 bg-zinc-800 border border-zinc-700 text-white text-sm rounded-lg hover:bg-zinc-700 transition"
                      >
                        {isFetching ? "..." : "Fetch"}
                      </button>
                    </div>
                  </div>
                </div>
              )}
            </div>

            <div className="px-6 py-4 border-t border-zinc-800 bg-zinc-950 flex justify-end gap-3">
              <button 
                onClick={() => setIsOpen(false)}
                className="px-4 py-2 text-sm font-medium text-zinc-300 bg-zinc-800 rounded-lg hover:bg-zinc-700 transition"
              >
                Cancel
              </button>
              <button 
                onClick={handleSave}
                className="px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 transition"
              >
                Save Settings
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
