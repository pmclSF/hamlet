// HAMLET-WARNING: Snapshot file location and format may differ between
// Jest (__snapshots__/*.snap) and Vitest. Run `vitest --update` to
// regenerate snapshots after migration.
import { describe, it, expect } from 'vitest';

describe('Serializer', () => {
  it('matches snapshot', () => {
    const output = serialize({ name: 'Alice', age: 30 });
    expect(output).toMatchSnapshot();
  });

  it('matches inline snapshot', () => {
    const output = serialize({ name: 'Bob' });
    expect(output).toMatchInlineSnapshot(`"{\\"name\\":\\"Bob\\"}"`);
  });
});
