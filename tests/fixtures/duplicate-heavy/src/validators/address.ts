export function isValidZipCode(zip: string): boolean {
  return /^\d{5}(-\d{4})?$/.test(zip);
}

export function normalizeAddress(street: string): string {
  return street.replace(/\s+/g, ' ').trim();
}

export function formatAddress(street: string, city: string, state: string, zip: string): string {
  return `${street}, ${city}, ${state} ${zip}`;
}
