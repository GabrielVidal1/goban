import { createContext, useContext, useState, useCallback, type ReactNode } from 'react'

interface ModalState {
  isOpen: boolean
  title: string
  content: ReactNode | null
}

interface ModalContextValue {
  modal: ModalState
  openModal: (title: string, content: ReactNode) => void
  closeModal: () => void
}

const ModalContext = createContext<ModalContextValue | null>(null)

export function ModalProvider({ children }: { children: ReactNode }) {
  const [modal, setModal] = useState<ModalState>({
    isOpen: false,
    title: '',
    content: null,
  })

  const openModal = useCallback((title: string, content: ReactNode) => {
    setModal({ isOpen: true, title, content })
  }, [])

  const closeModal = useCallback(() => {
    setModal(prev => ({ ...prev, isOpen: false }))
  }, [])

  return (
    <ModalContext.Provider value={{ modal, openModal, closeModal }}>
      {children}
    </ModalContext.Provider>
  )
}

export function useModal() {
  const ctx = useContext(ModalContext)
  if (!ctx) throw new Error('useModal must be used within ModalProvider')
  return ctx
}
