import { FloatingHeader } from '@/components/floating-header';

export default function Layout({ children }: LayoutProps<'/'>) {
  return (
    <>
      <FloatingHeader />
      {children}
    </>
  );
}
