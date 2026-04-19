import { useState, useEffect, useCallback } from 'react'
import { api } from '../api/kanban'
import type { BoardData } from '../types/kanban'

export function useBoard(projectName: string) {
  const [boardData, setBoardData] = useState<BoardData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchBoard = useCallback(async () => {
    try {
      const data = await api.getProject(projectName)
      setBoardData(data)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load board')
    } finally {
      setLoading(false)
    }
  }, [projectName])

  useEffect(() => {
    setLoading(true)
    fetchBoard()
  }, [fetchBoard])

  useEffect(() => {
    const handler = () => fetchBoard()
    window.addEventListener('kanban:refresh', handler)
    return () => window.removeEventListener('kanban:refresh', handler)
  }, [fetchBoard])

  return { boardData, setBoardData, loading, error, refetch: fetchBoard }
}
