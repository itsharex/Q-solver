import modelData from '../config/model-capabilities.json'

// 提供商Logo - 简洁清晰的图标设计
// 提供商Logo
export const PROVIDER_LOGOS = {
    // Google Gemini
    google: '/providers/google.svg',

    // OpenAI
    openai: '/providers/openai.svg',

    // Anthropic Claude
    anthropic: '/providers/anthropic.svg',

    // DeepSeek
    deepseek: '/providers/deepseek.svg',

    // Alibaba Qwen
    alibaba: '/providers/alibaba.svg',

    // Zhipu GLM
    zhipu: '/providers/zhipu.svg',

    // Moonshot Kimi
    moonshot: '/providers/moonshot.svg',

    // xAI Grok
    xai: '/providers/xai.svg',

    // Meta Llama
    meta: '/providers/meta.svg',


    // 自定义 - 齿轮 (保留 SVG 以防默认样式丢失)
    custom: `<svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
        <rect width="24" height="24" rx="6" fill="#6B7280"/>
        <circle cx="12" cy="12" r="3" stroke="white" stroke-width="2" fill="none"/>
        <path d="M12 5V7M12 17V19M5 12H7M17 12H19M7.05 7.05L8.46 8.46M15.54 15.54L16.95 16.95M7.05 16.95L8.46 15.54M15.54 8.46L16.95 7.05" stroke="white" stroke-width="2" stroke-linecap="round"/>
    </svg>`,

    // 默认 - 方块 (保留 SVG)
    default: `<svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
        <rect width="24" height="24" rx="6" fill="#4B5563"/>
        <rect x="7" y="7" width="10" height="10" rx="2" stroke="white" stroke-width="2" fill="none"/>
        <circle cx="12" cy="12" r="2" fill="white"/>
    </svg>`
}

// Maps provider codes to display names
const PROVIDER_NAMES = {
    google: 'Google Gemini',
    openai: 'OpenAI',
    anthropic: 'Anthropic',
    alibaba: 'Alibaba Cloud (Qwen)',
    xai: 'xAI (Grok)',
    meta: 'Meta (Llama)',
}

export const PROVIDER_BASE_URLS = {
    google: 'https://generativelanguage.googleapis.com',
    openai: 'https://api.openai.com/v1',
    anthropic: 'https://api.anthropic.com',
    alibaba: 'https://dashscope.aliyuncs.com/compatible-mode/v1',
    custom: ''
}

/**
 * Returns the capabilities object for a given model
 * @param {string} model - The model identifier (e.g. "gemini-1.5-pro")
 * @returns {object} Capability object { image: boolean, pdf: boolean, ... }
 */
export function getModelCapabilities(model) {
    if (!model) return { text: true, image: false, pdf: false, audio: false }

    // 先尝试精确匹配
    if (modelData.models[model]) {
        return modelData.models[model]
    }

    // 关键字匹配
    const lowerModel = model.toLowerCase()
    for (const [key, value] of Object.entries(modelData.models)) {
        if (lowerModel.includes(key.toLowerCase()) || key.toLowerCase().includes(lowerModel)) {
            return value
        }
    }

    return { text: true, image: false, pdf: false, audio: false }
}

/**
 * Returns the friendly provider name for a given model
 * @param {string} model - The model identifier
 * @returns {string} Provider display name
 */
export function getProviderName(model) {
    if (!model) return 'Unknown Provider'

    // 获取 provider code
    const providerCode = getProviderCodeFromModel(model)
    return PROVIDER_NAMES[providerCode] || providerCode
}

/**
 * 根据模型名称推断 provider code
 */
function getProviderCodeFromModel(model) {
    if (!model) return 'unknown'

    const lowerModel = model.toLowerCase()

    // 关键字匹配规则
    if (lowerModel.includes('gemini')) return 'google'
    if (lowerModel.includes('gpt') || lowerModel.includes('o1') || lowerModel.includes('o3') || lowerModel.includes('o4')) return 'openai'
    if (lowerModel.includes('claude')) return 'anthropic'
    if (lowerModel.includes('deepseek')) return 'deepseek'
    if (lowerModel.includes('qwen')) return 'alibaba'
    if (lowerModel.includes('glm') || lowerModel.includes('chatglm')) return 'zhipu'
    if (lowerModel.includes('moonshot') || lowerModel.includes('kimi')) return 'moonshot'
    if (lowerModel.includes('mistral') || lowerModel.includes('mixtral')) return 'mistral'
    if (lowerModel.includes('grok')) return 'xai'
    if (lowerModel.includes('llama')) return 'meta'
    if (lowerModel.includes('yi-')) return '01ai'

    // 尝试从配置中查找
    if (modelData.models[model]) {
        return modelData.models[model].provider
    }

    // 模糊匹配配置
    for (const [key, value] of Object.entries(modelData.models)) {
        if (lowerModel.includes(key.toLowerCase()) || key.toLowerCase().includes(lowerModel)) {
            return value.provider
        }
    }

    return 'unknown'
}

/**
 * Returns the SVG logo string for a given model's provider
 * @param {string} model - The model identifier
 * @returns {string} SVG string
 */
export function getProviderLogo(model) {
    const providerCode = getProviderCodeFromModel(model)

    // Check if we have a specific logo for this provider
    if (PROVIDER_LOGOS[providerCode]) {
        return PROVIDER_LOGOS[providerCode]
    }

    // Fallback to default
    return PROVIDER_LOGOS.default
}

/**
 * Get just the provider code for a model
 */
export function getProviderCode(model) {
    return getProviderCodeFromModel(model)
}

/**
 * Check if the model supports vision capabilities
 * @param {string} model - The model identifier
 * @returns {boolean} True if the model supports images
 */
export function supportsVision(model) {
    const capabilities = getModelCapabilities(model)
    return !!capabilities.image
}

/**
 * Check if the model supports PDF capabilities
 * @param {string} model - The model identifier
 * @returns {boolean} True if the model supports PDF
 */
export function supportsPDF(model) {
    const capabilities = getModelCapabilities(model)
    return !!capabilities.pdf
}
