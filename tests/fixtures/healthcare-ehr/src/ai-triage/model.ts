export function assessTriage(prompt: string) {
  return { triageLevel: 'moderate', confidence: 0.82, recommendation: 'Schedule within 48h' };
}
export function classifySymptoms(text: string) {
  return { category: 'respiratory', confidence: 0.88 };
}
