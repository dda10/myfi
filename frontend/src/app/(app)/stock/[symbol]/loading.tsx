import { CardSkeleton, ChartSkeleton } from "@/components/common/Skeleton";

export default function StockLoading() {
  return (
    <div className="space-y-6">
      <CardSkeleton />
      <ChartSkeleton />
    </div>
  );
}
