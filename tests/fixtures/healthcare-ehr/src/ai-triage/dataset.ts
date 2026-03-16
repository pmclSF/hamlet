export const triageDataset = [
  { symptoms: "chest pain", level: "emergency" },
  { symptoms: "mild headache", level: "low" },
  { symptoms: "fever 103F", level: "urgent" },
];
export function loadTriageDataset() { return triageDataset; }
export function splitTriageData(data: any[], ratio: number) {
  const idx = Math.floor(data.length * ratio);
  return { train: data.slice(0, idx), test: data.slice(idx) };
}
