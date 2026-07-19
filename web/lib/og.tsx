import { ImageResponse } from 'next/og';
import { readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { site } from '@/lib/site';

/**
 * Shared social-card renderer for every OG route on this site.
 *
 * This file is template-shared: it is byte-identical across the CLI suite and
 * takes all per-product identity from `lib/site.ts`. Fix it on jira-cli, then
 * copy verbatim.
 */

const OG_WIDTH = 1200;
const OG_HEIGHT = 630;

// satori cannot take woff2, and fetch(new URL(...)) is not implemented during a
// static export build, so the OFL faces are read off disk from node_modules.
const fontBuffer = async (...fontPath: string[]) => {
  const data = await readFile(join(process.cwd(), 'node_modules', ...fontPath));
  return data.buffer.slice(
    data.byteOffset,
    data.byteOffset + data.byteLength,
  ) as ArrayBuffer;
};

const displayFont = fontBuffer(
  '@fontsource',
  'inter',
  'files',
  'inter-latin-400-normal.woff',
);
const monoFont = fontBuffer(
  '@fontsource',
  'jetbrains-mono',
  'files',
  'jetbrains-mono-latin-500-normal.woff',
);

// satori has no oklch() support, so the card uses the hex twin of --site-accent.
const accent = site.accentHex ?? '#f3f4f1';

const ink = '#f3f4f1';
const dim = '#b6b8b3';
const muted = '#7f827b';
const surface = '#131412';
const raised = '#1e201b';

function clamp(value: string, limit: number) {
  const text = value.trim();
  if (text.length <= limit) return text;

  // Cut on a word boundary so the ellipsis never lands mid-word.
  const head = text.slice(0, limit - 1);
  const lastSpace = head.lastIndexOf(' ');
  const cut = lastSpace > limit * 0.6 ? head.slice(0, lastSpace) : head;

  return `${cut.replace(/[\s,.;:]+$/, '')}…`;
}

export interface SocialCardOptions {
  /** Large display line. */
  title: string;
  /** Supporting mono line under the title. */
  description?: string;
  /** Muted mono text on the right of the terminal header row. */
  context: string;
}

export async function socialCard({
  title,
  description,
  context,
}: SocialCardOptions) {
  const body = clamp(description?.trim() || site.description, 155);
  const [display, mono] = await Promise.all([displayFont, monoFont]);

  return new ImageResponse(
    (
      <div
        style={{
          width: '100%',
          height: '100%',
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'space-between',
          padding: '68px 76px',
          color: ink,
          background: surface,
          fontFamily: 'Inter',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            fontFamily: 'JetBrains Mono',
            fontSize: 24,
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <span style={{ color: accent }}>&gt;_</span>
            <span>{site.binary}</span>
          </div>
          <span style={{ color: muted }}>{context}</span>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 26 }}>
          <div
            style={{
              display: 'flex',
              maxWidth: 990,
              fontSize: 84,
              lineHeight: 0.96,
              letterSpacing: '-0.045em',
            }}
          >
            {clamp(title, 60)}
          </div>
          <div
            style={{
              display: 'flex',
              maxWidth: 900,
              color: dim,
              fontFamily: 'JetBrains Mono',
              fontSize: 25,
              lineHeight: 1.35,
              letterSpacing: '-0.01em',
            }}
          >
            {body}
          </div>
        </div>

        <div
          style={{
            display: 'flex',
            alignSelf: 'flex-start',
            alignItems: 'center',
            gap: 14,
            padding: '22px 26px',
            borderRadius: 8,
            background: raised,
            color: dim,
            fontFamily: 'JetBrains Mono',
            fontSize: 23,
          }}
        >
          <span style={{ color: accent }}>$</span>
          <span>{site.binary} --help</span>
          <span
            style={{
              display: 'flex',
              width: 12,
              height: 24,
              marginLeft: 2,
              background: accent,
            }}
          />
        </div>
      </div>
    ),
    {
      width: OG_WIDTH,
      height: OG_HEIGHT,
      fonts: [
        { name: 'Inter', data: display, weight: 400, style: 'normal' },
        { name: 'JetBrains Mono', data: mono, weight: 500, style: 'normal' },
      ],
    },
  );
}
