export const systemPrompt = "You are a document QA assistant.";
export function buildPrompt(query, context) { return query + context; }
