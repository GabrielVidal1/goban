---
title: Edit tickets
created: 2026-04-18
priority: medium
tags: frontend,ux
---
## Plan

**Backend**: No changes needed. `POST /api/tickets/{slug}/field` already handles per-field updates via `kanban.UpdateTicketField`.

**Frontend**:

1. Create `ui/src/components/ticket/EditTicketForm.tsx`
   - Mirror `NewTicketForm.tsx` structure but pre-populate all fields from the current ticket
   - Fields: title, priority, assignee, due, tags (comma-separated), body
   - On submit: call `api.updateField(slug, { project, field, value })` for each changed field
   - Tags: join array → comma string on load; split on save

2. Update `ui/src/components/ticket/TicketDetail.tsx`
   - Add an "Edit" button next to the existing "Move" / "Archive" buttons
   - Toggle `EditTicketForm` in a modal (reuse the existing modal pattern from `NewTicketModal`)
   - On save success: refetch ticket data and close modal

## Files
- `ui/src/components/ticket/EditTicketForm.tsx` — new file
- `ui/src/components/ticket/TicketDetail.tsx` — add edit button + modal
