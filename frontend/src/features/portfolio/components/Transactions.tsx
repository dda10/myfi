"use client";

import { useState } from "react";
import { useI18n } from "@/context/I18nContext";

// --- Types ---

export interface TransactionRow {
  id: number;
  transactionDate: string;
  transactionType: string; // "buy" | "sell" | "dividend"
  symbol: string;
  quantity: number;
  unitPrice: number;
  totalValue: number;
  realizedPL?: number;
}

interface TransactionsProps {
  transactions: TransactionRow[];
}

type TxFilter = "all" | "buy" | "sell" | "dividend";
const TX_FILTERS: TxFilter[] = ["all", "buy", "sell", "dividend"];

export function Transactions({ transactions }: TransactionsProps) {
  const { t, formatCurrency, formatDate } = useI18n();
  const [filter, setFilter] = useState<TxFilter>("all");

  const filtered = filter === "all"
    ? transactions
    : transactions.filter((tx) => tx.transactionType === filter);

  const sorted = [...filtered].sort(
    (a, b) => new Date(b.transactionDate).getTime() - new Date(a.transactionDate).getTime(),
  );

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden">
      {/* Filter bar */}
      <div className="px-4 py-3 border-b border-zinc-800 flex items-center gap-2 flex-wrap">
        {TX_FILTERS.map((f) => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`px-3 py-1 rounded-md text-xs font-medium capitalize transition ${
              filter === f
                ? "bg-indigo-600 text-white"
                : "bg-zinc-800 text-zinc-400 hover:text-white hover:bg-zinc-700"
            }`}
          >
            {f === "all" ? t("common.all") : f === "buy" ? t("portfolio.buy") : f === "sell" ? t("portfolio.sell") : "Dividend"}
          </button>
        ))}
      </div>

      {sorted.length === 0 ? (
        <div className="flex items-center justify-center h-48 text-zinc-500 text-sm">
          {t("common.no_data")}
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-zinc-800 text-zinc-400 text-left">
                <th className="px-4 py-3 font-medium">Date</th>
                <th className="px-4 py-3 font-medium">Type</th>
                <th className="px-4 py-3 font-medium">{t("table.symbol")}</th>
                <th className="px-4 py-3 font-medium text-right">{t("portfolio.quantity")}</th>
                <th className="px-4 py-3 font-medium text-right">{t("table.price")}</th>
                <th className="px-4 py-3 font-medium text-right">Total</th>
                <th className="px-4 py-3 font-medium text-right">{t("portfolio.realized_pl")}</th>
              </tr>
            </thead>
            <tbody>
              {sorted.map((tx) => {
                const isBuy = tx.transactionType === "buy" || tx.transactionType === "dividend";
                return (
                  <tr key={tx.id} className="border-b border-zinc-800/50 hover:bg-zinc-800/30 transition">
                    <td className="px-4 py-3 text-zinc-300">{formatDate(tx.transactionDate)}</td>
                    <td className="px-4 py-3">
                      <span className={`inline-block px-2 py-0.5 rounded text-xs font-semibold capitalize ${
                        isBuy ? "bg-green-500/10 text-green-400" : "bg-red-500/10 text-red-400"
                      }`}>
                        {tx.transactionType}
                      </span>
                    </td>
                    <td className="px-4 py-3 font-semibold text-white">{tx.symbol}</td>
                    <td className="px-4 py-3 text-right text-zinc-300">{tx.quantity.toLocaleString()}</td>
                    <td className="px-4 py-3 text-right text-zinc-300">{formatCurrency(tx.unitPrice)}</td>
                    <td className="px-4 py-3 text-right text-white font-medium">{formatCurrency(tx.totalValue)}</td>
                    <td className="px-4 py-3 text-right text-zinc-300">
                      {tx.realizedPL != null ? (
                        <span className={tx.realizedPL >= 0 ? "text-green-400" : "text-red-400"}>
                          {tx.realizedPL >= 0 ? "+" : ""}{formatCurrency(tx.realizedPL)}
                        </span>
                      ) : "—"}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
