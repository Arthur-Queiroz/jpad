import { ref, watch } from 'vue'

const STORAGE_KEY = 'jpad-theme'

const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
const stored = localStorage.getItem(STORAGE_KEY)

const isDark = ref(stored ? stored === 'dark' : prefersDark)

function apply() {
  document.documentElement.setAttribute('data-theme', isDark.value ? 'dark' : 'light')
}

function toggle() {
  isDark.value = !isDark.value
}

watch(isDark, (val) => {
  localStorage.setItem(STORAGE_KEY, val ? 'dark' : 'light')
  apply()
})

apply()

export function useTheme() {
  return { isDark, toggle }
}
