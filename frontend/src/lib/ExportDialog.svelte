<script lang="ts">
  import { ExportToFile, GetHomePath } from '../../wailsjs/go/main/App.js';
  import { onMount } from 'svelte';

  interface Props {
    content: string;
    defaultName: string;
    onClose: () => void;
  }

  let { content, defaultName, onClose }: Props = $props();

  let path = $state('');
  let saving = $state(false);
  let error = $state('');
  let inputEl = $state<HTMLInputElement | null>(null);

  onMount(async () => {
    const home = await GetHomePath();
    path = `${home}/Desktop/${defaultName}`;
    // Wait a tick for the input to render, then focus + select filename
    requestAnimationFrame(() => {
      if (inputEl) {
        inputEl.focus();
        const lastSlash = path.lastIndexOf('/');
        const dotPos = path.lastIndexOf('.');
        inputEl.setSelectionRange(lastSlash + 1, dotPos > lastSlash ? dotPos : path.length);
      }
    });
  });

  async function save() {
    if (!path.trim()) return;
    saving = true;
    error = '';
    try {
      await ExportToFile(content, path.trim());
      onClose();
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : String(e);
      saving = false;
    }
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') save();
    if (e.key === 'Escape') onClose();
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="overlay" onkeydown={onKeydown}>
  <div class="dialog">
    <label class="label">Save as</label>
    <input
      type="text"
      class="path-input"
      bind:value={path}
      bind:this={inputEl}
      spellcheck={false}
      autocomplete="off"
      disabled={saving}
    />
    {#if error}
      <p class="error">{error}</p>
    {/if}
    <div class="buttons">
      <button class="btn cancel" onclick={onClose} disabled={saving}>Cancel</button>
      <button class="btn save" onclick={save} disabled={saving || !path.trim()}>
        {saving ? 'Saving...' : 'Save'}
      </button>
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .dialog {
    background: rgba(30, 30, 38, 0.98);
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 10px;
    padding: 16px;
    width: 380px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .label {
    font-size: 11px;
    color: rgba(255, 255, 255, 0.45);
    font-weight: 500;
  }

  .path-input {
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 7px;
    color: rgba(255, 255, 255, 0.87);
    font-family: 'JetBrains Mono', monospace;
    font-size: 12px;
    padding: 7px 10px;
    outline: none;
    width: 100%;
    box-sizing: border-box;
  }

  .path-input:focus {
    border-color: rgba(124, 158, 245, 0.5);
  }

  .error {
    font-size: 11px;
    color: #e05252;
    margin: 0;
  }

  .buttons {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }

  .btn {
    border: none;
    border-radius: 6px;
    font-size: 12px;
    font-family: inherit;
    padding: 5px 14px;
    cursor: pointer;
  }

  .btn:disabled {
    opacity: 0.4;
    cursor: default;
  }

  .cancel {
    background: rgba(255, 255, 255, 0.06);
    color: rgba(255, 255, 255, 0.5);
  }

  .save {
    background: rgba(124, 158, 245, 0.2);
    color: #8cabff;
  }

  .save:hover:not(:disabled) {
    background: rgba(124, 158, 245, 0.3);
  }
</style>
