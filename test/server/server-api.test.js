import http from 'node:http';
import path, { dirname } from 'path';
import fs from 'fs/promises';
import { fileURLToPath } from 'url';
import { HamletServer } from '../../src/server/HamletServer.js';

const __dirname = dirname(fileURLToPath(import.meta.url));

const FIXTURES_ANALYZE = path.resolve(__dirname, '../fixtures/analyze');
const FIXTURES_CONVERT = path.resolve(__dirname, '../fixtures/convert');
const OUTPUT_DIR = path.resolve(__dirname, '../output/server-test');

describe('HamletServer API', () => {
  let server;
  let baseUrl;

  beforeAll(async () => {
    server = new HamletServer({ port: 0 });
    baseUrl = await server.start();
  });

  afterAll(async () => {
    await server.stop();
    // Clean up output dir
    try {
      await fs.rm(OUTPUT_DIR, { recursive: true, force: true });
    } catch (_e) {
      // Ignore
    }
  });

  describe('GET /api/health', () => {
    it('should return status, version, and uptime', async () => {
      const res = await fetch(`${baseUrl}/api/health`);
      expect(res.status).toBe(200);

      const body = await res.json();
      expect(body.status).toBe('ok');
      expect(body.version).toBeDefined();
      expect(typeof body.uptime).toBe('number');
    });
  });

  describe('POST /api/analyze', () => {
    it('should return a valid analysis report', async () => {
      const res = await fetch(`${baseUrl}/api/analyze`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ root: FIXTURES_ANALYZE }),
      });
      expect(res.status).toBe(200);

      const body = await res.json();
      expect(body.schemaVersion).toBe('1.0.0');
      expect(body.meta).toBeDefined();
      expect(body.summary).toBeDefined();
      expect(body.files).toBeDefined();
      expect(Array.isArray(body.files)).toBe(true);
    });

    it('should return 400 when root is missing', async () => {
      const res = await fetch(`${baseUrl}/api/analyze`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({}),
      });
      expect(res.status).toBe(400);

      const body = await res.json();
      expect(body.error).toContain('root');
    });
  });

  describe('POST /api/convert', () => {
    it('should return 202 with a jobId', async () => {
      const res = await fetch(`${baseUrl}/api/convert`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          root: FIXTURES_CONVERT,
          direction: { from: 'jest', to: 'vitest' },
          outputMode: 'out-dir',
          outputDir: OUTPUT_DIR,
        }),
      });
      expect(res.status).toBe(202);

      const body = await res.json();
      expect(body.jobId).toBeDefined();
      expect(typeof body.jobId).toBe('string');
    });

    it('should return 400 when required fields are missing', async () => {
      const res = await fetch(`${baseUrl}/api/convert`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ root: FIXTURES_CONVERT }),
      });
      expect(res.status).toBe(400);

      const body = await res.json();
      expect(body.error).toBeDefined();
    });
  });

  describe('GET /api/jobs/:id', () => {
    it('should return job status after conversion completes', async () => {
      // Start a conversion job
      const createRes = await fetch(`${baseUrl}/api/convert`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          root: FIXTURES_CONVERT,
          direction: { from: 'jest', to: 'vitest' },
          outputMode: 'out-dir',
          outputDir: OUTPUT_DIR,
        }),
      });
      const { jobId } = await createRes.json();

      // Poll until completed or failed (max 10 seconds)
      let job;
      const deadline = Date.now() + 10000;
      while (Date.now() < deadline) {
        const res = await fetch(`${baseUrl}/api/jobs/${jobId}`);
        job = await res.json();
        if (job.status === 'completed' || job.status === 'failed') break;
        await new Promise((r) => setTimeout(r, 200));
      }

      expect(job.status).toBe('completed');
      expect(job.result).toBeDefined();
      expect(job.result.direction).toEqual({ from: 'jest', to: 'vitest' });
      expect(typeof job.result.filesConverted).toBe('number');
    });

    it('should return 404 for unknown job id', async () => {
      const res = await fetch(`${baseUrl}/api/jobs/nonexistent-id`);
      expect(res.status).toBe(404);
    });
  });

  describe('GET /api/jobs/:id/stream', () => {
    it('should return SSE content-type and event data', async () => {
      // Start a conversion job
      const createRes = await fetch(`${baseUrl}/api/convert`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          root: FIXTURES_CONVERT,
          direction: { from: 'jest', to: 'vitest' },
          outputMode: 'out-dir',
          outputDir: OUTPUT_DIR,
        }),
      });
      const { jobId } = await createRes.json();

      // Wait a moment for the job to produce events
      await new Promise((r) => setTimeout(r, 500));

      const res = await fetch(`${baseUrl}/api/jobs/${jobId}/stream`);
      expect(res.headers.get('content-type')).toBe('text/event-stream');

      const text = await res.text();
      expect(text).toContain('event:');
      expect(text).toContain('data:');
    });
  });

  describe('GET /api/artifacts/:jobId', () => {
    it('should return file paths from a completed job', async () => {
      // Start a conversion job
      const createRes = await fetch(`${baseUrl}/api/convert`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          root: FIXTURES_CONVERT,
          direction: { from: 'jest', to: 'vitest' },
          outputMode: 'out-dir',
          outputDir: path.join(OUTPUT_DIR, 'artifacts-test'),
        }),
      });
      const { jobId } = await createRes.json();

      // Poll until done
      const deadline = Date.now() + 10000;
      let job;
      while (Date.now() < deadline) {
        const res = await fetch(`${baseUrl}/api/jobs/${jobId}`);
        job = await res.json();
        if (job.status === 'completed' || job.status === 'failed') break;
        await new Promise((r) => setTimeout(r, 200));
      }

      const res = await fetch(`${baseUrl}/api/artifacts/${jobId}`);
      expect(res.status).toBe(200);

      const body = await res.json();
      expect(body.jobId).toBe(jobId);
      expect(Array.isArray(body.files)).toBe(true);
    });
  });

  describe('error handling', () => {
    it('should return 404 for unknown routes', async () => {
      const res = await fetch(`${baseUrl}/api/nonexistent`);
      expect(res.status).toBe(404);
    });

    it('should return 400 for invalid JSON body', async () => {
      const res = await fetch(`${baseUrl}/api/analyze`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: '{invalid json',
      });
      expect(res.status).toBe(400);

      const body = await res.json();
      expect(body.error).toContain('Invalid JSON');
    });

    it('should return 403 for non-localhost Host header', async () => {
      // fetch() does not allow overriding Host, so use raw http.get
      const port = server.address;
      const status = await new Promise((resolve, reject) => {
        const req = http.get(
          {
            hostname: '127.0.0.1',
            port,
            path: '/api/health',
            headers: { Host: 'evil.example.com' },
          },
          (res) => resolve(res.statusCode)
        );
        req.on('error', reject);
      });
      expect(status).toBe(403);
    });

    it('should return 413 for oversized request body', async () => {
      // 2 MB payload exceeds the 1 MB limit
      const largeBody = JSON.stringify({ root: 'x'.repeat(2 * 1024 * 1024) });
      const res = await fetch(`${baseUrl}/api/analyze`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: largeBody,
      });
      expect(res.status).toBe(413);

      const body = await res.json();
      expect(body.error).toContain('maximum allowed size');
    });
  });
});
