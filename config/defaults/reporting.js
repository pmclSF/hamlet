/**
 * Default reporting configuration
 */
export const reportingConfig = {
    /**
     * Report generation settings
     */
    settings: {
      enabled: true,
      outputDir: './reports',
      formats: ['html', 'json', 'markdown'],
      timestamp: true,
      includeMetadata: true,
      preserveHistory: true
    },
  
    /**
     * HTML report configuration
     */
    html: {
      template: 'default',
      styling: {
        theme: 'light',
        customCSS: false,
        responsive: true,
        collapseDetails: true
      },
      sections: {
        summary: true,
        testDetails: true,
        errors: true,
        warnings: true,
        visualDiff: true,
        coverage: true,
        timeline: true
      },
      visualization: {
        charts: true,
        trends: true,
        diffView: true
      }
    },
  
    /**
     * JSON report configuration
     */
    json: {
      pretty: true,
      includeStackTrace: true,
      includeDiff: true,
      includeMetrics: true,
      grouping: {
        byType: true,
        byStatus: true,
        byDirectory: true
      }
    },
  
    /**
     * Markdown report configuration
     */
    markdown: {
      includeTOC: true,
      includeStats: true,
      includeExamples: true,
      codeBlocks: true,
      emojis: true
    },
  
    /**
     * Content configuration
     */
    content: {
      /**
       * Summary section
       */
      summary: {
        totals: true,
        successRate: true,
        duration: true,
        timing: true,
        testTypes: true,
        coverage: true
      },
  
      /**
       * Test details
       */
      testDetails: {
        fileName: true,
        testName: true,
        status: true,
        duration: true,
        errors: true,
        screenshots: true,
        diff: true
      },
  
      /**
       * Error reporting
       */
      errors: {
        stackTrace: true,
        context: true,
        suggestions: true,
        recovery: true,
        grouping: true
      },
  
      /**
       * Visual comparison
       */
      visual: {
        sideBySide: true,
        diffHighlight: true,
        metrics: true,
        thresholds: true
      },
  
      /**
       * Metrics collection
       */
      metrics: {
        conversionTime: true,
        testCount: true,
        fileCount: true,
        lineCount: true,
        coverage: true
      }
    },
  
    /**
     * Notification configuration
     */
    notifications: {
      enabled: true,
      channels: {
        console: true,
        slack: false,
        email: false,
        customWebhook: false
      },
      triggers: {
        onComplete: true,
        onError: true,
        onWarning: false
      }
    },
  
    /**
     * History tracking
     */
    history: {
      enabled: true,
      storageDir: './reports/history',
      maxEntries: 30,
      tracking: {
        successRate: true,
        errorRate: true,
        conversionTime: true,
        testCount: true
      }
    },
  
    /**
     * Report templates
     */
    templates: {
      html: {
        default: `
  <!DOCTYPE html>
  <html>
  <head>
    <title>Conversion Report</title>
    <style>
      /* Default styles */
      body { font-family: Arial, sans-serif; margin: 2rem; }
      .summary { margin-bottom: 2rem; }
      .success { color: green; }
      .error { color: red; }
      .warning { color: orange; }
    </style>
  </head>
  <body>
    <h1>Conversion Report</h1>
    {{content}}
  </body>
  </html>
        `,
        minimal: `
  <!DOCTYPE html>
  <html>
  <body>
    {{content}}
  </body>
  </html>
        `
      },
      markdown: {
        default: `
  # Conversion Report
  
  Generated on: {{date}}
  
  ## Summary
  {{summary}}
  
  ## Details
  {{details}}
  
  ## Errors
  {{errors}}
        `
      }
    },
  
    /**
     * Helper functions
     */
    helpers: {
      /**
       * Format duration
       * @param {number} ms - Duration in milliseconds
       * @returns {string} - Formatted duration
       */
      formatDuration(ms) {
        const seconds = Math.floor(ms / 1000);
        const minutes = Math.floor(seconds / 60);
        const hours = Math.floor(minutes / 60);
  
        if (hours > 0) {
          return `${hours}h ${minutes % 60}m ${seconds % 60}s`;
        }
        if (minutes > 0) {
          return `${minutes}m ${seconds % 60}s`;
        }
        return `${seconds}s`;
      },
  
      /**
       * Calculate success rate
       * @param {Object} stats - Statistics object
       * @returns {string} - Formatted success rate
       */
      calculateSuccessRate(stats) {
        const total = stats.total || 1;
        const success = stats.success || 0;
        const rate = (success / total) * 100;
        return `${rate.toFixed(2)}%`;
      },
  
      /**
       * Generate chart data
       * @param {Object} data - Report data
       * @returns {Object} - Chart configuration
       */
      generateChartData(data) {
        return {
          labels: Object.keys(data),
          datasets: [{
            label: 'Conversion Results',
            data: Object.values(data),
            backgroundColor: ['#4CAF50', '#F44336', '#FFC107']
          }]
        };
      },
  
      /**
       * Format error for display
       * @param {Error} error - Error object
       * @returns {string} - Formatted error
       */
      formatError(error) {
        return {
          message: error.message,
          type: error.name,
          stack: error.stack,
          location: error.location,
          suggestions: this.getSuggestions(error)
        };
      },
  
      /**
       * Get error suggestions
       * @param {Error} error - Error object
       * @returns {string[]} - Array of suggestions
       */
      getSuggestions(error) {
        const suggestions = {
          SyntaxError: [
            'Check for missing brackets or parentheses',
            'Verify correct usage of async/await',
            'Ensure proper function syntax'
          ],
          ReferenceError: [
            'Check variable declarations',
            'Verify import statements',
            'Check scope of variables'
          ],
          TypeError: [
            'Verify correct method names',
            'Check parameter types',
            'Ensure proper object properties'
          ]
        };
  
        return suggestions[error.name] || ['Review the error message and context'];
      }
    }
  };
  
  export default reportingConfig;