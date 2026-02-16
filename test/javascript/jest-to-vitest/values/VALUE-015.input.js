describe('deep equality with nested objects', () => {
  it('should compare deeply nested objects', () => {
    const actual = {
      user: {
        name: 'Alice',
        address: {
          street: '123 Main St',
          city: 'Springfield',
          state: 'IL',
          zip: '62701',
        },
        preferences: {
          theme: 'dark',
          notifications: {
            email: true,
            sms: false,
            push: {
              enabled: true,
              frequency: 'daily',
            },
          },
        },
      },
    };

    expect(actual).toEqual({
      user: {
        name: 'Alice',
        address: {
          street: '123 Main St',
          city: 'Springfield',
          state: 'IL',
          zip: '62701',
        },
        preferences: {
          theme: 'dark',
          notifications: {
            email: true,
            sms: false,
            push: {
              enabled: true,
              frequency: 'daily',
            },
          },
        },
      },
    });
  });

  it('should detect differences in deeply nested values', () => {
    const obj1 = { a: { b: { c: { d: 1 } } } };
    const obj2 = { a: { b: { c: { d: 2 } } } };
    expect(obj1).not.toEqual(obj2);
  });

  it('should handle arrays within nested objects', () => {
    const data = {
      items: [
        { id: 1, tags: ['a', 'b'] },
        { id: 2, tags: ['c'] },
      ],
    };
    expect(data).toEqual({
      items: [
        { id: 1, tags: ['a', 'b'] },
        { id: 2, tags: ['c'] },
      ],
    });
  });
});
