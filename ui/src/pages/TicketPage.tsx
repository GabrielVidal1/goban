import { useState, useEffect } from 'react'
import { useParams } from 'react-router'
import { api } from '../api/kanban'
import { TicketPageView } from '../components/ticket/TicketPageView'
import type { Ticket } from '../types/kanban'

export function TicketPage() {
  const params = useParams<{ name: string; slug: string }>()
  const project = params.name!
  const slug = params.slug!
  const [ticket, setTicket] = useState<Ticket | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)

    api.getTicket(project, slug)
      .then((data) => {
        if (!cancelled) {
          setTicket(data)
          setLoading(false)
        }
      })
      .catch((err: Error) => {
        if (!cancelled) {
          setError(err.message ?? 'Failed to load ticket')
          setLoading(false)
        }
      })

    return () => { cancelled = true }
  }, [project, slug])

  if (loading) {
    return <div className="py-24 text-center text-fg-faint text-[13.5px]">Loading ticket…</div>
  }

  if (error || !ticket) {
    return (
      <div className="py-24 text-center text-fg-muted">
        <p className="text-[18px] font-semibold text-fg mb-1.5">Ticket not found</p>
        <p className="text-[13.5px]">{error ?? 'The requested ticket could not be loaded.'}</p>
      </div>
    )
  }

  return <TicketPageView ticket={ticket} />
}
