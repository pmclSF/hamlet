export function createSubscription(orgId: string, planId: string) {
  return { subscriptionId: 'sub_' + Date.now(), orgId, planId, status: 'active' };
}

export function cancelSubscription(subscriptionId: string) {
  return { subscriptionId, status: 'cancelled' };
}

export function changeplan(subscriptionId: string, newPlanId: string) {
  return { subscriptionId, planId: newPlanId, status: 'active' };
}
