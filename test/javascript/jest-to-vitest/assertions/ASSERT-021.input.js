describe('EventEmitter', () => {
  it('should call the listener when an event is emitted', () => {
    const listener = jest.fn();
    const emitter = new EventEmitter();
    emitter.on('data', listener);
    emitter.emit('data');
    expect(listener).toHaveBeenCalled();
  });

  it('should call multiple listeners', () => {
    const first = jest.fn();
    const second = jest.fn();
    const emitter = new EventEmitter();
    emitter.on('update', first);
    emitter.on('update', second);
    emitter.emit('update');
    expect(first).toHaveBeenCalled();
    expect(second).toHaveBeenCalled();
  });

  it('should not call removed listeners', () => {
    const listener = jest.fn();
    const emitter = new EventEmitter();
    emitter.on('close', listener);
    emitter.off('close', listener);
    emitter.emit('close');
    expect(listener).not.toHaveBeenCalled();
  });
});
