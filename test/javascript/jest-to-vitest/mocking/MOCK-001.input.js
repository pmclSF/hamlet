describe('UserService', () => {
  it('should call the callback on success', () => {
    const callback = jest.fn();
    const service = new UserService();
    service.onSuccess(callback);
    service.execute();
    expect(callback).toHaveBeenCalled();
  });
});
