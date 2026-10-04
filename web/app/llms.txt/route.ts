import { source } from '@/lib/source';
import { llms } from 'fumadocs-core/source';
import { site } from '@/lib/site';
import { siteUrl } from '@/lib/shared';
import { getOtherSuiteProjects } from '@/lib/suite';

export const revalidate = false;

export async function GET() {
  const preamble =
    'CubeAPM CLI is an agent-ready, harness-agnostic terminal tool: any coding agent or agent harness that can run shell commands can manage traces, metrics, and logs with structured JSON/YAML output, read-only safety, and no-input automation. It is an independent, unofficial open-source project and is not affiliated with CubeAPM or its vendor.';
  // index() returns a Promise since fumadocs-core 16.15.17. Its links are
  // root-relative, so make them absolute to include the site's basePath.
  const index = (await llms(source).index()).replace(/\]\((\/[^)]+)\)/g, (_match, path: string) => `](${siteUrl}${path})`);
  const llmsIndex = index.replace('\n\n', `\n\n> ${preamble}\n\n`);
  const related = getOtherSuiteProjects(site.repo)
    .map(({ name, website }) => `- ${name}: ${website}`)
    .join('\n');

  return new Response(
    `${llmsIndex}\n\n## Related CLI sites\n\n${related}\n`,
  );
}
