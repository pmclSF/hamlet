export const systemPrompt = "You are a helpful gaming-backend assistant.";
export function buildPrompt(input: string) { return 'Process: ' + input; }
export function buildSafetyPrompt(input: string) { return 'Safety: ' + input; }
