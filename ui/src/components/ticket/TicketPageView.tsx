import type { Ticket } from '../../types/kanban'

interface TicketPageViewProps {
  ticket: Ticket | null
}

export function TicketPageView({ ticket }: TicketPageViewProps) {
  if (!ticket) {
    return (
      <div className="py-24 text-center text-fg-muted">
        <p className="text-[18px] font-semibold text-fg mb-1.5">No ticket data</p>
      </div>
    )
  }

  // Placeholder — full layout to be built in the next child ticket.
  return (
    <div className="py-6">
      <h2 className="text-[18px] font-semibold text-fg mb-2">{ticket.title}</h2>
      <p className="text-fg-faint text-sm italic">Ticket layout coming soon.</p>
    </div>
  )
}
