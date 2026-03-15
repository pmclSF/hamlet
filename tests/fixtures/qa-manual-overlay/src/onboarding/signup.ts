export interface SignupData {
  email: string;
  password: string;
  name: string;
  acceptTerms: boolean;
}

export function validateSignup(data: SignupData): string[] {
  const errors: string[] = [];
  if (!data.email.includes('@')) errors.push('invalid email');
  if (data.password.length < 8) errors.push('password too short');
  if (!data.name.trim()) errors.push('name required');
  if (!data.acceptTerms) errors.push('must accept terms');
  return errors;
}

export function normalizeEmail(email: string): string {
  return email.toLowerCase().trim();
}
