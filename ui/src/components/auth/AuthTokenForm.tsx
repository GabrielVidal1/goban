import { useState, type FormEvent } from 'react'
import { setAuthToken } from '../../api/client'

interface AuthTokenFormProps {
  onCancel: () => void
}

export function AuthTokenForm({ onCancel }: AuthTokenFormProps) {
  const [token, setToken] = useState('')

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const trimmed = token.trim()
    if (!trimmed) return
    setAuthToken(trimmed)
    window.location.reload()
  }

  return (
    <form onSubmit={handleSubmit}>
      <p className="text-[12.5px] text-fg-muted mb-3 leading-[1.5]">
        The server printed an auth token in its console on startup. Paste it below to continue.
      </p>
      <div className="mb-3">
        <label className="block text-[11.5px] font-medium text-fg-muted mb-1">Auth token</label>
        <input
          type="text"
          value={token}
          onChange={e => setToken(e.target.value)}
          autoFocus
          required
          placeholder="hex token from server logs"
          className={inputClass}
        />
      </div>
      <div className="flex gap-1.5 justify-end mt-4 pt-3 border-t border-border">
        <button type="button" onClick={onCancel} className={secondaryBtn}>Cancel</button>
        <button type="submit" className={primaryBtn}>Save token</button>
      </div>
    </form>
  )
}

const inputClass = 'w-full px-[10px] py-[7px] border border-border rounded-sm text-[13px] bg-bg-elev text-fg min-h-[32px] transition-[border-color,box-shadow] duration-[120ms] focus:outline-none focus:border-accent focus:shadow-[0_0_0_3px_var(--ring)] tracking-[-0.005em] font-mono'
const primaryBtn = 'inline-flex items-center justify-center h-[30px] px-[10px] rounded-sm text-[12.5px] font-medium cursor-pointer transition-all bg-fg text-bg border border-fg hover:bg-fg-muted hover:border-fg-muted disabled:opacity-50 disabled:cursor-not-allowed'
const secondaryBtn = 'inline-flex items-center justify-center h-[30px] px-[10px] rounded-sm text-[12.5px] font-medium cursor-pointer transition-all bg-bg-elev border border-border text-fg hover:bg-bg-hover hover:border-border-strong'
