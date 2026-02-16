// HAMLET-TODO [UNCONVERTIBLE-VIRTUAL-MOCK]: Vitest does not support { virtual: true } option in vi.mock()
// Original: vi.mock('virtual-module', () => ({ key: 'value' }), { virtual: true })
// Manual action required: Create the module file or use vi.mock with a manual factory
import { describe, it, expect, vi } from 'vitest';

vi.mock('virtual-module', () => ({ key: 'value' }));

import { key } from 'virtual-module';

describe('Virtual module', () => {
  it('imports the virtual module', () => {
    expect(key).toBe('value');
  });

  it('uses virtual module in computation', () => {
    const result = `prefix-${key}`;
    expect(result).toBe('prefix-value');
  });
});
