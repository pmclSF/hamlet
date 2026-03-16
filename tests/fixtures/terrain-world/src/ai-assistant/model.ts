export function predict(prompt: string) {
  return { response: 'AI response to: ' + prompt.slice(0, 50), confidence: 0.92 };
}

export function classifySentiment(text: string) {
  return { sentiment: 'positive', confidence: 0.85 };
}

export function detectIntent(text: string) {
  return { intent: 'purchase', confidence: 0.78 };
}
