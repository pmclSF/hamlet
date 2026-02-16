import fs from "fs/promises";
import path from "path";
import chalk from "chalk";

/**
 * Manages bidirectional mapping between Cypress and Playwright tests
 */
export class TestMapper {
  constructor() {
    this.mappings = new Map();
    this.metaData = {
      version: "1.0",
      lastUpdated: new Date().toISOString(),
      statistics: {
        totalMappings: 0,
        activeMappings: 0,
        pendingSync: 0,
      },
    };
  }

  /**
   * Set up initial mappings between Cypress and Playwright tests
   * @param {string} cypressPath - Path to Cypress project
   * @param {string} playwrightPath - Path to Playwright project
   */
  async setupMappings(cypressPath, playwrightPath) {
    try {
      // Find all test files
      const cypressTests = await this.findTestFiles(cypressPath);
      const playwrightTests = await this.findTestFiles(playwrightPath);

      // Create mappings based on file names and content similarity
      for (const cypressTest of cypressTests) {
        const matchingTest = await this.findMatchingTest(
          cypressTest,
          playwrightTests,
        );
        if (matchingTest) {
          this.addMapping(cypressTest, matchingTest);
        }
      }

      // Update statistics
      this.updateStatistics();

      console.log(chalk.green(`✓ Created ${this.mappings.size} test mappings`));
    } catch (error) {
      console.error(chalk.red("Error setting up test mappings:"), error);
      throw error;
    }
  }

  /**
   * Add a new mapping between Cypress and Playwright tests
   * @param {string} cypressTest - Path to Cypress test
   * @param {string} playwrightTest - Path to Playwright test
   */
  async addMapping(cypressTest, playwrightTest) {
    this.mappings.set(cypressTest, {
      playwrightPath: playwrightTest,
      timestamp: Date.now(),
      status: "active",
      syncStatus: "synced",
      lastSync: new Date().toISOString(),
      checksum: await this.calculateChecksum(cypressTest),
    });
  }

  /**
   * Find matching Playwright test for a given Cypress test
   * @param {string} cypressTest - Path to Cypress test
   * @param {string[]} playwrightTests - Array of Playwright test paths
   * @returns {string|null} - Matching Playwright test path or null
   */
  async findMatchingTest(cypressTest, playwrightTests) {
    const cypressName = path.basename(cypressTest, path.extname(cypressTest));
    const cypressContent = await fs.readFile(cypressTest, "utf8");

    let bestMatch = null;
    let highestSimilarity = 0;

    for (const playwrightTest of playwrightTests) {
      const playwrightName = path.basename(
        playwrightTest,
        path.extname(playwrightTest),
      );

      // Check name similarity
      const nameSimilarity = this.calculateNameSimilarity(
        cypressName,
        playwrightName,
      );

      // If names are very similar, check content
      if (nameSimilarity > 0.8) {
        const playwrightContent = await fs.readFile(playwrightTest, "utf8");
        const contentSimilarity = this.calculateContentSimilarity(
          cypressContent,
          playwrightContent,
        );

        const totalSimilarity = (nameSimilarity + contentSimilarity) / 2;
        if (totalSimilarity > highestSimilarity) {
          highestSimilarity = totalSimilarity;
          bestMatch = playwrightTest;
        }
      }
    }

    return bestMatch;
  }

  /**
   * Calculate similarity between two file names
   * @param {string} name1 - First file name
   * @param {string} name2 - Second file name
   * @returns {number} - Similarity score between 0 and 1
   */
  calculateNameSimilarity(name1, name2) {
    // Remove common test file suffixes
    name1 = name1.replace(/\.(spec|test|cy)/, "");
    name2 = name2.replace(/\.(spec|test|cy)/, "");

    const distance = this.levenshteinDistance(name1, name2);
    const maxLength = Math.max(name1.length, name2.length);
    return 1 - distance / maxLength;
  }

  /**
   * Calculate similarity between test contents
   * @param {string} content1 - First test content
   * @param {string} content2 - Second test content
   * @returns {number} - Similarity score between 0 and 1
   */
  calculateContentSimilarity(content1, content2) {
    // Extract test descriptions and assertions for comparison
    const pattern = /(describe|it|test)\s*\(\s*['"`](.*?)['"`]/g;
    const descriptions1 = [...content1.matchAll(pattern)].map((m) => m[2]);
    const descriptions2 = [...content2.matchAll(pattern)].map((m) => m[2]);

    // Compare descriptions
    let matchingDescriptions = 0;
    for (const desc1 of descriptions1) {
      if (
        descriptions2.some(
          (desc2) => this.calculateNameSimilarity(desc1, desc2) > 0.8,
        )
      ) {
        matchingDescriptions++;
      }
    }

    return (
      matchingDescriptions /
      Math.max(descriptions1.length, descriptions2.length)
    );
  }

  /**
   * Calculate Levenshtein distance between two strings
   * @param {string} str1 - First string
   * @param {string} str2 - Second string
   * @returns {number} - Levenshtein distance
   */
  levenshteinDistance(str1, str2) {
    const matrix = Array(str2.length + 1)
      .fill(null)
      .map(() => Array(str1.length + 1).fill(null));

    for (let i = 0; i <= str1.length; i++) {
      matrix[0][i] = i;
    }
    for (let j = 0; j <= str2.length; j++) {
      matrix[j][0] = j;
    }

    for (let j = 1; j <= str2.length; j++) {
      for (let i = 1; i <= str1.length; i++) {
        const substitute =
          matrix[j - 1][i - 1] + (str1[i - 1] !== str2[j - 1] ? 1 : 0);
        matrix[j][i] = Math.min(
          matrix[j - 1][i] + 1,
          matrix[j][i - 1] + 1,
          substitute,
        );
      }
    }

    return matrix[str2.length][str1.length];
  }

  /**
   * Find all test files in a directory
   * @param {string} dir - Directory to search
   * @returns {Promise<string[]>} - Array of test file paths
   */
  async findTestFiles(dir) {
    const testFiles = [];

    async function scan(directory) {
      const entries = await fs.readdir(directory, { withFileTypes: true });

      for (const entry of entries) {
        const fullPath = path.join(directory, entry.name);

        if (entry.isDirectory()) {
          await scan(fullPath);
        } else if (
          entry.isFile() &&
          /\.(spec|test|cy)\.(js|ts)$/.test(entry.name)
        ) {
          testFiles.push(fullPath);
        }
      }
    }

    await scan(dir);
    return testFiles;
  }

  /**
   * Calculate checksum for a file
   * @param {string} filePath - Path to file
   * @returns {string} - Checksum
   */
  async calculateChecksum(filePath) {
    const content = await fs.readFile(filePath, "utf8");
    let hash = 0;

    for (let i = 0; i < content.length; i++) {
      const char = content.charCodeAt(i);
      hash = (hash << 5) - hash + char;
      hash = hash & hash;
    }

    return hash.toString(16);
  }

  /**
   * Update mapping statistics
   */
  updateStatistics() {
    this.metaData.statistics = {
      totalMappings: this.mappings.size,
      activeMappings: Array.from(this.mappings.values()).filter(
        (m) => m.status === "active",
      ).length,
      pendingSync: Array.from(this.mappings.values()).filter(
        (m) => m.syncStatus === "pending",
      ).length,
    };
    this.metaData.lastUpdated = new Date().toISOString();
  }

  /**
   * Get all current mappings
   * @returns {Object} - Current mappings and metadata
   */
  getMappings() {
    return {
      ...this.metaData,
      mappings: Array.from(this.mappings.entries()).map(([cypress, data]) => ({
        cypressTest: cypress,
        playwrightTest: data.playwrightPath,
        status: data.status,
        syncStatus: data.syncStatus,
        lastSync: data.lastSync,
      })),
    };
  }

  /**
   * Save mappings to a file
   * @param {string} outputPath - Path to save mappings
   */
  async saveMappings(outputPath) {
    try {
      const mappingData = {
        ...this.metaData,
        mappings: Array.from(this.mappings.entries()).map(
          ([cypress, data]) => ({
            cypressTest: cypress,
            ...data,
          }),
        ),
      };

      await fs.writeFile(outputPath, JSON.stringify(mappingData, null, 2));

      console.log(chalk.green(`✓ Saved mappings to ${outputPath}`));
    } catch (error) {
      console.error(chalk.red("Error saving mappings:"), error);
      throw error;
    }
  }

  /**
   * Load mappings from a file
   * @param {string} inputPath - Path to load mappings from
   */
  async loadMappings(inputPath) {
    try {
      const content = await fs.readFile(inputPath, "utf8");
      const data = JSON.parse(content);

      this.metaData = {
        version: data.version,
        lastUpdated: data.lastUpdated,
        statistics: data.statistics,
      };

      this.mappings.clear();
      for (const mapping of data.mappings) {
        this.mappings.set(mapping.cypressTest, {
          playwrightPath: mapping.playwrightTest,
          timestamp: mapping.timestamp,
          status: mapping.status,
          syncStatus: mapping.syncStatus,
          lastSync: mapping.lastSync,
          checksum: mapping.checksum,
        });
      }

      console.log(chalk.green(`✓ Loaded mappings from ${inputPath}`));
    } catch (error) {
      console.error(chalk.red("Error loading mappings:"), error);
      throw error;
    }
  }
}
