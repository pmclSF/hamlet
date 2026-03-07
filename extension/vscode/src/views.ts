// Extension view descriptions for VS Code / Cursor sidebar.
//
// Each view consumes `hamlet analyze --json` output and renders
// a tree view. The extension does NOT re-implement business logic.
//
// This file defines the view structure and data flow.
// Actual VS Code TreeDataProvider implementations would import from here.

import { TestSuiteSnapshot, Signal, RiskSurface } from "./types";
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
  totalCount: number;
}

export function buildReview(snap: TestSuiteSnapshot): ReviewData {
  const worthy = reviewWorthy(snap.signals || []);
  return {
    signals: worthy,
    byType: groupByType(worthy),
    byOwner: groupByOwner(worthy),
    byDirectory: groupByDirectory(worthy),
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
  frameworkSummary: string[];
  totalBlockers: number;
}

export function buildMigration(snap: TestSuiteSnapshot): MigrationData {
  const migSignals = migrationSignals(snap.signals || []);
  const frameworks = (snap.frameworks || []).map(
    (f) => `${f.name} (${f.fileCount} files)`
  );

  return {
    signals: migSignals,
    blockerGroups: groupByType(migSignals),
    byOwner: groupByOwner(migSignals),
    frameworkSummary: frameworks,
    totalBlockers: migSignals.length,
  };
}
