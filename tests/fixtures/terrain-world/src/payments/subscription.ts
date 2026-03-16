export function createSubscription(userId: string, planId: string) {
  return { subscriptionId: 'sub_' + Date.now(), userId, planId, status: 'active' };
}

export function cancelSubscription(subscriptionId: string) {
  return { subscriptionId, status: 'cancelled' };
}

export function upgradeSubscription(subscriptionId: string, newPlanId: string) {
  return { subscriptionId, planId: newPlanId, status: 'active' };
}
