// PM2 ecosystem config -- reads from system.cfg.json
const path = require('path')
const fs = require('fs')
const cfg = require('./system.cfg.json')

const { execSync } = require('child_process')
const backendCwd = path.resolve(__dirname, cfg.backend.cwd)
const frontendCwd = path.resolve(__dirname, cfg.frontend.cwd)
const hasBinary = fs.existsSync(path.join(backendCwd, 'server'))
const hasNextBuild = fs.existsSync(path.join(frontendCwd, '.next'))
// Detect package manager: prefer pnpm, fallback to npm
let pkg = 'npm'
try { execSync('pnpm --version', { stdio: 'ignore' }); pkg = 'pnpm' } catch {}

module.exports = {
  apps: [
    {
      name: cfg.backend.name,
      cwd: backendCwd,
      script: hasBinary ? './server' : 'go',
      args: hasBinary ? '' : 'run ./cmd/server',
      interpreter: 'none',
      env: {
        BLUEPRINT_DIR: path.resolve(__dirname, cfg.backend.env.BLUEPRINT_DIR),
        MODEL_DIR: path.resolve(__dirname, cfg.backend.env.MODEL_DIR),
      },
      watch: false,
      autorestart: true,
      max_restarts: 5,
    },
    {
      name: cfg.frontend.name,
      cwd: frontendCwd,
      script: pkg,
      args: hasNextBuild ? `start --port ${cfg.frontend.port}` : `dev --port ${cfg.frontend.port}`,
      interpreter: 'none',
      watch: false,
      autorestart: true,
      max_restarts: 5,
    },
  ],
}
