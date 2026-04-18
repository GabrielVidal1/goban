import { useState, useEffect } from 'react'
import { Link } from 'react-router'
import { api } from '../api/kanban'
import type { Project } from '../types/kanban'

function ArrowIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" width="13" height="13">
      <polyline points="9 18 15 12 9 6" />
    </svg>
  )
}

export function HomePage() {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.listProjects().then(setProjects).finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="py-24 text-center text-fg-faint text-[13.5px]">Loading…</div>

  if (projects.length === 0) {
    return (
      <div className="text-center py-24 px-6 text-fg-muted">
        <p className="text-[18px] font-semibold text-fg mb-1.5 tracking-[-0.01em]">No projects found</p>
        <p className="text-[13.5px] max-w-[340px] mx-auto">
          Create a project directory inside your kanban directory to get started.
        </p>
      </div>
    )
  }

  return (
    <div className="px-1 py-6 pb-10">
      <div className="flex items-baseline justify-between mb-[18px] gap-3">
        <h1 className="text-[22px] font-semibold tracking-[-0.02em]">Projects</h1>
        <span className="font-mono text-[12px] text-fg-faint">{projects.length}</span>
      </div>

      <div className="grid gap-2" style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))' }}>
        {projects.map(p => (
          <Link
            key={p.name}
            to={`/project/${p.name}`}
            className="bg-bg-elev border border-border rounded-md px-4 py-[14px] flex flex-col gap-1 transition-[border-color,transform,box-shadow] duration-[150ms] relative min-h-[88px] group hover:border-border-strong hover:shadow-pop"
          >
            <h3 className="text-[14px] font-semibold text-fg tracking-[-0.01em]">{p.name}</h3>
            <div className="flex gap-[10px] font-mono text-[11.5px] text-fg-faint mt-auto pt-2">
              <span>{p.columns.length} columns</span>
              <span className="before:content-['·'] before:mr-[10px] before:text-fg-dim">{p.ticket_count} tickets</span>
            </div>
            <span className="absolute top-[14px] right-[14px] w-[22px] h-[22px] rounded-xs bg-bg-sunken text-fg-faint grid place-items-center opacity-0 group-hover:opacity-100 group-hover:translate-x-0.5 transition-all">
              <ArrowIcon />
            </span>
          </Link>
        ))}
      </div>
    </div>
  )
}
