import { getPageImage, source } from '@/lib/source';
import { notFound } from 'next/navigation';
import { ImageResponse } from 'next/og';
import { site } from '@/lib/site';

import { readFile } from 'node:fs/promises';
import { join } from 'node:path';

const fontBuffer = async (...fontPath: string[]) => {
  const data = await readFile(join(process.cwd(), 'node_modules', ...fontPath));
  return data.buffer.slice(data.byteOffset, data.byteOffset + data.byteLength) as ArrayBuffer;
};

export const revalidate = false;

const inter = fontBuffer('@fontsource', 'inter', 'files', 'inter-latin-400-normal.woff');

const jetbrainsMono = fontBuffer('@fontsource', 'jetbrains-mono', 'files', 'jetbrains-mono-latin-500-normal.woff');

export async function GET(_req: Request, { params }: RouteContext<'/og/docs/[...slug]'>) {
  const { slug } = await params;
  const page = source.getPage(slug.slice(0, -1));
  if (!page) notFound();

  return new ImageResponse(
    <div
      style={{
        display: 'flex',
        width: '100%',
        height: '100%',
        flexDirection: 'column',
        justifyContent: 'space-between',
        padding: '64px 72px',
        color: '#f3f4f1',
        background: '#131412',
        fontFamily: 'Inter',
      }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          color: site.accent,
          fontFamily: 'JetBrains Mono',
          fontSize: 26,
          letterSpacing: '-0.04em',
        }}
      >
        &gt;_ {site.binary}
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
        <div
          style={{
            display: 'flex',
            maxWidth: 980,
            fontSize: 78,
            lineHeight: 0.94,
            letterSpacing: '-0.055em',
          }}
        >
          {page.data.title}
        </div>
        <div
          style={{
            display: 'flex',
            maxWidth: 880,
            color: '#b6b8b3',
            fontSize: 30,
            lineHeight: 1.2,
            letterSpacing: '-0.02em',
          }}
        >
          {page.data.description}
        </div>
      </div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          color: '#7f827b',
          fontFamily: 'JetBrains Mono',
          fontSize: 19,
          letterSpacing: '0.08em',
          textTransform: 'uppercase',
        }}
      >
        <span>Documentation</span>
        <span style={{ color: site.accent }}>projects.piyushgambhir.com/cubeapm-cli</span>
      </div>
    </div>,
    {
      width: 1200,
      height: 630,
      fonts: [
        {
          name: 'Inter',
          data: await inter,
          style: 'normal',
          weight: 400,
        },
        {
          name: 'JetBrains Mono',
          data: await jetbrainsMono,
          style: 'normal',
          weight: 500,
        },
      ],
    },
  );
}

export function generateStaticParams() {
  return source.getPages().map((page) => ({
    lang: page.locale,
    slug: getPageImage(page).segments,
  }));
}
