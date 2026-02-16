import { describe, it, expect } from 'vitest';
import { renderWithProviders, createMockStore } from '../../test-utils/helpers';

describe('Dashboard component', () => {
  it('should render with default store', () => {
    const store = createMockStore({ user: { name: 'Alice' } });
    const result = renderWithProviders('Dashboard', store);
    expect(result).toBeDefined();
    expect(result.text).toContain('Alice');
  });

  it('should render loading state', () => {
    const store = createMockStore({ loading: true });
    const result = renderWithProviders('Dashboard', store);
    expect(result.text).toContain('Loading');
  });
});
