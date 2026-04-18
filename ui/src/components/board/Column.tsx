import { useDroppable } from '@dnd-kit/core'
import type { ColumnWithTickets, Ticket } from '../../types/kanban'
import { TicketCard } from './TicketCard'

interface ColumnProps {
  column: ColumnWithTickets
  onTicketClick: (ticket: Ticket) => void
  onAddTicket: (columnName: string) => void
}

function PlusIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" width="13" height="13">
      <line x1="12" y1="5" x2="12" y2="19" />
      <line x1="5" y1="12" x2="19" y2="12" />
    </svg>
  )
}

export function Column({ column, onTicketClick, onAddTicket }: ColumnProps) {
  const { setNodeRef, isOver } = useDroppable({ id: column.name })

  return (
    <div
      className={`flex-shrink-0 w-[288px] bg-bg-sunken border rounded-md flex flex-col transition-[border-color,background] duration-[150ms] max-h-[calc(100vh-130px)] ${
        isOver ? 'border-accent bg-accent-soft' : 'border-border'
      }`}
    >
      <div className="px-3 pt-[10px] pb-2 flex items-center gap-2">
        <h3 className="text-[12px] font-semibold text-fg uppercase tracking-[0.06em] leading-[1.2]">
          {column.name}
        </h3>
        <span className="font-mono text-[11px] text-fg-faint px-1.5 bg-bg-hover rounded-[10px] h-[18px] inline-flex items-center font-medium">
          {column.tickets.length}
        </span>
        <button
          onClick={() => onAddTicket(column.name)}
          className="ml-auto w-[22px] h-[22px] rounded-xs bg-transparent border-0 text-fg-faint cursor-pointer grid place-items-center hover:bg-bg-hover hover:text-fg transition-all"
          title={`Add ticket to ${column.name}`}
        >
          <PlusIcon />
        </button>
      </div>

      <div
        ref={setNodeRef}
        className="flex-1 overflow-y-auto px-2 pb-2 flex flex-col gap-1.5 min-h-[60px]"
      >
        {column.tickets.length === 0 ? (
          <div className="text-center py-[18px] px-[10px] text-fg-dim text-[12px] border border-dashed border-border rounded-sm">
            No tickets
          </div>
        ) : (
          column.tickets.map(ticket => (
            <TicketCard key={ticket.slug} ticket={ticket} onClick={onTicketClick} />
          ))
        )}
      </div>
    </div>
  )
}
