/*
 * MIT License
 *
 * Copyright (c) 2024 Acme Corp
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software.
 */

import { describe, it, expect } from 'vitest';

describe('licensed module', () => {
  it('should perform basic arithmetic', () => {
    expect(2 + 2).toBe(4);
  });

  it('should concatenate strings', () => {
    expect('hello' + ' ' + 'world').toBe('hello world');
  });
});
