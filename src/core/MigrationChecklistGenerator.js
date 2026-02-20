/**
 * Generates a migration checklist markdown string from conversion results.
 *
 * Sections: Summary, Fully Converted, Needs Review, Manual Steps, Config Changes.
 */

export class MigrationChecklistGenerator {
  /**
   * Generate a migration checklist from conversion results.
   *
   * @param {Object} projectGraph - The dependency graph (from DependencyGraphBuilder)
   * @param {Array<{path: string, confidence: number, warnings: string[], todos: string[], type: string}>} conversionResults
   * @returns {string} Markdown checklist
   */
  generate(projectGraph, conversionResults) {
    const sections = [];

    // Summary section
    sections.push(this._generateSummary(conversionResults));

    // Fully converted (>= 90% confidence)
    const fullyConverted = conversionResults.filter((r) => r.confidence >= 90);
    if (fullyConverted.length > 0) {
      sections.push(
        this._generateSection('Fully Converted', fullyConverted, true)
      );
    }

    // Needs review (< 90% confidence, > 0%)
    const needsReview = conversionResults.filter(
      (r) => r.confidence > 0 && r.confidence < 90
    );
    if (needsReview.length > 0) {
      sections.push(this._generateSection('Needs Review', needsReview, false));
    }

    // Manual steps (0% or failed)
    const manual = conversionResults.filter(
      (r) => r.confidence === 0 || r.status === 'failed'
    );
    if (manual.length > 0) {
      sections.push(this._generateManualSection(manual));
    }

    // Config changes
    const configs = conversionResults.filter((r) => r.type === 'config');
    if (configs.length > 0) {
      sections.push(this._generateConfigSection(configs));
    }

    return sections.join('\n\n');
  }

  /**
   * @param {Array} results
   * @returns {string}
   */
  _generateSummary(results) {
    const total = results.length;
    const high = results.filter((r) => r.confidence >= 90).length;
    const medium = results.filter(
      (r) => r.confidence >= 70 && r.confidence < 90
    ).length;
    const low = results.filter(
      (r) => r.confidence > 0 && r.confidence < 70
    ).length;
    const failed = results.filter(
      (r) => r.confidence === 0 || r.status === 'failed'
    ).length;

    const lines = [
      '# Migration Checklist',
      '',
      `- **Total files:** ${total}`,
      `- **High confidence (>=90%):** ${high}`,
      `- **Medium confidence (70-89%):** ${medium}`,
      `- **Low confidence (<70%):** ${low}`,
      `- **Failed/Manual:** ${failed}`,
    ];

    return lines.join('\n');
  }

  /**
   * @param {string} title
   * @param {Array} results
   * @param {boolean} checked
   * @returns {string}
   */
  _generateSection(title, results, checked) {
    const lines = [`## ${title}`, ''];

    for (const r of results) {
      const check = checked ? '[x]' : '[ ]';
      const conf = `(${r.confidence}%)`;
      lines.push(`- ${check} \`${r.path}\` ${conf}`);

      if (r.warnings && r.warnings.length > 0) {
        for (const w of r.warnings) {
          lines.push(`  - WARNING: ${w}`);
        }
      }

      if (r.todos && r.todos.length > 0) {
        for (const t of r.todos) {
          lines.push(`  - TODO: ${t}`);
        }
      }
    }

    return lines.join('\n');
  }

  /**
   * @param {Array} results
   * @returns {string}
   */
  _generateManualSection(results) {
    const lines = ['## Manual Steps Required', ''];

    for (const r of results) {
      lines.push(`- [ ] \`${r.path}\``);
      if (r.status === 'failed' && r.error) {
        lines.push(`  - Error: ${r.error}`);
      }
      if (r.todos && r.todos.length > 0) {
        for (const t of r.todos) {
          lines.push(`  - TODO: ${t}`);
        }
      }
    }

    return lines.join('\n');
  }

  /**
   * @param {Array} configs
   * @returns {string}
   */
  _generateConfigSection(configs) {
    const lines = ['## Config Changes', ''];

    for (const c of configs) {
      const check = c.confidence >= 90 ? '[x]' : '[ ]';
      lines.push(`- ${check} \`${c.path}\` (${c.confidence}%)`);
      if (c.todos && c.todos.length > 0) {
        for (const t of c.todos) {
          lines.push(`  - TODO: ${t}`);
        }
      }
    }

    return lines.join('\n');
  }
}
