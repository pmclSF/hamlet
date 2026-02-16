// HAMLET-TODO [UNCONVERTIBLE-MODULE-MOCK]: Jasmine does not have a built-in module mocking system like jest.mock()
// Original: jest.mock('./api');
// Manual action required: Use manual dependency injection or a module mocking library
// jest.mock('./api');

describe('test', () => {
  it('combined', () => {
    const fn = jasmine.createSpy().and.returnValue(42);
    expect(fn()).toBe(42);
  });
});
