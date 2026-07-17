'use client';

import { useEffect, useRef, type CSSProperties, type ElementType, type HTMLAttributes, type ReactNode } from 'react';
import { cn } from '@/lib/utils';

interface RevealProps extends Omit<HTMLAttributes<HTMLElement>, 'children'> {
  as?: ElementType;
  children: ReactNode;
  delay?: number | string;
}

export function Reveal({
  as: Component = 'div',
  children,
  className,
  delay,
  style,
  ...props
}: RevealProps) {
  const elementRef = useRef<HTMLElement>(null);

  useEffect(() => {
    const element = elementRef.current;

    if (!element) return;

    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      element.classList.add('is-inview');
      return;
    }

    if (!('IntersectionObserver' in window)) {
      element.classList.add('is-inview');
      return;
    }

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (!entry?.isIntersecting) return;

        element.classList.add('is-inview');
        observer.disconnect();
      },
      { threshold: 0.2 },
    );

    observer.observe(element);

    return () => observer.disconnect();
  }, []);

  const revealStyle = {
    ...style,
    ...(delay === undefined
      ? {}
      : {
          '--reveal-delay':
            typeof delay === 'number' ? `${delay}ms` : delay,
        }),
  } as CSSProperties;

  return (
    <Component
      ref={elementRef}
      className={cn('reveal', className)}
      style={revealStyle}
      {...props}
    >
      {children}
    </Component>
  );
}
