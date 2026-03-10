import { execFileSync } from 'child_process';
import fs from 'fs/promises';
import path from 'path';
import os from 'os';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '../..');

describe('Package version resolution', () => {
  let extractDir;
  let consumerDir;
  let packageDir;
  let tarball;
  let npmCacheDir;
  let npmEnv;

  beforeAll(async () => {
    npmCacheDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-npm-cache-'));
    npmEnv = {
      ...process.env,
      NPM_CONFIG_CACHE: npmCacheDir,
      npm_config_cache: npmCacheDir,
      npm_config_audit: 'false',
      npm_config_fund: 'false',
      npm_config_update_notifier: 'false',
    };

    // Pack the project into a tarball
    const packOutput = execFileSync('npm', ['pack', '--pack-destination', os.tmpdir()], {
      cwd: rootDir,
      encoding: 'utf8',
      env: npmEnv,
    }).trim();

    // npm pack prints the filename on the last line
    const tarballName = packOutput.split('\n').pop().trim();
    tarball = path.join(os.tmpdir(), tarballName);

    // Extract package tarball
    extractDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-pack-test-'));
    execFileSync('tar', ['-xzf', tarball, '-C', extractDir], {
      encoding: 'utf8',
    });
    packageDir = path.join(extractDir, 'package');

    // Create a temp consumer project that installs the packed artifact.
    consumerDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-consumer-'));
    const consumerNodeModules = path.join(consumerDir, 'node_modules');
    await fs.mkdir(consumerNodeModules, { recursive: true });
    await fs.symlink(
      packageDir,
      path.join(consumerNodeModules, 'hamlet-testframework')
    );

    // Allow extracted package to resolve dependencies from this workspace.
    await fs.symlink(
      path.join(rootDir, 'node_modules'),
      path.join(packageDir, 'node_modules')
    );
  }, 60000);

  afterAll(async () => {
    // Clean up temp dirs and tarball
    if (extractDir) {
      await fs.rm(extractDir, { recursive: true, force: true });
    }
    if (consumerDir) {
      await fs.rm(consumerDir, { recursive: true, force: true });
    }
    if (tarball) {
      await fs.rm(tarball, { force: true });
    }
    if (npmCacheDir) {
      await fs.rm(npmCacheDir, { recursive: true, force: true });
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
        "import { VERSION } from 'hamlet-testframework'; process.stdout.write(VERSION);",
      ],
      {
        cwd: consumerDir,
        encoding: 'utf8',
      }
    );

    expect(result).toBe(pkg.version);
  });

  it('should resolve core subpath exports from packaged artifact', () => {
    const result = execFileSync(
      'node',
      [
        '--input-type=module',
        '-e',
        "import { ConverterFactory, FRAMEWORKS } from 'hamlet-testframework/core'; process.stdout.write(String(Boolean(ConverterFactory && FRAMEWORKS)));",
      ],
      {
        cwd: consumerDir,
        encoding: 'utf8',
      }
    );

    expect(result).toBe('true');
  });
});
