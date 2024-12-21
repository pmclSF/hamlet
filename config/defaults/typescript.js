/**
 * TypeScript conversion and configuration settings
 */
export const typescriptConfig = {
    /**
     * General TypeScript settings
     */
    settings: {
      enabled: true,
      strict: true,
      generateTypes: true,
      inferTypes: true,
      preserveComments: true
    },
  
    /**
     * Compiler options
     */
    compiler: {
      target: 'ES2020',
      module: 'ESNext',
      moduleResolution: 'node',
      esModuleInterop: true,
      allowJs: true,
      checkJs: false,
      jsx: 'react',
      declaration: true,
      sourceMap: true,
      outDir: './dist',
      baseUrl: '.',
      paths: {
        '@/*': ['src/*']
      }
    },
  
    /**
     * Type definitions
     */
    types: {
      /**
       * Common type mappings from Cypress to Playwright
       */
      mappings: {
        // Cypress namespace types
        'Cypress.Chainable': 'Locator',
        'Cypress.Element': 'ElementHandle',
        'Cypress.Browser': 'BrowserContext',
        'Cypress.Config': 'PlaywrightTestConfig',
        'Cypress.Viewport': 'ViewportSize',
        'Cypress.Cookie': 'Cookie',
        'Cypress.Response': 'APIResponse',
        'Cypress.AUTWindow': 'Page',
  
        // Command return types
        'cy.get': 'Locator',
        'cy.find': 'Locator',
        'cy.contains': 'Locator',
        'cy.wait': 'void',
        'cy.url': 'string',
        'cy.title': 'string',
        'cy.location': 'Location',
  
        // Custom types
        'cy.mount': 'Promise<void>',
        'cy.intercept': 'Route',
        'cy.request': 'APIRequestContext'
      },
  
      /**
       * Interface definitions
       */
      interfaces: {
        // Test context interface
        TestContext: `
  interface TestContext {
    page: Page;
    context: BrowserContext;
    request: APIRequestContext;
  }`,
  
        // Test fixture interface
        TestFixture: `
  interface TestFixture {
    name: string;
    data: any;
    encoding?: string;
  }`,
  
        // Test options interface
        TestOptions: `
  interface TestOptions {
    viewport?: ViewportSize;
    locale?: string;
    timezone?: string;
    permissions?: string[];
    geolocation?: { latitude: number; longitude: number };
  }`
      },
  
      /**
       * Type generation settings
       */
      generation: {
        outputDir: './types',
        declarationFiles: true,
        mergeDeclarations: true,
        addJSDoc: true
      }
    },
  
    /**
     * Type assertion settings
     */
    assertions: {
      enabled: true,
      strictNullChecks: true,
      noImplicitAny: true,
      checkReturnTypes: true
    },
  
    /**
     * Import/Export handling
     */
    modules: {
      // Import preferences
      imports: {
        preferTypeImports: true,
        addPlaywrightTypes: true,
        organizeImports: true,
        removeUnused: true
      },
  
      // Export preferences
      exports: {
        addTypeAnnotations: true,
        generateDeclarations: true,
        preserveJSDoc: true
      }
    },
  
    /**
     * Helper functions
     */
    helpers: {
      /**
       * Convert Cypress type to Playwright type
       * @param {string} cypressType - Cypress type name
       * @returns {string} - Playwright type name
       */
      convertType(cypressType) {
        return this.types.mappings[cypressType] || cypressType;
      },
  
      /**
       * Generate TypeScript declaration
       * @param {Object} component - Component information
       * @returns {string} - TypeScript declaration
       */
      generateDeclaration(component) {
        const { name, props, methods } = component;
        
        const propsInterface = props ? `
  interface ${name}Props {
    ${props.map(prop => `${prop.name}${prop.optional ? '?' : ''}: ${prop.type};`).join('\n  ')}
  }` : '';
  
        const methodSignatures = methods ? 
          methods.map(method => 
            `${method.name}(${method.params.map(p => `${p.name}: ${p.type}`).join(', ')}): ${method.returnType};`
          ).join('\n  ') : '';
  
        return `
  ${propsInterface}
  
  export class ${name} {
    constructor(props: ${name}Props);
    ${methodSignatures}
  }`;
      },
  
      /**
       * Generate JSDoc comment
       * @param {Object} metadata - Method metadata
       * @returns {string} - JSDoc comment
       */
      generateJSDoc(metadata) {
        const { description, params, returns, example } = metadata;
        
        const paramDocs = params ? 
          params.map(p => `@param {${p.type}} ${p.name} - ${p.description}`).join('\n * ') : '';
        
        return `
  /**
   * ${description}
   * ${paramDocs}
   * @returns {${returns.type}} ${returns.description}
   * ${example ? `@example\n * ${example}` : ''}
   */`;
      },
  
      /**
       * Check if type exists
       * @param {string} type - Type name to check
       * @returns {boolean} - Whether type exists
       */
      typeExists(type) {
        return Object.values(this.types.mappings).includes(type) ||
               Object.keys(this.types.interfaces).includes(type);
      },
  
      /**
       * Convert function to TypeScript
       * @param {Object} func - Function information
       * @returns {string} - TypeScript function
       */
      convertFunction(func) {
        const { name, params, returnType, async } = func;
        const convertedParams = params.map(p => `${p.name}: ${this.convertType(p.type)}`).join(', ');
        const convertedReturn = this.convertType(returnType);
  
        return `${async ? 'async ' : ''}function ${name}(${convertedParams}): Promise<${convertedReturn}>`;
      }
    }
  };
  
  export default typescriptConfig;