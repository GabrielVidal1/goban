import { useCallback } from 'react'
import {
  DndContext,
  PointerSensor,
  KeyboardSensor,
  useSensor,
  useSensors,
  closestCenter,
  type DragEndEvent,
} from '@dnd-kit/core'
import { sortableKeyboardCoordinates } from '@dnd-kit/sortable'
import type { BoardData, ColumnWithTickets, Ticket } from '../../types/kanban'
import { Column } from './Column'
import { api } from '../../api/kanban'
import { useToast } from '../../context/ToastContext'
import { ApiError } from '../../api/client'

interface BoardProps {
  boardData: BoardData
  setBoardData: (data: BoardData | ((prev: BoardData | null) => BoardData | null)) => void
  onTicketClick: (ticket: Ticket) => void
  onAddTicket: (columnName: string) => void
  onRefresh: () => void
}

function optimisticallyMove(board: BoardData, slug: string, targetColumn: string): BoardData {
  let movedTicket: Ticket | null = null
  const newColumns: ColumnWithTickets[] = board.columns.map(col => ({
    ...col,
    tickets: col.tickets.filter(t => {
      if (t.slug === slug) {
        movedTicket = { ...t, column: targetColumn }
        return false
      }
      return true
    }),
  }))
  if (!movedTicket) return board
  const ticket = movedTicket
  return {
    ...board,
    columns: newColumns.map(col =>
      col.name === targetColumn
        ? { ...col, tickets: [...col.tickets, ticket] }
        : col
    ),
  }
}

export function Board({ boardData, setBoardData, onTicketClick, onAddTicket, onRefresh }: BoardProps) {
  const { showToast } = useToast()

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates })
  )

  const handleDragEnd = useCallback(async (event: DragEndEvent) => {
    const { active, over } = event
    if (!over) return

    const ticket = active.data.current?.ticket as Ticket | undefined
    if (!ticket) return

    const targetColumn = over.id as string
    if (ticket.column === targetColumn) return

    setBoardData(prev => prev ? optimisticallyMove(prev, ticket.slug, targetColumn) : prev)

    try {
      await api.moveTicket(ticket.slug, { project: boardData.project, target_column: targetColumn })
      showToast(`Moved to ${targetColumn}`)
    } catch (err) {
      showToast(err instanceof ApiError ? err.message : 'Failed to move ticket', 'error')
      onRefresh()
    }
  }, [boardData.project, setBoardData, showToast, onRefresh])

  return (
    <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
      <div className="flex gap-2.5 overflow-x-auto px-1 pb-10 pt-1 items-start" style={{ scrollSnapType: 'x proximity', minHeight: 'calc(100vh - 150px)' }}>
        {boardData.columns.map(col => (
          <Column
            key={col.name}
            column={col}
            onTicketClick={onTicketClick}
            onAddTicket={onAddTicket}
          />
        ))}
      </div>
    </DndContext>
  )
}
