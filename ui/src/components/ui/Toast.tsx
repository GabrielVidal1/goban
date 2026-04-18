import { useToast } from '../../context/ToastContext'

export function ToastList() {
  const { toasts } = useToast()

  if (toasts.length === 0) return null

  return (
    <div className="fixed bottom-5 right-5 z-[2000] flex flex-col-reverse gap-1.5">
      {toasts.map(toast => (
        <div
          key={toast.id}
          className="px-[14px] py-[9px] rounded-sm bg-fg text-bg text-[12.5px] font-medium max-w-[320px] border border-border-strong"
          style={{ boxShadow: 'var(--shadow-pop)', animation: 'toastIn 0.25s cubic-bezier(0.16,1,0.3,1)' }}
        >
          {toast.message}
        </div>
      ))}
    </div>
  )
}
