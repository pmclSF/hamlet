export const triageSystemPrompt = "You are a medical triage assistant. Assess symptom severity.";
export function buildTriagePrompt(symptoms: string) {
  return 'Assess triage level for symptoms: ' + symptoms;
}
export function buildSafetyPrompt(input: string) {
  return 'Evaluate medical safety: ' + input;
}
