import { createContext, useContext, useState, useCallback, useRef, type ReactNode } from 'react'

interface ModalState {
  isOpen: boolean
  title: string
  content: ReactNode | null
}

interface ModalContextValue {
  modal: ModalState
  /**
   * Open the modal. An optional `onClose` runs once when the modal is next
   * closed (Escape / backdrop / X / programmatic `closeModal`) — used to keep
   * deep-link URLs in sync. The callback persists across content swaps within
   * the same session (e.g. opening Edit from the ticket modal) until the modal
   * actually closes; pass `onClose` again only to replace it.
   */
  openModal: (title: string, content: ReactNode, onClose?: () => void) => void
  closeModal: () => void
}

const ModalContext = createContext<ModalContextValue | null>(null)

export function ModalProvider({ children }: { children: ReactNode }) {
  const [modal, setModal] = useState<ModalState>({
    isOpen: false,
    title: '',
    content: null,
  })
  const onCloseRef = useRef<(() => void) | undefined>(undefined)

  const openModal = useCallback((title: string, content: ReactNode, onClose?: () => void) => {
    if (onClose !== undefined) onCloseRef.current = onClose
    setModal({ isOpen: true, title, content })
  }, [])

  const closeModal = useCallback(() => {
    const onClose = onCloseRef.current
    onCloseRef.current = undefined
    setModal(prev => ({ ...prev, isOpen: false }))
    onClose?.()
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
