import type {
  BoardData,
  CreateTicketRequest,
  MoveTicketRequest,
  Project,
  ProjectConfig,
  Ticket,
  UpdateFieldRequest,
} from "../types/kanban";
import { apiFetch } from "./client";

export const api = {
  listProjects(): Promise<Project[]> {
    return apiFetch("/api/projects");
  },

  getProject(name: string): Promise<BoardData> {
    return apiFetch(`/api/projects/${encodeURIComponent(name)}`);
  },

  listColumns(project: string): Promise<string[]> {
    return apiFetch(`/api/projects/${encodeURIComponent(project)}/columns`);
  },

  getTicket(project: string, slug: string): Promise<Ticket> {
    return apiFetch(
      `/api/projects/${encodeURIComponent(project)}/tickets/${encodeURIComponent(slug)}`,
    );
  },

  createTicket(data: CreateTicketRequest): Promise<Ticket> {
    return apiFetch("/api/tickets", {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  moveTicket(slug: string, data: MoveTicketRequest): Promise<Ticket> {
    return apiFetch(`/api/tickets/${encodeURIComponent(slug)}/move`, {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  updateField(slug: string, data: UpdateFieldRequest): Promise<Ticket> {
    return apiFetch(`/api/tickets/${encodeURIComponent(slug)}/field`, {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  archiveTicket(slug: string, project: string): Promise<void> {
    return apiFetch(
      `/api/tickets/${encodeURIComponent(slug)}?project=${encodeURIComponent(project)}`,
      { method: "DELETE" },
    );
  },

  runScript(
    slug: string,
    data: { project: string },
  ): Promise<{ output: string }> {
    return apiFetch(`/api/tickets/${encodeURIComponent(slug)}/run`, {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  async getProjectConfig(project: string): Promise<ProjectConfig> {
    const res = await apiFetch<Partial<ProjectConfig>>(
      `/api/projects/${encodeURIComponent(project)}/config`,
    );
    return {
      columnsOrder: res.columnsOrder ?? [],
    };
  },

  updateProjectConfig(project: string, cfg: ProjectConfig): Promise<void> {
    return apiFetch(`/api/projects/${encodeURIComponent(project)}/config`, {
      method: "PUT",
      body: JSON.stringify(cfg),
    });
  },
};
