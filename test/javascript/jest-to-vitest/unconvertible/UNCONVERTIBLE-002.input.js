jest.mock('virtual-module', () => ({ key: 'value' }), { virtual: true });

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
