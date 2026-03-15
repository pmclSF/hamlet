export const validEmail = 'user@example.com';
export const validPhone = '+1-555-123-4567';
export const validZip = '94105';
export const validStreet = '123 Main St';
export const validCity = 'San Francisco';
export const validState = 'CA';

export function createValidInput(type: string) {
  switch (type) {
    case 'email': return validEmail;
    case 'phone': return validPhone;
    case 'zip': return validZip;
    default: return '';
  }
}
