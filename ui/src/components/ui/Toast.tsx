import { Link } from 'react-router'
import { useToast } from '../../context/ToastContext'

export function ToastList() {
  const { toasts } = useToast()

  if (toasts.length === 0) return null

  return (
    <div className="fixed bottom-5 right-5 z-[2000] flex flex-col-reverse gap-1.5">
      {toasts.map(toast => (
        <div
          key={toast.id}
          className="px-[14px] py-[9px] rounded-sm bg-fg text-bg text-[12.5px] font-medium max-w-[360px] border border-border-strong flex items-center gap-2"
          style={{ boxShadow: 'var(--shadow-pop)', animation: 'toastIn 0.25s cubic-bezier(0.16,1,0.3,1)' }}
        >
          <span className="flex-1">{toast.message}</span>
          {toast.action && (
            <Link
              to={toast.action.href}
              className="underline underline-offset-2 opacity-80 hover:opacity-100 shrink-0"
            >
              {toast.action.label}
            </Link>
          )}
        </div>
      ))}
    </div>
  )
}
