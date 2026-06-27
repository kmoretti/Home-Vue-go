import { getFrontendConfig } from '../api'

/**
 * 从API获取前端配置并更新页面
 */
export async function loadAndApplyFrontendConfig() {
  try {
    const res = await getFrontendConfig()
    const config = res.data

    // 更新页面标题
    // 防御：如果 title 看起来像 URL（错误的用户配置），忽略它
    if (config.title && isTitleValid(config.title)) {
      document.title = config.title
    } else if (config.siteName) {
      document.title = config.siteName
    }

		// 更新meta标签
		if (config.keywords) {
			updateMetaTag('keywords', config.keywords)
		}
		if (config.description) {
			updateMetaTag('description', config.description)
		}

    // 更新favicon
    if (config.favicon) {
      updateFavicon(config.favicon)
    }

    // 动态加载图标库
    if (config.iconLibrary) {
      loadStylesheet(config.iconLibrary, 'icon-library')
    }

    // 动态加载字体库
    if (config.fontLibrary) {
      loadStylesheet(config.fontLibrary, 'font-library')
    }

    // 动态加载Umami统计脚本
    if (config.umamiScript && config.umamiWebsiteId) {
      loadUmamiScript(config.umamiScript, config.umamiWebsiteId)
    }
  } catch (error) {
    console.error('加载前端配置失败:', error)
    // 如果API失败，使用默认值（从环境变量或index.html中的占位符）
  }
}

function updateMetaTag(name, content) {
  if (!content) return

  let meta = document.querySelector(`meta[name="${name}"]`)
  if (!meta) {
    meta = document.createElement('meta')
    meta.setAttribute('name', name)
    document.head.appendChild(meta)
  }
  meta.setAttribute('content', content)
}

function updateFavicon(href) {
  let link = document.querySelector("link[rel*='icon']")
  if (!link) {
    link = document.createElement('link')
    link.rel = 'icon'
    document.head.appendChild(link)
  }
  link.href = href
}

function loadStylesheet(href, id) {
  // 检查是否已加载
  if (document.getElementById(id)) {
    return
  }

  const link = document.createElement('link')
  link.id = id
  link.rel = 'stylesheet'
  link.href = href.startsWith('//') ? `https:${href}` : href
  document.head.appendChild(link)
}

function loadUmamiScript(src, websiteId) {
  // 检查是否已加载
  if (document.querySelector(`script[data-website-id="${websiteId}"]`)) {
    return
  }

  const script = document.createElement('script')
  script.defer = true
  script.src = src
  script.setAttribute('data-website-id', websiteId)
  document.head.appendChild(script)
}

/**
 * 验证标题是否有效（不是URL或其他无效内容）
 */
function isTitleValid(title) {
  if (!title) return false
  // 如果以 http:// 或 https:// 开头，说明是误填的URL
  if (/^https?:\/\//i.test(title)) return false
  // 标题长度不能超过100个字符（URL通常更长）
  if (title.length > 100) return false
  // 不包含换行符
  if (/\n/.test(title)) return false
  return true
}
