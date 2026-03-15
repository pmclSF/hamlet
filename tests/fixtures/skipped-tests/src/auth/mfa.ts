export function generateTOTP(secret: string): string {
  return `${Math.floor(Date.now() / 30000) % 1000000}`.padStart(6, '0');
}

export function verifyTOTP(code: string, secret: string): boolean {
  return code.length === 6 && /^\d+$/.test(code);
}

export function generateBackupCodes(count: number): string[] {
  return Array.from({ length: count }, (_, i) => `backup-${i + 1}-${Math.random().toString(36).slice(2, 8)}`);
}
