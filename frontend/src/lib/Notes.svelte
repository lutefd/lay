<script lang="ts">
  import { onMount } from 'svelte';
  import { GetNotes, SaveNotes } from '../../wailsjs/go/main/App.js';
  import Markdown from './Markdown.svelte';

  let content = $state('');
  let status = $state<'idle' | 'saving' | 'saved' | 'error'>('idle');
  let mode = $state<'edit' | 'preview'>('edit');
  let saveTimer: ReturnType<typeof setTimeout> | null = null;

  onMount(async () => {
    content = await GetNotes();
  });

  function onInput() {
    if (saveTimer) clearTimeout(saveTimer);
    status = 'idle';
    saveTimer = setTimeout(save, 800);
  }

  async function save() {
    status = 'saving';
    try {
      await SaveNotes(content);
      status = 'saved';
      setTimeout(() => (status = 'idle'), 1500);
    } catch {
      status = 'error';
    }
  }

  function copyRaw() {
    navigator.clipboard.writeText(content);
  }
</script>

<div class="notes-panel">
  <div class="notes-toolbar">
    <div class="mode-tabs">
      <button class="mode-btn" class:active={mode === 'edit'} onclick={() => (mode = 'edit')}>Edit</button>
      <button class="mode-btn" class:active={mode === 'preview'} onclick={() => (mode = 'preview')}>Preview</button>
    </div>

    <div class="toolbar-right">
      {#if mode === 'preview'}
        <button class="action-btn" onclick={copyRaw} title="Copy raw markdown">Copy md</button>
      {/if}
      <span class="save-status">
        {#if status === 'saving'}
          <span class="dot saving"></span> saving…
        {:else if status === 'saved'}
          <span class="dot saved"></span> saved
        {:else if status === 'error'}
          <span class="dot error"></span> error
        {/if}
      </span>
    </div>
  </div>

  {#if mode === 'edit'}
    <textarea
      class="notes-textarea"
      bind:value={content}
      oninput={onInput}
      placeholder="Meeting notes… (auto-saved, supports markdown)"
      spellcheck={false}
    ></textarea>
  {:else}
    <div class="preview-scroll">
      {#if content.trim()}
        <Markdown raw={content} class="notes-md" />
      {:else}
        <p class="empty-hint">Nothing here yet — switch to Edit to write notes.</p>
      {/if}
    </div>
  {/if}
</div>

<style>
  .notes-panel {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-height: 0;
  }

  .notes-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 4px 10px;
    min-height: 30px;
    flex-shrink: 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }

  .mode-tabs {
    display: flex;
    gap: 2px;
  }

  .mode-btn {
    background: none;
    border: none;
    color: rgba(255, 255, 255, 0.35);
    font-size: 11px;
    font-family: inherit;
    padding: 2px 8px;
    border-radius: 4px;
    cursor: pointer;
    transition: color 0.15s, background 0.15s;
  }

  .mode-btn:hover {
    color: rgba(255, 255, 255, 0.7);
    background: rgba(255, 255, 255, 0.05);
  }

  .mode-btn.active {
    color: rgba(255, 255, 255, 0.85);
    background: rgba(255, 255, 255, 0.09);
  }

  .toolbar-right {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .action-btn {
    background: none;
    border: none;
    color: rgba(255, 255, 255, 0.3);
    font-size: 11px;
    font-family: inherit;
    padding: 2px 7px;
    border-radius: 4px;
    cursor: pointer;
    transition: color 0.15s, background 0.15s;
  }

  .action-btn:hover {
    color: rgba(255, 255, 255, 0.7);
    background: rgba(255, 255, 255, 0.06);
  }

  .save-status {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 11px;
    color: rgba(255, 255, 255, 0.3);
    min-width: 60px;
  }

  .dot {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    display: inline-block;
    flex-shrink: 0;
  }
  .dot.saving { background: #f5a623; }
  .dot.saved  { background: #4caf82; }
  .dot.error  { background: #e05252; }

  /* Edit mode */
  .notes-textarea {
    flex: 1;
    width: 100%;
    box-sizing: border-box;
    background: transparent;
    border: none;
    outline: none;
    resize: none;
    color: rgba(255, 255, 255, 0.87);
    font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace;
    font-size: 13px;
    line-height: 1.6;
    padding: 10px 14px 14px;
    caret-color: #7c9ef5;
  }

  .notes-textarea::placeholder {
    color: rgba(255, 255, 255, 0.18);
  }

  /* Preview mode */
  .preview-scroll {
    flex: 1;
    overflow-y: auto;
    padding: 12px 16px 20px;
  }

  .preview-scroll::-webkit-scrollbar { width: 4px; }
  .preview-scroll::-webkit-scrollbar-track { background: transparent; }
  .preview-scroll::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.1); border-radius: 2px; }

  .empty-hint {
    color: rgba(255, 255, 255, 0.2);
    font-size: 13px;
    text-align: center;
    margin-top: 40px;
  }

  /* Let Markdown.svelte's scoped styles work; add any overrides here */
  .preview-scroll :global(.notes-md) {
    max-width: 100%;
  }
</style>
