/**
 * Azure DevOps integration configuration
 */
export const azureConfig = {
    /**
     * Azure DevOps connection settings
     */
    connection: {
      organization: process.env.AZURE_ORGANIZATION || '',
      project: process.env.AZURE_PROJECT || '',
      pat: process.env.AZURE_PAT || '',
      apiVersion: '7.0'
    },
  
    /**
     * Test plan configuration
     */
    testPlan: {
      name: 'Automated Tests',
      areaPath: 'YourProject\\Testing',
      iterationPath: 'YourProject\\Sprint 1',
      suiteType: 'StaticTestSuite',
      defaultState: 'Design'
    },
  
    /**
     * Test case configuration
     */
    testCase: {
      workItemType: 'Test Case',
      defaultFields: {
        'System.Title': '',
        'System.Description': '',
        'Microsoft.VSTS.TCM.AutomatedTestName': '',
        'Microsoft.VSTS.TCM.AutomatedTestStorage': '',
        'Microsoft.VSTS.TCM.AutomationStatus': 'Automated',
        'Microsoft.VSTS.TCM.TestSuite': ''
      },
      customFields: {
        // Add any custom fields here
      }
    },
  
    /**
     * Test result configuration
     */
    testResult: {
      outcomes: {
        passed: 'Passed',
        failed: 'Failed',
        skipped: 'Not Executed',
        blocked: 'Blocked',
        notApplicable: 'Not Applicable'
      },
      attachScreenshots: true,
      attachLogs: true,
      includeRunDetails: true
    },
  
    /**
     * Mapping configuration for test attributes
     */
    mapping: {
      testSuites: {
        'e2e': 'End-to-End Tests',
        'integration': 'Integration Tests',
        'component': 'Component Tests',
        'api': 'API Tests',
        'accessibility': 'Accessibility Tests',
        'performance': 'Performance Tests',
        'visual': 'Visual Tests'
      },
      priorities: {
        high: 1,
        medium: 2,
        low: 3
      },
      categories: {
        smoke: 'Smoke Test',
        regression: 'Regression Test',
        functional: 'Functional Test'
      }
    },
  
    /**
     * Synchronization settings
     */
    sync: {
      enabled: true,
      direction: 'bidirectional', // 'toAzure', 'fromAzure', 'bidirectional'
      frequency: 'onUpdate', // 'onUpdate', 'scheduled'
      schedule: '0 0 * * *', // Daily at midnight (if frequency is 'scheduled')
      includeMetadata: true,
      includeAttachments: true
    },
  
    /**
     * Retry configuration for API calls
     */
    retry: {
      maxAttempts: 3,
      initialDelay: 1000,
      maxDelay: 5000,
      backoff: 2
    },
  
    /**
     * API endpoints
     */
    endpoints: {
      baseUrl: 'https://dev.azure.com/',
      testPlans: '_apis/testplan/plans',
      testSuites: '_apis/testplan/suites',
      testCases: '_apis/wit/workitems',
      testResults: '_apis/test/runs',
      attachments: '_apis/wit/attachments'
    },
  
    /**
     * Logging configuration
     */
    logging: {
      level: 'info', // error, warn, info, debug
      includeTimestamp: true,
      format: 'json',
      destination: 'console' // console, file, both
    },
  
    /**
     * Error handling configuration
     */
    errorHandling: {
      ignoredErrors: [],
      retryableErrors: [
        'ETIMEDOUT',
        'ECONNRESET',
        'ECONNREFUSED'
      ],
      throwOnError: true
    }
  };
  
  /**
   * Helper functions for Azure DevOps integration
   */
  export const azureHelpers = {
    /**
     * Generate test case fields from Playwright test
     * @param {Object} test - Playwright test metadata
     * @returns {Object} - Azure DevOps test case fields
     */
    generateTestCaseFields(test) {
      return {
        'System.Title': test.title,
        'System.Description': test.description,
        'Microsoft.VSTS.TCM.AutomatedTestName': test.name,
        'Microsoft.VSTS.TCM.AutomatedTestStorage': test.file,
        'Microsoft.VSTS.TCM.AutomationStatus': 'Automated',
        'System.Tags': test.tags.join('; ')
      };
    },
  
    /**
     * Get test result outcome
     * @param {Object} result - Test result
     * @returns {string} - Azure DevOps outcome
     */
    getTestOutcome(result) {
      if (result.status === 'passed') return azureConfig.testResult.outcomes.passed;
      if (result.status === 'failed') return azureConfig.testResult.outcomes.failed;
      if (result.status === 'skipped') return azureConfig.testResult.outcomes.skipped;
      return azureConfig.testResult.outcomes.notApplicable;
    },
  
    /**
     * Get suite path for test type
     * @param {string} testType - Type of test
     * @returns {string} - Suite path
     */
    getSuitePath(testType) {
      return azureConfig.mapping.testSuites[testType] || 'Other Tests';
    }
  };
  
  export default azureConfig;