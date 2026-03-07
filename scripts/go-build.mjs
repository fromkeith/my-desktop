import { spawn } from 'child_process';
import path from 'path';
import { fileURLToPath } from 'url';
import fs from 'fs';

/**
 * Finds the project root by searching upwards from a starting directory
 * for a marker file, in this case `pnpm-workspace.yaml`.
 * @param {string} startDir The directory to start searching from.
 * @returns {string} The absolute path to the project root.
 */
function findProjectRoot(startDir) {
  let currentDir = startDir;
  while (true) {
    if (fs.existsSync(path.join(currentDir, 'pnpm-workspace.yaml'))) {
      return currentDir;
    }
    const parentDir = path.dirname(currentDir);
    // If we've reached the filesystem root, we can't go any higher.
    if (parentDir === currentDir) {
      throw new Error('Could not find project root. Make sure "pnpm-workspace.yaml" exists in the root.');
    }
    currentDir = parentDir;
  }
}

async function main() {
  try {
    // Get the directory of the current script.
    const __dirname = path.dirname(fileURLToPath(import.meta.url));

    // Get the arguments to forward to `go build` (e.g., "-o", "my-app.exe", ".").
    const buildArgs = process.argv.slice(2);

    // Find the monorepo root directory.
    const projectRoot = findProjectRoot(__dirname);

    // Construct the absolute path for the Go build cache.
    const goCachePath = path.join(projectRoot, '.turbo', 'gocache');

    // Prepare the command and arguments.
    const command = 'go';
    const args = ['build', ...buildArgs];

    console.log(`> go build ${buildArgs.join(' ')}`);

    // Execute the `go build` command.
    const goBuildProcess = spawn(command, args, {
      // Set the environment for the child process.
      env: {
        ...process.env,
        GOCACHE: goCachePath,
      },
      // Inherit stdio to see build output/errors in real-time.
      stdio: 'inherit',
      // Run in a shell to ensure `go` is found in PATH on all systems.
      shell: true,
    });

    // Wait for the process to exit and propagate the exit code.
    goBuildProcess.on('close', (code) => {
      process.exit(code);
    });

    goBuildProcess.on('error', (err) => {
        console.error('Failed to start the go build process.', err);
        process.exit(1);
    });

  } catch (error) {
    console.error('An error occurred in the build script:', error);
    process.exit(1);
  }
}

main();
