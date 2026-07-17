export interface SuiteProject {
  name: string;
  website: string;
  repository: string;
}

export const suite: readonly SuiteProject[] = [
  'jira-cli',
  'jenkins-cli',
  'es-cli',
  'grafana-cli',
  'cubeapm-cli',
  'nginxpm-cli',
  'reckon',
].map((name) => ({
  name,
  website: `https://projects.piyushgambhir.com/${name}`,
  repository: `https://github.com/piyush-gambhir/${name}`,
}));

export function getOtherSuiteProjects(currentRepo: string): SuiteProject[] {
  const currentName = currentRepo.split('/').at(-1)?.replace(/\.git$/, '');

  return suite.filter(({ name }) => name !== currentName);
}
