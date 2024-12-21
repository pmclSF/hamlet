/**
 * TestRail integration configuration
 */
export const testrailConfig = {
    /**
     * TestRail connection settings
     */
    connection: {
      host: process.env.TESTRAIL_HOST || '',
      username: process.env.TESTRAIL_USERNAME || '',
      apiKey: process.env.TESTRAIL_API_KEY || '',
      projectId: parseInt(process.env.TESTRAIL_PROJECT_ID) || null,
      secure: true // Use HTTPS
    },
  
    /**
     * Test suite configuration
     */
    suite: {
      name: 'Automated Tests',
      description: 'Tests converted from Cypress to Playwright',
      defaultSections: {
        e2e: 'End-to-End Tests',
        api: 'API Tests',
        component: 'Component Tests',
        visual: 'Visual Tests',
        performance: 'Performance Tests',
        accessibility: 'Accessibility Tests'
      },
      includeBaseline: true
    },
  
    /**
     * Test case configuration
     */
    testCase: {
      template: {
        type_id: 1, // Automated
        priority_id: 2, // Medium
        custom_steps_separated: [], // Steps will be added programmatically
        custom_preconds: '', // Preconditions
        custom_expected: '', // Expected results
        custom_automation_type: 'playwright'
      },
      customFields: {
        test_type: null,
        automation_status: 'automated',
        automation_engine: 'playwright'
      }
    },
  
    /**
     * Test run configuration
     */
    testRun: {
      name: 'Playwright Automated Tests',
      description: 'Automated test execution from Playwright',
      includeAll: false,
      assignedToId: null,
      suiteId: null,
      milestoneId: null,
      customStatus: {
        passed: 1,
        blocked: 2,
        untested: 3,
        retest: 4,
        failed: 5
      }
    },
  
    /**
     * Result configuration
     */
    results: {
      statusMap: {
        passed: 1,
        failed: 5,
        skipped: 3,
        blocked: 2
      },
      includeScreenshots: true,
      includeVideos: true,
      includeTrace: true,
      maxAttachmentSize: 25 * 1024 * 1024, // 25MB
      screenshotOnFailure: true
    },
  
    /**
     * Synchronization settings
     */
    sync: {
      enabled: true,
      mode: 'push', // 'push', 'pull', 'bidirectional'
      autoSync: true,
      onFailure: 'continue', // 'continue', 'stop'
      batchSize: 100,
      retryAttempts: 3
    },
  
    /**
     * Milestone configuration
     */
    milestone: {
      enabled: false,
      name: 'Sprint {n}',
      description: 'Automated test results for Sprint {n}',
      dueOn: null, // Set dynamically
      parentId: null
    },
  
    /**
     * API configuration
     */
    api: {
      baseUrl: '/index.php?/api/v2',
      endpoints: {
        addRun: 'add_run/{project_id}',
        addResults: 'add_results_for_cases/{run_id}',
        addResultsForCases: 'add_results_for_cases/{run_id}',
        getCase: 'get_case/{case_id}',
        addCase: 'add_case/{section_id}',
        updateCase: 'update_case/{case_id}'
      },
      timeout: 30000,
      rateLimit: {
        maxRequests: 180,
        perSeconds: 60
      }
    },
  
    /**
     * Reporting configuration
     */
    reporting: {
      enabled: true,
      format: 'html', // 'html', 'json', 'xml'
      outputDir: './reports/testrail',
      includeScreenshots: true,
      customFields: true
    },
  
    /**
     * Error handling
     */
    errorHandling: {
      continueOnError: true,
      retryableErrors: [
        'ETIMEDOUT',
        'ECONNRESET',
        'ECONNREFUSED',
        'rate limit exceeded'
      ],
      maxRetries: 3,
      retryDelay: 1000
    }
  };
  
  /**
   * Helper functions for TestRail integration
   */
  export const testrailHelpers = {
    /**
     * Generate test case data from Playwright test
     * @param {Object} test - Playwright test metadata
     * @returns {Object} - TestRail test case data
     */
    generateTestCase(test) {
      return {
        title: test.title,
        type_id: testrailConfig.testCase.template.type_id,
        priority_id: testrailConfig.testCase.template.priority_id,
        custom_automation_type: testrailConfig.testCase.customFields.automation_engine,
        custom_steps_separated: test.steps.map(step => ({
          content: step.description,
          expected: step.expected
        })),
        custom_preconds: test.preconditions || '',
        custom_expected: test.expectedResult || ''
      };
    },
  
    /**
     * Map test result to TestRail status
     * @param {Object} result - Test result
     * @returns {number} - TestRail status ID
     */
    mapResultStatus(result) {
      const status = result.status.toLowerCase();
      return testrailConfig.results.statusMap[status] || 
             testrailConfig.results.statusMap.failed;
    },
  
    /**
     * Generate test run name
     * @param {string} prefix - Run name prefix
     * @param {Object} options - Additional options
     * @returns {string} - Test run name
     */
    generateRunName(prefix = '', options = {}) {
      const date = new Date().toISOString().split('T')[0];
      const env = options.environment || 'test';
      return `${prefix} [${env}] - ${date}`;
    },
  
    /**
     * Format error for TestRail
     * @param {Error} error - Error object
     * @returns {string} - Formatted error message
     */
    formatError(error) {
      return `Error: ${error.message}\n\nStack:\n${error.stack}`;
    }
  };
  
  export default testrailConfig;