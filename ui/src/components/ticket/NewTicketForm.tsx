import { useState, type FormEvent } from 'react'
import type { CreateTicketRequest } from '../../types/kanban'

interface NewTicketFormProps {
  project: string
  columns: string[]
  defaultColumn?: string
  onSubmit: (data: CreateTicketRequest) => Promise<void>
  onCancel: () => void
}

export function NewTicketForm({ project, columns, defaultColumn, onSubmit, onCancel }: NewTicketFormProps) {
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const fd = new FormData(e.currentTarget)
    setSubmitting(true)
    try {
      await onSubmit({
        project,
        title: fd.get('title') as string,
        column: fd.get('column') as string,
        priority: fd.get('priority') as string || undefined,
        assignee: fd.get('assignee') as string || undefined,
        due: fd.get('due') as string || undefined,
        tags: fd.get('tags') as string || undefined,
        body: fd.get('body') as string || undefined,
      })
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      <FormGroup label="Title *">
        <input name="title" type="text" required autoFocus placeholder="Short description" className={inputClass} />
      </FormGroup>

      <FormGroup label="Column *">
        <select name="column" required defaultValue={defaultColumn ?? ''} className={inputClass}>
          <option value="" disabled>Select column</option>
          {columns.map(c => <option key={c} value={c}>{c}</option>)}
        </select>
      </FormGroup>

      <div className="grid grid-cols-2 gap-2.5 mb-3">
        <FormGroup label="Priority">
          <select name="priority" className={inputClass}>
            <option value="">None</option>
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
            <option value="critical">Critical</option>
          </select>
        </FormGroup>
        <FormGroup label="Assignee">
          <input name="assignee" type="text" placeholder="@username" className={inputClass} />
        </FormGroup>
      </div>

      <div className="grid grid-cols-2 gap-2.5 mb-3">
        <FormGroup label="Due date">
          <input name="due" type="date" className={inputClass} />
        </FormGroup>
        <FormGroup label="Tags">
          <input name="tags" type="text" placeholder="bug, feature" className={inputClass} />
        </FormGroup>
      </div>

      <FormGroup label="Description">
        <textarea name="body" rows={3} placeholder="Optional details..." className={`${inputClass} resize-y min-h-[72px] leading-[1.5] py-2`} />
      </FormGroup>

      <div className="flex gap-1.5 justify-end mt-4 pt-3 border-t border-border">
        <button type="button" onClick={onCancel} className={secondaryBtn}>Cancel</button>
        <button type="submit" disabled={submitting} className={primaryBtn}>
          {submitting ? 'Creating…' : 'Create ticket'}
        </button>
      </div>
    </form>
  )
}

function FormGroup({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="mb-3">
      <label className="block text-[11.5px] font-medium text-fg-muted mb-1">{label}</label>
      {children}
    </div>
  )
}

const inputClass = 'w-full px-[10px] py-[7px] border border-border rounded-sm text-[13px] bg-bg-elev text-fg min-h-[32px] transition-[border-color,box-shadow] duration-[120ms] focus:outline-none focus:border-accent focus:shadow-[0_0_0_3px_var(--ring)] tracking-[-0.005em]'
const primaryBtn = 'inline-flex items-center justify-center h-[30px] px-[10px] rounded-sm text-[12.5px] font-medium cursor-pointer transition-all bg-fg text-bg border border-fg hover:bg-fg-muted hover:border-fg-muted disabled:opacity-50 disabled:cursor-not-allowed'
const secondaryBtn = 'inline-flex items-center justify-center h-[30px] px-[10px] rounded-sm text-[12.5px] font-medium cursor-pointer transition-all bg-bg-elev border border-border text-fg hover:bg-bg-hover hover:border-border-strong'
