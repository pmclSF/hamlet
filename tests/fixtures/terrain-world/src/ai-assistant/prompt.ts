export const systemPrompt = "You are a helpful shopping assistant.";

export function buildUserPrompt(query: string, context: any) {
  return `User asks: ${query}\nContext: ${JSON.stringify(context)}`;
}

export function buildSafetyPrompt(input: string) {
  return `Evaluate safety of: ${input}`;
}
