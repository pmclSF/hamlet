import { describe, it, expect } from 'vitest';
import { logger } from '../../src/utils';
describe('utils', () => { it('should log', () => { expect(logger).toBeDefined(); }); });
