// Helpers for transforming snapshot data into grouped view items.
// Keep this thin — grouping only, no inference or business logic.

import { Signal, RiskSurface, TestSuiteSnapshot } from "./types";

export interface GroupedItem {
  key: string;
  signals: Signal[];
  count: number;
}

/** Group signals by their type field. */
export function groupByType(signals: Signal[]): GroupedItem[] {
  return groupBy(signals, (s) => s.type);
}

/** Group signals by their owner field. */
export function groupByOwner(signals: Signal[]): GroupedItem[] {
  return groupBy(signals, (s) => s.owner || "unknown");
}

/** Group signals by their category field. */
export function groupByCategory(signals: Signal[]): GroupedItem[] {
  return groupBy(signals, (s) => s.category);
}

/** Group signals by the directory of their file location. */
export function groupByDirectory(signals: Signal[]): GroupedItem[] {
  return groupBy(signals, (s) => {
    const file = s.location?.file || "";
    const lastSlash = file.lastIndexOf("/");
    return lastSlash > 0 ? file.substring(0, lastSlash) : "(repo-level)";
  });
}

/** Filter signals to review-worthy items (medium+ severity). */
export function reviewWorthy(signals: Signal[]): Signal[] {
  const reviewSeverities = new Set(["medium", "high", "critical"]);
  return signals.filter((s) => reviewSeverities.has(s.severity));
}

/** Filter signals to migration-related types. */
export function migrationSignals(signals: Signal[]): Signal[] {
  const migrationTypes = new Set([
    "frameworkMigration",
    "migrationBlocker",
    "deprecatedTestPattern",
    "dynamicTestGeneration",
    "customMatcherRisk",
    "legacyFrameworkUsage",
  ]);
  return signals.filter((s) => migrationTypes.has(s.type));
}

/** Get a severity icon for display. */
export function severityIcon(severity: string): string {
  switch (severity) {
    case "critical":
      return "error";
    case "high":
      return "warning";
    case "medium":
      return "info";
    case "low":
      return "debug-stackframe-dot";
    default:
      return "circle-outline";
  }
}

/** Get a risk band icon for display. */
export function riskBandIcon(band: string): string {
  switch (band) {
    case "critical":
      return "error";
    case "high":
      return "warning";
    case "medium":
      return "info";
    default:
      return "pass";
  }
}

function groupBy(
  signals: Signal[],
  keyFn: (s: Signal) => string
): GroupedItem[] {
  const map = new Map<string, Signal[]>();
  for (const s of signals) {
    const key = keyFn(s);
    const existing = map.get(key);
    if (existing) {
      existing.push(s);
    } else {
      map.set(key, [s]);
    }
  }

  const groups: GroupedItem[] = [];
  for (const [key, sigs] of map) {
    groups.push({ key, signals: sigs, count: sigs.length });
  }
  groups.sort((a, b) => b.count - a.count);
  return groups;
}
