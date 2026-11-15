import { createMDX } from 'fumadocs-mdx/next';

const withMDX = createMDX();

const basePath = '/cubeapm-cli';

const config = {
  experimental: { agentUpgrade: 'security' },
  output: 'export',
  basePath,
  env: { NEXT_PUBLIC_BASE_PATH: basePath },
  reactStrictMode: true,
};

export default withMDX(config);
