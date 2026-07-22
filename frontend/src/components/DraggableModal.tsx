import { Modal, type ModalProps } from 'antd'
import { type CSSProperties, type ReactNode, useRef, useState } from 'react'

type DragState = {
  pointerId: number
  startX: number
  startY: number
  originX: number
  originY: number
}

const passthroughModalRender = (node: ReactNode) => node

export function DraggableModal(props: ModalProps) {
  const { modalRender, afterClose, ...rest } = props
  const [position, setPosition] = useState({ x: 0, y: 0 })
  const dragState = useRef<DragState | null>(null)

  function resetPosition() {
    dragState.current = null
    setPosition({ x: 0, y: 0 })
    afterClose?.()
  }

  return (
    <Modal
      {...rest}
      afterClose={resetPosition}
      modalRender={(node) =>
        renderDraggableModal(
          modalRender ? modalRender(node) : passthroughModalRender(node),
          position,
          setPosition,
          dragState
        )
      }
    />
  )
}

export function draggableModalRender(node: ReactNode) {
  return <DraggableModalShell>{node}</DraggableModalShell>
}

function renderDraggableModal(
  node: ReactNode,
  position: { x: number; y: number },
  setPosition: (position: { x: number; y: number }) => void,
  dragState: React.MutableRefObject<DragState | null>
) {
  return (
    <DraggableModalShell position={position} setPosition={setPosition} dragState={dragState}>
      {node}
    </DraggableModalShell>
  )
}

function DraggableModalShell({
  children,
  position,
  setPosition,
  dragState
}: {
  children: ReactNode
  position?: { x: number; y: number }
  setPosition?: (position: { x: number; y: number }) => void
  dragState?: React.MutableRefObject<DragState | null>
}) {
  const [localPosition, setLocalPosition] = useState({ x: 0, y: 0 })
  const localDragState = useRef<DragState | null>(null)
  const activePosition = position ?? localPosition
  const updatePosition = setPosition ?? setLocalPosition
  const activeDragState = dragState ?? localDragState
  const style: CSSProperties = {
    transform: `translate(${activePosition.x}px, ${activePosition.y}px)`
  }

  return (
    <div
      className="draggable-modal-shell"
      style={style}
      onPointerDownCapture={(event) => {
        if (event.button !== 0 || !isDraggableModalTarget(event.target)) {
          return
        }
        activeDragState.current = {
          pointerId: event.pointerId,
          startX: event.clientX,
          startY: event.clientY,
          originX: activePosition.x,
          originY: activePosition.y
        }
        event.currentTarget.setPointerCapture(event.pointerId)
        event.preventDefault()
      }}
      onPointerMove={(event) => {
        const state = activeDragState.current
        if (!state || state.pointerId !== event.pointerId) {
          return
        }
        updatePosition({
          x: state.originX + event.clientX - state.startX,
          y: state.originY + event.clientY - state.startY
        })
      }}
      onPointerUp={(event) => {
        if (activeDragState.current?.pointerId === event.pointerId) {
          activeDragState.current = null
        }
      }}
      onPointerCancel={() => {
        activeDragState.current = null
      }}
    >
      {children}
    </div>
  )
}

function isDraggableModalTarget(target: EventTarget) {
  if (!(target instanceof HTMLElement)) {
    return false
  }
  if (target.closest('button, a, input, textarea, select, [role="button"]')) {
    return false
  }
  return Boolean(target.closest('.ant-modal-header'))
}
