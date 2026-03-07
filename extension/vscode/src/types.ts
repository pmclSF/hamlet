// Types aligned with the Hamlet CLI JSON snapshot contract.
// These types are consumed from `hamlet analyze --json` output.
// Do NOT duplicate business logic here — the CLI is the source of truth.

export interface SignalLocation {
  repository?: string;
  package?: string;
  file?: string;
  symbol?: string;
  line?: number;
}

export interface Signal {
  type: string;
  category: "structure" | "health" | "quality" | "migration" | "governance";
  severity: "info" | "low" | "medium" | "high" | "critical";
  confidence?: number;
  location: SignalLocation;
  owner?: string;
  explanation: string;
  suggestedAction?: string;
  metadata?: Record<string, unknown>;
}

export interface RiskSurface {
  type: string;
  scope: string;
  scopeName: string;
  band: "low" | "medium" | "high" | "critical";
  score?: number;
  contributingSignals?: Signal[];
  explanation?: string;
  suggestedAction?: string;
}

export interface Framework {
  name: string;
  type?: string;
  fileCount: number;
}

export interface RuntimeStats {
  avgRuntimeMs?: number;
  p95RuntimeMs?: number;
  passRate?: number;
  retryRate?: number;
  runtimeVariance?: number;
}

export interface TestFile {
  path: string;
  framework?: string;
  owner?: string;
  testCount: number;
  assertionCount: number;
  mockCount: number;
  snapshotCount: number;
  runtimeStats?: RuntimeStats;
  linkedCodeUnits?: string[];
  signals?: Signal[];
}

export interface RepositoryMetadata {
  name: string;
  rootPath: string;
  languages?: string[];
  packageManagers?: string[];
  ciSystems?: string[];
  snapshotTimestamp?: string;
  commitSHA?: string;
  branch?: string;
}

export interface TestSuiteSnapshot {
  repository: RepositoryMetadata;
  frameworks?: Framework[];
  testFiles?: TestFile[];
  codeUnits?: unknown[];
  signals?: Signal[];
  risk?: RiskSurface[];
  ownership?: Record<string, string[]>;
  policies?: Record<string, unknown>;
  metadata?: Record<string, unknown>;
  generatedAt: string;
}
