<template>
    <div class="history-item" :class="{ active: isActive }" @click="$emit('select')">
        <div class="history-header">
            <span class="history-tag">{{ isFirst ? '当前问题' : '历史问题' }}</span>
            <div class="menu-trigger" @click.stop="toggleMenu" ref="menuTriggerRef">
                <span class="dots">⋮</span>
            </div>
        </div>

        <div class="history-preview" v-html="previewHtml"></div>
        <div class="history-time">{{ time }}</div>
    </div>

    <!-- 使用 Teleport 将菜单渲染到 body，避免被父容器裁剪 -->
    <Teleport to="body">
        <Transition name="menu-fade">
            <div v-if="menuOpen" class="history-menu" :style="menuStyle" @click.stop>
                <div class="menu-item" @click="handleExportMarkdown">
                    <span class="menu-icon">📋</span>
                    <span>导出 Markdown</span>
                </div>
                <div class="menu-item" @click="handleExportImage">
                    <span class="menu-icon">🖼️</span>
                    <span>导出为图片</span>
                </div>
                <div class="menu-divider"></div>
                <div class="menu-item danger" @click="handleDelete">
                    <span class="menu-icon">🗑️</span>
                    <span>删除此会话</span>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue'

const props = defineProps({
    summary: { type: String, default: '' },
    time: { type: String, default: '' },
    isActive: { type: Boolean, default: false },
    isFirst: { type: Boolean, default: false },
    previewHtml: { type: String, default: '' }
})

const emit = defineEmits(['select', 'delete', 'export-markdown', 'export-image'])

const menuOpen = ref(false)
const menuTriggerRef = ref(null)
const menuStyle = reactive({ top: '0px', left: '0px' })

function toggleMenu() {
    if (!menuOpen.value && menuTriggerRef.value) {
        // 计算菜单位置
        const rect = menuTriggerRef.value.getBoundingClientRect()
        menuStyle.top = `${rect.top}px`
        menuStyle.left = `${rect.right + 8}px`
    }
    menuOpen.value = !menuOpen.value
}

function closeMenu() {
    menuOpen.value = false
}

function handleDelete() {
    emit('delete')
    closeMenu()
}

function handleExportMarkdown() {
    emit('export-markdown')
    closeMenu()
}

function handleExportImage() {
    emit('export-image')
    closeMenu()
}

// 点击外部关闭菜单
function handleClickOutside(event) {
    if (menuTriggerRef.value && !menuTriggerRef.value.contains(event.target)) {
        closeMenu()
    }
}

onMounted(() => {
    document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
    document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
/* ========================================
   History Item Card
   ======================================== */

.history-item {
    background: var(--bg-card);
    border-radius: var(--radius-md);
    padding: var(--space-3);
    cursor: pointer;
    transition: all var(--transition-fast);
    border: 1px solid var(--border-subtle);
    position: relative;
    overflow: visible;
    max-height: 120px;
}

/* Left accent bar */
.history-item::before {
    content: '';
    position: absolute;
    left: 0;
    top: 50%;
    transform: translateY(-50%);
    width: 3px;
    height: 0;
    background: var(--color-primary);
    border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
    transition: height var(--transition-fast);
}

.history-item:hover {
    background: var(--bg-card-hover);
    border-color: var(--border-default);
    transform: translateX(2px);
}

.history-item:hover::before {
    height: 40%;
}

.history-item.active {
    background: var(--color-primary-light);
    border-color: var(--color-primary);
}

.history-item.active::before {
    height: 60%;
}

/* Header Row */
.history-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--space-2);
}

.history-tag {
    font-size: var(--text-xs);
    padding: 2px var(--space-2);
    border-radius: var(--radius-sm);
    background: var(--bg-card-hover);
    color: var(--text-muted);
    font-weight: 600;
    letter-spacing: 0.3px;
}

.history-item.active .history-tag {
    background: rgba(16, 185, 129, 0.2);
    color: var(--color-primary);
}

/* Content Preview */
.history-preview {
    font-size: var(--text-sm);
    color: var(--text-primary);
    line-height: 1.5;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    font-weight: 400;
}

/* Timestamp */
.history-time {
    font-size: var(--text-xs);
    color: var(--text-muted);
    margin-top: var(--space-2);
    text-align: right;
    font-family: var(--font-mono);
}

/* ========================================
   Menu Trigger & Dropdown
   ======================================== */

.menu-trigger {
    position: relative;
}

.dots {
    font-size: 14px;
    color: var(--text-muted);
    cursor: pointer;
    padding: var(--space-1);
    border-radius: var(--radius-sm);
    transition: all var(--transition-fast);
    line-height: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
}

.dots:hover {
    background: rgba(255, 255, 255, 0.12);
    color: var(--text-primary);
}

/* Dropdown Menu */
.history-menu {
    position: fixed;
    background: var(--bg-elevated);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-lg);
    z-index: 9999;
    min-width: 160px;
    padding: var(--space-1);
    backdrop-filter: blur(16px);
}

.menu-item {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    font-size: var(--text-sm);
    color: var(--text-secondary);
    cursor: pointer;
    transition: all var(--transition-fast);
    border-radius: var(--radius-sm);
}

.menu-item:hover {
    background: rgba(255, 255, 255, 0.08);
    color: var(--text-primary);
}

.menu-item.danger {
    color: var(--color-error);
}

.menu-item.danger:hover {
    background: rgba(239, 68, 68, 0.12);
}

.menu-icon {
    font-size: 14px;
    width: 18px;
    text-align: center;
}

.menu-divider {
    height: 1px;
    background: var(--border-subtle);
    margin: var(--space-1) 0;
}

/* Menu Animation */
.menu-fade-enter-active,
.menu-fade-leave-active {
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.menu-fade-enter-from,
.menu-fade-leave-to {
    opacity: 0;
    transform: translateY(-6px) scale(0.95);
}
</style>
