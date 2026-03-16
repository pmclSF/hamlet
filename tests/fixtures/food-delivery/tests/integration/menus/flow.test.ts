import { describe, it, expect } from 'vitest';
import { menusCreate } from '../../../src/menus/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('menus integration', () => { it('should flow', () => { connect(); seed(); menusCreate('test'); cleanup(); }); });
