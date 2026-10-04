import { createMDX } from 'fumadocs-mdx/next';

const withMDX = createMDX();

// The site is served under projects.piyushgambhir.com/cubeapm-cli.
const basePath = '/cubeapm-cli';

/** @type {import('next').NextConfig} */
const config = {
  output: 'export',
  basePath,
  // Client code that builds URLs by hand (the search index) needs the prefix too.
  env: { NEXT_PUBLIC_BASE_PATH: basePath },
  reactStrictMode: true,
};

export default withMDX(config);
