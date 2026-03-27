"use client";

import { Clock } from "lucide-react";

interface Props {
  symbol: string;
  name: string;
  price: number;
  change: number;
  changePercent: number;
  activeView: string;
  onViewChange: (view: string) => void;
}

export function StockDetailHeader({ symbol, name, price, change, changePercent, activeView, onViewChange }: Props) {
  const isPositive = change >= 0;
  
  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden shadow-lg mb-8">
      {/* Top Header Card */}
      <div className="p-6">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-10 h-10 rounded-full bg-red-600 flex items-center justify-center font-bold text-white text-lg border border-red-500">
            {symbol.substring(0, 1)}
          </div>
          <div>
            <h1 className="text-xl font-bold flex items-center gap-2 text-white">
              {symbol} <span className="text-xs font-normal bg-zinc-800 px-2 py-0.5 rounded text-zinc-400">HOSE</span>
            </h1>
            <p className="text-sm text-zinc-400">{name}</p>
          </div>
        </div>

        <div className="flex items-baseline gap-4 mt-2">
          <span className="text-4xl font-bold text-white">{price.toLocaleString()} <span className="text-lg text-zinc-500 font-medium">VND</span></span>
          <span className={`text-lg font-semibold ${isPositive ? 'text-green-500' : 'text-red-500'}`}>
            {isPositive ? '+' : ''}{change.toLocaleString()} ({isPositive ? '+' : ''}{changePercent.toFixed(2)}%)
          </span>
        </div>
        
        <div className="flex items-center gap-2 mt-2 text-xs text-zinc-500">
          <Clock size={12} />
          <span>Market open • As of today</span>
        </div>
      </div>

      {/* Tabs Menu */}
      <div className="flex px-6 border-t border-zinc-800 bg-zinc-950 overflow-x-auto no-scrollbar">
        {[
          "Overview", "Financials", "News", "Community", "Technicals", "Forecasts", "Seasonals", "ETFs"
        ].map((tab) => (
          <button 
            key={tab}
            onClick={() => onViewChange(tab)}
            className={`py-4 px-1 mr-6 text-sm font-medium border-b-2 whitespace-nowrap outline-none transition-colors ${activeView === tab ? 'text-white border-blue-600' : 'text-zinc-500 border-transparent hover:text-zinc-300 hover:border-zinc-700 focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2 focus-visible:ring-offset-zinc-950 rounded-sm'}`}
          >
            {tab}
          </button>
        ))}
      </div>
    </div>
  );
}
