// PM2 ecosystem config -- reads from system.cfg.json with env override support
const path = require('path')
const fs = require('fs')
const cfg = require('./system.cfg.json')

// Apply production overrides if binary exists (prod build indicator)
const backendBinExists = fs.existsSync(path.resolve(__dirname, cfg.backend.cwd, 'server'))
const nextBuildExists = fs.existsSync(path.resolve(__dirname, cfg.frontend.cwd, '.next'))
const isProd = backendBinExists && nextBuildExists
if (isProd && cfg.environments && cfg.environments.production) {
  const env = cfg.environments.production
  if (env.backend) Object.assign(cfg.backend, env.backend)
  if (env.frontend) Object.assign(cfg.frontend, env.frontend)
  if (env.api) Object.assign(cfg.api, env.api)
  if (env.host) cfg.host = env.host
}

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
