import {
  createJob,
  getJob,
  updateJob,
  appendLog,
  MAX_COMPLETED_JOBS,
  MAX_LOG_LINES,
  _resetStore,
} from '../../src/server/jobStore.js';

describe('jobStore', () => {
  beforeEach(() => {
    _resetStore();
  });

  afterAll(() => {
    _resetStore();
  });

  describe('log truncation', () => {
    it('should truncate logs when MAX_LOG_LINES is exceeded', () => {
      const job = createJob({ test: true });

      // Append more than MAX_LOG_LINES entries
      for (let i = 0; i < MAX_LOG_LINES + 100; i++) {
        appendLog(job.id, `line ${i}`);
      }

      const stored = getJob(job.id);
      expect(stored.log.length).toBe(MAX_LOG_LINES);
      expect(stored.log[0]).toMatch(/\[\.\.\..*earlier lines truncated\]/);
    });

    it('should not truncate logs within the limit', () => {
      const job = createJob({ test: true });

      for (let i = 0; i < 10; i++) {
        appendLog(job.id, `line ${i}`);
      }

      const stored = getJob(job.id);
      expect(stored.log.length).toBe(10);
      expect(stored.log[0]).toBe('line 0');
    });
  });

  describe('completed job eviction', () => {
    it('should evict oldest completed jobs when cap is exceeded', () => {
      const jobIds = [];

      // Create MAX_COMPLETED_JOBS + 5 jobs and complete them all
      for (let i = 0; i < MAX_COMPLETED_JOBS + 5; i++) {
        const job = createJob({ index: i });
        jobIds.push(job.id);
        updateJob(job.id, {
          status: 'completed',
          finishedAt: new Date(Date.now() + i).toISOString(),
        });
      }

      // The first 5 should be evicted
      for (let i = 0; i < 5; i++) {
        expect(getJob(jobIds[i])).toBeNull();
      }

      // The rest should still exist
      for (let i = 5; i < jobIds.length; i++) {
        expect(getJob(jobIds[i])).not.toBeNull();
      }
    });

    it('should not evict active (running/queued) jobs', () => {
      // Fill up with completed jobs
      for (let i = 0; i < MAX_COMPLETED_JOBS; i++) {
        const job = createJob({ index: i });
        updateJob(job.id, {
          status: 'completed',
          finishedAt: new Date(Date.now() + i).toISOString(),
        });
      }

      // Create a running job
      const activeJob = createJob({ type: 'active' });
      updateJob(activeJob.id, { status: 'running' });

      // Add more completed jobs to trigger eviction
      for (let i = 0; i < 5; i++) {
        const job = createJob({ extra: i });
        updateJob(job.id, {
          status: 'completed',
          finishedAt: new Date(
            Date.now() + MAX_COMPLETED_JOBS + i
          ).toISOString(),
        });
      }

      // The active job must still exist
      const stored = getJob(activeJob.id);
      expect(stored).not.toBeNull();
      expect(stored.status).toBe('running');
    });
  });

  describe('failed jobs', () => {
    it('should count failed jobs toward the completed cap', () => {
      const jobIds = [];

      for (let i = 0; i < MAX_COMPLETED_JOBS + 3; i++) {
        const job = createJob({ index: i });
        jobIds.push(job.id);
        updateJob(job.id, {
          status: 'failed',
          finishedAt: new Date(Date.now() + i).toISOString(),
          error: 'test failure',
        });
      }

      // First 3 should be evicted
      for (let i = 0; i < 3; i++) {
        expect(getJob(jobIds[i])).toBeNull();
      }
    });
  });
});
