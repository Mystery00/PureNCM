import { onMounted, onUnmounted } from 'vue'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { ConvertFiles } from '../../wailsjs/go/main/App'
import { useFiles } from '@/composables/useFiles'
import { useConfig } from '@/composables/useConfig'

const EVENT_PROGRESS = 'ncm:progress'

interface ProgressPayload {
    path: string
    status: 'converting' | 'done' | 'error'
    size?: number
    outputPath?: string
    error?: string
}

export function useConvert() {
    const { files, isConverting } = useFiles()
    const { config } = useConfig()

    const startConvert = async () => {
        const pending = files.value.filter(f => f.status === 'pending')
        if (pending.length === 0 || isConverting.value) return

        const paths = pending.map(f => f.path)
        const outputDir = config.value.outputDir
        const pattern = config.value.filenamePattern || '{title}'

        // Fire-and-forget: Go runs synchronously on its goroutine and emits events
        ConvertFiles(paths, outputDir, pattern)
    }

    const handleProgress = (payload: ProgressPayload) => {
        const item = files.value.find(f => f.path === payload.path)
        if (!item) return

        item.status = payload.status
        if (payload.size && payload.size > 0) {
            item.size = payload.size
        }
        if (payload.status === 'error') {
            item.error = payload.error ?? '未知错误'
        } else if (payload.status === 'done') {
            item.error = undefined
        }
    }

    onMounted(() => {
        EventsOn(EVENT_PROGRESS, handleProgress)
    })

    onUnmounted(() => {
        EventsOff(EVENT_PROGRESS)
    })

    return { startConvert }
}
