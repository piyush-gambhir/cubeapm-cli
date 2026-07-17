import Link from 'next/link';
import type {
  AnchorHTMLAttributes,
  ButtonHTMLAttributes,
  ReactNode,
} from 'react';
import { cn } from '@/lib/utils';

export type OsmoButtonTheme =
  | 'electric'
  | 'coral'
  | 'dark'
  | 'purple'
  | 'neutral'
  | 'light';
export type OsmoButtonRadius = 'square' | 'pill';

type SharedOsmoButtonProps = {
  children: ReactNode;
  className?: string;
  theme?: OsmoButtonTheme;
  radius?: OsmoButtonRadius;
  icon?: ReactNode;
};

type LinkOsmoButtonProps = SharedOsmoButtonProps &
  Omit<
    AnchorHTMLAttributes<HTMLAnchorElement>,
    'children' | 'className' | 'href'
  > & {
    href: string;
    type?: never;
  };

type NativeOsmoButtonProps = SharedOsmoButtonProps &
  Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'children' | 'className'> & {
    href?: never;
  };

export type OsmoButtonProps = LinkOsmoButtonProps | NativeOsmoButtonProps;

function OsmoButtonInner({
  children,
  icon,
}: Pick<SharedOsmoButtonProps, 'children' | 'icon'>) {
  return (
    <>
      <span className="button-bg" aria-hidden="true" />
      <span className="button-label-wrap" aria-hidden="true">
        <span className="button-label is--primary">{children}</span>
        <span className="button-label is--secondary">{children}</span>
      </span>
      {icon ? (
        <span className="button-icon" aria-hidden="true">
          {icon}
        </span>
      ) : null}
    </>
  );
}

export function OsmoButton(props: OsmoButtonProps) {
  const {
    children,
    className,
    theme = 'electric',
    radius = 'square',
    icon,
    ...elementProps
  } = props;
  const classes = cn(
    'osmo-button',
    `is--${theme}`,
    `is--${radius}`,
    className,
  );
  const accessibleLabel = typeof children === 'string' ? children : undefined;

  if ('href' in props && props.href) {
    const { href, ...linkProps } =
      elementProps as Omit<LinkOsmoButtonProps, keyof SharedOsmoButtonProps>;

    return (
      <Link
        {...linkProps}
        href={href}
        className={classes}
        aria-label={linkProps['aria-label'] ?? accessibleLabel}
      >
        <OsmoButtonInner icon={icon}>{children}</OsmoButtonInner>
      </Link>
    );
  }

  const { type = 'button', ...buttonProps } =
    elementProps as Omit<
      NativeOsmoButtonProps,
      keyof SharedOsmoButtonProps
    >;

  return (
    <button
      {...buttonProps}
      type={type}
      className={classes}
      aria-label={buttonProps['aria-label'] ?? accessibleLabel}
    >
      <OsmoButtonInner icon={icon}>{children}</OsmoButtonInner>
    </button>
  );
}
