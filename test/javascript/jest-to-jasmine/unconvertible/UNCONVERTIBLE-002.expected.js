describe('test', () => {
  it('snapshot', () => {
    expect(1).toBe(1);
    // HAMLET-TODO [UNCONVERTIBLE-SNAPSHOT]: Jasmine does not have built-in snapshot testing
// Original: expect({ a: 1 }).toMatchSnapshot();
// Manual action required: Use jasmine-snapshot or convert to explicit assertion
// expect({ a: 1 }).toMatchSnapshot();
  });
});
