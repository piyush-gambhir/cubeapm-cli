// Verify the production search asset with the same client used by the UI.
import assert from 'node:assert/strict';
import { readdir, readFile } from 'node:fs/promises';
import { staticClient } from 'fumadocs-core/search/client/orama-static';

const exported = await readFile(new URL('../out/api/search', import.meta.url), 'utf8');
JSON.parse(exported);
const endpoint = 'https://search.test/api/search';
const originalFetch = globalThis.fetch;
globalThis.fetch = async (url) => {
  assert.equal(String(url), endpoint);
  return new Response(exported, { headers: { 'Content-Type': 'application/json' } });
};
try {
  const results = await staticClient({ from: endpoint }).search('installation');
  assert.ok(results.length > 0, 'exported search index returned no installation results');
  assert.ok(results.some((result) => /^\/docs(?:[\/#?]|$)/.test(result.url)), 'missing documentation links');
  console.log(`Static search verified: ${results.length} installation results`);
  // The UI must request the index under the site's basePath, not at the domain root.
  const chunks = new URL('../out/_next/static/chunks/', import.meta.url);
  const files = (await readdir(chunks, { recursive: true })).filter((f) => f.endsWith('.js'));
  const prefixed = await Promise.all(files.map(async (f) => (await readFile(new URL(f, chunks), 'utf8')).includes('/cubeapm-cli/api/search')));
  assert.ok(prefixed.some(Boolean), 'the search client does not request /cubeapm-cli/api/search');
  console.log('Search client uses the basePath endpoint');
} finally {
  globalThis.fetch = originalFetch;
}
