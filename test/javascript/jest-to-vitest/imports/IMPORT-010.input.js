import { render as customRender } from '@testing-library/react';

describe('Component', () => {
  it('renders with custom render', () => {
    const { container } = customRender('<div />');
    expect(container).toBeDefined();
  });

  it('renders content inside the container', () => {
    const { container } = customRender('<span>Hello</span>');
    expect(container.textContent).toBe('Hello');
  });

  it('returns a rerender function', () => {
    const { rerender } = customRender('<div />');
    expect(typeof rerender).toBe('function');
  });
});
