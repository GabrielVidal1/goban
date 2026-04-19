import { useState, type FormEvent } from 'react'
import { api } from '../../api/kanban'
import { useToast } from '../../context/ToastContext'
import { ApiError } from '../../api/client'

interface EditTicketFormProps {
  ticket: {
    slug: string
    project: string
    title: string
    priority: string
    assignee: string
    due: string
    tags: string[]
    body: string
  }
  onSubmit: () => void
  onCancel: () => void
}

export function EditTicketForm({ ticket, onSubmit, onCancel }: EditTicketFormProps) {
  const [submitting, setSubmitting] = useState(false)
  const [title, setTitle] = useState(ticket.title)
  const [priority, setPriority] = useState(ticket.priority)
  const [assignee, setAssignee] = useState(ticket.assignee)
  const [due, setDue] = useState(ticket.due)
  const [tags, setTags] = useState(ticket.tags.join(', '))
  const [body, setBody] = useState(ticket.body)

  const { showToast } = useToast()

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const slug = ticket.slug
    const project = ticket.project
    setSubmitting(true)
    try {
      // Only send fields that actually changed
      if (title !== ticket.title) {
        await api.updateField(slug, { project, field: 'title', value: title })
      }
      if (priority !== ticket.priority) {
        await api.updateField(slug, { project, field: 'priority', value: priority || '' })
      }
      if (assignee !== ticket.assignee) {
        await api.updateField(slug, { project, field: 'assignee', value: assignee || '' })
      }
      if (due !== ticket.due) {
        await api.updateField(slug, { project, field: 'due', value: due || '' })
      }
      const tagsValue = tags.trim()
      if (tagsValue && tagsValue !== ticket.tags.join(',')) {
        await api.updateField(slug, { project, field: 'tags', value: tagsValue })
      }
      if (body !== ticket.body) {
        await api.updateField(slug, { project, field: 'body', value: body || '' })
      }

      showToast('Ticket updated')
      onSubmit()
    } catch (err) {
      showToast(err instanceof ApiError ? err.message : 'Failed to update ticket', 'error')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      <FormGroup label="Title *">
        <input name="title" type="text" required autoFocus value={title} onChange={e => setTitle(e.target.value)} placeholder="Short description" className={inputClass} />
      </FormGroup>

      <div className="grid grid-cols-2 gap-2.5 mb-3">
        <FormGroup label="Priority">
          <select value={priority} onChange={e => setPriority(e.target.value)} className={inputClass}>
            <option value="">None</option>
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
            <option value="critical">Critical</option>
          </select>
        </FormGroup>
        <FormGroup label="Assignee">
          <input name="assignee" type="text" value={assignee} onChange={e => setAssignee(e.target.value)} placeholder="@username" className={inputClass} />
        </FormGroup>
      </div>

      <div className="grid grid-cols-2 gap-2.5 mb-3">
        <FormGroup label="Due date">
          <input name="due" type="date" value={due} onChange={e => setDue(e.target.value)} className={inputClass} />
        </FormGroup>
        <FormGroup label="Tags">
          <input name="tags" type="text" value={tags} onChange={e => setTags(e.target.value)} placeholder="bug, feature" className={inputClass} />
        </FormGroup>
      </div>

      <FormGroup label="Description">
        <textarea name="body" rows={3} value={body} onChange={e => setBody(e.target.value)} placeholder="Optional details..." className={`${inputClass} resize-y min-h-[72px] leading-[1.5] py-2`} />
      </FormGroup>

      <div className="flex gap-1.5 justify-end mt-4 pt-3 border-t border-border">
        <button type="button" onClick={onCancel} className={secondaryBtn}>Cancel</button>
        <button type="submit" disabled={submitting} className={primaryBtn}>
          {submitting ? 'Saving…' : 'Save changes'}
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
