import { describe, it, expect } from 'vitest';
import { analyzerAction } from '../../../src/risk/analyzer';
import { connectDB, getAccount, cleanupDB } from '../../../src/shared/db';
describe('risk check', () => { it('should analyze', () => { connectDB(); analyzerAction('test'); cleanupDB(); }); });
