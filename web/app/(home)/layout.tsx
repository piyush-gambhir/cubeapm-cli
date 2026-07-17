import { FloatingHeader } from '@/components/floating-header';
import { LenisProvider } from '@/lib/motion/LenisProvider';

export default function Layout({ children }: LayoutProps<'/'>) {
  return (
    <LenisProvider>
      <div className="marketing-shell flex min-h-screen flex-col">
        <FloatingHeader />
        {children}
      </div>
    </LenisProvider>
  );
}
