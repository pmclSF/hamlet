/**
 * Default conversion settings and patterns
 */
export const conversionConfig = {
    /**
     * General conversion settings
     */
    settings: {
      // Output configuration
      output: {
        directory: './playwright-tests',
        preserveStructure: true,
        fileExtension: '.spec.ts',
        createMissingDirs: true
      },
  
      // Conversion options
      options: {
        typescript: true,
        addComments: true,
        preserveDescriptions: true,
        includeOriginalCode: false,
        generateMigrationGuide: true
      },
  
      // Test structure configuration
      structure: {
        wrapInDescribe: true,
        useArrowFunctions: true,
        addPageParameter: true,
        addRequestParameter: true,
        addContextParameter: true
      }
    },
  
    /**
     * File handling patterns
     */
    files: {
      // File name conversions
      naming: {
        'cy.js': '.spec.js',
        'cy.ts': '.spec.ts',
        'cypress.config.js': 'playwright.config.js',
        'cypress.json': 'playwright.config.js'
      },
  
      // Directory mappings
      directories: {
        'cypress/integration': 'tests/e2e',
        'cypress/component': 'tests/component',
        'cypress/fixtures': 'tests/fixtures',
        'cypress/support': 'tests/support',
        'cypress/plugins': 'tests/plugins'
      },
  
      // Files to ignore
      ignore: [
        'node_modules/**',
        '**/cypress/plugins/**',
        '**/cypress.env.json'
      ]
    },
  
    /**
     * Code formatting settings
     */
    formatting: {
      // Basic formatting
      style: {
        indent: 2,
        quotes: 'single',
        semicolons: true,
        trailingComma: 'es5'
      },
  
      // Import organization
      imports: {
        addPlaywrightImport: true,
        sortImports: true,
        removeUnused: true,
        grouping: [
          '@playwright/test',
          'external modules',
          'local modules',
          'test utils'
        ]
      },
  
      // Comments
      comments: {
        preserveCypress: true,
        addConversionNotes: true,
        addTodoComments: true
      }
    },
  
    /**
     * Error handling settings
     */
    errorHandling: {
      // Error behaviors
      behavior: {
        onSyntaxError: 'warn',
        onConversionError: 'skip',
        onValidationError: 'fail',
        continueOnError: true
      },
  
      // Recovery strategies
      recovery: {
        attemptPartialConversion: true,
        preserveOriginalOnError: true,
        addTodoComments: true
      },
  
      // Error reporting
      reporting: {
        includeStackTrace: true,
        groupByErrorType: true,
        suggestFixes: true
      }
    },
  
    /**
     * Code transformation settings
     */
    transformation: {
      // Function transformations
      functions: {
        // Arrow function preferences
        arrowFunctions: {
          enabled: true,
          singleParam: true,
          multipleParams: true
        },
  
        // Async/await handling
        asyncAwait: {
          addAsync: true,
          convertPromises: true,
          preserveChaining: false
        },
  
        // Parameter handling
        parameters: {
          addTypes: true,
          defaultValues: true,
          destructuring: true
        }
      },
  
      // Variable transformations
      variables: {
        // Const/let preferences
        constLet: {
          preferConst: true,
          convertVar: true
        },
  
        // Destructuring
        destructuring: {
          enabled: true,
          objects: true,
          arrays: true
        }
      },
  
      // Module transformations
      modules: {
        // Import/export handling
        imports: {
          addPlaywright: true,
          removeUnused: true,
          convertRequire: true
        },
  
        // Export handling
        exports: {
          convertCommonJs: true,
          addTypeAnnotations: true
        }
      }
    },
  
    /**
     * Context handling
     */
    context: {
      // Page context
      page: {
        parameter: 'page',
        addToTests: true,
        addToHooks: true
      },
  
      // Request context
      request: {
        parameter: 'request',
        addToApiTests: true
      },
  
      // Browser context
      browser: {
        parameter: 'context',
        addToBrowserTests: true
      }
    },
  
    /**
     * Helper functions
     */
    helpers: {
      /**
       * Get output path for converted file
       * @param {string} inputPath - Original file path
       * @returns {string} - Converted file path
       */
      getOutputPath(inputPath) {
        const relativePath = path.relative(process.cwd(), inputPath);
        let outputPath = relativePath;
  
        // Apply directory mappings
        for (const [cypressDir, playwrightDir] of Object.entries(this.files.directories)) {
          if (outputPath.includes(cypressDir)) {
            outputPath = outputPath.replace(cypressDir, playwrightDir);
            break;
          }
        }
  
        // Apply file name conversions
        for (const [cypressExt, playwrightExt] of Object.entries(this.files.naming)) {
          if (outputPath.endsWith(cypressExt)) {
            outputPath = outputPath.replace(new RegExp(cypressExt + '$'), playwrightExt);
            break;
          }
        }
  
        return path.join(this.settings.output.directory, outputPath);
      },
  
      /**
       * Should ignore file
       * @param {string} filePath - File path to check
       * @returns {boolean} - Whether to ignore the file
       */
      shouldIgnoreFile(filePath) {
        return this.files.ignore.some(pattern => {
          if (pattern instanceof RegExp) {
            return pattern.test(filePath);
          }
          return filePath.includes(pattern);
        });
      },
  
      /**
       * Get error handling strategy
       * @param {string} errorType - Type of error
       * @returns {string} - Error handling strategy
       */
      getErrorStrategy(errorType) {
        return this.errorHandling.behavior[`on${errorType}`] || 'warn';
      }
    }
  };
  
  export default conversionConfig;