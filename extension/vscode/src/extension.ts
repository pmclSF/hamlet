// Terrain VS Code Extension
//
// Thin client over the CLI. All intelligence lives in `terrain analyze --json`.
// The extension renders structured views from the snapshot — no business logic.

import * as vscode from "vscode";
import { TestSuiteSnapshot, Signal } from "./types";
import {
  buildOverview,
  buildHealth,
  buildQuality,
  buildMigration,
  buildReview,
  OverviewData,
  HealthData,
  QualityData,
  MigrationData,
  ReviewData,
} from "./views";
import { severityIcon, riskBandIcon, GroupedItem } from "./signal_renderer";
import { execFile } from "child_process";

// ── State ──────────────────────────────────────────────────

type ExtensionState =
  | { kind: "empty" }
  | { kind: "loading" }
  | { kind: "error"; message: string }
  | { kind: "loaded"; snapshot: TestSuiteSnapshot };

let state: ExtensionState = { kind: "empty" };

// ── Activation ─────────────────────────────────────────────

export function activate(context: vscode.ExtensionContext) {
  const overviewProvider = new OverviewTreeProvider();
  const healthProvider = new HealthTreeProvider();
  const qualityProvider = new QualityTreeProvider();
  const migrationProvider = new MigrationTreeProvider();
  const reviewProvider = new ReviewTreeProvider();

  const providers = [
    overviewProvider,
    healthProvider,
    qualityProvider,
    migrationProvider,
    reviewProvider,
  ];

  context.subscriptions.push(
    vscode.window.registerTreeDataProvider("terrain.overview", overviewProvider),
    vscode.window.registerTreeDataProvider("terrain.health", healthProvider),
    vscode.window.registerTreeDataProvider("terrain.quality", qualityProvider),
    vscode.window.registerTreeDataProvider(
      "terrain.migration",
      migrationProvider
    ),
    vscode.window.registerTreeDataProvider("terrain.review", reviewProvider)
  );

  context.subscriptions.push(
    vscode.commands.registerCommand("terrain.refresh", () => {
      runAnalysis(providers);
    }),
    vscode.commands.registerCommand("terrain.openSummary", () => {
      runCliInTerminal("summary");
    }),
    vscode.commands.registerCommand("terrain.openMigrationBlockers", () => {
      runCliInTerminal("migration blockers");
    }),
    vscode.commands.registerCommand(
      "terrain.revealFile",
      (filePath: string) => {
        if (filePath && vscode.workspace.workspaceFolders?.[0]) {
          const uri = vscode.Uri.joinPath(
            vscode.workspace.workspaceFolders[0].uri,
            filePath
          );
          vscode.window.showTextDocument(uri);
        }
      }
    )
  );

  // Initial analysis on activation.
  runAnalysis(providers);
}

export function deactivate() {}

// ── CLI Integration ────────────────────────────────────────

function getTerrainBinary(): string {
  return (
    vscode.workspace.getConfiguration("terrain").get<string>("binaryPath") ||
    "terrain"
  );
}

function getWorkspaceRoot(): string | undefined {
  return vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
}

function runAnalysis(providers: TerrainTreeProvider[]) {
  const root = getWorkspaceRoot();
  if (!root) {
    state = { kind: "error", message: "No workspace folder open" };
    providers.forEach((p) => p.refresh());
    return;
  }

  state = { kind: "loading" };
  providers.forEach((p) => p.refresh());

  const binary = getTerrainBinary();
  execFile(binary, ["analyze", "--json", "--root", root], (err, stdout) => {
    if (err) {
      state = {
        kind: "error",
        message: `terrain analyze failed: ${err.message}`,
      };
      providers.forEach((p) => p.refresh());
      return;
    }

    try {
      const snapshot = JSON.parse(stdout) as TestSuiteSnapshot;
      state = { kind: "loaded", snapshot };
    } catch (e) {
      state = { kind: "error", message: `Failed to parse snapshot JSON` };
    }
    providers.forEach((p) => p.refresh());
  });
}

function runCliInTerminal(subcommand: string) {
  const root = getWorkspaceRoot();
  const binary = getTerrainBinary();
  const terminal = vscode.window.createTerminal("Terrain");
  const rootFlag = root ? ` --root "${root}"` : "";
  terminal.sendText(`${binary} ${subcommand}${rootFlag}`);
  terminal.show();
}

// ── Base TreeDataProvider ──────────────────────────────────

interface TerrainTreeProvider extends vscode.TreeDataProvider<TerrainTreeItem> {
  refresh(): void;
}

class TerrainTreeItem extends vscode.TreeItem {
  constructor(
    label: string,
    collapsibleState: vscode.TreeItemCollapsibleState = vscode
      .TreeItemCollapsibleState.None,
    public children: TerrainTreeItem[] = []
  ) {
    super(label, collapsibleState);
  }
}

function stateItems(): TerrainTreeItem[] | null {
  switch (state.kind) {
    case "empty":
      return [
        makeItem(
          "Run terrain.refresh to analyze",
          "$(play)",
          "No analysis data yet. Click refresh or run 'Terrain: Refresh Analysis'."
        ),
      ];
    case "loading":
      return [makeItem("Analyzing...", "$(sync~spin)")];
    case "error":
      return [
        makeItem("Analysis failed", "$(error)"),
        makeItem(state.message, "$(info)"),
        makeItem(
          "Is terrain installed?",
          "$(question)",
          "Install: go install github.com/pmclSF/terrain/cmd/terrain@latest"
        ),
      ];
    case "loaded":
      return null; // proceed to real data
  }
}

function makeItem(
  label: string,
  icon?: string,
  tooltip?: string
): TerrainTreeItem {
  const item = new TerrainTreeItem(label);
  if (icon) {
    item.iconPath = new vscode.ThemeIcon(icon.replace("$(", "").replace(")", ""));
  }
  if (tooltip) {
    item.tooltip = tooltip;
  }
  return item;
}

function signalItem(s: Signal): TerrainTreeItem {
  const loc = s.location?.file
    ? `${s.location.file}${s.location.line ? `:${s.location.line}` : ""}`
    : "";
  const label = loc ? `${s.type} - ${loc}` : s.type;
  const item = new TerrainTreeItem(label);
  item.iconPath = new vscode.ThemeIcon(severityIcon(s.severity));
  item.tooltip = s.explanation;
  item.description = s.evidenceStrength
    ? `${s.severity} (${s.evidenceStrength})`
    : s.severity;
  if (s.location?.file) {
    item.command = {
      command: "terrain.revealFile",
      title: "Open File",
      arguments: [s.location.file],
    };
  }
  return item;
}

function groupItems(groups: GroupedItem[]): TerrainTreeItem[] {
  return groups.map((g) => {
    const children = g.signals.map(signalItem);
    const item = new TerrainTreeItem(
      `${g.key} (${g.count})`,
      vscode.TreeItemCollapsibleState.Collapsed,
      children
    );
    return item;
  });
}

// ── Overview Provider ──────────────────────────────────────

class OverviewTreeProvider implements TerrainTreeProvider {
  private _onDidChangeTreeData = new vscode.EventEmitter<void>();
  readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

  refresh() {
    this._onDidChangeTreeData.fire();
  }

  getTreeItem(element: TerrainTreeItem): TerrainTreeItem {
    return element;
  }

  getChildren(element?: TerrainTreeItem): TerrainTreeItem[] {
    if (element) return element.children;
    const items = stateItems();
    if (items) return items;
    const snap = (state as { kind: "loaded"; snapshot: TestSuiteSnapshot })
      .snapshot;
    const data = buildOverview(snap);
    return this.buildItems(data);
  }

  private buildItems(data: OverviewData): TerrainTreeItem[] {
    const items: TerrainTreeItem[] = [
      makeItem(`Repository: ${data.repoName}`, "$(repo)"),
      makeItem(`Frameworks: ${data.frameworkCount}`, "$(package)"),
      makeItem(`Test files: ${data.testFileCount}`, "$(file-code)"),
      makeItem(`Signals: ${data.signalCount}`, "$(warning)"),
    ];

    if (data.riskSurfaces.length > 0) {
      const riskChildren = data.riskSurfaces
        .filter((r) => r.scope === "repository")
        .map((r) => {
          const item = makeItem(
            `${r.type}: ${r.band}`,
            riskBandIcon(r.band)
          );
          item.tooltip = r.explanation || "";
          return item;
        });
      if (riskChildren.length > 0) {
        items.push(
          new TerrainTreeItem(
            "Risk Surfaces",
            vscode.TreeItemCollapsibleState.Expanded,
            riskChildren
          )
        );
      }
    }

    if (data.topIssues.length > 0) {
      const issueChildren = data.topIssues.map(signalItem);
      items.push(
        new TerrainTreeItem(
          "Top Issues",
          vscode.TreeItemCollapsibleState.Collapsed,
          issueChildren
        )
      );
    }

    return items;
  }
}

// ── Health Provider ────────────────────────────────────────

class HealthTreeProvider implements TerrainTreeProvider {
  private _onDidChangeTreeData = new vscode.EventEmitter<void>();
  readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

  refresh() {
    this._onDidChangeTreeData.fire();
  }

  getTreeItem(element: TerrainTreeItem): TerrainTreeItem {
    return element;
  }

  getChildren(element?: TerrainTreeItem): TerrainTreeItem[] {
    if (element) return element.children;
    const items = stateItems();
    if (items) return items;
    const snap = (state as { kind: "loaded"; snapshot: TestSuiteSnapshot })
      .snapshot;
    const data = buildHealth(snap);
    return this.buildItems(data);
  }

  private buildItems(data: HealthData): TerrainTreeItem[] {
    if (data.signals.length === 0) {
      return [
        makeItem(
          "No health signals detected",
          "$(pass)",
          "Health signals require runtime artifacts. Use --runtime flag with terrain analyze."
        ),
      ];
    }
    return groupItems(data.byType);
  }
}

// ── Quality Provider ───────────────────────────────────────

class QualityTreeProvider implements TerrainTreeProvider {
  private _onDidChangeTreeData = new vscode.EventEmitter<void>();
  readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

  refresh() {
    this._onDidChangeTreeData.fire();
  }

  getTreeItem(element: TerrainTreeItem): TerrainTreeItem {
    return element;
  }

  getChildren(element?: TerrainTreeItem): TerrainTreeItem[] {
    if (element) return element.children;
    const items = stateItems();
    if (items) return items;
    const snap = (state as { kind: "loaded"; snapshot: TestSuiteSnapshot })
      .snapshot;
    const data = buildQuality(snap);
    return this.buildItems(data);
  }

  private buildItems(data: QualityData): TerrainTreeItem[] {
    if (data.signals.length === 0) {
      return [makeItem("No quality signals detected", "$(pass)")];
    }
    return groupItems(data.byType);
  }
}

// ── Migration Provider ─────────────────────────────────────

class MigrationTreeProvider implements TerrainTreeProvider {
  private _onDidChangeTreeData = new vscode.EventEmitter<void>();
  readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

  refresh() {
    this._onDidChangeTreeData.fire();
  }

  getTreeItem(element: TerrainTreeItem): TerrainTreeItem {
    return element;
  }

  getChildren(element?: TerrainTreeItem): TerrainTreeItem[] {
    if (element) return element.children;
    const items = stateItems();
    if (items) return items;
    const snap = (state as { kind: "loaded"; snapshot: TestSuiteSnapshot })
      .snapshot;
    const data = buildMigration(snap);
    return this.buildItems(data);
  }

  private buildItems(data: MigrationData): TerrainTreeItem[] {
    const items: TerrainTreeItem[] = [];

    // Framework summary
    if (data.frameworkSummary.length > 0) {
      const fwChildren = data.frameworkSummary.map((fw) =>
        makeItem(fw, "$(package)")
      );
      items.push(
        new TerrainTreeItem(
          "Frameworks",
          vscode.TreeItemCollapsibleState.Expanded,
          fwChildren
        )
      );
    }

    // Blocker count
    if (data.totalBlockers === 0) {
      items.push(
        makeItem("No migration blockers found", "$(pass)")
      );
    } else {
      items.push(
        makeItem(`${data.totalBlockers} migration signals`, "$(warning)")
      );
    }

    // Blocker groups
    if (data.blockerGroups.length > 0) {
      items.push(
        new TerrainTreeItem(
          "Blockers by Type",
          vscode.TreeItemCollapsibleState.Collapsed,
          groupItems(data.blockerGroups)
        )
      );
    }

    // Area assessments
    if (data.areaAssessments.length > 0) {
      const areaChildren = data.areaAssessments.map((a) => {
        const icon =
          a.classification === "risky"
            ? "error"
            : a.classification === "caution"
            ? "warning"
            : "pass";
        const item = makeItem(
          `${a.directory} [${a.classification.toUpperCase()}]`,
          icon
        );
        item.description = `${a.blockerCount} blockers, ${a.qualityIssueCount} quality issues`;
        return item;
      });
      items.push(
        new TerrainTreeItem(
          "Area Assessments",
          vscode.TreeItemCollapsibleState.Collapsed,
          areaChildren
        )
      );
    }

    return items;
  }
}

// ── Review Provider ────────────────────────────────────────

class ReviewTreeProvider implements TerrainTreeProvider {
  private _onDidChangeTreeData = new vscode.EventEmitter<void>();
  readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

  refresh() {
    this._onDidChangeTreeData.fire();
  }

  getTreeItem(element: TerrainTreeItem): TerrainTreeItem {
    return element;
  }

  getChildren(element?: TerrainTreeItem): TerrainTreeItem[] {
    if (element) return element.children;
    const items = stateItems();
    if (items) return items;
    const snap = (state as { kind: "loaded"; snapshot: TestSuiteSnapshot })
      .snapshot;
    const data = buildReview(snap);
    return this.buildItems(data);
  }

  private buildItems(data: ReviewData): TerrainTreeItem[] {
    if (data.totalCount === 0) {
      return [makeItem("No review-worthy findings", "$(pass)")];
    }

    const items: TerrainTreeItem[] = [
      makeItem(`${data.totalCount} findings need attention`, "$(warning)"),
    ];

    if (data.byType.length > 0) {
      items.push(
        new TerrainTreeItem(
          "By Type",
          vscode.TreeItemCollapsibleState.Collapsed,
          groupItems(data.byType)
        )
      );
    }

    if (data.byOwner.length > 0) {
      items.push(
        new TerrainTreeItem(
          "By Owner",
          vscode.TreeItemCollapsibleState.Collapsed,
          groupItems(data.byOwner)
        )
      );
    }

    if (data.byDirectory.length > 0) {
      items.push(
        new TerrainTreeItem(
          "By Directory",
          vscode.TreeItemCollapsibleState.Collapsed,
          groupItems(data.byDirectory)
        )
      );
    }

    if (data.migrationBlockers.length > 0) {
      items.push(
        new TerrainTreeItem(
          "Migration Blockers",
          vscode.TreeItemCollapsibleState.Collapsed,
          groupItems(data.migrationBlockers)
        )
      );
    }

    return items;
  }
}
