'use client';

import Lenis from 'lenis';
import { type ReactNode, useEffect } from 'react';
import { ScrollTrigger } from './gsap';

const REDUCED_MOTION_QUERY = '(prefers-reduced-motion: reduce)';

export function LenisProvider({ children }: { children: ReactNode }) {
  useEffect(() => {
    const root = document.documentElement;
    const reducedMotion = window.matchMedia(REDUCED_MOTION_QUERY);
    let lenis: Lenis | null = null;

    const destroyLenis = () => {
      if (!lenis) return;
      lenis.off('scroll', ScrollTrigger.update);
      lenis.destroy();
      lenis = null;
    };

    const syncMotionPreference = () => {
      destroyLenis();

      if (reducedMotion.matches) {
        root.removeAttribute('data-lenis');
        return;
      }

      root.setAttribute('data-lenis', 'true');
      lenis = new Lenis({
        autoRaf: true,
        lerp: 0.165,
        wheelMultiplier: 1.25,
        prevent: (element) => Boolean(element.closest('.is--textarea:focus')),
      });
      lenis.on('scroll', ScrollTrigger.update);
      ScrollTrigger.refresh();
    };

    syncMotionPreference();
    reducedMotion.addEventListener('change', syncMotionPreference);

    return () => {
      reducedMotion.removeEventListener('change', syncMotionPreference);
      destroyLenis();
      root.removeAttribute('data-lenis');
    };
  }, []);

  return children;
}
