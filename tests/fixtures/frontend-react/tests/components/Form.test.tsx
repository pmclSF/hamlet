import { Form, validateField } from '../../src/components/Form';

describe('Form', () => {
  it('should match snapshot with text fields', () => {
    const fields = [
      { name: 'username', type: 'text' as const, required: true },
      { name: 'email', type: 'email' as const, required: true },
    ];
    const result = Form({ fields, onSubmit: () => {} });
    expect(result).toMatchSnapshot();
  });

  it('should match snapshot with password field', () => {
    const fields = [
      { name: 'password', type: 'password' as const, required: true },
    ];
    const result = Form({ fields, onSubmit: () => {} });
    expect(result).toMatchSnapshot();
  });

  it('should validate email format', () => {
    expect(validateField('user@example.com', 'email')).toBe(true);
    expect(validateField('invalid', 'email')).toBe(false);
  });

  it('should validate password length', () => {
    expect(validateField('short', 'password')).toBe(false);
    expect(validateField('longenough', 'password')).toBe(true);
  });
});
