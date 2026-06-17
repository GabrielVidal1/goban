import { useState, useEffect, useCallback } from 'react'
import { Routes, Route, useLocation } from 'react-router'
import { AppHeader } from './components/layout/AppHeader'
import { Modal } from './components/ui/Modal'
import { ToastList } from './components/ui/Toast'
import { HomePage } from './pages/HomePage'
import { BoardPage } from './pages/BoardPage'
import { ProjectConfigPage } from './pages/ProjectConfigPage'
import { useModal } from './context/ModalContext'
import { NewTicketForm } from './components/ticket/NewTicketForm'
import { AuthTokenForm } from './components/auth/AuthTokenForm'
import { useToast } from './context/ToastContext'
import { useEventStream } from './hooks/useEventStream'
import { api } from './api/kanban'
import { ApiError } from './api/client'
import type { Project } from './types/kanban'

function useTheme() {
  const [theme, setTheme] = useState<'light' | 'dark'>(() => {
    const saved = localStorage.getItem('theme')
    if (saved === 'dark' || saved === 'light') return saved
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  })

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
    localStorage.setItem('theme', theme)
  }, [theme])

  const toggle = useCallback(() => setTheme(t => t === 'dark' ? 'light' : 'dark'), [])
  return { theme, toggle }
}

function extractCurrentProject(pathname: string): string | null {
  const match = pathname.match(/^\/project\/([^/]+)/)
  return match ? decodeURIComponent(match[1]) : null
}

export function App() {
  const { theme, toggle } = useTheme()
  const { openModal, closeModal } = useModal()
  const { showToast } = useToast()
  useEventStream()
  const [projects, setProjects] = useState<Project[]>([])
  const location = useLocation()
  const currentProject = extractCurrentProject(location.pathname)

  const refreshProjects = useCallback(() => {
    api.listProjects().then(setProjects).catch(() => {})
  }, [])

  useEffect(() => {
    refreshProjects()
  }, [refreshProjects])

  useEffect(() => {
    let promptOpen = false
    const handler = () => {
      if (promptOpen) return
      promptOpen = true
      openModal('Authentication required', (
        <AuthTokenForm
          onCancel={() => {
            promptOpen = false
            closeModal()
          }}
        />
      ))
    }
    window.addEventListener('kanban:unauthorized', handler)
    return () => window.removeEventListener('kanban:unauthorized', handler)
  }, [openModal, closeModal, showToast, refreshProjects])

  const handleNewTicket = useCallback(() => {
    if (!currentProject) return
    const columns = projects.find(p => p.name === currentProject)?.columns ?? []
    openModal('New ticket', (
      <NewTicketForm
        project={currentProject}
        columns={columns}
        onSubmit={async (data) => {
          try {
            await api.createTicket(data)
            showToast('Ticket created')
            closeModal()
            window.dispatchEvent(new CustomEvent('kanban:refresh'))
          } catch (err) {
            showToast(err instanceof ApiError ? err.message : 'Failed to create ticket', 'error')
            throw err
          }
        }}
        onCancel={closeModal}
      />
    ))
  }, [currentProject, projects, openModal, closeModal, showToast])

  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement).tagName
      if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return
      if (e.key === 'c' && currentProject) handleNewTicket()
    }
    document.addEventListener('keydown', handleKey)
    return () => document.removeEventListener('keydown', handleKey)
  }, [handleNewTicket, currentProject])

  return (
    <>
      <AppHeader
        projects={projects}
        currentProject={currentProject}
        theme={theme}
        onToggleTheme={toggle}
        onNewTicket={handleNewTicket}
      />
      <div className="max-w-full mx-auto px-4">
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/project/:name" element={<BoardPage />} />
          <Route path="/project/:name/ticket/:slug" element={<BoardPage />} />
          <Route path="/project/:name/config" element={<ProjectConfigPage />} />
        </Routes>
      </div>
      <Modal />
      <ToastList />
    </>
  )
}
