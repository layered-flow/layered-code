const { execSync } = require('child_process');

// Function to check if a command exists
function commandExists(command) {
  try {
    execSync(`${command} --version`, { stdio: 'ignore' });
    return true;
  } catch (error) {
    return false;
  }
}

// Determine which package manager to use
const packageManager = commandExists('pnpm') ? 'pnpm' : 'npm';

console.log(`Using package manager: ${packageManager}`);

module.exports = {
  apps: [
    {
      name: '{{.AppName}}',
      script: packageManager,
      args: 'run dev',
      pid_file: './.layered-code/server.pid',
      error_file: './logs/error.log',
      out_file: './logs/output.log'
    }
  ]
}