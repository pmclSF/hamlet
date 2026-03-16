export const supportDataset = [
  { input: "How do I upgrade my plan?", intent: "billing" },
  { input: "Add a new team member", intent: "admin" },
  { input: "Check our API usage", intent: "entitlements" },
];

export function loadEvalDataset() { return supportDataset; }

export function splitDataset(data: any[], ratio: number) {
  const idx = Math.floor(data.length * ratio);
  return { train: data.slice(0, idx), test: data.slice(idx) };
}
