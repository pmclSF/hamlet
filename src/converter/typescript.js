import fs from 'fs/promises';
import path from 'path';
import chalk from 'chalk';

let _ts = null;

/**
 * Lazily load the TypeScript compiler. It is an optional dependency — only
 * needed when converting TypeScript source files or generating type defs.
 * @returns {Promise<import('typescript')>}
 */
async function loadTs() {
  if (_ts) return _ts;
  try {
    _ts = (await import('typescript')).default;
  } catch (_e) {
    throw new Error(
      'Optional dependency "typescript" is required for TypeScript conversion. ' +
        'Install it with: npm install typescript'
    );
  }
  return _ts;
}

/**
 * Handles TypeScript conversion and type generation for Cypress to Playwright
 */
export class TypeScriptConverter {
  constructor() {
    this.typeMap = new Map([
      // Cypress type mappings to Playwright equivalents
      ['Cypress.Chainable', 'Locator'],
      ['cy.wrap', 'Promise'],
      ['Cypress.Config', 'PlaywrightTestConfig'],
      ['cy.stub', 'Mock'],
      ['cy.spy', 'Mock'],
      ['Cypress.Browser', 'BrowserContext'],
      ['Cypress.ElementHandle', 'ElementHandle'],
      ['Cypress.FileReference', 'FilePayload'],
      ['cy.clock', 'Page["setDefaultTimeout"]'],
      ['Cypress.Cookie', 'Cookie'],
      ['Cypress.Response', 'APIResponse'],
      ['Cypress.AUTWindow', 'Page'],
      ['Cypress.Keyboard', 'Keyboard'],
      ['Cypress.Mouse', 'Mouse'],
      ['Cypress.ViewportPosition', 'Position'],
      ['Cypress.SelectOptions', 'SelectOption'],
    ]);

    // Common type interfaces that need conversion
    this.interfaceMap = new Map([
      ['CypressConfiguration', 'PlaywrightTestConfig'],
      ['CypressPlugin', 'PlaywrightPlugin'],
      ['CypressCommand', 'PlaywrightTest'],
      ['CypressFixture', 'TestFixture'],
    ]);
  }

  /**
   * Convert content string by applying type mappings
   * @param {string} content - Source content string
   * @returns {string} - Content with Cypress types replaced by Playwright types
   */
  convertContent(content) {
    let result = content;
    for (const [cypressType, playwrightType] of this.typeMap) {
      result = result.replaceAll(cypressType, playwrightType);
    }
    for (const [cypressInterface, playwrightInterface] of this.interfaceMap) {
      result = result.replaceAll(cypressInterface, playwrightInterface);
    }
    return result;
  }

  /**
   * Convert TypeScript files and generate type definitions
   * @param {string} sourcePath - Source directory path
   * @param {string} outputPath - Output directory path
   */
  async convertProject(sourcePath, outputPath) {
    await loadTs();
    try {
      console.log(chalk.blue('\nStarting TypeScript conversion...'));

      // Create program
      const program = this.createProgram(sourcePath);
      const typeChecker = program.getTypeChecker();

      // Process all source files
      const sourceFiles = program
        .getSourceFiles()
        .filter((file) => !file.fileName.includes('node_modules'));

      for (const sourceFile of sourceFiles) {
        await this.convertSourceFile(sourceFile, typeChecker, outputPath);
      }

      // Generate type definitions
      await this.generateTypeDefinitions(program, outputPath);

      console.log(chalk.green('✓ TypeScript conversion completed'));
    } catch (error) {
      console.error(chalk.red('Error during TypeScript conversion:'), error);
      throw error;
    }
  }

  /**
   * Create TypeScript program
   * @param {string} sourcePath - Source directory path
   * @returns {_ts.Program} - TypeScript program
   */
  createProgram(sourcePath) {
    const configPath = _ts.findConfigFile(
      sourcePath,
      _ts.sys.fileExists,
      'tsconfig.json'
    );

    if (!configPath) {
      throw new Error('Could not find tsconfig.json');
    }

    const { config } = _ts.readConfigFile(configPath, _ts.sys.readFile);
    const { options, fileNames } = _ts.parseJsonConfigFileContent(
      config,
      _ts.sys,
      path.dirname(configPath)
    );

    return _ts.createProgram(fileNames, options);
  }

  /**
   * Convert a single TypeScript source file
   * @param {_ts.SourceFile} sourceFile - TypeScript source file
   * @param {_ts.TypeChecker} typeChecker - Type checker
   * @param {string} outputPath - Output directory path
   */
  async convertSourceFile(sourceFile, typeChecker, outputPath) {
    const relativePath = path.relative(process.cwd(), sourceFile.fileName);
    const outputFile = path.join(outputPath, relativePath);

    try {
      // Transform AST
      const result = this.transformSourceFile(sourceFile, typeChecker);

      // Write converted file
      await fs.mkdir(path.dirname(outputFile), { recursive: true });
      await fs.writeFile(outputFile, result);

      console.log(
        chalk.green(`✓ Converted ${path.basename(sourceFile.fileName)}`)
      );
    } catch (error) {
      console.error(
        chalk.red(`✗ Failed to convert ${sourceFile.fileName}:`),
        error
      );
      throw error;
    }
  }

  /**
   * Transform TypeScript source file
   * @param {_ts.SourceFile} sourceFile - Source file
   * @param {_ts.TypeChecker} typeChecker - Type checker
   * @returns {string} - Transformed source code
   */
  transformSourceFile(sourceFile, typeChecker) {
    const transformer = (context) => {
      const visit = (node) => {
        // Convert type references
        if (_ts.isTypeReferenceNode(node)) {
          return this.transformTypeReference(node);
        }

        // Convert interfaces
        if (_ts.isInterfaceDeclaration(node)) {
          return this.transformInterface(node);
        }

        // Convert method calls
        if (_ts.isCallExpression(node)) {
          return this.transformMethodCall(node, typeChecker);
        }

        // Convert imports
        if (_ts.isImportDeclaration(node)) {
          return this.transformImport(node);
        }

        return _ts.visitEachChild(node, visit, context);
      };

      return (node) => _ts.visitNode(node, visit);
    };

    const result = _ts.transform(sourceFile, [transformer]);
    const printer = _ts.createPrinter({ newLine: _ts.NewLineKind.LineFeed });

    return printer.printFile(result.transformed[0]);
  }

  /**
   * Transform type reference
   * @param {_ts.TypeReferenceNode} node - Type reference node
   * @returns {_ts.Node} - Transformed node
   */
  transformTypeReference(node) {
    const typeName = node.typeName.getText();
    const mappedType = this.typeMap.get(typeName);

    if (mappedType) {
      return _ts.factory.createTypeReferenceNode(
        mappedType,
        node.typeArguments
      );
    }

    return node;
  }

  /**
   * Transform interface declaration
   * @param {_ts.InterfaceDeclaration} node - Interface declaration
   * @returns {_ts.Node} - Transformed node
   */
  transformInterface(node) {
    const interfaceName = node.name.getText();
    const mappedInterface = this.interfaceMap.get(interfaceName);

    if (mappedInterface) {
      return _ts.factory.createInterfaceDeclaration(
        node.decorators,
        node.modifiers,
        mappedInterface,
        node.typeParameters,
        node.heritageClauses,
        node.members
      );
    }

    return node;
  }

  /**
   * Transform method call
   * @param {_ts.CallExpression} node - Call expression node
   * @param {_ts.TypeChecker} typeChecker - Type checker
   * @returns {_ts.Node} - Transformed node
   */
  transformMethodCall(node, typeChecker) {
    const signature = typeChecker.getResolvedSignature(node);
    if (signature) {
      const { declaration } = signature;
      if (declaration && _ts.isMethodDeclaration(declaration)) {
        const methodName = declaration.name.getText();
        // Transform Cypress method calls to Playwright equivalents
        const transformedName = this.transformMethodName(methodName);
        if (transformedName !== methodName) {
          return _ts.factory.createCallExpression(
            _ts.factory.createIdentifier(transformedName),
            node.typeArguments,
            node.arguments
          );
        }
      }
    }
    return node;
  }

  /**
   * Transform import declaration
   * @param {_ts.ImportDeclaration} node - Import declaration
   * @returns {_ts.Node} - Transformed node
   */
  transformImport(node) {
    const importPath = node.moduleSpecifier.getText().replace(/['"]/g, '');

    // Transform Cypress imports to Playwright
    if (importPath.includes('cypress')) {
      return _ts.factory.createImportDeclaration(
        node.decorators,
        node.modifiers,
        node.importClause,
        _ts.factory.createStringLiteral('@playwright/test')
      );
    }

    return node;
  }

  /**
   * Transform Cypress method name to Playwright equivalent
   * @param {string} methodName - Cypress method name
   * @returns {string} - Playwright method name
   */
  transformMethodName(methodName) {
    const methodMap = {
      visit: 'goto',
      get: 'locator',
      find: 'locator',
      click: 'click',
      type: 'fill',
      should: 'expect',
      contains: 'getByText',
      first: 'first',
      last: 'last',
      eq: 'nth',
      parent: "locator('..')",
      children: "locator('>*')",
      siblings: "locator('~')",
      next: "locator('+')",
      prev: "locator('-')",
    };

    return methodMap[methodName] || methodName;
  }

  /**
   * Generate type definitions
   * @param {_ts.Program} program - TypeScript program
   * @param {string} outputPath - Output directory path
   */
  async generateTypeDefinitions(program, outputPath) {
    const typeDefinitions = new Map();
    const typeChecker = program.getTypeChecker();

    // Collect type information from all source files
    for (const sourceFile of program.getSourceFiles()) {
      if (!sourceFile.isDeclarationFile) {
        _ts.forEachChild(sourceFile, (node) => {
          if (
            _ts.isInterfaceDeclaration(node) ||
            _ts.isTypeAliasDeclaration(node)
          ) {
            const symbol = typeChecker.getSymbolAtLocation(node.name);
            if (symbol) {
              const type = typeChecker.getTypeAtLocation(node);
              typeDefinitions.set(symbol.getName(), {
                kind: node.kind,
                type: typeChecker.typeToString(type),
              });
            }
          }
        });
      }
    }

    // Generate type definition file
    const dtsContent =
      await this.generateDefinitionFileContent(typeDefinitions);
    const dtsPath = path.join(outputPath, 'playwright.d.ts');

    await fs.writeFile(dtsPath, dtsContent);
    console.log(chalk.green(`✓ Generated type definitions at ${dtsPath}`));
  }

  /**
   * Generate content for type definition file
   * @param {Map} typeDefinitions - Collected type definitions
   * @returns {string} - Type definition file content
   */
  async generateDefinitionFileContent(typeDefinitions) {
    await loadTs();
    let content = '// Generated type definitions for Playwright tests\n\n';
    content +=
      "import { test, expect, Page, Locator } from '@playwright/test';\n\n";

    for (const [name, { kind, type }] of typeDefinitions) {
      if (kind === _ts.SyntaxKind.InterfaceDeclaration) {
        content += `interface ${name} ${type}\n\n`;
      } else if (kind === _ts.SyntaxKind.TypeAliasDeclaration) {
        content += `type ${name} = ${type};\n\n`;
      }
    }

    return content;
  }
}
