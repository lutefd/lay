<script lang="ts">
  import { onMount } from 'svelte';
  import { GetNotes, SaveNotes } from '../../wailsjs/go/main/App.js';

  let content = $state('');
  let status = $state<'idle' | 'saving' | 'saved' | 'error'>('idle');
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
</script>

<div class="notes-panel">
  <div class="notes-status">
    {#if status === 'saving'}
      <span class="dot saving"></span> saving…
    {:else if status === 'saved'}
      <span class="dot saved"></span> saved
    {:else if status === 'error'}
      <span class="dot error"></span> error saving
    {/if}
  </div>
  <textarea
    class="notes-textarea"
    bind:value={content}
    oninput={onInput}
    placeholder="Meeting notes… (auto-saved)"
    spellcheck={false}
  ></textarea>
</div>

<style>
  .notes-panel {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-height: 0;
  }

  .notes-status {
    display: flex;
    align-items: center;
    gap: 5px;
    padding: 4px 12px;
    font-size: 11px;
    color: rgba(255, 255, 255, 0.35);
    min-height: 22px;
    flex-shrink: 0;
  }

  .dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    display: inline-block;
  }

  .dot.saving { background: #f5a623; }
  .dot.saved  { background: #4caf82; }
  .dot.error  { background: #e05252; }

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
    padding: 8px 14px 14px;
    caret-color: #7c9ef5;
  }

  .notes-textarea::placeholder {
    color: rgba(255, 255, 255, 0.2);
  }
</style>
