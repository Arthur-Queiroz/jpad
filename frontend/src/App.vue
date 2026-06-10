<script setup>
import { ref, onMounted, watch } from 'vue'
import { useTheme } from './composables/useTheme.js'

const isHome = ref(window.location.pathname === '/')
const inputPath = ref('')
const content = ref('')
const saving = ref(false)
const saved = ref(false)
const error = ref('')

const { isDark, toggle } = useTheme()

let saveTimeout = null

const notePath = window.location.pathname

// Mantém o path alinhado ao que o backend aceita (^[a-zA-Z0-9_/-]+$):
// espaços viram hífen e o resto dos caracteres inválidos é descartado.
const sanitizePath = (raw) =>
  raw
    .replace(/\s+/g, '-')
    .replace(/[^a-zA-Z0-9_/-]/g, '')
    .replace(/^\/+/, '')

const onPathInput = () => {
  inputPath.value = sanitizePath(inputPath.value)
}

const navigateTo = () => {
  const val = sanitizePath(inputPath.value)
  if (!val) return
  window.location.href = '/' + val
}

const loadNote = async () => {
  error.value = ''
  try {
    const res = await fetch(`/api${notePath}`)
    if (!res.ok) {
      if (res.status === 404) {
        content.value = ''
        return
      }
      throw new Error(`HTTP ${res.status}`)
    }
    const data = await res.json()
    content.value = data.content || ''
  } catch (err) {
    error.value = 'Erro ao carregar nota'
    console.error(err)
  }
}

const saveNote = async () => {
  if (saving.value) return
  saving.value = true
  saved.value = false
  error.value = ''
  try {
    const res = await fetch(`/api${notePath}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ content: content.value }),
    })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    saved.value = true
    setTimeout(() => { saved.value = false }, 2000)
  } catch (err) {
    error.value = 'Erro ao salvar nota'
    console.error(err)
  } finally {
    saving.value = false
  }
}

const debouncedSave = () => {
  clearTimeout(saveTimeout)
  saveTimeout = setTimeout(saveNote, 500)
}

onMounted(() => {
  if (!isHome.value) loadNote()
})

watch(content, debouncedSave)
</script>

<template>
  <div v-if="isHome" class="home">
    <button class="theme-btn" @click="toggle" aria-label="Alternar tema">
      <svg v-if="isDark" class="theme-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="5"></circle>
        <line x1="12" y1="1" x2="12" y2="3"></line>
        <line x1="12" y1="21" x2="12" y2="23"></line>
        <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line>
        <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line>
        <line x1="1" y1="12" x2="3" y2="12"></line>
        <line x1="21" y1="12" x2="23" y2="12"></line>
        <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line>
        <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>
      </svg>
      <svg v-else class="theme-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>
      </svg>
    </button>
    <div class="home-brand-wrap">
      <h1 class="home-brand">JPAD</h1>
    </div>
    <div class="home-card">
      <div class="home-form">
        <span class="home-prefix">jpad.com/</span>
        <input
          v-model="inputPath"
          type="text"
          class="home-input"
          placeholder="nome-da-nota"
          autofocus
          @input="onPathInput"
          @keydown.enter="navigateTo"
        />
        <button class="home-btn" @click="navigateTo">go</button>
      </div>
    </div>
  </div>

  <div v-else class="editor">
    <textarea
      v-model="content"
      class="editor-area"
      spellcheck="false"
      autofocus
    />
    <div class="status-bar">
      <span class="path">{{ notePath }}</span>
      <span v-if="saving" class="saving">salvando...</span>
      <span v-else-if="saved" class="saved">salvo</span>
      <span v-if="error" class="error">{{ error }}</span>
      <button class="theme-btn" @click="toggle" aria-label="Alternar tema">
        <svg v-if="isDark" class="theme-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="5"></circle>
          <line x1="12" y1="1" x2="12" y2="3"></line>
          <line x1="12" y1="21" x2="12" y2="23"></line>
          <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line>
          <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line>
          <line x1="1" y1="12" x2="3" y2="12"></line>
          <line x1="21" y1="12" x2="23" y2="12"></line>
          <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line>
          <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>
        </svg>
        <svg v-else class="theme-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>
        </svg>
      </button>
    </div>
  </div>
</template>

<style>
*, *::before, *::after {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: var(--bg);
  color: var(--text);
}
</style>

<style scoped>
.home {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  padding: 1rem;
  position: relative;
  gap: 2rem;
}

.home-brand-wrap {
  text-align: center;
  transform: translateY(-6rem);
}

.home-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 2rem;
  box-shadow: 0 4px 24px var(--shadow);
}

.home-brand {
  font-size: 2.5rem;
  font-weight: 800;
  text-align: center;
  color: var(--accent);
  letter-spacing: -0.02em;
}

.home-tagline {
  text-align: center;
  color: var(--text-muted);
  font-size: 0.9rem;
  margin-top: 0.25rem;
}

.home-form {
  display: flex;
  align-items: stretch;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--border-input);
}

.home-prefix {
  display: flex;
  align-items: center;
  padding: 10px 14px;
  background: var(--bg-status);
  font-size: 0.95rem;
  white-space: nowrap;
  color: var(--text-muted);
  font-weight: 500;
}

.home-input {
  padding: 10px 14px;
  font-size: 0.95rem;
  border: none;
  outline: none;
  flex: 1;
  min-width: 0;
  background: var(--bg-input);
  color: var(--text);
  font-family: 'Inter', sans-serif;
}

.home-input::placeholder {
  color: var(--text-dim);
}

.home-btn {
  padding: 10px 22px;
  background: var(--accent);
  color: #fff;
  border: none;
  font-size: 0.95rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s ease;
}

.home-btn:hover {
  background: var(--accent-hover);
}

.editor {
  display: flex;
  flex-direction: column;
  height: 100vh;
}

.editor-area {
  flex: 1;
  width: 100%;
  padding: 24px;
  font-size: 0.95rem;
  line-height: 1.7;
  border: none;
  outline: none;
  resize: none;
  background: var(--bg);
  color: var(--text);
  font-family: 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
  caret-color: var(--caret);
}

.status-bar {
  display: flex;
  gap: 1rem;
  align-items: center;
  padding: 6px 24px;
  border-top: 1px solid var(--border);
  background: var(--bg-status);
  font-size: 0.75rem;
  color: var(--text-dim);
  font-family: 'Inter', sans-serif;
}

.path { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.saving { color: var(--saving); }
.saved { color: var(--saved); }
.error { color: var(--error); }

.theme-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  color: var(--text-dim);
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.15s ease;
  margin-left: auto;
}

.theme-btn:hover {
  color: var(--text);
}

.theme-icon {
  width: 16px;
  height: 16px;
}

.home .theme-btn {
  position: absolute;
  top: 1rem;
  right: 1rem;
}

@media (max-width: 480px) {
  .home {
    padding: 0.75rem;
  }

  .home-card {
    padding: 1.25rem;
    width: 100%;
  }

  .home-prefix {
    font-size: 0.8rem;
    padding: 8px 10px;
  }

  .home-input {
    font-size: 0.85rem;
    padding: 8px 10px;
  }

  .home-btn {
    font-size: 0.85rem;
    padding: 8px 14px;
  }

  .editor-area {
    padding: 16px;
    font-size: 0.9rem;
  }

  .status-bar {
    padding: 6px 12px;
    gap: 0.5rem;
  }

  .home .theme-btn {
    top: 0.5rem;
    right: 0.5rem;
  }
}
</style>
