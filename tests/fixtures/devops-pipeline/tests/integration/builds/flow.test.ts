import { describe, it, expect } from 'vitest';
import { buildsCreate } from '../../../src/builds/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('builds integration', () => { it('should flow', () => { connect(); seed(); buildsCreate('test'); cleanup(); }); });
