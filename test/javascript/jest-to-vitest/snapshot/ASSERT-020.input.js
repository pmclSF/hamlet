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
