import cypress from '../../../../src/languages/javascript/frameworks/cypress.js';
import {
  Assertion,
  Navigation,
  MockCall,
  TestFile,
  TestSuite,
  TestCase,
  Hook,
  walkIR,
} from '../../../../src/core/ir.js';

const { parse } = cypress;

/**
 * Helper: parse source and return all Assertion nodes from the IR tree.
 */
function parseAssertions(source) {
  const ir = parse(source);
  const nodes = [];
  walkIR(ir, (n) => {
    if (n instanceof Assertion) nodes.push(n);
  });
  return nodes;
}

describe('cypress.parse — assertion IR fields', () => {
  describe('.should() patterns', () => {
    it('should extract kind and subject from cy.get().should()', () => {
      const assertions = parseAssertions(
        "cy.get('#submit-btn').should('be.visible')"
      );
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('be.visible');
      expect(assertions[0].subject).toBe('#submit-btn');
      expect(assertions[0].expected).toBeNull();
      expect(assertions[0].isNegated).toBe(false);
    });

    it('should extract expected value from should with two args', () => {
      const assertions = parseAssertions(
        "cy.get('.message').should('contain', 'Hello world')"
      );
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('contain');
      expect(assertions[0].subject).toBe('.message');
      expect(assertions[0].expected).toBe("'Hello world'");
      expect(assertions[0].isNegated).toBe(false);
    });

    it('should handle have.length with numeric expected', () => {
      const assertions = parseAssertions(
        "cy.get('.item').should('have.length', 5)"
      );
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('have.length');
      expect(assertions[0].subject).toBe('.item');
      expect(assertions[0].expected).toBe('5');
    });

    it('should handle have.text', () => {
      const assertions = parseAssertions(
        "cy.get('#title').should('have.text', 'Welcome')"
      );
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('have.text');
      expect(assertions[0].subject).toBe('#title');
      expect(assertions[0].expected).toBe("'Welcome'");
    });

    it('should handle have.attr with key and value', () => {
      const assertions = parseAssertions(
        "cy.get('input').should('have.attr', 'type', 'email')"
      );
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('have.attr');
      expect(assertions[0].subject).toBe('input');
      // expected captures everything after the matcher
      expect(assertions[0].expected).toMatch(/type/);
    });

    it('should handle have.class', () => {
      const assertions = parseAssertions(
        "cy.get('#btn').should('have.class', 'active')"
      );
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('have.class');
      expect(assertions[0].expected).toBe("'active'");
    });

    it('should handle have.value', () => {
      const assertions = parseAssertions(
        "cy.get('#input').should('have.value', 'test')"
      );
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('have.value');
      expect(assertions[0].expected).toBe("'test'");
    });

    it('should handle be.checked', () => {
      const assertions = parseAssertions(
        "cy.get('#checkbox').should('be.checked')"
      );
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('be.checked');
      expect(assertions[0].expected).toBeNull();
    });

    it('should handle be.disabled', () => {
      const assertions = parseAssertions(
        "cy.get('#btn').should('be.disabled')"
      );
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('be.disabled');
    });

    it('should handle exist', () => {
      const assertions = parseAssertions("cy.get('#modal').should('exist')");
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('exist');
    });
  });

  describe('negated .should() patterns', () => {
    it('should detect not.be.visible', () => {
      const assertions = parseAssertions(
        "cy.get('#hidden').should('not.be.visible')"
      );
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('be.visible');
      expect(assertions[0].isNegated).toBe(true);
    });

    it('should detect not.exist', () => {
      const assertions = parseAssertions(
        "cy.get('#deleted').should('not.exist')"
      );
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('exist');
      expect(assertions[0].isNegated).toBe(true);
    });

    it('should detect not.contain', () => {
      const assertions = parseAssertions(
        "cy.get('.msg').should('not.contain', 'error')"
      );
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('contain');
      expect(assertions[0].isNegated).toBe(true);
      expect(assertions[0].expected).toBe("'error'");
    });
  });

  describe('chained selectors', () => {
    it('should combine cy.get().find() into subject', () => {
      const assertions = parseAssertions(
        "cy.get('#form').find('.input').should('be.visible')"
      );
      expect(assertions).toHaveLength(1);
      expect(assertions[0].subject).toBe('#form .input');
      expect(assertions[0].kind).toBe('be.visible');
    });
  });

  describe('cy.url() assertions', () => {
    it('should extract url.include', () => {
      const assertions = parseAssertions(
        "cy.url().should('include', '/dashboard')"
      );
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('url.include');
      expect(assertions[0].subject).toBe('cy.url()');
      expect(assertions[0].expected).toBe("'/dashboard'");
    });

    it('should extract url.equal', () => {
      const assertions = parseAssertions(
        "cy.url().should('eq', 'http://localhost/home')"
      );
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('url.equal');
      expect(assertions[0].subject).toBe('cy.url()');
      expect(assertions[0].expected).toBe("'http://localhost/home'");
    });
  });

  describe('cy.title() assertions', () => {
    it('should extract title.equal', () => {
      const assertions = parseAssertions("cy.title().should('eq', 'My App')");
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('title.equal');
      expect(assertions[0].subject).toBe('cy.title()');
      expect(assertions[0].expected).toBe("'My App'");
    });
  });

  describe('expect() patterns', () => {
    it('should extract expect().to.equal()', () => {
      const assertions = parseAssertions("expect(result).to.equal('success')");
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('equal');
      expect(assertions[0].subject).toBe('result');
      expect(assertions[0].expected).toBe("'success'");
      expect(assertions[0].isNegated).toBe(false);
    });

    it('should extract expect().to.be.true', () => {
      const assertions = parseAssertions('expect(isValid).to.be.true');
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('be.true');
      expect(assertions[0].subject).toBe('isValid');
      expect(assertions[0].isNegated).toBe(false);
    });

    it('should extract expect().to.be.false', () => {
      const assertions = parseAssertions('expect(isEmpty).to.be.false');
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('be.false');
      expect(assertions[0].subject).toBe('isEmpty');
    });

    it('should extract expect().to.be.null', () => {
      const assertions = parseAssertions('expect(value).to.be.null');
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('be.null');
      expect(assertions[0].subject).toBe('value');
    });

    it('should extract expect().to.be.undefined', () => {
      const assertions = parseAssertions('expect(result).to.be.undefined');
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('be.undefined');
    });

    it('should extract expect().to.include()', () => {
      const assertions = parseAssertions("expect(list).to.include('item')");
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('include');
      expect(assertions[0].subject).toBe('list');
      expect(assertions[0].expected).toBe("'item'");
    });

    it('should extract expect().to.contain()', () => {
      const assertions = parseAssertions("expect(text).to.contain('hello')");
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('include');
      expect(assertions[0].expected).toBe("'hello'");
    });

    it('should extract expect().to.have.property()', () => {
      const assertions = parseAssertions(
        "expect(obj).to.have.property('name')"
      );
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('have.property');
      expect(assertions[0].subject).toBe('obj');
      expect(assertions[0].expected).toBe("'name'");
    });
  });

  describe('negated expect() patterns', () => {
    it('should detect expect().to.not.equal()', () => {
      const assertions = parseAssertions(
        "expect(status).to.not.equal('error')"
      );
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('equal');
      expect(assertions[0].isNegated).toBe(true);
      expect(assertions[0].expected).toBe("'error'");
    });

    it('should detect expect().to.not.be.true', () => {
      const assertions = parseAssertions('expect(isAdmin).to.not.be.true');
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('be.true');
      expect(assertions[0].isNegated).toBe(true);
    });

    it('should detect expect().to.not.include()', () => {
      const assertions = parseAssertions(
        "expect(list).to.not.include('removed')"
      );
      expect(assertions).toHaveLength(1);
      expect(assertions[0].kind).toBe('include');
      expect(assertions[0].isNegated).toBe(true);
    });
  });

  describe('multi-line files', () => {
    it('should parse assertions among other code', () => {
      const source = `
describe('Login', () => {
  it('shows the form', () => {
    cy.visit('/login');
    cy.get('#form').should('be.visible');
    cy.get('#username').type('admin');
    cy.get('#password').type('secret');
    cy.get('#submit').click();
    cy.url().should('include', '/dashboard');
    expect(true).to.be.true;
  });
});`;
      const assertions = parseAssertions(source);
      expect(assertions).toHaveLength(3);
      expect(assertions[0].kind).toBe('be.visible');
      expect(assertions[0].subject).toBe('#form');
      expect(assertions[1].kind).toBe('url.include');
      expect(assertions[1].expected).toBe("'/dashboard'");
      expect(assertions[2].kind).toBe('be.true');
    });
  });

  describe('backward compatibility', () => {
    it('should still produce TestFile with body array', () => {
      const ir = parse("cy.get('#btn').should('exist')");
      expect(ir).toBeInstanceOf(TestFile);
      expect(Array.isArray(ir.body)).toBe(true);
    });

    it('should preserve sourceLocation on assertions', () => {
      const ir = parse("// comment\ncy.get('#btn').should('exist')");
      const assertion = ir.body.find((n) => n instanceof Assertion);
      expect(assertion.sourceLocation).toEqual({ line: 2, column: 0 });
    });

    it('should preserve originalSource on assertions', () => {
      const line = "    cy.get('#btn').should('be.visible')";
      const ir = parse(line);
      const assertion = ir.body.find((n) => n instanceof Assertion);
      expect(assertion.originalSource).toBe(line);
    });

    it('should set confidence to converted', () => {
      const ir = parse("cy.get('#btn').should('exist')");
      const assertion = ir.body.find((n) => n instanceof Assertion);
      expect(assertion.confidence).toBe('converted');
    });
  });

  describe('navigation parsing', () => {
    function parseNavigations(source) {
      const ir = parse(source);
      const nodes = [];
      walkIR(ir, (n) => {
        if (n instanceof Navigation) nodes.push(n);
      });
      return nodes;
    }

    it('should parse cy.visit() with string URL', () => {
      const navs = parseNavigations("cy.visit('/login');");
      expect(navs).toHaveLength(1);
      expect(navs[0].action).toBe('visit');
      expect(navs[0].url).toBe('/login');
    });

    it('should parse cy.visit() with double-quoted URL', () => {
      const navs = parseNavigations('cy.visit("/dashboard");');
      expect(navs).toHaveLength(1);
      expect(navs[0].url).toBe('/dashboard');
    });

    it('should parse cy.visit() with options', () => {
      const navs = parseNavigations(
        "cy.visit('/page', { timeout: 5000 });",
      );
      expect(navs).toHaveLength(1);
      expect(navs[0].url).toBe('/page');
      expect(navs[0].options).toBe('{ timeout: 5000 }');
    });

    it('should parse cy.visit() with variable', () => {
      const navs = parseNavigations('cy.visit(baseUrl);');
      expect(navs).toHaveLength(1);
      expect(navs[0].url).toBe('baseUrl');
    });

    it("should parse cy.go('back')", () => {
      const navs = parseNavigations("cy.go('back');");
      expect(navs).toHaveLength(1);
      expect(navs[0].action).toBe('goBack');
    });

    it("should parse cy.go('forward')", () => {
      const navs = parseNavigations("cy.go('forward');");
      expect(navs).toHaveLength(1);
      expect(navs[0].action).toBe('goForward');
    });

    it('should parse cy.go(-1) as goBack', () => {
      const navs = parseNavigations('cy.go(-1);');
      expect(navs).toHaveLength(1);
      expect(navs[0].action).toBe('goBack');
    });

    it('should parse cy.go(1) as goForward', () => {
      const navs = parseNavigations('cy.go(1);');
      expect(navs).toHaveLength(1);
      expect(navs[0].action).toBe('goForward');
    });

    it('should parse cy.reload()', () => {
      const navs = parseNavigations('cy.reload();');
      expect(navs).toHaveLength(1);
      expect(navs[0].action).toBe('reload');
    });

    it('should parse cy.reload(true)', () => {
      const navs = parseNavigations('cy.reload(true);');
      expect(navs).toHaveLength(1);
      expect(navs[0].action).toBe('reload');
      expect(navs[0].options).toBe('forceReload');
    });

    it('should preserve sourceLocation and originalSource', () => {
      const ir = parse("// comment\ncy.visit('/home');");
      const nav = ir.body.find((n) => n instanceof Navigation);
      expect(nav.sourceLocation.line).toBe(2);
      expect(nav.originalSource).toContain("cy.visit('/home')");
      expect(nav.confidence).toBe('converted');
    });

    it('should parse navigation among other nodes', () => {
      const source = `
describe('Nav', () => {
  it('navigates', () => {
    cy.visit('/home');
    cy.get('#link').click();
    cy.go('back');
    cy.reload();
  });
});`;
      const navs = parseNavigations(source);
      expect(navs).toHaveLength(3);
      expect(navs[0].action).toBe('visit');
      expect(navs[1].action).toBe('goBack');
      expect(navs[2].action).toBe('reload');
    });
  });

  describe('nested IR structure', () => {
    it('should extract TestSuite name from describe()', () => {
      const ir = parse("describe('My Suite', () => {\n});");
      expect(ir.body).toHaveLength(1);
      expect(ir.body[0]).toBeInstanceOf(TestSuite);
      expect(ir.body[0].name).toBe('My Suite');
    });

    it('should extract TestCase name from it()', () => {
      const source = `
describe('Suite', () => {
  it('should do something', () => {
    cy.get('#btn').should('exist');
  });
});`;
      const ir = parse(source);
      const suite = ir.body[0];
      expect(suite).toBeInstanceOf(TestSuite);
      expect(suite.tests).toHaveLength(1);
      expect(suite.tests[0]).toBeInstanceOf(TestCase);
      expect(suite.tests[0].name).toBe('should do something');
    });

    it('should extract hookType from hooks', () => {
      const source = `
describe('Suite', () => {
  beforeEach(() => {
    cy.visit('/app');
  });
  afterEach(() => {
    cy.clearCookies();
  });
});`;
      const ir = parse(source);
      const suite = ir.body[0];
      expect(suite.hooks).toHaveLength(2);
      expect(suite.hooks[0]).toBeInstanceOf(Hook);
      expect(suite.hooks[0].hookType).toBe('beforeEach');
      expect(suite.hooks[1].hookType).toBe('afterEach');
    });

    it('should map before() to beforeAll hookType', () => {
      const source = `
describe('Suite', () => {
  before(() => {
    cy.visit('/setup');
  });
  after(() => {
    cy.clearCookies();
  });
});`;
      const ir = parse(source);
      const suite = ir.body[0];
      expect(suite.hooks[0].hookType).toBe('beforeAll');
      expect(suite.hooks[1].hookType).toBe('afterAll');
    });

    it('should nest assertions inside TestCase body', () => {
      const source = `
describe('Suite', () => {
  it('checks', () => {
    cy.get('#btn').should('be.visible');
    cy.get('#input').should('have.value', 'test');
  });
});`;
      const ir = parse(source);
      const tc = ir.body[0].tests[0];
      expect(tc.body).toHaveLength(2);
      expect(tc.body[0]).toBeInstanceOf(Assertion);
      expect(tc.body[0].kind).toBe('be.visible');
      expect(tc.body[1]).toBeInstanceOf(Assertion);
      expect(tc.body[1].kind).toBe('have.value');
    });

    it('should nest navigation inside Hook body', () => {
      const source = `
describe('Suite', () => {
  beforeEach(() => {
    cy.visit('/app');
  });
  it('test', () => {
    cy.get('#btn').should('exist');
  });
});`;
      const ir = parse(source);
      const hook = ir.body[0].hooks[0];
      expect(hook.body).toHaveLength(1);
      expect(hook.body[0]).toBeInstanceOf(Navigation);
      expect(hook.body[0].action).toBe('visit');
    });

    it('should handle nested describe blocks', () => {
      const source = `
describe('Outer', () => {
  describe('Inner', () => {
    it('nested test', () => {
      cy.get('#x').should('exist');
    });
  });
});`;
      const ir = parse(source);
      const outer = ir.body[0];
      expect(outer.name).toBe('Outer');
      expect(outer.tests).toHaveLength(1);
      const inner = outer.tests[0];
      expect(inner).toBeInstanceOf(TestSuite);
      expect(inner.name).toBe('Inner');
      expect(inner.tests).toHaveLength(1);
      expect(inner.tests[0].name).toBe('nested test');
    });

    it('should detect .only modifier on describe', () => {
      const ir = parse("describe.only('Focused', () => {\n});");
      expect(ir.body[0].modifiers).toHaveLength(1);
      expect(ir.body[0].modifiers[0].modifierType).toBe('only');
    });

    it('should detect .skip modifier on it', () => {
      const source = `
describe('Suite', () => {
  it.skip('skipped test', () => {
    cy.get('#x').should('exist');
  });
});`;
      const ir = parse(source);
      const tc = ir.body[0].tests[0];
      expect(tc.modifiers).toHaveLength(1);
      expect(tc.modifiers[0].modifierType).toBe('skip');
    });
  });

  describe('MockCall parsing', () => {
    function parseMocks(source) {
      const ir = parse(source);
      const nodes = [];
      walkIR(ir, (n) => {
        if (n instanceof MockCall) nodes.push(n);
      });
      return nodes;
    }

    it('should parse cy.intercept with 3 args as networkIntercept', () => {
      const mocks = parseMocks(
        "cy.intercept('GET', '/api/users', { fixture: 'users.json' }).as('getUsers');"
      );
      expect(mocks).toHaveLength(1);
      expect(mocks[0].kind).toBe('networkIntercept');
      expect(mocks[0].target).toBe('/api/users');
      expect(mocks[0].args).toEqual(['GET']);
      expect(mocks[0].returnValue).toBe("{ fixture: 'users.json' }");
    });

    it('should parse cy.intercept with 2 args as networkIntercept spy', () => {
      const mocks = parseMocks(
        "cy.intercept('GET', '/api/data').as('getData');"
      );
      expect(mocks).toHaveLength(1);
      expect(mocks[0].kind).toBe('networkIntercept');
      expect(mocks[0].target).toBe('/api/data');
      expect(mocks[0].args).toEqual(['GET']);
      expect(mocks[0].returnValue).toBeNull();
    });

    it('should parse cy.intercept with 1 arg as bare spy', () => {
      const mocks = parseMocks("cy.intercept('/api/health');");
      expect(mocks).toHaveLength(1);
      expect(mocks[0].kind).toBe('networkIntercept');
      expect(mocks[0].target).toBe('/api/health');
      expect(mocks[0].args).toEqual([]);
    });

    it('should parse cy.stub as createStub', () => {
      const mocks = parseMocks('cy.stub(win, "open");');
      expect(mocks).toHaveLength(1);
      expect(mocks[0].kind).toBe('createStub');
    });

    it('should parse cy.spy as createSpy', () => {
      const mocks = parseMocks('cy.spy(console, "log");');
      expect(mocks).toHaveLength(1);
      expect(mocks[0].kind).toBe('createSpy');
    });

    it('should parse cy.clock as fakeTimers', () => {
      const mocks = parseMocks('cy.clock();');
      expect(mocks).toHaveLength(1);
      expect(mocks[0].kind).toBe('fakeTimers');
    });

    it('should parse cy.tick as advanceTimers', () => {
      const mocks = parseMocks('cy.tick(5000);');
      expect(mocks).toHaveLength(1);
      expect(mocks[0].kind).toBe('advanceTimers');
      expect(mocks[0].args).toEqual(['5000']);
    });

    it('should nest MockCall inside TestCase body', () => {
      const source = `
describe('API', () => {
  it('intercepts request', () => {
    cy.intercept('GET', '/api/users', { fixture: 'users.json' }).as('getUsers');
    cy.visit('/users');
  });
});`;
      const ir = parse(source);
      const tc = ir.body[0].tests[0];
      expect(tc.body).toHaveLength(2);
      expect(tc.body[0]).toBeInstanceOf(MockCall);
      expect(tc.body[0].kind).toBe('networkIntercept');
    });
  });
});
