'use client';

import {
  type DependencyList,
  type RefObject,
  useLayoutEffect,
} from 'react';
import { gsap } from './gsap';

type GsapEffect = (context: gsap.Context) => void | (() => void);

export function useGsap<T extends Element>(
  effect: GsapEffect,
  dependencies: DependencyList = [],
  scope?: RefObject<T | null>,
) {
  useLayoutEffect(() => {
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;

    let effectCleanup: void | (() => void);
    const context = gsap.context((self) => {
      effectCleanup = effect(self);
    }, scope);

    return () => {
      effectCleanup?.();
      context.revert();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, dependencies);
}
