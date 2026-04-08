// PM2 ecosystem config -- reads ports from system.cfg.json (URLs managed via .env files)
const path = require('path')
const fs = require('fs')
const cfg = require('./system.cfg.json')

// Detect if production binary exists (determines run mode: compiled binary vs go run)
const backendBinExists = fs.existsSync(path.resolve(__dirname, cfg.backend.cwd, 'server'))
const nextBuildExists = fs.existsSync(path.resolve(__dirname, cfg.frontend.cwd, '.next'))

const { execSync } = require('child_process')
const backendCwd = path.resolve(__dirname, cfg.backend.cwd)
const frontendCwd = path.resolve(__dirname, cfg.frontend.cwd)
const hasBinary = backendBinExists
const hasNextBuild = nextBuildExists
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
      script: 'npx',
      args: hasNextBuild
        ? `next start --hostname 0.0.0.0 --port ${cfg.frontend.port}`
        : `next dev --hostname 0.0.0.0 --port ${cfg.frontend.port}`,
      interpreter: 'none',
      watch: false,
      autorestart: true,
      max_restarts: 5,
    },
  ],
}
