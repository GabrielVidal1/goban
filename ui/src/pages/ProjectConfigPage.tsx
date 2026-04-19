import { useState, useCallback, useEffect } from 'react'
import { useParams, Link } from 'react-router'
import {
  DndContext,
  PointerSensor,
  KeyboardSensor,
  useSensor,
  useSensors,
  closestCenter,
  type DragEndEvent,
} from '@dnd-kit/core'
import {
  SortableContext,
  arrayMove,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { useToast } from '../context/ToastContext'
import { api } from '../api/kanban'
import type { ProjectConfig } from '../types/kanban'

interface SortableColumnItemProps {
  column: string
  existsOnDisk: boolean
  onRemove: (column: string) => void
}

function SortableColumnItem({ column, existsOnDisk, onRemove }: SortableColumnItemProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: column })
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : existsOnDisk ? 1 : 0.5,
  }
  return (
    <div
      ref={setNodeRef}
      style={style}
      className="flex items-center gap-2 px-3 py-2 bg-bg-elev border border-border rounded-sm hover:border-border-strong transition-colors"
    >
      <span
        {...attributes}
        {...listeners}
        className="text-fg-dim text-[14px] cursor-grab active:cursor-grabbing"
        title="Drag to reorder"
      >
        ⠿
      </span>
      <span className={`text-[13px] flex-1 ${!existsOnDisk ? 'text-fg-faint italic' : 'text-fg'}`}>
        {column}
        {!existsOnDisk && (
          <span className="ml-2 text-[10.5px] text-fg-dim">(not on disk)</span>
        )}
      </span>
      <button
        onClick={() => onRemove(column)}
        className="w-[20px] h-[20px] grid place-items-center rounded-xs bg-transparent border-0 text-fg-dim hover:text-fg hover:bg-bg-hover transition-all cursor-pointer"
        title="Remove"
      >
        ×
      </button>
    </div>
  )
}

export function ProjectConfigPage() {
  const { name } = useParams<{ name: string }>()
  const { showToast } = useToast()
  const [columns, setColumns] = useState<string[]>([])
  const [config, setConfig] = useState<ProjectConfig>({ columnsOrder: [] })
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [shortnameInput, setShortnameInput] = useState('')

  // For adding new column to order
  const [newColumnInput, setNewColumnInput] = useState('')

  const loadData = useCallback(async () => {
    if (!name) return
    try {
      const [cols, cfg] = await Promise.all([
        api.listColumns(name),
        api.getProjectConfig(name),
      ])
      setColumns(cols)
      setConfig(cfg)
      setShortnameInput(cfg.shortname ?? '')
    } catch (err) {
      showToast(err instanceof Error ? err.message : 'Failed to load config', 'error')
    } finally {
      setLoading(false)
    }
  }, [name])

  useEffect(() => { loadData() }, [loadData])

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates })
  )

  const handleSave = useCallback(async () => {
    if (!name) return
    setSaving(true)
    try {
      await api.updateProjectConfig(name, { ...config, shortname: shortnameInput.toUpperCase() || undefined })
      showToast('Configuration saved')
      window.dispatchEvent(new CustomEvent('kanban:refresh'))
    } catch (err) {
      showToast(err instanceof Error ? err.message : 'Failed to save configuration', 'error')
    } finally {
      setSaving(false)
    }
  }, [name, config])

  const handleDragEnd = useCallback((event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id) return
    const oldIndex = config.columnsOrder.indexOf(active.id as string)
    const newIndex = config.columnsOrder.indexOf(over.id as string)
    if (oldIndex === -1 || newIndex === -1) return
    setConfig({ ...config, columnsOrder: arrayMove(config.columnsOrder, oldIndex, newIndex) })
  }, [config])

  const removeColumnFromOrder = useCallback((column: string) => {
    setConfig({
      ...config,
      columnsOrder: config.columnsOrder.filter(c => c !== column),
    })
  }, [config])

  const addColumnToOrder = useCallback((column: string) => {
    if (!column || config.columnsOrder.includes(column)) return
    setConfig({
      ...config,
      columnsOrder: [...config.columnsOrder, column],
    })
    setNewColumnInput('')
  }, [config])

  const unorderedColumns = columns.filter(c => !config.columnsOrder.includes(c))

  if (loading) {
    return <div className="py-24 text-center text-fg-faint text-[13.5px]">Loading…</div>
  }

  return (
    <div className="px-1 max-w-xl mx-auto py-8">
      <div className="flex items-center gap-3 mb-6">
        <Link
          to={`/project/${name}`}
          className="text-fg-muted hover:text-fg text-[12.5px] transition-colors"
        >
          ← Back to board
        </Link>
        <h2 className="text-[17px] font-semibold tracking-[-0.02em] text-fg">Project Config</h2>
      </div>

      <section className="mb-8">
        <h3 className="text-[14px] font-semibold text-fg mb-1.5">Project Shortname</h3>
        <p className="text-[12.5px] text-fg-muted mb-3">
          2–6 uppercase alphanumeric characters (e.g. <code className="font-mono">KBN</code>). When set, new tickets will be named <code className="font-mono">KBN-1</code>, <code className="font-mono">KBN-2</code>, etc.
        </p>
        <input
          type="text"
          value={shortnameInput}
          onChange={(e) => setShortnameInput(e.target.value.toUpperCase().replace(/[^A-Z0-9]/g, '').slice(0, 6))}
          placeholder="e.g. KBN"
          maxLength={6}
          className="w-32 px-3 py-2 bg-bg-elev border border-border rounded-sm text-[13px] font-mono text-fg placeholder:text-fg-dim focus:outline-none focus:border-accent transition-colors"
        />
      </section>

      <section className="mb-8">
        <h3 className="text-[14px] font-semibold text-fg mb-1.5">Column Order</h3>
        <p className="text-[12.5px] text-fg-muted mb-3">
          Drag to reorder columns on the board. Columns not listed below will appear alphabetically at the end.
        </p>

        {config.columnsOrder.length === 0 ? (
          <div className="text-center py-8 px-4 text-fg-dim text-[12px] border border-dashed border-border rounded-sm mb-3">
            No columns ordered yet. Add columns below or from the list on disk.
          </div>
        ) : (
          <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
            <SortableContext items={config.columnsOrder} strategy={verticalListSortingStrategy}>
              <div className="flex flex-col gap-1 mb-3">
                {config.columnsOrder.map(col => (
                  <SortableColumnItem
                    key={col}
                    column={col}
                    existsOnDisk={columns.includes(col)}
                    onRemove={removeColumnFromOrder}
                  />
                ))}
              </div>
            </SortableContext>
          </DndContext>
        )}

        {/* Add column input */}
        <div className="flex gap-2 mb-3">
          <input
            type="text"
            value={newColumnInput}
            onChange={(e) => setNewColumnInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.preventDefault()
                addColumnToOrder(newColumnInput.trim())
              }
            }}
            placeholder="Add column name..."
            className="flex-1 px-3 py-2 bg-bg-elev border border-border rounded-sm text-[13px] text-fg placeholder:text-fg-dim focus:outline-none focus:border-accent transition-colors"
          />
          <button
            onClick={() => addColumnToOrder(newColumnInput.trim())}
            className="px-3 py-2 bg-bg-elev border border-border rounded-sm text-[12.5px] font-medium text-fg hover:bg-bg-hover hover:border-border-strong transition-all cursor-pointer"
          >
            Add
          </button>
        </div>

        {/* Unordered columns preview */}
        {unorderedColumns.length > 0 && (
          <div>
            <h4 className="text-[12px] font-medium text-fg-muted mb-1.5">Unordered (will appear alphabetically)</h4>
            <div className="flex flex-wrap gap-1.5">
              {unorderedColumns.map(col => (
                <span
                  key={col}
                  className="inline-flex items-center gap-1 px-2 py-1 bg-bg-hover border border-border rounded-sm text-[11.5px] text-fg-muted"
                >
                  {col}
                  <button
                    onClick={() => addColumnToOrder(col)}
                    className="text-fg-dim hover:text-accent transition-colors cursor-pointer"
                    title={`Add "${col}" to order`}
                  >
                    +
                  </button>
                </span>
              ))}
            </div>
          </div>
        )}

        <button
          onClick={handleSave}
          disabled={saving}
          className="mt-4 px-4 py-2 bg-fg text-bg border border-fg rounded-sm text-[12.5px] font-medium hover:bg-fg-muted hover:border-fg-muted transition-all cursor-pointer disabled:opacity-50"
        >
          {saving ? 'Saving…' : 'Save'}
        </button>
      </section>

      <section>
        <h3 className="text-[14px] font-semibold text-fg mb-1.5">All Columns on Disk</h3>
        <p className="text-[12.5px] text-fg-muted mb-3">
          These are the actual column directories for this project.
        </p>
        <div className="flex flex-wrap gap-1.5">
          {columns.map(col => (
            <span
              key={col}
              className={`inline-flex items-center px-2 py-1 rounded-sm text-[11.5px] ${
                config.columnsOrder.includes(col)
                  ? 'bg-accent-soft text-fg border border-border'
                  : 'bg-bg-hover text-fg-muted border border-border'
              }`}
            >
              {col}
            </span>
          ))}
        </div>
      </section>
    </div>
  )
}
