import { source } from '@/lib/source';
import { llms } from 'fumadocs-core/source';
import { site } from '@/lib/site';
import { getOtherSuiteProjects } from '@/lib/suite';

export const revalidate = false;

export function GET() {
  const preamble =
    'CubeAPM CLI is an agent-ready, harness-agnostic terminal tool: any coding agent or agent harness that can run shell commands can manage traces, metrics, and logs with structured JSON/YAML output, read-only safety, and no-input automation. It is an independent, unofficial open-source project and is not affiliated with CubeAPM or its vendor.';
  const index = llms(source).index();
  const llmsIndex = index.replace('\n\n', `\n\n> ${preamble}\n\n`);
  const related = getOtherSuiteProjects(site.repo)
    .map(({ name, website }) => `- ${name}: ${website}`)
    .join('\n');

  return new Response(
    `${llmsIndex}\n\n## Related CLI sites\n\n${related}\n`,
  );
}
