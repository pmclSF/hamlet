export interface FormField {
  name: string;
  type: 'text' | 'email' | 'password';
  required: boolean;
}

export function Form({ fields, onSubmit }: { fields: FormField[]; onSubmit: (data: Record<string, string>) => void }) {
  const inputs = fields.map(f => `<input name="${f.name}" type="${f.type}" ${f.required ? 'required' : ''} />`).join('');
  return `<form onsubmit="${onSubmit}">${inputs}<button type="submit">Submit</button></form>`;
}

export function validateField(value: string, type: string): boolean {
  if (type === 'email') return /^[^@]+@[^@]+\.[^@]+$/.test(value);
  if (type === 'password') return value.length >= 8;
  return value.length > 0;
}
