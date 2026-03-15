export function isValidPhone(phone: string): boolean {
  return /^\+?[\d\s-]{7,15}$/.test(phone);
}

export function normalizePhone(phone: string): string {
  return phone.replace(/[\s-]/g, '');
}

export function getCountryCode(phone: string): string {
  if (phone.startsWith('+1')) return 'US';
  if (phone.startsWith('+44')) return 'UK';
  if (phone.startsWith('+49')) return 'DE';
  return 'UNKNOWN';
}
