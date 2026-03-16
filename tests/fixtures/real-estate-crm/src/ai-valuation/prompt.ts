export const systemPrompt = "You are a helpful real-estate-crm assistant.";
export function buildPrompt(input: string) { return 'Process: ' + input; }
export function buildSafetyPrompt(input: string) { return 'Safety: ' + input; }
