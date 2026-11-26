import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  workers: 1,
  timeout: 60_000,
  expect: { timeout: 10_000 },
  reporter: 'list',
  use: {
    baseURL: process.env.E2E_BASE_URL ?? 'http://127.0.0.1:5178',
    browserName: 'chromium',
    headless: true,
    viewport: { width: 390, height: 844 },
    launchOptions: {
      executablePath: process.env.CHROME_PATH ?? '/usr/bin/google-chrome-stable',
      args: ['--no-sandbox'],
    },
  },
  webServer: [
    {
      command: 'go run ./cmd/api',
      cwd: '../backend',
      url: 'http://127.0.0.1:4000/api/health/ready',
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
    },
    {
      command: 'bunx --bun vite --host 127.0.0.1 --port 5178 --strictPort',
      url: 'http://127.0.0.1:5178',
      env: { VITE_API_BASE_URL: '/api' },
      reuseExistingServer: !process.env.CI,
      timeout: 60_000,
    },
  ],
})
