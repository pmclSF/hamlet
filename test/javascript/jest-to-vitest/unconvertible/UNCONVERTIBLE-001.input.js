const mySerializer = {
  test: (val) => typeof val === 'string',
  print: (val) => `"${val}"`,
};

expect.addSnapshotSerializer(mySerializer);

describe('Custom serializer', () => {
  it('uses custom serializer', () => {
    expect('hello').toMatchSnapshot();
  });

  it('serializes another string', () => {
    expect('world').toMatchSnapshot();
  });
});
