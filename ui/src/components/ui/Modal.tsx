import { useEffect, type ReactNode } from 'react'
import { useModal } from '../../context/ModalContext'

function CloseIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" width="14" height="14">
      <line x1="18" y1="6" x2="6" y2="18" />
      <line x1="6" y1="6" x2="18" y2="18" />
    </svg>
  )
}

export function Modal() {
  const { modal, closeModal } = useModal()

  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') closeModal()
    }
    document.addEventListener('keydown', handleKey)
    return () => document.removeEventListener('keydown', handleKey)
  }, [closeModal])

  if (!modal.isOpen) return null

  return (
    <div
      className="fixed inset-0 z-[1000] flex items-center justify-center p-6 backdrop-blur-sm"
      style={{ background: 'rgba(17,17,19,0.4)', animation: 'overlayIn 0.15s ease' }}
      onClick={e => { if (e.target === e.currentTarget) closeModal() }}
    >
      <div
        className="bg-bg-elev border border-border rounded-lg w-full max-w-[520px] max-h-[85vh] overflow-y-auto"
        style={{ boxShadow: 'var(--shadow-modal)', animation: 'modalIn 0.18s cubic-bezier(0.16,1,0.3,1)' }}
        onClick={e => e.stopPropagation()}
      >
        <div className="flex items-center justify-between px-[18px] py-[14px] border-b border-border sticky top-0 bg-bg-elev z-10 rounded-t-lg">
          <h2 className="text-[13.5px] font-semibold tracking-[-0.01em] text-fg">{modal.title}</h2>
          <button
            onClick={closeModal}
            className="w-[26px] h-[26px] grid place-items-center text-fg-faint hover:bg-bg-hover hover:text-fg rounded-xs transition-all cursor-pointer border-0 bg-transparent"
          >
            <CloseIcon />
          </button>
        </div>
        <div className="p-[18px]">
          {modal.content}
        </div>
      </div>
    </div>
  )
}

export function ModalDark({ children }: { children: ReactNode }) {
  return <>{children}</>
}
