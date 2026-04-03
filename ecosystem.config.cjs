// PM2 ecosystem config -- reads from system.cfg.json
const path = require('path')
const fs = require('fs')
const cfg = require('./system.cfg.json')

const backendCwd = path.resolve(__dirname, cfg.backend.cwd)
const hasBinary = fs.existsSync(path.join(backendCwd, 'server'))

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
      cwd: path.resolve(__dirname, cfg.frontend.cwd),
      script: 'pnpm',
      args: `dev --port ${cfg.frontend.port}`,
      interpreter: 'none',
      watch: false,
      autorestart: true,
      max_restarts: 5,
    },
  ],
}
