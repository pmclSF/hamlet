import { describe, it, expect, vi } from 'vitest';

describe('EventEmitter', () => {
  it('should call the listener when an event is emitted', () => {
    const listener = vi.fn();
    const emitter = new EventEmitter();
    emitter.on('data', listener);
    emitter.emit('data');
    expect(listener).toHaveBeenCalled();
  });

  it('should call multiple listeners', () => {
    const first = vi.fn();
    const second = vi.fn();
    const emitter = new EventEmitter();
    emitter.on('update', first);
    emitter.on('update', second);
    emitter.emit('update');
    expect(first).toHaveBeenCalled();
    expect(second).toHaveBeenCalled();
  });

  it('should not call removed listeners', () => {
    const listener = vi.fn();
    const emitter = new EventEmitter();
    emitter.on('close', listener);
    emitter.off('close', listener);
    emitter.emit('close');
    expect(listener).not.toHaveBeenCalled();
  });
});
