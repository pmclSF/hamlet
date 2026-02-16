jest.mock('./utils', () => {
  const actual = jest.requireActual('./utils');
  return {
    ...actual,
    formatDate: jest.fn(() => '2024-01-01'),
  };
});

describe('Utils', () => {
  it('uses real helpers but mocked formatDate', () => {
    expect(formatDate()).toBe('2024-01-01');
    expect(parseInput('test')).toBeDefined();
  });
});
