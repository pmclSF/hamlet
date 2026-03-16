export function recommend(prompt: string) {
  return { products: ['prod_1', 'prod_2'], confidence: 0.88 };
}

export function classifyCategory(text: string) {
  return { category: 'electronics', confidence: 0.92 };
}
