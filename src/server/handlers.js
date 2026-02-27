import fs from 'fs/promises';
import path from 'path';
import { execFile } from 'node:child_process';
import { createRequire } from 'module';
import { sendJson } from './router.js';
import { safePath } from './pathUtils.js';
import {
  createJob,
  getJob,
  updateJob,
  appendLog,
  onJobEvent,
  offJobEvent,
} from './jobStore.js';

const __require = createRequire(import.meta.url);
const version = __require('../../package.json').version;

const serverStart = Date.now();

// ── Duplicated helpers (see plan — avoids touching bin/hamlet.js) ────

function countTodos(content) {
  const matches = content.match(/HAMLET-TODO/g);
  return matches ? matches.length : 0;
}

function buildOutputFilename(sourceBasename, toFramework) {
  const ext = path.extname(sourceBasename);
  const base = path.basename(sourceBasename, ext);
  const cleanBase = base.replace(/\.(cy|spec|test)$/, '');
  if (ext === '.py' || ext === '.java') return cleanBase + ext;
  if (toFramework === 'cypress') return cleanBase + '.cy.js';
  if (toFramework === 'playwright') return cleanBase + '.spec.js';
  return cleanBase + '.test.js';
}

// ── Handlers ─────────────────────────────────────────────────────────

export function handleHealth(req, res) {
  sendJson(res, 200, {
    status: 'ok',
    version,
    uptime: Math.round((Date.now() - serverStart) / 1000),
    root: req.serverRoot || '.',
  });
}

export async function handleAnalyze(req, res) {
  const { root, maxFiles, include, exclude } = req.body;
  if (!root) {
    return sendJson(res, 400, { error: 'Missing required field: root' });
  }

  const { ProjectAnalyzer } = await import('../core/ProjectAnalyzer.js');
  const analyzer = new ProjectAnalyzer();
  const report = await analyzer.analyze(root, {
    maxFiles: maxFiles || 5000,
    include: include || [],
    exclude: exclude || [],
  });
  sendJson(res, 200, report);
}

export function handleConvert(req, res) {
  const { root, direction, outputMode, outputDir, includeFiles, excludeGlobs } =
    req.body;

  if (!root || !direction || !direction.from || !direction.to || !outputMode) {
    return sendJson(res, 400, {
      error:
        'Missing required fields: root, direction.from, direction.to, outputMode',
    });
  }

  const job = createJob({
    root,
    direction,
    outputMode,
    outputDir,
    includeFiles,
    excludeGlobs,
  });

  // Fire-and-forget — errors are captured on the job
  _runConversionJob(job.id);

  sendJson(res, 202, { jobId: job.id });
}

export function handleGetJob(req, res) {
  const job = getJob(req.params.id);
  if (!job) {
    return sendJson(res, 404, { error: 'Job not found' });
  }
  sendJson(res, 200, job);
}

export function handleJobStream(req, res) {
  const job = getJob(req.params.id);
  if (!job) {
    sendJson(res, 404, { error: 'Job not found' });
    return;
  }

  res.writeHead(200, {
    'Content-Type': 'text/event-stream',
    'Cache-Control': 'no-cache',
    Connection: 'keep-alive',
  });

  // Replay existing log entries
  for (const line of job.log) {
    res.write(`event: log\ndata: ${JSON.stringify(line)}\n\n`);
  }

  // If already done, send final event and close
  if (job.status === 'completed' || job.status === 'failed') {
    res.write(
      `event: done\ndata: ${JSON.stringify({ status: job.status })}\n\n`
    );
    res.end();
    return;
  }

  const listener = (event) => {
    if (event.type === 'log') {
      res.write(`event: log\ndata: ${JSON.stringify(event.data)}\n\n`);
    } else if (event.type === 'status') {
      res.write(`event: status\ndata: ${JSON.stringify(event.data)}\n\n`);
      if (event.data.status === 'completed' || event.data.status === 'failed') {
        res.write(`event: done\ndata: ${JSON.stringify(event.data)}\n\n`);
        res.end();
        offJobEvent(job.id, listener);
      }
    }
  };

  onJobEvent(job.id, listener);

  req.on('close', () => {
    offJobEvent(job.id, listener);
  });
}

export function handleArtifacts(req, res) {
  const job = getJob(req.params.jobId);
  if (!job) {
    return sendJson(res, 404, { error: 'Job not found' });
  }

  const files =
    job.result && job.result.files
      ? job.result.files
          .filter((f) => f.status === 'converted')
          .map((f) => f.outputPath)
      : [];

  sendJson(res, 200, { jobId: job.id, files });
}

export async function handleOpen(req, res) {
  const { path: filePath } = req.body;
  if (!filePath) {
    return sendJson(res, 400, { error: 'Missing required field: path' });
  }

  // Reject URL schemes — only allow filesystem paths
  if (/^[a-zA-Z][a-zA-Z0-9+.-]*:/.test(filePath)) {
    return sendJson(res, 400, {
      error: 'URL schemes are not allowed, only filesystem paths',
    });
  }

  // Restrict to project root
  let resolved;
  try {
    resolved = safePath(filePath, req.serverRoot);
  } catch (_e) {
    return sendJson(res, 403, { error: 'Path outside project root' });
  }

  const platform = process.platform;
  let cmd;
  let args;
  if (platform === 'darwin') {
    cmd = 'open';
    args = [resolved];
  } else if (platform === 'win32') {
    cmd = 'cmd';
    args = ['/c', 'start', '', resolved];
  } else {
    cmd = 'xdg-open';
    args = [resolved];
  }

  execFile(cmd, args, (err) => {
    if (err) {
      return sendJson(res, 500, { error: `Failed to open: ${err.message}` });
    }
    sendJson(res, 200, { opened: resolved });
  });
}

export async function handleFile(req, res) {
  const url = new URL(req.url, 'http://localhost');
  const filePath = url.searchParams.get('path');
  if (!filePath) {
    return sendJson(res, 400, { error: 'Missing path query parameter' });
  }

  let resolved;
  try {
    resolved = safePath(filePath, req.serverRoot);
  } catch (_e) {
    return sendJson(res, 403, { error: 'Path outside project root' });
  }

  try {
    const content = await fs.readFile(resolved, 'utf8');
    sendJson(res, 200, { path: filePath, content });
  } catch (err) {
    sendJson(res, 404, { error: `Cannot read file: ${err.message}` });
  }
}

export async function handlePreview(req, res) {
  const { sourcePath, from, to } = req.body;
  if (!sourcePath || !from || !to) {
    return sendJson(res, 400, {
      error: 'Missing required fields: sourcePath, from, to',
    });
  }

  let resolved;
  try {
    resolved = safePath(sourcePath, req.serverRoot);
  } catch (_e) {
    return sendJson(res, 403, { error: 'Path outside project root' });
  }

  try {
    const source = await fs.readFile(resolved, 'utf8');

    const { ConverterFactory } = await import('../core/ConverterFactory.js');
    const converter = await ConverterFactory.createConverter(from, to);
    const converted = await converter.convert(source);

    sendJson(res, 200, { sourcePath, from, to, source, converted });
  } catch (err) {
    sendJson(res, 500, { error: err.message });
  }
}

// ── Conversion job runner ────────────────────────────────────────────

async function _runConversionJob(jobId) {
  const job = getJob(jobId);
  if (!job) return;

  const { root, direction, outputMode, outputDir, includeFiles, excludeGlobs } =
    job.params;
  const { from, to } = direction;

  try {
    updateJob(jobId, {
      status: 'running',
      startedAt: new Date().toISOString(),
    });
    appendLog(jobId, `Starting conversion: ${from} → ${to}`);

    const { Scanner } = await import('../core/Scanner.js');
    const { FileClassifier } = await import('../core/FileClassifier.js');
    const { ConverterFactory } = await import('../core/ConverterFactory.js');

    // Scan files
    const scanner = new Scanner();
    const classifier = new FileClassifier();
    const resolvedRoot = path.resolve(root);

    const scanOptions = {};
    if (excludeGlobs) scanOptions.exclude = excludeGlobs;

    const allFiles = await scanner.scan(resolvedRoot, scanOptions);
    appendLog(jobId, `Scanned ${allFiles.length} files`);

    // Filter to source framework test files
    let testFiles = [];
    for (const file of allFiles) {
      if (includeFiles && !includeFiles.some((p) => file.path.includes(p))) {
        continue;
      }
      try {
        const content = await fs.readFile(file.path, 'utf8');
        const classification = classifier.classify(file.path, content);
        if (
          classification.type === 'test' &&
          classification.framework === from
        ) {
          testFiles.push(file);
        }
      } catch (_e) {
        // Skip unreadable files
      }
    }

    if (testFiles.length === 0) {
      // Fallback: try all JS/TS files
      testFiles = allFiles.filter((f) =>
        /\.(js|ts|tsx|jsx|py|java)$/.test(f.path)
      );
    }

    appendLog(jobId, `Found ${testFiles.length} files to convert`);

    // Create converter
    const converter = await ConverterFactory.createConverter(from, to);
    const resultFiles = [];

    for (const file of testFiles) {
      const relPath =
        file.relativePath || path.relative(resolvedRoot, file.path);
      try {
        const content = await fs.readFile(file.path, 'utf8');
        const converted = await converter.convert(content);
        const todos = countTodos(converted);
        const newFilename = buildOutputFilename(path.basename(file.path), to);

        let outputPath;
        if (outputMode === 'out-dir') {
          const relDir = path.dirname(relPath);
          outputPath = path.join(
            path.resolve(outputDir || path.join(resolvedRoot, 'hamlet-output')),
            relDir === '.' ? '' : relDir,
            newFilename
          );
        } else {
          // in-place: write alongside source
          outputPath = path.join(path.dirname(file.path), newFilename);
        }

        await fs.mkdir(path.dirname(outputPath), { recursive: true });
        await fs.writeFile(outputPath, converted, 'utf8');

        resultFiles.push({
          source: relPath,
          outputPath,
          status: 'converted',
          todosAdded: todos,
        });
        appendLog(jobId, `Converted: ${relPath}`);
      } catch (err) {
        resultFiles.push({
          source: relPath,
          outputPath: null,
          status: 'failed',
          error: err.message,
          todosAdded: 0,
        });
        appendLog(jobId, `Failed: ${relPath} — ${err.message}`);
      }
    }

    const filesConverted = resultFiles.filter(
      (f) => f.status === 'converted'
    ).length;
    const filesFailed = resultFiles.filter((f) => f.status === 'failed').length;
    const totalTodos = resultFiles.reduce((sum, f) => sum + f.todosAdded, 0);

    updateJob(jobId, {
      status: 'completed',
      finishedAt: new Date().toISOString(),
      result: {
        direction: { from, to },
        filesConverted,
        filesFailed,
        todosAdded: totalTodos,
        files: resultFiles,
      },
    });
  } catch (err) {
    updateJob(jobId, {
      status: 'failed',
      finishedAt: new Date().toISOString(),
      error: err.message,
    });
    appendLog(jobId, `Job failed: ${err.message}`);
  }
}
