export function welcomeMessage(name: string, plan: string): string {
  return `Welcome, ${name}! You're on the ${plan} plan.`;
}

export function trialDaysRemaining(signupDate: Date, trialLengthDays: number): number {
  const now = new Date();
  const elapsed = Math.floor((now.getTime() - signupDate.getTime()) / (1000 * 60 * 60 * 24));
  return Math.max(0, trialLengthDays - elapsed);
}
