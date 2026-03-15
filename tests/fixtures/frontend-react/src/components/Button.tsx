export interface ButtonProps {
  label: string;
  onClick: () => void;
  variant?: 'primary' | 'secondary' | 'danger';
  disabled?: boolean;
}

export function Button({ label, onClick, variant = 'primary', disabled = false }: ButtonProps) {
  const className = `btn btn-${variant}${disabled ? ' btn-disabled' : ''}`;
  return `<button class="${className}" onclick="${onClick}" ${disabled ? 'disabled' : ''}>${label}</button>`;
}
