// Extension view descriptions for VS Code / Cursor sidebar.
//
// Each view consumes `hamlet analyze --json` output and renders
// a tree view. The extension does NOT re-implement business logic.
//
// This file defines the view structure and data flow.
// Actual VS Code TreeDataProvider implementations would import from here.

import {
  TestSuiteSnapshot,
  Signal,
  RiskSurface,
  MigrationReadiness,
  MigrationAreaAssessment,
} from "./types";
import {
  groupByType,
  groupByOwner,
  groupByDirectory,
  migrationSignals,
  reviewWorthy,
  GroupedItem,
} from "./signal_renderer";

// ---------- Overview View ----------

export interface OverviewData {
  repoName: string;
  branch: string;
  frameworkCount: number;
  testFileCount: number;
  signalCount: number;
  riskSurfaces: RiskSurface[];
  topIssues: Signal[];
}

export function buildOverview(snap: TestSuiteSnapshot): OverviewData {
  const signals = snap.signals || [];
  return {
    repoName: snap.repository.name,
    branch: snap.repository.branch || "",
    frameworkCount: snap.frameworks?.length || 0,
    testFileCount: snap.testFiles?.length || 0,
    signalCount: signals.length,
    riskSurfaces: snap.risk || [],
    topIssues: signals.slice(0, 5),
  };
}

// ---------- Health View ----------

export interface HealthData {
  signals: Signal[];
  byType: GroupedItem[];
}

export function buildHealth(snap: TestSuiteSnapshot): HealthData {
  const healthSignals = (snap.signals || []).filter(
    (s) => s.category === "health"
  );
  return {
    signals: healthSignals,
    byType: groupByType(healthSignals),
  };
}

// ---------- Quality View ----------

export interface QualityData {
  signals: Signal[];
  byType: GroupedItem[];
}

export function buildQuality(snap: TestSuiteSnapshot): QualityData {
  const qualitySignals = (snap.signals || []).filter(
    (s) => s.category === "quality"
  );
  return {
    signals: qualitySignals,
    byType: groupByType(qualitySignals),
  };
}

// ---------- Review View ----------
// Shows findings that likely require human attention,
// grouped by signal type, owner, or directory.

export interface ReviewData {
  /** All review-worthy signals (medium+ severity). */
  signals: Signal[];
  byType: GroupedItem[];
  byOwner: GroupedItem[];
  byDirectory: GroupedItem[];
  /** Migration blockers surfaced as a first-class review grouping. */
  migrationBlockers: GroupedItem[];
  totalCount: number;
}

export function buildReview(snap: TestSuiteSnapshot): ReviewData {
  const worthy = reviewWorthy(snap.signals || []);
  const migBlockers = migrationSignals(snap.signals || []).filter((s) =>
    new Set(["medium", "high", "critical"]).has(s.severity)
  );
  return {
    signals: worthy,
    byType: groupByType(worthy),
    byOwner: groupByOwner(worthy),
    byDirectory: groupByDirectory(worthy),
    migrationBlockers: groupByType(migBlockers),
    totalCount: worthy.length,
  };
}

// ---------- Migration View ----------
// Surfaces migration readiness, blockers, and modernization status.

export interface MigrationData {
  /** All migration-related signals. */
  signals: Signal[];
  blockerGroups: GroupedItem[];
  byOwner: GroupedItem[];
  byDirectory: GroupedItem[];
  frameworkSummary: string[];
  totalBlockers: number;
  /** Area assessments from readiness model, if available via separate CLI call. */
  areaAssessments: MigrationAreaItem[];
  /** Whether the preview command is available for file-level drill-down. */
  previewAvailable: boolean;
}

/** Simplified area assessment for the extension view. */
export interface MigrationAreaItem {
  directory: string;
  classification: "safe" | "caution" | "risky";
  blockerCount: number;
  qualityIssueCount: number;
}

export function buildMigration(snap: TestSuiteSnapshot): MigrationData {
  const migSignals = migrationSignals(snap.signals || []);
  const frameworks = (snap.frameworks || []).map(
    (f) => `${f.name} (${f.fileCount} files)`
  );

  // Derive area assessments from migration signals grouped by directory.
  const dirGroups = groupByDirectory(migSignals);
  const qualitySignals = (snap.signals || []).filter(
    (s) => s.category === "quality"
  );
  const qualityByDir = new Map<string, number>();
  for (const s of qualitySignals) {
    const file = s.location?.file || "";
    const lastSlash = file.lastIndexOf("/");
    const dir = lastSlash > 0 ? file.substring(0, lastSlash) : "(repo-level)";
    qualityByDir.set(dir, (qualityByDir.get(dir) || 0) + 1);
  }

  const areaAssessments: MigrationAreaItem[] = dirGroups.map((g) => {
    const qualityCount = qualityByDir.get(g.key) || 0;
    let classification: "safe" | "caution" | "risky" = "safe";
    if (g.count > 0 && qualityCount > 0) {
      classification = "risky";
    } else if (g.count > 0 || qualityCount > 0) {
      classification = "caution";
    }
    return {
      directory: g.key,
      classification,
      blockerCount: g.count,
      qualityIssueCount: qualityCount,
    };
  });

  // Sort: risky first, then caution.
  const classOrder: Record<string, number> = { risky: 0, caution: 1, safe: 2 };
  areaAssessments.sort(
    (a, b) =>
      (classOrder[a.classification] ?? 2) -
      (classOrder[b.classification] ?? 2)
  );

  return {
    signals: migSignals,
    blockerGroups: groupByType(migSignals),
    byOwner: groupByOwner(migSignals),
    byDirectory: dirGroups,
    frameworkSummary: frameworks,
    totalBlockers: migSignals.length,
    areaAssessments,
    previewAvailable: true,
  };
}
