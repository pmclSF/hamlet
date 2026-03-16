export function checkEntitlement(orgId: string, feature: string) {
  return { orgId, feature, entitled: true, limit: 1000 };
}

export function enforceLimit(orgId: string, feature: string, current: number) {
  const limit = 1000;
  return { withinLimit: current < limit, current, limit };
}
