/**
 * Drawing tools data model, serialization, and localStorage persistence.
 * Supports: TrendLine, HorizontalLine, FibonacciRetracement, Rectangle.
 */

import type { Time } from "lightweight-charts";

// ── Drawing types ──────────────────────────────────────────────────────────

export type DrawingType = "trendline" | "horizontal" | "fibonacci" | "rectangle";

export interface DrawingPoint {
  time: Time;
  price: number;
}

export interface BaseDrawing {
  id: string;
  type: DrawingType;
  color: string;
  lineWidth: number;
}

export interface TrendLineDrawing extends BaseDrawing {
  type: "trendline";
  start: DrawingPoint;
  end: DrawingPoint;
}

export interface HorizontalLineDrawing extends BaseDrawing {
  type: "horizontal";
  price: number;
}

export interface FibonacciDrawing extends BaseDrawing {
  type: "fibonacci";
  start: DrawingPoint;
  end: DrawingPoint;
  /** Standard Fibonacci levels */
  levels: number[];
}

export interface RectangleDrawing extends BaseDrawing {
  type: "rectangle";
  start: DrawingPoint;
  end: DrawingPoint;
}

export type Drawing =
  | TrendLineDrawing
  | HorizontalLineDrawing
  | FibonacciDrawing
  | RectangleDrawing;

// ── Default values ─────────────────────────────────────────────────────────

const DEFAULT_COLOR = "#2962ff";
const DEFAULT_LINE_WIDTH = 2;
const FIBONACCI_LEVELS = [0, 0.236, 0.382, 0.5, 0.618, 0.786, 1];

let drawingIdCounter = 0;

export function generateDrawingId(): string {
  drawingIdCounter++;
  return `drawing-${Date.now()}-${drawingIdCounter}`;
}

// ── Factory helpers ────────────────────────────────────────────────────────

export function createTrendLine(start: DrawingPoint, end: DrawingPoint): TrendLineDrawing {
  return {
    id: generateDrawingId(),
    type: "trendline",
    color: DEFAULT_COLOR,
    lineWidth: DEFAULT_LINE_WIDTH,
    start,
    end,
  };
}

export function createHorizontalLine(price: number): HorizontalLineDrawing {
  return {
    id: generateDrawingId(),
    type: "horizontal",
    color: "#ff6d00",
    lineWidth: DEFAULT_LINE_WIDTH,
    price,
  };
}

export function createFibonacci(start: DrawingPoint, end: DrawingPoint): FibonacciDrawing {
  return {
    id: generateDrawingId(),
    type: "fibonacci",
    color: "#aa00ff",
    lineWidth: 1,
    start,
    end,
    levels: FIBONACCI_LEVELS,
  };
}

export function createRectangle(start: DrawingPoint, end: DrawingPoint): RectangleDrawing {
  return {
    id: generateDrawingId(),
    type: "rectangle",
    color: "rgba(41, 98, 255, 0.2)",
    lineWidth: 1,
    start,
    end,
  };
}

// ── localStorage persistence ───────────────────────────────────────────────

const STORAGE_PREFIX = "ezistock-drawings-";

function storageKey(symbol: string): string {
  return `${STORAGE_PREFIX}${symbol}`;
}

/** Serialize drawings to JSON string. */
export function serializeDrawings(drawings: Drawing[]): string {
  return JSON.stringify(drawings);
}

/** Deserialize drawings from JSON string. Returns empty array on failure. */
export function deserializeDrawings(json: string): Drawing[] {
  try {
    const parsed = JSON.parse(json);
    if (!Array.isArray(parsed)) return [];
    return parsed as Drawing[];
  } catch {
    return [];
  }
}

/** Save drawings for a symbol to localStorage. */
export function saveDrawings(symbol: string, drawings: Drawing[]): void {
  if (typeof window === "undefined") return;
  try {
    localStorage.setItem(storageKey(symbol), serializeDrawings(drawings));
  } catch {
    // localStorage may be full or unavailable
  }
}

/** Load drawings for a symbol from localStorage. */
export function loadDrawings(symbol: string): Drawing[] {
  if (typeof window === "undefined") return [];
  try {
    const raw = localStorage.getItem(storageKey(symbol));
    if (!raw) return [];
    return deserializeDrawings(raw);
  } catch {
    return [];
  }
}

/** Clear all drawings for a symbol. */
export function clearDrawings(symbol: string): void {
  if (typeof window === "undefined") return;
  try {
    localStorage.removeItem(storageKey(symbol));
  } catch {
    // ignore
  }
}
