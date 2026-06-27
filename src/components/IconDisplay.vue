<template>
  <img
    v-if="isImageUrl && !imgError"
    :src="icon"
    class="icon-img"
    :class="imgClass"
    :style="imgStyle"
    @error="onError"
  />
  <iconify-icon
    v-else-if="isIconify"
    :icon="icon"
    :class="iconClass"
    :style="iconifyStyle"
    width="1em"
    height="1em"
  ></iconify-icon>
  <i
    v-else
    :class="[icon, iconClass]"
    :style="iconStyle"
  ></i>
</template>

<script setup>
import { computed, ref, watch } from 'vue'

const props = defineProps({
  icon: { type: String, default: '' },
  hoverColor: { type: String, default: '' },
  size: { type: [String, Number], default: '' },
})

const imgError = ref(false)

// icon 变化时重置错误状态，支持重试
watch(() => props.icon, () => {
  imgError.value = false
})

const isImageUrl = computed(() => {
  return props.icon && (props.icon.startsWith('http://') || props.icon.startsWith('https://'))
})

const isIconify = computed(() => {
  if (!props.icon || isImageUrl.value) return false
  // Iconify 格式： "mdi:home", "logos:github-icon", "fa:home"
  // Font Awesome 类名格式： "fa fa-blog", "fab fa-qq"
  // 判断依据：包含冒号但不是 Font Awesome 类
  return props.icon.includes(':')
})

const iconClass = computed(() => {
  return props.hoverColor ? '' : ''
})

const iconStyle = computed(() => {
  const style = {}
  if (props.hoverColor) style.color = props.hoverColor
  if (props.size) {
    const px = typeof props.size === 'number' ? `${props.size}px` : props.size
    style.fontSize = px
  }
  return style
})

const iconifyStyle = computed(() => {
  const style = {}
  if (props.hoverColor) style.color = props.hoverColor
  return style
})

const imgClass = computed(() => {
  return props.hoverColor ? '' : ''
})

const imgStyle = computed(() => {
  const style = {}
  if (props.size) {
    const px = typeof props.size === 'number' ? `${props.size}px` : props.size
    style.width = px
    style.height = px
  }
  return style
})

const emit = defineEmits(['error'])
const onError = () => {
  imgError.value = true
  emit('error')
}
</script>

<style scoped>
.icon-img {
  width: 1em;
  height: 1em;
  vertical-align: middle;
  object-fit: contain;
  display: inline-block;
}
</style>
