import { useState } from 'react'
import type { Ticket } from '../../types/kanban'
import { Badge, getPriorityVariant, getTagVariant } from '../ui/Badge'
import { useModal } from '../../context/ModalContext'
import { MoveTicketForm } from './MoveTicketForm'
import { EditTicketForm } from './EditTicketForm'
import { api } from '../../api/kanban'
import { useToast } from '../../context/ToastContext'
import { ApiError } from '../../api/client'

interface TicketDetailProps {
  ticket: Ticket
  columns: string[]
  onRefresh: () => void
}

export function TicketDetail({ ticket, columns, onRefresh }: TicketDetailProps) {
  const { openModal, closeModal } = useModal()
  const { showToast } = useToast()
  const [isRunning, setIsRunning] = useState(false)

  const handleRunScript = async () => {
    if (isRunning) return
    setIsRunning(true)
    try {
      await api.runScript(ticket.slug, { project: ticket.project })
      showToast(`Script completed for ${ticket.title}`)
      closeModal()
      onRefresh()
    } catch (err) {
      showToast(err instanceof ApiError ? err.message : 'Failed to run script', 'error')
    } finally {
      setIsRunning(false)
    }
  }

  const handleMove = () => {
    openModal('Move ticket', (
      <MoveTicketForm
        ticket={ticket}
        columns={columns}
        onSubmit={async (targetColumn) => {
          try {
            await api.moveTicket(ticket.slug, { project: ticket.project, target_column: targetColumn })
            showToast(`Moved to ${targetColumn}`)
            closeModal()
            onRefresh()
          } catch (err) {
            showToast(err instanceof ApiError ? err.message : 'Failed to move ticket', 'error')
          }
        }}
        onCancel={() => openModal('Ticket', <TicketDetail ticket={ticket} columns={columns} onRefresh={onRefresh} />)}
      />
    ))
  }

  const handleEdit = () => {
    openModal('Edit ticket', (
      <EditTicketForm
        ticket={ticket}
        onSubmit={() => {
          closeModal()
          onRefresh()
        }}
        onCancel={() => openModal('Ticket', <TicketDetail ticket={ticket} columns={columns} onRefresh={onRefresh} />)}
      />
    ))
  }

  const handleArchive = async () => {
    if (!confirm('Archive this ticket?')) return
    try {
      await api.archiveTicket(ticket.slug, ticket.project)
      showToast('Ticket archived')
      closeModal()
      onRefresh()
    } catch (err) {
      showToast(err instanceof ApiError ? err.message : 'Failed to archive', 'error')
    }
  }

  return (
    <div>
      <div className="mb-[14px]">
        <div className="flex items-center gap-2 w-full font-mono text-[11.5px] text-fg-faint mb-0.5">
          <span>{ticket.project}</span>
          <span>/</span>
          <span>{ticket.column}</span>
        </div>
        <h2 className="text-[17px] font-semibold leading-[1.35] tracking-[-0.015em] text-fg w-full">
          {ticket.title}
        </h2>
      </div>

      <div className="border-t border-border pt-1 mb-[14px]">
        {ticket.priority && (
          <MetaRow label="Priority">
            <Badge label={ticket.priority} variant={getPriorityVariant(ticket.priority)} />
          </MetaRow>
        )}
        <MetaRow label="Status">
          <span className="text-fg text-[12.5px]">{ticket.column}</span>
        </MetaRow>
        {ticket.assignee && (
          <MetaRow label="Assignee">
            <span className="bg-bg-hover px-1.5 py-0.5 rounded-lg font-sans text-[11px] text-fg-muted font-medium">
              {ticket.assignee}
            </span>
          </MetaRow>
        )}
        {ticket.due && (
          <MetaRow label="Due">
            <span className={`font-mono text-[11.5px] ${isOverdue(ticket.due) ? 'text-s-red font-medium' : 'text-fg-faint'}`}>
              {ticket.due}
            </span>
          </MetaRow>
        )}
        {ticket.tags && ticket.tags.length > 0 && (
          <MetaRow label="Tags">
            <div className="flex gap-1 flex-wrap">
              {ticket.tags.map(tag => (
                <Badge key={tag} label={tag} variant={getTagVariant(tag)} small />
              ))}
            </div>
          </MetaRow>
        )}
        {ticket.created && (
          <MetaRow label="Created">
            <span className="font-mono text-[11.5px] text-fg-faint">{ticket.created}</span>
          </MetaRow>
        )}
      </div>

      {ticket.body && (
        <div className="mt-1.5">
          <h3 className="text-[11.5px] mb-1.5 font-medium text-fg-faint">Description</h3>
          <pre className="whitespace-pre-wrap break-words leading-[1.6] p-3 bg-bg-sunken border border-border-subtle rounded-sm min-h-12 text-[13px] text-fg-muted font-sans">
            {ticket.body}
          </pre>
        </div>
      )}

      <div className="border-t border-border pt-3 mt-[14px]">
        <div className="flex gap-1.5 flex-wrap items-center">
          <button
            onClick={handleRunScript}
            disabled={isRunning}
            className="inline-flex items-center gap-[5px] h-[30px] px-[10px] rounded-sm text-[12.5px] font-medium cursor-pointer transition-all bg-bg-elev border border-border text-fg hover:bg-bg-hover hover:border-border-strong disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isRunning ? 'Running...' : 'Start'}
          </button>
          <button
            onClick={handleEdit}
            className="inline-flex items-center gap-[5px] h-[30px] px-[10px] rounded-sm text-[12.5px] font-medium cursor-pointer transition-all bg-bg-elev border border-border text-fg hover:bg-bg-hover hover:border-border-strong"
          >
            Edit
          </button>
          <button
            onClick={handleMove}
            className="inline-flex items-center gap-[5px] h-[30px] px-[10px] rounded-sm text-[12.5px] font-medium cursor-pointer transition-all bg-bg-elev border border-border text-fg hover:bg-bg-hover hover:border-border-strong"
          >
            Move
          </button>
          <div className="flex-1" />
          <button
            onClick={handleArchive}
            className="inline-flex items-center gap-[5px] h-[30px] px-[10px] rounded-sm text-[12.5px] font-medium cursor-pointer transition-all bg-transparent border border-border text-s-red hover:bg-s-red-bg hover:border-s-red"
          >
            Archive
          </button>
        </div>
      </div>
    </div>
  )
}

function MetaRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="grid gap-3 py-2 text-[12.5px] border-b border-border-subtle last:border-0" style={{ gridTemplateColumns: '90px 1fr' }}>
      <label className="text-[11.5px] font-medium text-fg-faint">{label}</label>
      <div className="flex items-center gap-1.5 flex-wrap text-fg">{children}</div>
    </div>
  )
}

function isOverdue(due: string): boolean {
  if (!due) return false
  try {
    return new Date(due) < new Date(new Date().toDateString())
  } catch {
    return false
  }
}
