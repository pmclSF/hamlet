import { getUser } from '../shared/db-helper';

export function analyzeRisk(userId: string, amount: number) {
  const score = amount > 5000 ? 0.85 : 0.15;
  return { userId, riskScore: score, flagged: score > 0.7 };
}

export function checkVelocity(userId: string) {
  return { userId, txPerHour: 2, flagged: false };
}

export function reportFraud(transactionId: string, reason: string) {
  return { transactionId, reported: true, reason };
}
