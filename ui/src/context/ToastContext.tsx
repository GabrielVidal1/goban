import { createContext, useContext, useState, useCallback, type ReactNode } from 'react'

interface ToastAction {
  label: string
  href: string
}

interface Toast {
  id: string
  message: string
  type: 'info' | 'error'
  action?: ToastAction
}

interface ShowToastOptions {
  type?: 'info' | 'error'
  action?: ToastAction
}

interface ToastContextValue {
  toasts: Toast[]
  showToast: (message: string, typeOrOpts?: 'info' | 'error' | ShowToastOptions) => void
}

const ToastContext = createContext<ToastContextValue | null>(null)

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([])

  const showToast = useCallback((message: string, typeOrOpts: 'info' | 'error' | ShowToastOptions = 'info') => {
    const opts: ShowToastOptions = typeof typeOrOpts === 'string' ? { type: typeOrOpts } : typeOrOpts
    const id = crypto.randomUUID()
    setToasts(prev => [...prev, { id, message, type: opts.type ?? 'info', action: opts.action }])
    setTimeout(() => {
      setToasts(prev => prev.filter(t => t.id !== id))
    }, 4500)
  }, [])

  return (
    <ToastContext.Provider value={{ toasts, showToast }}>
      {children}
    </ToastContext.Provider>
  )
}

export function useToast() {
  const ctx = useContext(ToastContext)
  if (!ctx) throw new Error('useToast must be used within ToastProvider')
  return ctx
}
