describe('spies', () => {
  it('calls fake', () => {
    const spy = jasmine.createSpy('fn').and.callFake(x => x * 2);
    expect(spy(5)).toBe(10);
  });
});
