import { describe, it, expect } from 'vitest';

const mySerializer = {
  test: (val) => typeof val === 'string',
  print: (val) => `"${val}"`,
};

// HAMLET-TODO [UNCONVERTIBLE-SNAPSHOT-SERIALIZER]: Vitest snapshot serializers are configured in vitest.config.ts, not in test files
// Original: expect.addSnapshotSerializer(mySerializer);
// Manual action required: Move serializer config to vitest.config.ts snapshotSerializers option
// expect.addSnapshotSerializer(mySerializer);

// HAMLET-WARNING: Snapshot file location and format may differ between
// Jest (__snapshots__/*.snap) and Vitest. Run `vitest --update` to
// regenerate snapshots after migration.
describe('Custom serializer', () => {
  it('uses custom serializer', () => {
    expect('hello').toMatchSnapshot();
  });

  it('serializes another string', () => {
    expect('world').toMatchSnapshot();
  });
});
