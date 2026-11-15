import {
  Bot,
  GitBranch,
  KeyRound,
  ListChecks,
  Search,
  Zap,
  type LucideIcon,
} from 'lucide-react';

export interface Feature {
  icon: LucideIcon;
  title: string;
  body: string;
}

export interface SiteConfig {
  name: string;
  binary: string;
  repo: string;
  tagline: string;
  description: string;
  badge: string;
  installCommand: string;
  features: Feature[];
  exampleTitle: string;
  example: string;
  compatible?: string[];
  featuresTitle?: string;
  featuresSubtitle?: string;
  ctaBody?: string;
  accent?: string;
  accentName?: string;
  accentHex?: string;
}

export const site: SiteConfig = {
  name: 'CubeAPM CLI',
  binary: 'cubeapm',
  repo: 'piyush-gambhir/cubeapm-cli',
  tagline: 'Observe CubeAPM from your terminal',
  description:
    'An independent, unofficial open-source CLI for CubeAPM that is not affiliated with CubeAPM or its vendor. Query distributed traces, Prometheus-compatible metrics, and VictoriaLogs-compatible logs from one binary.',
  badge: 'Open-source · Traces, metrics & logs',
  accent: 'oklch(0.72 0.14 300)',
  accentName: 'violet',
  accentHex: '#b28fef',
  installCommand:
    'curl -sSfL https://raw.githubusercontent.com/piyush-gambhir/cubeapm-cli/main/install.sh | sh',
  features: [
    {
      icon: Search,
      title: 'Trace investigations',
      body: 'Discover services, search Jaeger-compatible traces, inspect span waterfalls, map dependencies, and rank callers.',
    },
    {
      icon: KeyRound,
      title: 'Profiles & sessions',
      body: 'Connect with Ory Kratos email/password sessions or no authentication, with named profiles and automatic re-authentication.',
    },
    {
      icon: Bot,
      title: 'Agent-friendly',
      body: '-o json|yaml, --read-only safety mode, --no-input, --quiet, structured errors, and environment-based configuration.',
    },
    {
      icon: GitBranch,
      title: 'Three-signal workflows',
      body: 'Move from a failing trace to PromQL metrics and related LogsQL results without leaving the terminal.',
    },
    {
      icon: Zap,
      title: 'PromQL & LogsQL',
      body: 'Run instant and range metric queries, search logs, aggregate statistics, inspect fields, and probe retention.',
    },
    {
      icon: ListChecks,
      title: 'Ingest & administer',
      body: 'Push metrics and logs through dedicated ingest endpoints, and safely manage asynchronous log deletion tasks.',
    },
  ],
  exampleTitle: 'A cross-signal investigation',
  example: `# Configure a connection profile
cubeapm login
# Find recent errors and inspect one trace
cubeapm traces search --service api-gateway --status error --last 1h -o json
cubeapm traces get <trace-id> -o json
# Correlate logs and check current metrics
cubeapm logs query 'trace_id:<trace-id>' --last 1h -o json
cubeapm metrics query 'up' -o json`,
  compatible: [
    'PromQL',
    'LogsQL',
    'Jaeger',
    'OpenTelemetry',
    'Prometheus',
    'VictoriaLogs',
  ],
};
