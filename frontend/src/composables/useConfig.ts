import { ref } from 'vue'
import { GetConfig, SetOutputDir, SetFilenamePattern } from '../../wailsjs/go/main/App'

export interface AppConfig {
    outputDir: string
    filenamePattern: string
}

const config = ref<AppConfig>({ outputDir: '', filenamePattern: '{title}' })

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

    return { config, load, updateOutputDir, updateFilenamePattern }
}
