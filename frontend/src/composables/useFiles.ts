import { ref, computed } from 'vue'

export type FileStatus = 'pending' | 'converting' | 'done' | 'error'

export interface FileItem {
    id: string
    path: string
    name: string
    size: number
    status: FileStatus
    error?: string
}

const files = ref<FileItem[]>([])

let _idCounter = 0
function genId() { return `f-${++_idCounter}` }

function formatBytes(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function useFiles() {
    const addPaths = (paths: string[]) => {
        const existing = new Set(files.value.map(f => f.path))
        for (const p of paths) {
            if (existing.has(p)) continue
            // Extract filename — strip trailing separators first to handle "C:\path\" style paths
            const name = p.replace(/[\\/]+$/, '').split(/[\\/]/).pop() || p

            files.value.push({
                id: genId(),
                path: p,
                name,
                size: 0,
                status: 'pending',
            })
        }
    }

    const removeFile = (id: string) => {
        files.value = files.value.filter(f => f.id !== id)
    }

    const clearAll = () => {
        files.value = []
    }

    const clearDone = () => {
        files.value = files.value.filter(f => f.status !== 'done' && f.status !== 'error')
    }

    const pendingCount = computed(() => files.value.filter(f => f.status === 'pending').length)
    const hasFiles = computed(() => files.value.length > 0)
    const isConverting = computed(() => files.value.some(f => f.status === 'converting'))

    return {
        files,
        addPaths,
        removeFile,
        clearAll,
        clearDone,
        pendingCount,
        hasFiles,
        isConverting,
        formatBytes,
    }
}
