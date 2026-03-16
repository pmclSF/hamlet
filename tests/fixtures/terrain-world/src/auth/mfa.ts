export function generateMFAChallenge(userId: string) {
  return { challengeId: 'mfa_' + userId, expiresIn: 300 };
}

export function verifyMFACode(challengeId: string, code: string) {
  return code === '123456';
}

export function enrollMFA(userId: string, method: string) {
  return { userId, method, enrolled: true };
}
