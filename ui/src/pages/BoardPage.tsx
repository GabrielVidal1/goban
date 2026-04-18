import { useParams } from 'react-router'
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
  const { name } = useParams<{ name: string }>()
  const { boardData, setBoardData, loading, error, refetch } = useBoard(name!)
  const { openModal, closeModal } = useModal()
  const { showToast } = useToast()

  const columns = boardData?.columns.map(c => c.name) ?? []

  const handleTicketClick = (ticket: Ticket) => {
    openModal('Ticket', (
      <TicketDetail
        ticket={ticket}
        columns={columns}
        onRefresh={() => {
          closeModal()
          refetch()
        }}
      />
    ))
  }

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
