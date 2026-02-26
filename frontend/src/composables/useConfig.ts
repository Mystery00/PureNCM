import { ref } from 'vue'
import { GetConfig, SetOutputDir, SetFilenamePattern, SetCopyLrc } from '../../wailsjs/go/main/App'

export interface AppConfig {
    outputDir: string
    filenamePattern: string
    copyLrc: boolean
}

const config = ref<AppConfig>({ outputDir: '', filenamePattern: '{title}', copyLrc: false })

export function useConfig() {
    const load = async () => {
        try {
            const c = await GetConfig()
            config.value = c as AppConfig
        } catch (e) {
            console.error('Failed to load config:', e)
        }
    }

    const updateOutputDir = async (dir: string) => {
        await SetOutputDir(dir)
        config.value.outputDir = dir
    }

    const updateFilenamePattern = async (pattern: string) => {
        await SetFilenamePattern(pattern)
        config.value.filenamePattern = pattern
    }

    const updateCopyLrc = async (enabled: boolean) => {
        await SetCopyLrc(enabled)
        config.value.copyLrc = enabled
    }

    return { config, load, updateOutputDir, updateFilenamePattern, updateCopyLrc }
}
