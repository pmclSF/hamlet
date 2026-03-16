export const riskSystemPrompt = "You are a trading risk assessment model.";
export function buildRiskPrompt(position: any) { return 'Assess risk: ' + JSON.stringify(position); }
export function buildSafetyPrompt(input: string) { return 'Safety check: ' + input; }
