jest.mock('./utils', () => ({
  ...jest.requireActual('./utils'),
  formatDate: jest.fn(() => '2024-01-01'),
  generateId: jest.fn(() => 'mock-id-123'),
}));

const { formatDate, generateId, parseInput } = require('./utils');

describe('Partial module mock', () => {
  it('uses mocked formatDate', () => {
    expect(formatDate(new Date())).toBe('2024-01-01');
  });

  it('uses mocked generateId', () => {
    expect(generateId()).toBe('mock-id-123');
  });

  it('uses real parseInput', () => {
    expect(parseInput('hello')).toBeDefined();
  });
});
