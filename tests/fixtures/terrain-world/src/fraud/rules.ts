export function evaluateRules(transaction: any) {
  const rules = ['amount_check', 'velocity_check', 'geo_check'];
  return rules.map(r => ({ rule: r, passed: true }));
}

export function addRule(name: string, condition: string) {
  return { name, condition, active: true };
}

export function disableRule(name: string) {
  return { name, active: false };
}
