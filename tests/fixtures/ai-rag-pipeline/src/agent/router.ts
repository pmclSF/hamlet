export function agentRouter(input) { return "search"; }
export const fallbackStrategy = { model: "gpt-3.5-turbo", maxRetries: 3 };
export const stepBudget = { maxSteps: 10, maxTokens: 4000 };
