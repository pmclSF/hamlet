export const trainingDataset = [
  { input: "I want to buy shoes", intent: "purchase" },
  { input: "Track my order", intent: "tracking" },
  { input: "I need a refund", intent: "refund" },
];

export function loadEvalDataset() {
  return trainingDataset;
}

export function splitDataset(data: any[], ratio: number) {
  const split = Math.floor(data.length * ratio);
  return { train: data.slice(0, split), test: data.slice(split) };
}
