"use client";

import { type ReactNode } from "react";

/** Base skeleton pulse animation block. */
function SkeletonPulse({ className = "" }: { className?: string }) {
  return (
    <div
      className={`animate-pulse bg-zinc-700/50 rounded ${className}`}
      role="status"
      aria-label="Loading"
    />
  );
}

/** Card-shaped loading skeleton. */
export function CardSkeleton() {
  return (
    <div className="p-4 bg-zinc-900 border border-zinc-800 rounded-xl space-y-3">
      <SkeletonPulse className="h-4 w-1/3" />
      <SkeletonPulse className="h-8 w-2/3" />
      <SkeletonPulse className="h-3 w-1/2" />
    </div>
  );
}

/** Table row loading skeleton. */
export function TableRowSkeleton({ columns = 5 }: { columns?: number }) {
  return (
    <div className="flex items-center gap-4 px-4 py-3 border-b border-zinc-800">
      {Array.from({ length: columns }).map((_, i) => (
        <SkeletonPulse key={i} className="h-4 flex-1" />
      ))}
    </div>
  );
}

/** Chart area loading skeleton. */
export function ChartSkeleton() {
  return (
    <div className="p-4 bg-zinc-900 border border-zinc-800 rounded-xl">
      <SkeletonPulse className="h-4 w-1/4 mb-4" />
      <SkeletonPulse className="h-48 w-full rounded-lg" />
    </div>
  );
}

/** Generic list skeleton with configurable row count. */
export function ListSkeleton({ rows = 5 }: { rows?: number }) {
  return (
    <div className="space-y-2">
      {Array.from({ length: rows }).map((_, i) => (
        <TableRowSkeleton key={i} />
      ))}
    </div>
  );
}

/** Wraps children and shows skeleton while loading. */
export function SkeletonWrapper({
  loading,
  skeleton,
  children,
}: {
  loading: boolean;
  skeleton: ReactNode;
  children: ReactNode;
}) {
  return loading ? <>{skeleton}</> : <>{children}</>;
}
