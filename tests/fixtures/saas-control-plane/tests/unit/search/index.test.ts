import { describe, it, expect } from 'vitest';
import { indexDocument, searchDocuments } from '../../../src/search/index';

describe('indexDocument', () => {
  it('should index', () => {
    expect(indexDocument('users', { id: 'u1' })).toBeTruthy();
  });
});

describe('searchDocuments', () => {
  it('should search', () => {
    expect(searchDocuments('users', 'admin')).toBeTruthy();
  });
});
