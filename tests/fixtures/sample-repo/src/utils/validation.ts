export function validateEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

export function validatePassword(password: string): boolean {
  return password.length >= 8;
}

export function sanitizeInput(input: string): string {
  return input.replace(/[<>&"']/g, '');
}
