/**
 * v-dialog-drag 弹窗拖拽指令
 * 用法：在 <el-dialog> 上添加 v-dialog-drag 即可
 * 效果：鼠标按住弹窗标题栏可拖拽移动弹窗
 */
export const dialogDragDirective = {
  mounted(el) {
    // Element Plus 的 el-dialog 使用 teleport 传送到 body
    // 需要等待 dialog 渲染完成后再绑定事件
    const initDrag = () => {
      // 如果 el 是 overlay，dialog 可能在 body 下
      // 先尝试在 el 内部找，找不到则在全局找
      let dialogEl = el.querySelector('.el-dialog')
      let dialogHeader = el.querySelector('.el-dialog__header')
      
      // 如果没找到，可能是 teleport 到 body 了
      if (!dialogEl || !dialogHeader) {
        // 通过 data-v-xxx 属性或特定标记找到对应的 dialog
        // 或者通过查找当前显示的 dialog
        const allDialogs = document.querySelectorAll('.el-dialog')
        for (const d of allDialogs) {
          const header = d.querySelector('.el-dialog__header')
          if (header && !d._dragInitialized) {
            dialogEl = d
            dialogHeader = header
            break
          }
        }
      }
      
      if (!dialogEl || !dialogHeader || dialogEl._dragInitialized) {
        // 还没渲染好，稍后重试
        setTimeout(initDrag, 100)
        return
      }
      
      dialogEl._dragInitialized = true
      dialogHeader.style.cursor = 'move'
      dialogHeader.style.userSelect = 'none'

      let isDragging = false
      let offsetX = 0
      let offsetY = 0

      const onMouseDown = (e) => {
        if (e.target.closest('.el-dialog__headerbtn')) return
        isDragging = true

        const rect = dialogEl.getBoundingClientRect()
        offsetX = e.clientX - rect.left
        offsetY = e.clientY - rect.top

        // 切换为绝对定位，保持当前位置
        dialogEl.style.position = 'absolute'
        dialogEl.style.margin = '0'
        dialogEl.style.top = rect.top + 'px'
        dialogEl.style.left = rect.left + 'px'
        dialogEl.style.transform = 'none'

        document.addEventListener('mousemove', onMouseMove)
        document.addEventListener('mouseup', onMouseUp)
        e.preventDefault()
      }

      const onMouseMove = (e) => {
        if (!isDragging) return
        const newLeft = e.clientX - offsetX
        const newTop = e.clientY - offsetY
        dialogEl.style.left = newLeft + 'px'
        dialogEl.style.top = newTop + 'px'
      }

      const onMouseUp = () => {
        isDragging = false
        document.removeEventListener('mousemove', onMouseMove)
        document.removeEventListener('mouseup', onMouseUp)
      }

      dialogHeader.addEventListener('mousedown', onMouseDown)

      el._dialogDragCleanup = () => {
        dialogHeader.removeEventListener('mousedown', onMouseDown)
        document.removeEventListener('mousemove', onMouseMove)
        document.removeEventListener('mouseup', onMouseUp)
        if (dialogEl) delete dialogEl._dragInitialized
      }
    }
    
    // 延迟执行，等待 dialog 渲染
    setTimeout(initDrag, 0)
  },

  unmounted(el) {
    if (el._dialogDragCleanup) {
      el._dialogDragCleanup()
      delete el._dialogDragCleanup
    }
  },
}
