import { useEffect, useRef } from 'react'
import { useToast } from '../context/ToastContext'

interface FileEvent {
  op: 'create' | 'write' | 'remove'
  project: string
  column: string
  slug: string
  file: string
}

function humanizeSlug(slug: string): string {
  return slug.replace(/-\d{9,}$/, '').replace(/-/g, ' ')
}

export function useEventStream() {
  const { showToast } = useToast()
  const pendingRemovesRef = useRef<Map<string, { ev: FileEvent; timer: number }>>(new Map())

  useEffect(() => {
    const es = new EventSource('/events')
    const pending = pendingRemovesRef.current

    es.addEventListener('filechange', (e: MessageEvent) => {
      window.dispatchEvent(new CustomEvent('kanban:refresh'))

      let ev: FileEvent
      try {
        ev = JSON.parse(e.data) as FileEvent
      } catch {
        showToast('Board updated')
        return
      }

      const title = humanizeSlug(ev.slug)
      const ticketHref = `/project/${encodeURIComponent(ev.project)}/ticket/${encodeURIComponent(ev.slug)}`
      const boardHref = `/project/${encodeURIComponent(ev.project)}`

      if (ev.op === 'remove') {
        const timer = window.setTimeout(() => {
          if (pending.delete(ev.slug)) {
            showToast(`Ticket removed: ${title}`, { action: { label: ev.project, href: boardHref } })
          }
        }, 250)
        pending.set(ev.slug, { ev, timer })
        return
      }

      if (ev.op === 'create') {
        const prior = pending.get(ev.slug)
        if (prior && prior.ev.column !== ev.column) {
          window.clearTimeout(prior.timer)
          pending.delete(ev.slug)
          showToast(`Moved to ${ev.column}: ${title}`, { action: { label: 'Open', href: ticketHref } })
          return
        }
        if (prior) {
          window.clearTimeout(prior.timer)
          pending.delete(ev.slug)
        }
        showToast(`New ticket: ${title}`, { action: { label: 'Open', href: ticketHref } })
        return
      }

      showToast(`Ticket updated: ${title}`, { action: { label: 'Open', href: ticketHref } })
    })

    return () => {
      es.close()
      pending.forEach(p => window.clearTimeout(p.timer))
      pending.clear()
    }
  }, [showToast])
}
