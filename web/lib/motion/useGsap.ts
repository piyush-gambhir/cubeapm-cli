'use client';

import {
  type DependencyList,
  type RefObject,
  useLayoutEffect,
} from 'react';
import { gsap } from './gsap';

type GsapEffect = (context: gsap.Context) => void | (() => void);

/** Runs a scoped GSAP effect only when motion is allowed, then reverts it. */
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
    // The caller controls when the animation is rebuilt, like React's effects.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, dependencies);
}
