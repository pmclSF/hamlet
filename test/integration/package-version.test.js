import { execFileSync } from 'child_process';
import fs from 'fs/promises';
import path from 'path';
import os from 'os';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '../..');

describe('Package version resolution', () => {
  let tmpDir;
  let tarball;

  beforeAll(async () => {
    // Pack the project into a tarball
    const packOutput = execFileSync('npm', ['pack', '--pack-destination', os.tmpdir()], {
      cwd: rootDir,
      encoding: 'utf8',
    }).trim();

    // npm pack prints the filename on the last line
    const tarballName = packOutput.split('\n').pop().trim();
    tarball = path.join(os.tmpdir(), tarballName);

    // Create a temp directory to simulate a consumer project
    tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-pack-test-'));

    // Initialize a package.json in the temp dir
    execFileSync('npm', ['init', '-y'], {
      cwd: tmpDir,
      encoding: 'utf8',
    });

    // Install the packed tarball
    execFileSync('npm', ['install', tarball], {
      cwd: tmpDir,
      encoding: 'utf8',
    });
  }, 60000);

  afterAll(async () => {
    // Clean up temp dir and tarball
    if (tmpDir) {
      await fs.rm(tmpDir, { recursive: true, force: true });
    }
    if (tarball) {
      await fs.rm(tarball, { force: true });
    }
  });

  it('should export VERSION matching package.json version', async () => {
    const pkg = JSON.parse(
      await fs.readFile(path.join(rootDir, 'package.json'), 'utf8')
    );

    const result = execFileSync(
      'node',
      [
        '--input-type=module',
        '-e',
        "import { VERSION } from 'hamlet-converter'; process.stdout.write(VERSION);",
      ],
      {
        cwd: tmpDir,
        encoding: 'utf8',
      }
    );

    expect(result).toBe(pkg.version);
  });

  it('should export VERSION equal to 2.0.0', async () => {
    const result = execFileSync(
      'node',
      [
        '--input-type=module',
        '-e',
        "import { VERSION } from 'hamlet-converter'; process.stdout.write(VERSION);",
      ],
      {
        cwd: tmpDir,
        encoding: 'utf8',
      }
    );

    const pkg = JSON.parse(
      await fs.readFile(path.join(rootDir, 'package.json'), 'utf8')
    );
    expect(result).toBe(pkg.version);
  });
});
