"use client";

import { Flame, Clock } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { usePolling, isVNTradingHours } from "@/hooks/usePolling";
import { apiFetch } from "@/lib/api";

interface HotTopic {
  name: string;
  description: string;
  timestamp?: string;
}

// Fallback mock topics in MiQuant style
const MOCK_TOPICS: HotTopic[] = [
  { name: "Giá dầu thô giảm sau tin hiệu kết thúc chiến tranh Iran", description: "Thị trường năng lượng phản ứng tích cực trước các tín hiệu hòa đàm.", timestamp: "2d ago" },
  { name: "Giá dầu tăng mạnh do chiến sự Iran và đồng Strait of Hormuz", description: "Lo ngại nguồn cung bị gián đoạn đẩy giá dầu Brent vượt $90.", timestamp: "2d ago" },
  { name: "Khối ngoại bán ròng liên tiếp, đảo chiều nhẹ phiên 01/04", description: "Dòng tiền ngoại vẫn thận trọng trong bối cảnh tỷ giá biến động.", timestamp: "2d ago" },
  { name: "Nhóm Vingroup dẫn dắt VN-Index vượt mốc 1.700", description: "VHM, VIC, VRE đồng loạt tăng mạnh trong phiên sáng.", timestamp: "2d ago" },
  { name: "NHNN bơm tiền mạnh qua kênh OMO hỗ trợ thanh khoản", description: "Lãi suất liên ngân hàng qua đêm giảm về 4.2% sau khi NHNN bơm.", timestamp: "3d ago" },
];

export function HotTopics() {
  const { t } = useI18n();
  const interval = isVNTradingHours() ? 5 * 60_000 : 30 * 60_000;

  const { data, loading } = usePolling<HotTopic[]>(
    () => apiFetch<HotTopic[]>("/api/market/hot-topics"),
    interval,
  );

  const topics = Array.isArray(data) && data.length > 0 ? data : MOCK_TOPICS;

  return (
    <section className="bg-card-bg border border-border-theme rounded-xl overflow-hidden">
      <div className="px-4 py-3 border-b border-border-theme flex items-center gap-2">
        <Flame size={14} className="text-orange-400" />
        <h2 className="text-xs font-bold text-foreground">Chủ đề nổi bật trên thị trường</h2>
      </div>

      {loading && !data ? (
        <div className="divide-y divide-border-theme">
          {[1, 2, 3].map(i => (
            <div key={i} className="px-4 py-3 animate-pulse">
              <div className="h-3 w-3/4 bg-surface rounded mb-1.5" />
              <div className="h-2.5 w-1/2 bg-surface rounded" />
            </div>
          ))}
        </div>
      ) : (
        <div className="divide-y divide-border-theme">
          {topics.slice(0, 5).map((topic, i) => (
            <div
              key={i}
              className="px-4 py-3 hover:bg-surface transition cursor-pointer group"
            >
              <div className="flex items-start justify-between gap-3">
                <div className="flex-1 min-w-0">
                  <p className="text-xs font-semibold text-foreground group-hover:text-indigo-400 transition line-clamp-1">
                    {i + 1}. {topic.name}
                  </p>
                  {topic.description && (
                    <p className="text-[11px] text-text-muted mt-0.5 line-clamp-1">{topic.description}</p>
                  )}
                </div>
                {topic.timestamp && (
                  <div className="flex items-center gap-1 text-[10px] text-text-muted flex-shrink-0">
                    <Clock size={10} />
                    {topic.timestamp}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
