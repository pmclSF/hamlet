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
  evidenceStrength?: "strong" | "moderate" | "weak";
  evidenceSource?:
    | "ast"
    | "structural-pattern"
    | "path-name"
    | "runtime"
    | "coverage"
    | "policy"
    | "codeowners";
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

export interface ParameterizationInfo {
  isTemplate: boolean;
  paramSignature?: string;
  estimatedInstances?: number;
}

export interface TestCase {
  testId: string;
  canonicalIdentity: string;
  filePath: string;
  suiteHierarchy?: string[];
  testName: string;
  framework: string;
  language: string;
  line?: number;
  extractionKind: string;
  confidence: number;
  testType?: string;
  testTypeConfidence?: number;
  testTypeEvidence?: string[];
  parameterized?: ParameterizationInfo;
}

export interface CodeUnit {
  unitId?: string;
  name: string;
  path: string;
  kind: "function" | "method" | "class" | "module" | "unknown";
  exported: boolean;
  parentName?: string;
  language?: string;
  startLine?: number;
  endLine?: number;
  complexity?: number;
  coverage?: number;
  linkedTestFiles?: string[];
  owner?: string;
}

export interface CoverageInsight {
  type: string;
  severity: string;
  description: string;
  path?: string;
  unitId?: string;
  suggestedAction?: string;
}

export interface CoverageSummary {
  totalCodeUnits: number;
  coveredByUnitTests: number;
  coveredByIntegration: number;
  coveredByE2e: number;
  coveredOnlyByE2e: number;
  uncoveredExported: number;
  uncovered: number;
  lineCoveragePct?: number;
  branchCoveragePct?: number;
}

// Migration preview types aligned with `hamlet migration preview --json`.
export interface MigrationPreviewBlocker {
  type: string;
  pattern: string;
  explanation: string;
  remediation: string;
}

export interface MigrationPreviewResult {
  file: string;
  sourceFramework: string;
  suggestedTarget?: string;
  difficulty: "low" | "medium" | "high" | "unknown";
  blockers?: MigrationPreviewBlocker[];
  safePatterns?: string[];
  previewAvailable: boolean;
  explanation: string;
  limitations?: string[];
}

// Migration readiness types aligned with `hamlet migration readiness --json`.
export interface MigrationBlockerExample {
  type: string;
  file: string;
  explanation: string;
}

export interface MigrationAreaAssessment {
  directory: string;
  classification: "safe" | "caution" | "risky";
  migrationBlockers: number;
  qualityIssues: number;
  testFileCount: number;
  explanation: string;
}

export interface MigrationReadiness {
  frameworks: Framework[];
  totalBlockers: number;
  blockersByType: Record<string, number>;
  representativeBlockers?: MigrationBlockerExample[];
  readinessLevel: "low" | "medium" | "high" | "unknown";
  explanation: string;
  areaAssessments?: MigrationAreaAssessment[];
}

export interface SnapshotMeta {
  schemaVersion: string;
  engineVersion?: string;
  detectorCount?: number;
  detectors?: string[];
}

export interface TestSuiteSnapshot {
  snapshotMeta?: SnapshotMeta;
  repository: RepositoryMetadata;
  frameworks?: Framework[];
  testFiles?: TestFile[];
  testCases?: TestCase[];
  codeUnits?: CodeUnit[];
  signals?: Signal[];
  risk?: RiskSurface[];
  coverageSummary?: CoverageSummary;
  coverageInsights?: CoverageInsight[];
  ownership?: Record<string, string[]>;
  policies?: Record<string, unknown>;
  metadata?: Record<string, unknown>;
  generatedAt: string;
}
