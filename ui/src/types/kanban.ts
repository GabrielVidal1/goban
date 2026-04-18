export interface Ticket {
  title: string
  priority: string
  assignee: string
  due: string
  tags: string[]
  created: string
  body: string
  slug: string
  column: string
  project: string
}

export interface ColumnWithTickets {
  name: string
  tickets: Ticket[]
}

export interface BoardData {
  project: string
  columns: ColumnWithTickets[]
}

export interface Project {
  name: string
  columns: string[]
  ticket_count: number
}

export interface CreateTicketRequest {
  project: string
  column: string
  title: string
  priority?: string
  assignee?: string
  due?: string
  tags?: string
  body?: string
}

export interface MoveTicketRequest {
  project: string
  target_column: string
}

export interface UpdateFieldRequest {
  project: string
  field: string
  value: string
}
