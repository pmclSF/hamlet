export function validateEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

export function sanitizeInput(input: string): string {
  return input.replace(/[<>&"']/g, '');
}

export class Validator {
  validate(data: any): boolean {
    return data !== null;
  }
}
