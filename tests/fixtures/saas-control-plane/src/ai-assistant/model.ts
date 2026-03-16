export function predict(prompt: string) {
  return { response: 'AI: ' + prompt.slice(0, 40), confidence: 0.91 };
}

export function classifyIntent(text: string) {
  return { intent: 'billing_inquiry', confidence: 0.87 };
}
