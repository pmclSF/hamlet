export const systemPrompt = "You are a helpful logistics-platform assistant.";
export function buildPrompt(input: string) { return 'Process: ' + input; }
export function buildSafetyPrompt(input: string) { return 'Safety: ' + input; }
