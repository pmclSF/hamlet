import { TypeStripper } from '../../src/core/TypeStripper.js';

describe('TypeStripper', () => {
  describe('strip', () => {
    it('should remove import type statements', () => {
      const input = `import type { Foo, Bar } from './types';
import { test, expect } from '@playwright/test';`;
      const result = TypeStripper.strip(input);
      expect(result).not.toContain('import type');
      expect(result).toContain("import { test, expect } from '@playwright/test'");
    });

    it('should remove interface declarations', () => {
      const input = `interface LoginPage {
  username: string;
  password: string;
}

describe('test', () => {});`;
      const result = TypeStripper.strip(input);
      expect(result).not.toContain('interface LoginPage');
      expect(result).toContain("describe('test'");
    });

    it('should remove type alias declarations', () => {
      const input = `type Framework = 'cypress' | 'playwright';
const fw = 'cypress';`;
      const result = TypeStripper.strip(input);
      expect(result).not.toContain('type Framework');
      expect(result).toContain("const fw = 'cypress'");
    });

    it('should remove parameter type annotations', () => {
      const input = `function convert(content: string, options: ConvertOptions) {`;
      const result = TypeStripper.strip(input);
      expect(result).toContain('function convert(content, options)');
    });

    it('should remove return type annotations', () => {
      const input = `const fn = (x: number): string => {`;
      const result = TypeStripper.strip(input);
      expect(result).not.toContain(': string');
      expect(result).not.toContain(': number');
    });

    it('should remove as Type assertions', () => {
      const input = `const el = document.querySelector('#btn') as HTMLElement;`;
      const result = TypeStripper.strip(input);
      expect(result).not.toContain('as HTMLElement');
      expect(result).toContain("document.querySelector('#btn')");
    });

    it('should preserve as const', () => {
      const input = `const FRAMEWORKS = ['cypress', 'playwright'] as const;`;
      const result = TypeStripper.strip(input);
      expect(result).toContain('as const');
    });

    it('should remove generic type params on calls', () => {
      const input = `const result = convert<Framework>('test');`;
      const result = TypeStripper.strip(input);
      expect(result).toContain("convert('test')");
      expect(result).not.toContain('<Framework>');
    });

    it('should remove non-null assertions', () => {
      const input = `const value = element!.textContent;`;
      const result = TypeStripper.strip(input);
      expect(result).toContain('element.textContent');
      expect(result).not.toContain('element!.');
    });

    it('should remove variable type annotations', () => {
      const input = `const name: string = 'test';
let count: number = 0;`;
      const result = TypeStripper.strip(input);
      expect(result).toContain("const name = 'test'");
      expect(result).toContain('let count = 0');
    });

    it('should preserve non-TS code unchanged', () => {
      const input = `describe('test', () => {
  it('works', () => {
    cy.visit('/');
    cy.get('#btn').click();
  });
});`;
      const result = TypeStripper.strip(input);
      expect(result).toBe(input);
    });
  });

  describe('hasTypeAnnotations', () => {
    it('should detect interface declarations', () => {
      expect(
        TypeStripper.hasTypeAnnotations('interface Foo { bar: string; }')
      ).toBe(true);
    });

    it('should detect type aliases', () => {
      expect(
        TypeStripper.hasTypeAnnotations("type Foo = 'bar' | 'baz';")
      ).toBe(true);
    });

    it('should detect import type', () => {
      expect(
        TypeStripper.hasTypeAnnotations("import type { Foo } from './foo';")
      ).toBe(true);
    });

    it('should detect primitive type annotations', () => {
      expect(
        TypeStripper.hasTypeAnnotations('function foo(x: string) {}')
      ).toBe(true);
    });

    it('should return false for plain JS', () => {
      expect(
        TypeStripper.hasTypeAnnotations(
          "describe('test', () => { it('works', () => {}); });"
        )
      ).toBe(false);
    });

    it('should detect as Type assertions', () => {
      expect(
        TypeStripper.hasTypeAnnotations('const el = x as HTMLElement;')
      ).toBe(true);
    });

    it('should detect generic type params', () => {
      expect(
        TypeStripper.hasTypeAnnotations("convert<Framework>('test')")
      ).toBe(true);
    });
  });
});
