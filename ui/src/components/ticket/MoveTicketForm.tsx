import { useState, type FormEvent } from 'react'
import type { Ticket } from '../../types/kanban'

interface MoveTicketFormProps {
  ticket: Ticket
  columns: string[]
  onSubmit: (targetColumn: string) => Promise<void>
  onCancel: () => void
}

export function MoveTicketForm({ ticket, columns, onSubmit, onCancel }: MoveTicketFormProps) {
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const fd = new FormData(e.currentTarget)
    setSubmitting(true)
    try {
      await onSubmit(fd.get('target_column') as string)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      <p className="text-[13px] text-fg-muted mb-4">
        Move <strong className="text-fg font-medium">{ticket.title}</strong> to:
      </p>

      <div className="mb-4">
        <label className="block text-[11.5px] font-medium text-fg-muted mb-1">Target column</label>
        <select
          name="target_column"
          required
          defaultValue={ticket.column}
          className="w-full px-[10px] py-[7px] border border-border rounded-sm text-[13px] bg-bg-elev text-fg min-h-[32px] focus:outline-none focus:border-accent focus:shadow-[0_0_0_3px_var(--ring)]"
        >
          {columns.map(c => (
            <option key={c} value={c}>{c}</option>
          ))}
        </select>
      </div>

      <div className="flex gap-1.5 justify-end pt-3 border-t border-border">
        <button type="button" onClick={onCancel} className={secondaryBtn}>Cancel</button>
        <button type="submit" disabled={submitting} className={primaryBtn}>
          {submitting ? 'Moving…' : 'Move ticket'}
        </button>
      </div>
    </form>
  )
}

const primaryBtn = 'inline-flex items-center justify-center h-[30px] px-[10px] rounded-sm text-[12.5px] font-medium cursor-pointer transition-all bg-fg text-bg border border-fg hover:bg-fg-muted hover:border-fg-muted disabled:opacity-50'
const secondaryBtn = 'inline-flex items-center justify-center h-[30px] px-[10px] rounded-sm text-[12.5px] font-medium cursor-pointer transition-all bg-bg-elev border border-border text-fg hover:bg-bg-hover hover:border-border-strong'
