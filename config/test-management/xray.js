/**
 * Xray (Jira) integration configuration
 */
export const xrayConfig = {
    /**
     * Xray connection settings
     */
    connection: {
      // Jira settings
      jiraHost: process.env.JIRA_HOST || '',
      jiraUsername: process.env.JIRA_USERNAME || '',
      jiraToken: process.env.JIRA_API_TOKEN || '',
      // Xray specific settings
      xrayClientId: process.env.XRAY_CLIENT_ID || '',
      xrayClientSecret: process.env.XRAY_CLIENT_SECRET || '',
      cloudInstance: true,
      baseUrl: 'https://xray.cloud.getxray.app/api/v2'
    },
  
    /**
     * Jira project configuration
     */
    project: {
      key: process.env.JIRA_PROJECT_KEY || '',
      testIssueType: 'Test',
      executionIssueType: 'Test Execution',
      preconditionIssueType: 'Precondition',
      testPlanIssueType: 'Test Plan',
      testSetIssueType: 'Test Set',
      defaultAssignee: process.env.DEFAULT_ASSIGNEE
    },
  
    /**
     * Test case configuration
     */
    testCase: {
      // Required fields
      fields: {
        summary: '',
        description: '',
        testType: 'Automated',
        priority: {
          name: 'Medium'
        },
        labels: ['playwright', 'automated'],
        components: [],
        // Xray specific fields
        'xray_test_type': 'Automated',
        'xray_test_repository': 'Playwright'
      },
      // Custom field mappings
      customFields: {
        // Map your custom fields here
        'customfield_10100': 'testType',
        'customfield_10101': 'automationFramework'
      }
    },
  
    /**
     * Test execution configuration
     */
    execution: {
      // Test execution settings
      fields: {
        summary: 'Playwright Test Execution - {date}',
        description: 'Automated test execution from Playwright tests',
        assignee: null,
        priority: {
          name: 'Medium'
        },
        labels: ['playwright-execution']
      },
      // Status mappings
      statusMap: {
        PASSED: 'PASS',
        FAILED: 'FAIL',
        SKIPPED: 'TODO',
        BLOCKED: 'BLOCKED'
      },
      // Evidence settings
      evidence: {
        screenshots: true,
        videos: true,
        logs: true,
        traces: true,
        maxSize: 10 * 1024 * 1024 // 10MB
      }
    },
  
    /**
     * Test plan configuration
     */
    testPlan: {
      enabled: true,
      fields: {
        summary: 'Playwright Test Plan - {date}',
        description: 'Test plan for Playwright automated tests',
        labels: ['playwright-plan']
      },
      // Folder structure mapping
      folders: {
        'e2e': 'End-to-End Tests',
        'api': 'API Tests',
        'component': 'Component Tests',
        'visual': 'Visual Tests',
        'performance': 'Performance Tests',
        'accessibility': 'Accessibility Tests'
      }
    },
  
    /**
     * Precondition configuration
     */
    precondition: {
      enabled: true,
      fields: {
        summary: '',
        description: '',
        labels: ['playwright-precondition']
      }
    },
  
    /**
     * Test set configuration
     */
    testSet: {
      enabled: true,
      fields: {
        summary: 'Playwright Test Set - {date}',
        description: 'Test set for Playwright automated tests',
        labels: ['playwright-set']
      }
    },
  
    /**
     * Reporting configuration
     */
    reporting: {
      enabled: true,
      format: 'xray', // xray, junit, custom
      uploadResults: true,
      generatePDF: false,
      includeEvidence: true,
      customFields: {
        testInfo: true,
        environment: true,
        requirements: true
      }
    },
  
    /**
     * Synchronization settings
     */
    sync: {
      enabled: true,
      mode: 'twoWay', // oneWay, twoWay
      automatic: true,
      interval: '5m',
      fields: {
        sync: ['summary', 'description', 'labels', 'components'],
        ignore: ['comments', 'attachments']
      },
      conflict: 'newer' // newer, jira, local
    },
  
    /**
     * API configuration
     */
    api: {
      version: 'v2',
      endpoints: {
        authenticate: '/authenticate',
        importExecution: '/import/execution',
        exportExecution: '/export/execution',
        graphql: '/graphql',
        webhook: '/webhook'
      },
      rateLimiting: {
        enabled: true,
        maxRequests: 100,
        timeWindow: 60 // seconds
      },
      timeout: 30000
    },
  
    /**
     * Error handling configuration
     */
    errorHandling: {
      retryOnError: true,
      maxRetries: 3,
      retryDelay: 1000,
      ignoredErrors: [],
      criticalErrors: ['AUTHENTICATION_ERROR', 'PERMISSION_ERROR']
    }
  };
  
  /**
   * Helper functions for Xray integration
   */
  export const xrayHelpers = {
    /**
     * Generate test case data for Xray
     * @param {Object} test - Playwright test metadata
     * @returns {Object} - Xray test case data
     */
    generateTestCase(test) {
      return {
        fields: {
          ...xrayConfig.testCase.fields,
          summary: test.title,
          description: test.description || '',
          labels: [...xrayConfig.testCase.fields.labels, ...test.labels || []],
          'xray_test_type': 'Automated',
          steps: test.steps.map(step => ({
            action: step.action,
            data: step.data,
            expected: step.expected
          }))
        }
      };
    },
  
    /**
     * Map test result to Xray status
     * @param {Object} result - Test result
     * @returns {string} - Xray status
     */
    mapResultStatus(result) {
      return xrayConfig.execution.statusMap[result.status.toUpperCase()] || 
             xrayConfig.execution.statusMap.FAILED;
    },
  
    /**
     * Generate execution info
     * @param {Object} options - Execution options
     * @returns {Object} - Execution data
     */
    generateExecutionInfo(options = {}) {
      const date = new Date().toISOString().split('T')[0];
      return {
        fields: {
          ...xrayConfig.execution.fields,
          summary: xrayConfig.execution.fields.summary.replace('{date}', date),
          environment: options.environment || 'test',
          revision: options.revision || 'main',
          testPlanKey: options.testPlanKey
        }
      };
    },
  
    /**
     * Format evidence for Xray
     * @param {Object} evidence - Test evidence
     * @returns {Object} - Formatted evidence
     */
    formatEvidence(evidence) {
      return {
        filename: evidence.filename,
        contentType: evidence.contentType,
        data: evidence.data,
        timestamp: new Date().toISOString(),
        type: evidence.type // screenshot, video, log, trace
      };
    },
  
    /**
     * Generate GraphQL query for test results
     * @param {string} testKey - Test key
     * @returns {string} - GraphQL query
     */
    generateTestQuery(testKey) {
      return `
        query {
          getTest(issueId: "${testKey}") {
            issueId
            testType
            status
            executions {
              status
              startedOn
              finishedOn
              evidence {
                filename
                fileSize
                contentType
              }
            }
          }
        }
      `;
    }
  };
  
  export default xrayConfig;