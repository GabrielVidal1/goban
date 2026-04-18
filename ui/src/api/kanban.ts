import { apiFetch } from './client'
import type {
  Project,
  BoardData,
  Ticket,
  CreateTicketRequest,
  MoveTicketRequest,
  UpdateFieldRequest,
} from '../types/kanban'

export const api = {
  listProjects(): Promise<Project[]> {
    return apiFetch('/api/projects')
  },

  getProject(name: string): Promise<BoardData> {
    return apiFetch(`/api/projects/${encodeURIComponent(name)}`)
  },

  listColumns(project: string): Promise<string[]> {
    return apiFetch(`/api/projects/${encodeURIComponent(project)}/columns`)
  },

  getTicket(project: string, slug: string): Promise<Ticket> {
    return apiFetch(
      `/api/projects/${encodeURIComponent(project)}/tickets/${encodeURIComponent(slug)}`
    )
  },

  createTicket(data: CreateTicketRequest): Promise<Ticket> {
    return apiFetch('/api/tickets', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  moveTicket(slug: string, data: MoveTicketRequest): Promise<Ticket> {
    return apiFetch(`/api/tickets/${encodeURIComponent(slug)}/move`, {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  updateField(slug: string, data: UpdateFieldRequest): Promise<Ticket> {
    return apiFetch(`/api/tickets/${encodeURIComponent(slug)}/field`, {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  archiveTicket(slug: string, project: string): Promise<void> {
    return apiFetch(
      `/api/tickets/${encodeURIComponent(slug)}?project=${encodeURIComponent(project)}`,
      { method: 'DELETE' }
    )
  },
}
