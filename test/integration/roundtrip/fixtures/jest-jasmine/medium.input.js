import { describe, it, expect, beforeEach, jest } from '@jest/globals';
import { EventBus } from '../../src/event-bus.js';

describe('EventBus', () => {
  let bus;

  beforeEach(() => {
    bus = new EventBus();
  });

  it('should register a listener for an event', () => {
    const handler = jest.fn();
    bus.on('user:login', handler);
    expect(bus.listenerCount('user:login')).toBe(1);
  });

  it('should invoke the handler when an event is emitted', () => {
    const handler = jest.fn();
    bus.on('user:login', handler);
    bus.emit('user:login', { userId: 42 });
    expect(handler).toHaveBeenCalledWith({ userId: 42 });
  });

  it('should support multiple listeners on the same event', () => {
    const first = jest.fn();
    const second = jest.fn();
    bus.on('data:refresh', first);
    bus.on('data:refresh', second);
    bus.emit('data:refresh', null);
    expect(first).toHaveBeenCalledTimes(1);
    expect(second).toHaveBeenCalledTimes(1);
  });

  describe('unsubscribe', () => {
    it('should remove a specific listener', () => {
      const handler = jest.fn();
      const off = bus.on('click', handler);
      off();
      bus.emit('click');
      expect(handler).not.toHaveBeenCalled();
    });

    it('should not affect other listeners when one is removed', () => {
      const kept = jest.fn();
      const removed = jest.fn();
      bus.on('click', kept);
      const off = bus.on('click', removed);
      off();
      bus.emit('click');
      expect(kept).toHaveBeenCalledTimes(1);
      expect(removed).not.toHaveBeenCalled();
    });
  });
});
