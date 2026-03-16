export const merchandisingPrompt = "You are a product recommendation engine.";

export function buildRecommendationPrompt(userId: string, history: any[]) {
  return 'Recommend products for user ' + userId + ' based on ' + history.length + ' items';
}

export function buildSafetyPrompt(input: string) {
  return 'Check safety of recommendation: ' + input;
}
