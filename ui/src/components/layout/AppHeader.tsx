import { Link } from 'react-router'
import type { Project } from '../../types/kanban'

interface AppHeaderProps {
  projects: Project[]
  currentProject: string | null
  theme: 'light' | 'dark'
  onToggleTheme: () => void
  onNewTicket: () => void
}

function SunIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" width="15" height="15">
      <circle cx="12" cy="12" r="5" />
      <line x1="12" y1="1" x2="12" y2="3" />
      <line x1="12" y1="21" x2="12" y2="23" />
      <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" />
      <line x1="18.36" y1="18.36" x2="19.78" y2="19.78" />
      <line x1="1" y1="12" x2="3" y2="12" />
      <line x1="21" y1="12" x2="23" y2="12" />
      <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" />
      <line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
    </svg>
  )
}

function MoonIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" width="15" height="15">
      <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
    </svg>
  )
}

function PlusIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" width="13" height="13">
      <line x1="12" y1="5" x2="12" y2="19" />
      <line x1="5" y1="12" x2="19" y2="12" />
    </svg>
  )
}

export function AppHeader({ projects, currentProject, theme, onToggleTheme, onNewTicket }: AppHeaderProps) {
  return (
    <header
      className="sticky top-0 z-50 border-b border-border"
      style={{
        background: 'color-mix(in srgb, var(--bg) 85%, transparent)',
        backdropFilter: 'saturate(180%) blur(14px)',
        WebkitBackdropFilter: 'saturate(180%) blur(14px)',
      }}
    >
      <div className="flex items-center gap-2 h-11 px-[14px]">
        <Link to="/" className="flex items-center gap-2 font-semibold text-[13px] text-fg pr-[10px] mr-1 border-r border-border h-full tracking-[-0.01em]">
          <span
            className="w-5 h-5 rounded-[5px] grid place-items-center font-mono text-[11px] font-bold"
            style={{ background: 'var(--fg)', color: 'var(--bg)' }}
          >
            K
          </span>
          <span>Kanban</span>
        </Link>

        <nav className="flex gap-0.5 items-center overflow-x-auto flex-1 h-full px-0.5" style={{ scrollbarWidth: 'none' }}>
          {projects.length === 0 ? (
            <span className="text-fg-faint text-[12.5px] px-2 py-1">No projects</span>
          ) : (
            projects.map(p => (
              <Link
                key={p.name}
                to={`/project/${p.name}`}
                className={`text-[12.5px] font-medium px-[10px] py-[5px] rounded-sm whitespace-nowrap transition-all duration-[120ms] ${
                  currentProject === p.name
                    ? 'bg-bg-hover text-fg font-semibold'
                    : 'text-fg-muted hover:bg-bg-hover hover:text-fg'
                }`}
              >
                {p.name}
              </Link>
            ))
          )}
        </nav>

        <div className="flex items-center gap-1.5 h-full pl-2 border-l border-border">
          <button
            onClick={onToggleTheme}
            className="w-[30px] h-[30px] grid place-items-center border border-border bg-bg-elev rounded-sm text-fg-muted hover:text-fg hover:border-border-strong transition-all cursor-pointer"
            title="Toggle theme"
          >
            {theme === 'dark' ? <SunIcon /> : <MoonIcon />}
          </button>

          {currentProject && (
            <button
              onClick={onNewTicket}
              className="inline-flex items-center gap-[5px] h-[30px] px-[10px] rounded-sm text-[12.5px] font-medium cursor-pointer transition-all bg-fg text-bg border border-fg hover:bg-fg-muted hover:border-fg-muted"
            >
              <PlusIcon />
              <span>New ticket</span>
              <span className="inline-flex items-center px-[5px] border border-b-2 border-border-strong rounded font-mono text-[10.5px] text-fg-muted bg-bg-elev h-[18px]" style={{ color: 'var(--bg)', borderColor: 'var(--bg)', opacity: 0.5 }}>C</span>
            </button>
          )}
        </div>
      </div>
    </header>
  )
}
