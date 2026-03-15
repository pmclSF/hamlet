import { Button } from '../../src/components/Button';

describe('Button', () => {
  it('should match snapshot for primary variant', () => {
    const result = Button({ label: 'Click me', onClick: () => {}, variant: 'primary' });
    expect(result).toMatchSnapshot();
  });

  it('should match snapshot for secondary variant', () => {
    const result = Button({ label: 'Cancel', onClick: () => {}, variant: 'secondary' });
    expect(result).toMatchSnapshot();
  });

  it('should match snapshot for disabled state', () => {
    const result = Button({ label: 'Disabled', onClick: () => {}, disabled: true });
    expect(result).toMatchSnapshot();
  });

  it('should contain the label text', () => {
    const result = Button({ label: 'Test', onClick: () => {} });
    expect(result).toContain('Test');
  });
});
