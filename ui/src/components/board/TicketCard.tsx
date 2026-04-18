import { useDraggable } from '@dnd-kit/core'
import { CSS } from '@dnd-kit/utilities'
import type { Ticket } from '../../types/kanban'
import { Badge, getPriorityVariant, getTagVariant } from '../ui/Badge'

interface TicketCardProps {
  ticket: Ticket
  onClick: (ticket: Ticket) => void
}

function isOverdue(due: string): boolean {
  if (!due) return false
  try {
    return new Date(due) < new Date(new Date().toDateString())
  } catch {
    return false
  }
}

export function TicketCard({ ticket, onClick }: TicketCardProps) {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: ticket.slug,
    data: { ticket },
  })

  const style = transform
    ? { transform: CSS.Translate.toString(transform) }
    : undefined

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      onClick={() => onClick(ticket)}
      className={`bg-bg-elev border border-border rounded-sm px-[11px] py-[10px] cursor-grab transition-[border-color,box-shadow,transform] duration-[120ms] relative shadow-card hover:border-border-strong hover:shadow-pop active:cursor-grabbing ${
        isDragging ? 'opacity-35 rotate-[1.5deg] cursor-grabbing' : ''
      }`}
    >
      <span className="font-mono text-[10.5px] text-fg-faint font-medium tracking-[0.02em] mb-[3px] block">
        {ticket.slug.split('-').slice(-1)[0]}
      </span>

      <div className="flex items-start justify-between gap-1.5 mb-1.5">
        <span className="text-[13px] font-medium text-fg leading-[1.4] break-words flex-1 tracking-[-0.005em]">
          {ticket.title}
        </span>
        {ticket.priority && (
          <Badge label={ticket.priority} variant={getPriorityVariant(ticket.priority)} small />
        )}
      </div>

      <div className="flex gap-1.5 flex-wrap items-center">
        {ticket.tags && ticket.tags.slice(0, 3).map(tag => (
          <Badge key={tag} label={tag} variant={getTagVariant(tag)} small />
        ))}
        {ticket.assignee && (
          <span className="font-mono text-[11.5px] text-fg-faint inline-flex items-center gap-[3px]">
            <span className="bg-bg-hover px-1.5 py-0.5 rounded-lg font-sans text-[11px] text-fg-muted font-medium">
              {ticket.assignee}
            </span>
          </span>
        )}
        {ticket.due && (
          <span className={`font-mono text-[11.5px] inline-flex items-center gap-[3px] ${isOverdue(ticket.due) ? 'text-s-red font-medium' : 'text-fg-faint'}`}>
            {ticket.due}
          </span>
        )}
      </div>
    </div>
  )
}
