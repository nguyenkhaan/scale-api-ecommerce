import type { HTMLAttributes, ReactNode } from 'react'

type SurfaceElement = 'article' | 'div' | 'section'

export type SurfaceProps = HTMLAttributes<HTMLElement> & {
  as?: SurfaceElement
  children: ReactNode
}

export function Surface({ as = 'div', className = '', children, ...props }: SurfaceProps) {
  const Component = as
  return (
    <Component
      {...props}
      className={`rounded-[var(--ss-radius-card)] border border-[var(--ss-color-border)] bg-[var(--ss-color-surface)] ${className}`}
    >
      {children}
    </Component>
  )
}
