export const systemPrompt = "You are a helpful edtech-lms assistant.";
export function buildPrompt(input: string) { return 'Process: ' + input; }
export function buildSafetyPrompt(input: string) { return 'Safety: ' + input; }
