interface BadgeProps {
  label: string
  variant?: 'priority-critical' | 'priority-high' | 'priority-medium' | 'priority-low' | 'tag-bug' | 'tag-feature' | 'tag-improvement' | 'tag-default'
  small?: boolean
}

const variantStyles: Record<string, string> = {
  'priority-critical': 'bg-s-red-bg text-s-red',
  'priority-high':     'bg-s-amber-bg text-s-amber',
  'priority-medium':   'bg-s-yellow-bg text-s-yellow',
  'priority-low':      'bg-s-green-bg text-s-green',
  'tag-bug':           'bg-s-red-bg text-s-red',
  'tag-feature':       'bg-s-blue-bg text-s-blue',
  'tag-improvement':   'bg-s-green-bg text-s-green',
  'tag-default':       'bg-bg-hover text-fg-muted [&::before]:hidden',
}

export function Badge({ label, variant = 'tag-default', small = false }: BadgeProps) {
  const styles = variantStyles[variant] ?? variantStyles['tag-default']
  return (
    <span
      className={`inline-flex items-center gap-[3px] px-[6px] rounded-[10px] font-medium whitespace-nowrap h-[17px] before:content-[''] before:w-[5px] before:h-[5px] before:rounded-full before:bg-current before:opacity-90 ${styles} ${small ? 'text-[10px] px-[5px] h-[16px]' : 'text-[10.5px]'}`}
    >
      {label}
    </span>
  )
}

export function getPriorityVariant(priority: string): BadgeProps['variant'] {
  switch (priority.toLowerCase()) {
    case 'critical':
    case 'blocker':
      return 'priority-critical'
    case 'high':
      return 'priority-high'
    case 'medium':
      return 'priority-medium'
    case 'low':
      return 'priority-low'
    default:
      return 'tag-default'
  }
}

export function getTagVariant(tag: string): BadgeProps['variant'] {
  switch (tag.toLowerCase()) {
    case 'bug':
    case 'fix':
      return 'tag-bug'
    case 'feature':
    case 'feat':
      return 'tag-feature'
    case 'improvement':
    case 'enhancement':
      return 'tag-improvement'
    default:
      return 'tag-default'
  }
}
