"use client";

import { TrendingUp, TrendingDown, DollarSign, Activity } from "lucide-react";
import { useEffect, useState } from "react";

export function Stats() {
  const [data, setData] = useState<any>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const res = await fetch("http://localhost:8080/api/market/quote?symbols=VNINDEX");
        const json = await res.json();
        if (json.data && json.data.length > 0) {
          setData(json.data[0]);
        }
      } catch (err) {
        console.error("Failed to fetch VNINDEX stats", err);
      }
    };
    fetchData();
  }, []);

  const vniValue = data ? data.close.toFixed(2) : "Loading...";
  const refOrClose = data ? (data.ref || data.close * 0.99) : 1; 
  const vniChangeVal = data ? (data.close - refOrClose).toFixed(2) : "0.00";
  const vniChangePct = data ? (((data.close - refOrClose) / refOrClose) * 100).toFixed(2) : "0.00";
  const isPositive = data ? data.close >= refOrClose : true;
  const vol = data ? (data.volume / 1000000).toFixed(2) + "M" : "Loading...";

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
      <StatCard 
        title="VNINDEX" 
        value={vniValue} 
        change={`${vniChangeVal} (${vniChangePct}%)`} 
        isPositive={isPositive} 
        icon={<TrendingUp className="text-purple-400" size={20} />} 
      />
      <StatCard 
        title="Market Volume" 
        value={vol} 
        change="Shares" 
        isPositive={true} 
        icon={<Activity className="text-green-400" size={20} />} 
      />
      <StatCard 
        title="Available Cash" 
        value="$12,040.00" 
        change="-1.2%" 
        isPositive={false} 
        icon={<DollarSign className="text-zinc-400" size={20} />} 
      />
      <StatCard 
        title="Total NAV" 
        value="$124,563.00" 
        change="+2.4%" 
        isPositive={true} 
        icon={<DollarSign className="text-blue-400" size={20} />} 
      />
    </div>
  );
}

function StatCard({ title, value, change, isPositive, icon }: { title: string, value: string, change: string, isPositive: boolean, icon: React.ReactNode }) {
  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-5 hover:border-zinc-700 transition">
      <div className="flex items-start justify-between mb-4">
        <h3 className="text-zinc-400 text-sm font-medium">{title}</h3>
        <div className="p-2 bg-zinc-800 rounded-lg">
          {icon}
        </div>
      </div>
      <div className="flex items-baseline gap-2 mt-2">
        <span className="text-2xl font-bold text-white">{value}</span>
      </div>
      <div className="mt-2">
        {change !== "Shares" && (
          <span className={`text-xs font-semibold px-2 py-1 rounded-full ${isPositive ? 'bg-green-500/10 text-green-400' : 'bg-red-500/10 text-red-400'}`}>
            {isPositive && change !== 'Neutral' ? '+' : ''}{change}
          </span>
        )}
        {change === "Shares" && (
          <span className="text-zinc-500 text-xs">Today</span>
        )}
        <span className="text-zinc-500 text-xs ml-2">{change !== "Shares" ? "" : ""}</span>
      </div>
    </div>
  );
}
