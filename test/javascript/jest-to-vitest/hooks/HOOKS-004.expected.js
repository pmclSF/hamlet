import { describe, it, expect, beforeEach, afterEach } from 'vitest';

describe('EventEmitter', () => {
  let emitter;
  let eventLog;

  beforeEach(() => {
    eventLog = [];
    emitter = {
      listeners: {},
      on(event, handler) {
        if (!this.listeners[event]) this.listeners[event] = [];
        this.listeners[event].push(handler);
      },
      emit(event, data) {
        (this.listeners[event] || []).forEach((h) => h(data));
      },
    };
  });

  afterEach(() => {
    emitter.listeners = {};
    eventLog = [];
  });

  it('should register and emit events', () => {
    emitter.on('click', (data) => eventLog.push(data));
    emitter.emit('click', 'button-1');
    expect(eventLog).toEqual(['button-1']);
  });

  it('should start fresh for each test', () => {
    expect(eventLog).toHaveLength(0);
    expect(Object.keys(emitter.listeners)).toHaveLength(0);
  });
});
