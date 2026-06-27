<template>
  <div class="icon-picker">
    <!-- 模式切换标签页 -->
    <div class="picker-tabs">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        @click="switchTab(tab.key)"
        :class="['tab-btn', { active: activeTab === tab.key }]"
      >
        <i :class="tab.icon"></i>
        <span>{{ tab.label }}</span>
      </button>
    </div>

    <!-- Font Awesome 模式 -->
    <template v-if="activeTab === 'fa'">
      <div class="icon-picker-header">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="搜索 Font Awesome 图标..."
          class="icon-search"
          @input="filterIcons"
        />
      </div>
      <div class="icon-picker-body">
        <div class="icon-categories">
          <button
            v-for="category in categories"
            :key="category.name"
            @click="activeCategory = category.name"
            :class="['category-btn', { active: activeCategory === category.name }]"
          >
            <i :class="category.icon"></i>
            <span>{{ category.label }}</span>
          </button>
        </div>
        <div class="icon-grid" ref="iconGrid">
          <div
            v-for="icon in filteredIcons"
            :key="icon"
            @click="selectIcon(icon)"
            :class="['icon-item', { active: modelValue === icon }]"
            :title="icon"
          >
            <i :class="icon"></i>
            <span class="icon-name">{{ getIconName(icon) }}</span>
          </div>
        </div>
      </div>
    </template>

    <!-- Iconify 模式 -->
    <template v-if="activeTab === 'iconify'">
      <div class="icon-picker-header">
        <input
          v-model="iconifyQuery"
          type="text"
          placeholder="搜索 Iconify 图标（如 home, github, qq）..."
          class="icon-search"
          @input="onIconifySearch"
        />
        <div v-if="iconifySearching" class="search-hint">
          <i class="fas fa-spinner fa-spin"></i> 搜索中...
        </div>
        <div v-else-if="iconifyResults.length > 0" class="search-hint">
          找到 {{ iconifyTotal }} 个图标
        </div>
      </div>
      <div class="icon-picker-body">
        <div class="icon-grid">
          <div
            v-for="icon in iconifyResults"
            :key="icon"
            @click="selectIcon(icon)"
            :class="['icon-item', { active: modelValue === icon }]"
            :title="icon"
          >
            <iconify-icon :icon="icon" width="1em" height="1em"></iconify-icon>
            <span class="icon-name">{{ icon }}</span>
          </div>
          <div v-if="iconifyResults.length === 0 && iconifyQuery && !iconifySearching" class="no-results">
            <i class="fas fa-search"></i>
            <p>未找到匹配的图标，试试其他关键词</p>
          </div>
        </div>
      </div>
      <div v-if="iconifyHasMore" class="load-more-bar">
        <button @click="loadMoreIconify" class="load-more-btn">
          <i class="fas fa-chevron-down"></i> 加载更多
        </button>
      </div>
    </template>

    <!-- URL 模式 -->
    <template v-if="activeTab === 'url'">
      <div class="icon-picker-header">
        <label class="url-label">输入图片URL</label>
        <input
          v-model="urlInput"
          type="url"
          placeholder="https://example.com/icon.svg"
          class="icon-search"
          @input="onUrlChange"
        />
      </div>
      <div class="icon-picker-body url-preview-body">
        <div class="url-preview-area">
          <div v-if="urlInput" class="url-preview-box">
            <img :src="urlInput" class="url-preview-img" @error="urlError = true" @load="urlError = false" />
            <code class="url-preview-code">{{ urlInput }}</code>
          </div>
          <div v-else class="url-placeholder">
            <i class="fas fa-image"></i>
            <p>输入图片URL预览</p>
          </div>
        </div>
        <div class="url-actions">
          <button @click="selectIcon(urlInput)" :disabled="!urlInput" class="btn-confirm-url">
            <i class="fas fa-check"></i>
            使用此URL
          </button>
        </div>
      </div>
    </template>

    <!-- 底部 -->
    <div class="icon-picker-footer" v-if="modelValue">
      <div class="selected-icon">
        <span>已选择：</span>
        <IconDisplay :icon="modelValue" />
        <code>{{ modelValue }}</code>
      </div>
      <div class="icon-picker-actions">
        <button @click="clearIcon" class="btn-clear">
          <i class="fas fa-times"></i>
          清除
        </button>
        <button @click="closePicker" class="btn-close">
          <i class="fas fa-check"></i>
          确定
        </button>
      </div>
    </div>
    <div class="icon-picker-footer" v-else>
      <div class="selected-icon">
        <span class="text-muted">尚未选择图标</span>
      </div>
      <div class="icon-picker-actions">
        <button @click="closePicker" class="btn-close">
          <i class="fas fa-times"></i>
          关闭
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import IconDisplay from './IconDisplay.vue'

const props = defineProps({
  modelValue: { type: String, default: '' },
})

const emit = defineEmits(['update:modelValue', 'close'])

// ---- Tabs ----
const tabs = [
  { key: 'fa', label: 'Font Awesome', icon: 'fab fa-font-awesome' },
  { key: 'iconify', label: 'Iconify', icon: 'fas fa-search' },
  { key: 'url', label: '图片URL', icon: 'fas fa-link' },
]
const activeTab = ref('fa')

const switchTab = (key) => {
  activeTab.value = key
}

// ---- Font Awesome ----
const searchQuery = ref('')
const activeCategory = ref('all')

const categories = [
  { name: 'all', label: '全部', icon: 'fas fa-th' },
  { name: 'web', label: '网页', icon: 'fas fa-globe' },
  { name: 'social', label: '社交', icon: 'fas fa-share-alt' },
  { name: 'media', label: '媒体', icon: 'fas fa-photo-video' },
  { name: 'business', label: '商业', icon: 'fas fa-briefcase' },
  { name: 'tech', label: '技术', icon: 'fas fa-code' },
  { name: 'other', label: '其他', icon: 'fas fa-ellipsis-h' },
]

const iconLibrary = {
  web: [
    'fas fa-home', 'fas fa-globe', 'fas fa-link', 'fas fa-external-link-alt',
    'fas fa-bookmark', 'fas fa-star', 'fas fa-heart', 'fas fa-thumbs-up',
  ],
  social: [
    'fab fa-github', 'fab fa-twitter', 'fab fa-facebook', 'fab fa-instagram',
    'fab fa-linkedin', 'fab fa-youtube', 'fab fa-telegram', 'fab fa-discord',
    'fab fa-weixin', 'fab fa-qq', 'fab fa-weibo', 'fab fa-bilibili',
  ],
  media: [
    'fas fa-image', 'fas fa-video', 'fas fa-music', 'fas fa-film',
    'fas fa-camera', 'fas fa-microphone', 'fas fa-headphones',
  ],
  business: [
    'fas fa-briefcase', 'fas fa-building', 'fas fa-chart-line', 'fas fa-dollar-sign',
    'fas fa-shopping-cart', 'fas fa-credit-card', 'fas fa-handshake',
  ],
  tech: [
    'fas fa-code', 'fas fa-terminal', 'fas fa-server', 'fas fa-database',
    'fas fa-cloud', 'fas fa-mobile-alt', 'fas fa-laptop', 'fas fa-keyboard',
  ],
  other: [
    'fas fa-envelope', 'fas fa-phone', 'fas fa-map-marker-alt', 'fas fa-calendar',
    'fas fa-clock', 'fas fa-bell', 'fas fa-cog', 'fas fa-user', 'fas fa-users',
  ],
}

const allIcons = computed(() => {
  const icons = []
  Object.values(iconLibrary).forEach(categoryIcons => {
    icons.push(...categoryIcons)
  })
  return icons
})

const filteredIcons = computed(() => {
  let icons = activeCategory.value === 'all'
    ? allIcons.value
    : iconLibrary[activeCategory.value] || []

  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase()
    icons = icons.filter(icon =>
      icon.toLowerCase().includes(query) ||
      getIconName(icon).toLowerCase().includes(query)
    )
  }
  return icons
})

const getIconName = (icon) => {
  const parts = icon.split(' ')
  return parts[parts.length - 1] || icon
}

// ---- Iconify ----
const iconifyQuery = ref('')
const iconifyResults = ref([])
const iconifyTotal = ref(0)
const iconifyOffset = ref(0)
const iconifySearching = ref(false)
const iconifyHasMore = ref(false)
let iconifyTimer = null
const ICONIFY_LIMIT = 60

const onIconifySearch = () => {
  clearTimeout(iconifyTimer)
  iconifyTimer = setTimeout(() => {
    iconifyOffset.value = 0
    iconifyResults.value = []
    fetchIconify()
  }, 300)
}

const fetchIconify = async () => {
  const q = iconifyQuery.value.trim()
  if (!q) {
    iconifyResults.value = []
    iconifyTotal.value = 0
    iconifyHasMore.value = false
    return
  }

  iconifySearching.value = true
  try {
    const url = `https://api.iconify.design/search?query=${encodeURIComponent(q)}&limit=${ICONIFY_LIMIT}&offset=${iconifyOffset.value}`
    const res = await fetch(url)
    const data = await res.json()
    if (iconifyOffset.value === 0) {
      iconifyResults.value = data.icons || []
    } else {
      iconifyResults.value = [...iconifyResults.value, ...(data.icons || [])]
    }
    iconifyTotal.value = data.total || 0
    iconifyHasMore.value = iconifyResults.value.length < (data.total || 0)
  } catch (e) {
    console.error('Iconify search failed:', e)
  } finally {
    iconifySearching.value = false
  }
}

const loadMoreIconify = () => {
  iconifyOffset.value += ICONIFY_LIMIT
  fetchIconify()
}

// ---- URL ----
const urlInput = ref('')
const urlError = ref(false)

watch(activeTab, (tab) => {
  if (tab === 'url' && props.modelValue && (props.modelValue.startsWith('http://') || props.modelValue.startsWith('https://'))) {
    urlInput.value = props.modelValue
  } else {
    urlInput.value = ''
  }
})

const onUrlChange = () => {
  urlError.value = false
}

// ---- Common ----
const selectIcon = (icon) => {
  emit('update:modelValue', icon)
}

const clearIcon = () => {
  emit('update:modelValue', '')
  iconifyQuery.value = ''
  iconifyResults.value = []
  urlInput.value = ''
}

const closePicker = () => {
  emit('close')
}

const filterIcons = () => {
  if (searchQuery.value.trim() && activeCategory.value !== 'all') {
    activeCategory.value = 'all'
  }
}

watch(() => props.modelValue, (newVal) => {
  if (newVal && (newVal.startsWith('http://') || newVal.startsWith('https://'))) {
    urlInput.value = newVal
  }
})
</script>

<style scoped>
.icon-picker {
  display: flex;
  flex-direction: column;
  height: 100%;
  max-height: 100%;
  background: rgba(var(--background-color-rgb), 0.98);
  border-radius: var(--border-radius);
  overflow: hidden;
}

/* ---- Tabs ---- */
.picker-tabs {
  display: flex;
  gap: 0;
  border-bottom: 1px solid var(--border-color);
  background: rgba(var(--background-color-rgb), 0.9);
  flex-shrink: 0;
}

.tab-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 12px 8px;
  border: none;
  border-bottom: 2px solid transparent;
  background: transparent;
  color: var(--text-color);
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s ease;
  opacity: 0.6;
}

.tab-btn:hover {
  opacity: 0.9;
  background: rgba(var(--background-color-rgb), 0.5);
}

.tab-btn.active {
  opacity: 1;
  border-bottom-color: var(--hover-link-color);
  color: var(--hover-link-color);
}

/* ---- Header ---- */
.icon-picker-header {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.icon-search {
  width: 100%;
  padding: 10px 16px;
  border: 2px solid var(--border-color);
  border-radius: 8px;
  background: rgba(var(--background-color-rgb), 0.6);
  color: var(--text-color);
  font-size: 14px;
  transition: all 0.3s ease;
  box-sizing: border-box;
}

.icon-search:focus {
  outline: none;
  border-color: #007aff;
  background: rgba(var(--background-color-rgb), 0.8);
}

.search-hint {
  font-size: 12px;
  color: var(--text-color);
  opacity: 0.6;
  margin-top: 6px;
  padding-left: 4px;
}

.url-label {
  display: block;
  font-size: 13px;
  margin-bottom: 8px;
  opacity: 0.8;
}

/* ---- Body ---- */
.icon-picker-body {
  flex: 1;
  display: flex;
  overflow: hidden;
  min-height: 0;
}

.icon-categories {
  width: 160px;
  padding: 12px;
  border-right: 1px solid var(--border-color);
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex-shrink: 0;
  scrollbar-width: thin;
  scrollbar-color: var(--hover-link-color) rgba(var(--background-color-rgb), 0.3);
}

.icon-categories::-webkit-scrollbar { width: 6px; }
.icon-categories::-webkit-scrollbar-track { background: rgba(var(--background-color-rgb), 0.3); border-radius: 3px; }
.icon-categories::-webkit-scrollbar-thumb { background: var(--hover-link-color); border-radius: 3px; }

.category-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: rgba(var(--background-color-rgb), 0.4);
  color: var(--text-color);
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: left;
  font-size: 13px;
}

.category-btn:hover {
  background: rgba(var(--background-color-rgb), 0.7);
  border-color: var(--hover-link-color);
}

.category-btn.active {
  background: var(--hover-link-color);
  color: #333;
  border-color: var(--hover-link-color);
  font-weight: 500;
}

.icon-grid {
  flex: 1;
  padding: 12px;
  overflow-y: auto;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(90px, 1fr));
  gap: 10px;
  align-content: start;
  scrollbar-width: thin;
  scrollbar-color: var(--hover-link-color) rgba(var(--background-color-rgb), 0.3);
}

.icon-grid::-webkit-scrollbar { width: 8px; }
.icon-grid::-webkit-scrollbar-track { background: rgba(var(--background-color-rgb), 0.3); border-radius: 4px; }
.icon-grid::-webkit-scrollbar-thumb { background: var(--hover-link-color); border-radius: 4px; }

.icon-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 14px 6px;
  border: 2px solid var(--border-color);
  border-radius: 8px;
  background: rgba(var(--background-color-rgb), 0.4);
  cursor: pointer;
  transition: all 0.2s ease;
  min-height: 80px;
}

.icon-item:hover {
  background: rgba(var(--background-color-rgb), 0.7);
  border-color: var(--hover-link-color);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px var(--shadow-color);
}

.icon-item.active {
  background: var(--hover-link-color);
  border-color: var(--hover-link-color);
  color: #333;
}

.icon-item i,
.icon-item iconify-icon {
  font-size: 24px;
  margin-bottom: 6px;
  color: inherit;
}

.icon-item.active i,
.icon-item.active iconify-icon {
  color: #333;
}

.icon-name {
  font-size: 10px;
  text-align: center;
  word-break: break-all;
  color: inherit;
  opacity: 0.8;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.no-results {
  grid-column: 1 / -1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: var(--text-color);
  opacity: 0.5;
}

.no-results i { font-size: 36px; margin-bottom: 12px; }
.no-results p { margin: 0; font-size: 14px; }

/* ---- Load More ---- */
.load-more-bar {
  padding: 8px 16px;
  text-align: center;
  border-top: 1px solid var(--border-color);
  flex-shrink: 0;
}

.load-more-btn {
  padding: 8px 24px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: rgba(var(--background-color-rgb), 0.6);
  color: var(--text-color);
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s ease;
}

.load-more-btn:hover {
  background: rgba(var(--background-color-rgb), 0.8);
  border-color: var(--hover-link-color);
}

/* ---- URL Preview ---- */
.url-preview-body {
  flex-direction: column;
  padding: 20px;
}

.url-preview-area {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.url-preview-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.url-preview-img {
  max-width: 120px;
  max-height: 120px;
  object-fit: contain;
  border: 2px solid var(--border-color);
  border-radius: 12px;
  padding: 12px;
  background: rgba(var(--background-color-rgb), 0.4);
}

.url-preview-code {
  font-size: 12px;
  word-break: break-all;
  text-align: center;
  opacity: 0.7;
  max-width: 300px;
}

.url-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  opacity: 0.4;
}

.url-placeholder i { font-size: 48px; }
.url-placeholder p { margin: 0; font-size: 14px; }

.url-actions {
  text-align: center;
  padding-top: 16px;
}

.btn-confirm-url {
  padding: 10px 32px;
  border: none;
  border-radius: 8px;
  background: var(--hover-link-color);
  color: #333;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.btn-confirm-url:hover {
  background: #ffd700;
  transform: translateY(-1px);
}

.btn-confirm-url:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  transform: none;
}

/* ---- Footer ---- */
.icon-picker-footer {
  padding: 12px 16px;
  border-top: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  background: rgba(var(--background-color-rgb), 0.98);
  flex-shrink: 0;
}

.selected-icon {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  font-size: 13px;
  color: var(--text-color);
  min-width: 0;
}

.selected-icon iconify-icon,
.selected-icon i {
  font-size: 20px;
  color: var(--hover-link-color);
  flex-shrink: 0;
}

.selected-icon code {
  background: rgba(var(--background-color-rgb), 0.8);
  padding: 3px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-family: 'Courier New', monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 200px;
}

.text-muted { opacity: 0.5; }

.icon-picker-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.btn-clear,
.btn-close {
  padding: 8px 16px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: rgba(var(--background-color-rgb), 0.8);
  color: var(--text-color);
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}

.btn-clear:hover {
  background: rgba(244, 67, 54, 0.1);
  border-color: #f44336;
  color: #f44336;
}

.btn-close {
  background: var(--hover-link-color);
  color: #333;
  border-color: var(--hover-link-color);
  font-weight: 500;
}

.btn-close:hover {
  background: #ffd700;
}

@media (max-width: 768px) {
  .icon-picker-body {
    flex-direction: column;
  }
  .icon-categories {
    width: 100%;
    flex-direction: row;
    overflow-x: auto;
    border-right: none;
    border-bottom: 1px solid var(--border-color);
    padding: 8px 12px;
  }
  .category-btn {
    flex-shrink: 0;
    white-space: nowrap;
  }
  .icon-grid {
    grid-template-columns: repeat(auto-fill, minmax(72px, 1fr));
    gap: 8px;
  }
  .icon-item {
    min-height: 68px;
    padding: 10px 4px;
  }
  .tab-btn {
    font-size: 12px;
    padding: 10px 4px;
  }
}
</style>
