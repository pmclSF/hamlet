import { TestManagementSystem } from '../src/index';

// Mock test management system API calls
jest.mock('../src/tms/api', () => ({
  TestRailAPI: jest.fn(),
  AzureAPI: jest.fn(),
  XrayAPI: jest.fn(),
}));

// Global test setup
beforeAll(() => {
  // Set up test environment
});

afterAll(() => {
  // Clean up test environment
});