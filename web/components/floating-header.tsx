'use client';

import { Moon, Search, Sun } from 'lucide-react';
import Link from 'next/link';
import { useTheme } from 'next-themes';
import { useSearchContext } from 'fumadocs-ui/contexts/search';
import { useCallback, useEffect, useRef, useState } from 'react';
import { OsmoButton } from '@/components/ui/osmo-button';
import { site } from '@/lib/site';

type MenuLink = { label: string; href: string; external?: boolean };

export function FloatingHeader() {
  const { setTheme, resolvedTheme } = useTheme();
  const { setOpenSearch } = useSearchContext();
  const [open, setOpen] = useState(false);
  const [hidden, setHidden] = useState(false);
  const menuButton = useRef<HTMLButtonElement>(null);
  const menu = useRef<HTMLDivElement>(null);
  const focusTrap = useRef<HTMLDivElement>(null);
  const repoUrl = `https://github.com/${site.repo}`;
  const groups: { title: string; links: MenuLink[] }[] = [
    {
      title: 'Documentation',
      links: [
        { label: 'Introduction', href: '/docs' },
        { label: 'Installation', href: '/docs/installation' },
        { label: 'Authentication', href: '/docs/authentication' },
        { label: 'Quick start', href: '/docs/quickstart' },
      ],
    },
    {
      title: 'Project',
      links: [
        { label: 'GitHub', href: repoUrl, external: true },
        { label: 'Releases', href: `${repoUrl}/releases`, external: true },
        { label: 'Issues', href: `${repoUrl}/issues`, external: true },
        { label: 'License', href: `${repoUrl}/blob/main/LICENSE`, external: true },
      ],
    },
  ];

  const closeMenu = useCallback((restoreFocus = true) => {
    setOpen(false);
    if (restoreFocus) requestAnimationFrame(() => menuButton.current?.focus());
  }, []);

  useEffect(() => {
    const root = document.documentElement;
    const previousOverflow = document.body.style.overflow;
    root.toggleAttribute('data-menu-open', open);
    document.body.classList.toggle('menu-open', open);
    if (!open) return;

    document.body.style.overflow = 'hidden';
    requestAnimationFrame(() => getFocusableElements(menu.current)[0]?.focus());

    function onKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        event.preventDefault();
        closeMenu();
        return;
      }

      if (event.key !== 'Tab') return;
      const focusable = getFocusableElements(focusTrap.current);
      if (!focusable.length) return;
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    }

    document.addEventListener('keydown', onKeyDown);
    return () => {
      document.removeEventListener('keydown', onKeyDown);
      document.body.style.overflow = previousOverflow;
      document.body.classList.remove('menu-open');
      root.removeAttribute('data-menu-open');
    };
  }, [closeMenu, open]);

  useEffect(() => {
    if (open) setHidden(false);
  }, [open]);

  useEffect(() => {
    let lastY = window.scrollY;
    const onScroll = () => {
      const y = window.scrollY;
      const delta = y - lastY;
      // Small deltas are ignored so momentum jitter doesn't flicker the bar.
      if (Math.abs(delta) < 8) return;
      setHidden(y > 120 && delta > 0);
      lastY = y;
    };
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => window.removeEventListener('scroll', onScroll);
  }, []);

  const openSearch = () => {
    closeMenu(false);
    setOpenSearch(true);
  };

  return (
    <nav
      className="nav"
      data-nav-status={open ? 'active' : 'not-active'}
      data-nav-hidden={hidden || undefined}
      aria-label="Primary navigation"
    >
      <button
        type="button"
        className="nav__bg"
        aria-label="Close menu"
        tabIndex={open ? 0 : -1}
        onClick={() => closeMenu()}
      />
      <div className="nav-bar__wrap">
        <div className="nav-bar__width">
          <div ref={focusTrap} className="nav-bar">
            <div className="nav-bar__back" aria-hidden="true">
              <span className="nav-bar__bg" />
            </div>

            <div className="nav-bar__top">
              <div className="nav-bar__menu">
                <button
                  ref={menuButton}
                  type="button"
                  className="nav-menu"
                  aria-label={open ? 'Close menu' : 'Open menu'}
                  aria-expanded={open}
                  aria-controls="primary-menu-panel"
                  onClick={() => setOpen((value) => !value)}
                >
                  <span className="nav-menu__icon" aria-hidden="true">
                    <span />
                    <span />
                  </span>
                  <span className="nav-menu__labels" aria-hidden="true">
                    <span className="nav-menu__label is--menu">Menu</span>
                    <span className="nav-menu__label is--close">Close</span>
                  </span>
                </button>
              </div>

              <div className="nav-bar__logo">
                <Link href="/" className="nav-logo" aria-label={`${site.name} home`}>
                  <span className="nav-logo__wordmark">
                    <span aria-hidden>
                      &gt;<span className="nav-logo__cursor">_</span>
                    </span>{' '}
                    {site.binary}
                  </span>
                  <span className="nav-logo__icon" aria-hidden="true">&gt;_</span>
                </Link>
              </div>

              <div className="nav-bar__buttons">
                <button
                  type="button"
                  onClick={openSearch}
                  aria-label="Search"
                  className="nav-icon-button"
                >
                  <Search />
                </button>
                <button
                  type="button"
                  onClick={() => setTheme(resolvedTheme === 'dark' ? 'light' : 'dark')}
                  aria-label="Toggle theme"
                  className="nav-icon-button"
                >
                  <Sun className="hidden dark:block" />
                  <Moon className="dark:hidden" />
                </button>
                <OsmoButton
                  href={repoUrl}
                  target="_blank"
                  rel="noreferrer"
                  theme="dark"
                  radius="pill"
                  className="nav-github"
                >
                  GitHub
                </OsmoButton>
                <OsmoButton href="/docs" radius="pill" className="nav-get-started">
                  Get started
                </OsmoButton>
              </div>
            </div>

            <div
              id="primary-menu-panel"
              className="nav-bar__bottom"
              aria-hidden={!open}
              inert={!open}
            >
              <div className="nav-bar__bottom-overflow">
                <div ref={menu} className="nav-bar__bottom-inner" data-lenis-prevent="">
                  <div className="nav-desktop-menu">
                    {groups.map((group) => (
                      <section className="nav-desktop-menu__col" key={group.title}>
                        <p className="nav-menu__eyebrow">{group.title}</p>
                        <ul>
                          {group.links.map((link, index) => (
                            <li key={link.label}>
                              <Link
                                href={link.href}
                                target={link.external ? '_blank' : undefined}
                                rel={link.external ? 'noreferrer' : undefined}
                                onClick={() => closeMenu(false)}
                              >
                                <span aria-hidden className="nav-menu-link__index">
                                  {String(index + 1).padStart(2, '0')}
                                </span>
                                {link.label}
                              </Link>
                            </li>
                          ))}
                        </ul>
                      </section>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </nav>
  );
}

function getFocusableElements(root: HTMLElement | null) {
  if (!root) return [];

  return Array.from(
    root.querySelectorAll<HTMLElement>(
      'a[href], button:not([disabled]), input:not([disabled]), textarea:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])',
    ),
  ).filter((element) => {
    const style = window.getComputedStyle(element);
    return element.getClientRects().length > 0 && style.visibility !== 'hidden';
  });
}
