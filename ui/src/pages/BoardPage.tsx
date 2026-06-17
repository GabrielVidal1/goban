import { useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router'
import { useBoard } from '../hooks/useBoard'
import { Board } from '../components/board/Board'
import { TicketDetail } from '../components/ticket/TicketDetail'
import { NewTicketForm } from '../components/ticket/NewTicketForm'
import { useModal } from '../context/ModalContext'
import { useToast } from '../context/ToastContext'
import { api } from '../api/kanban'
import { ApiError } from '../api/client'
import type { Ticket } from '../types/kanban'

export function BoardPage() {
  const { name, slug } = useParams<{ name: string; slug?: string }>()
  const navigate = useNavigate()
  const { boardData, setBoardData, loading, error, refetch } = useBoard(name!)
  const { openModal, closeModal } = useModal()
  const { showToast } = useToast()

  const columns = boardData?.columns.map(c => c.name) ?? []

  const handleTicketClick = (ticket: Ticket) => {
    navigate(`/project/${encodeURIComponent(name!)}/ticket/${encodeURIComponent(ticket.slug)}`)
  }

  // Deep linking: the ticket modal is a function of the `:slug` URL segment.
  // Opening a ticket pushes the slug into the URL (above); this effect mirrors
  // the URL back into the modal so direct links / refresh / back-button all work.
  // We track the slug we last opened so a board refetch (SSE) doesn't reopen the
  // modal on top of a sub-flow like Edit/Move — we only react to slug changes.
  const openedSlugRef = useRef<string | null>(null)
  useEffect(() => {
    const boardUrl = `/project/${encodeURIComponent(name!)}`
    if (slug && boardData) {
      const ticket = boardData.columns
        .flatMap(c => c.tickets)
        .find(t => t.slug === slug)
      if (!ticket) {
        // Stale or unknown slug — drop back to the board so the URL stays valid.
        openedSlugRef.current = null
        navigate(boardUrl, { replace: true })
        return
      }
      if (openedSlugRef.current !== slug) {
        openedSlugRef.current = slug
        openModal('Ticket', (
          <TicketDetail
            ticket={ticket}
            columns={boardData.columns.map(c => c.name)}
            onRefresh={() => {
              closeModal()
              refetch()
            }}
          />
        ), () => navigate(boardUrl))
      }
    } else if (!slug && openedSlugRef.current !== null) {
      // Navigated away from the ticket (e.g. browser back) — close the modal.
      openedSlugRef.current = null
      closeModal()
    }
  }, [slug, boardData, name, navigate, openModal, closeModal, refetch])

  const handleAddTicket = (columnName: string) => {
    openModal('New ticket', (
      <NewTicketForm
        project={name!}
        columns={columns}
        defaultColumn={columnName}
        onSubmit={async (data) => {
          try {
            await api.createTicket(data)
            showToast('Ticket created')
            closeModal()
            refetch()
          } catch (err) {
            showToast(err instanceof ApiError ? err.message : 'Failed to create ticket', 'error')
            throw err
          }
        }}
        onCancel={closeModal}
      />
    ))
  }

  if (loading) {
    return <div className="py-24 text-center text-fg-faint text-[13.5px]">Loading board…</div>
  }

  if (error || !boardData) {
    return (
      <div className="py-24 text-center text-fg-muted">
        <p className="text-[18px] font-semibold text-fg mb-1.5">Board not found</p>
        <p className="text-[13.5px]">{error ?? 'Project does not exist.'}</p>
      </div>
    )
  }

  return (
    <div className="px-1">
      <div className="flex items-center gap-2.5 py-[18px] pb-3">
        <span className="w-1.5 h-1.5 rounded-full bg-accent flex-shrink-0" style={{ boxShadow: '0 0 0 3px var(--accent-soft)' }} />
        <h2 className="text-[17px] font-semibold tracking-[-0.02em] text-fg">{boardData.project}</h2>
        <span className="ml-auto font-mono text-[12px] text-fg-faint">{boardData.columns.length} columns</span>
      </div>

      <Board
        boardData={boardData}
        setBoardData={setBoardData as (data: typeof boardData | ((prev: typeof boardData | null) => typeof boardData | null)) => void}
        onTicketClick={handleTicketClick}
        onAddTicket={handleAddTicket}
        onRefresh={refetch}
      />
    </div>
  )
}
