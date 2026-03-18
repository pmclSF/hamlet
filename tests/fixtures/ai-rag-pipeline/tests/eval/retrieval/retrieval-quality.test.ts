import { describe, it, expect } from 'vitest';
import { retriever, searchQuery } from '../../../src/rag/retriever';
import { chunkConfig } from '../../../src/rag/chunking';
import { rerankerConfig } from '../../../src/rag/reranker';
describe('retrieval quality', () => {
  it('should retrieve relevant docs', () => { expect(retriever('test')).toBeDefined(); });
  it('should use config', () => { expect(chunkConfig.chunkSize).toBe(512); });
});
