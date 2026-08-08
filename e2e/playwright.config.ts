import { defineConfig } from '@playwright/test';

// The suite runs against the real binary serving the demo docs. `webServer`
// builds and starts it; tests talk to it like any user's browser would.
const PORT = 4665;

export default defineConfig({
  testDir: './tests',
  timeout: 30_000,
  retries: process.env.CI ? 2 : 0,
  reporter: process.env.CI ? 'github' : 'list',
  use: {
    baseURL: `http://127.0.0.1:${PORT}`,
  },
  webServer: {
    command: `sh -c "cd .. && go build -o byakugan . && exec ./byakugan --port ${PORT} testdata/demo"`,
    url: `http://127.0.0.1:${PORT}/api/index.json`,
    reuseExistingServer: false,
    timeout: 120_000,
  },
});
