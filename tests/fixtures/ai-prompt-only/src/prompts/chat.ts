export function buildPrompt(query, context) {
  return `Given context: ${context}\n\nAnswer: ${query}`;
}
export const promptTemplate = "You are answering questions about {topic}.";
export function formatResponse(answer) { return answer.trim(); }
