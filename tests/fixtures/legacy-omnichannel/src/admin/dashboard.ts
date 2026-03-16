import { getUser } from '../shared/db-helper';

export function getAdminStats() {
  return { totalOrders: 1500, totalRevenue: 250000, activeUsers: 340 };
}

export function searchOrders(query: string) {
  return { query, results: [], total: 0 };
}

export function exportReport(format: string) {
  return { format, url: 'https://export.example.com/report.csv' };
}
