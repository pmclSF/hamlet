import { describe, it, expect } from 'vitest';
import { splitDocument } from '../../src/rag/chunking';
describe('rag utils', () => { it('splits', () => { expect(splitDocument({})).toBeDefined(); }); });
