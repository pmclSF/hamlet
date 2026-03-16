export function analyzeTransaction(transactionId: string, amount: number) {
  const riskScore = amount > 10000 ? 0.9 : amount > 1000 ? 0.5 : 0.1;
  return { transactionId, riskScore, flagged: riskScore > 0.7 };
}

export function checkVelocity(userId: string) {
  return { userId, transactionsPerHour: 3, flagged: false };
}

export function reportFraud(transactionId: string, evidence: string) {
  return { transactionId, reported: true, evidence };
}
