import { describe, it, expect, vi } from 'vitest';
import { render } from './render';

// HAMLET-WARNING: vi.mock is hoisted like jest.mock, but factory function scoping differs. Variables defined above vi.mock() are not accessible inside the factory. Verify this mock works correctly.
// Original: vi.mock('./render');
vi.mock('./render');

describe('App', () => {
  it('uses the mock', () => {
    render();
    expect(render).toHaveBeenCalled();
  });
});
