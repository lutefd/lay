<script lang="ts">
  import { onMount } from 'svelte';
  import { GetConfig, SaveConfig } from '../../wailsjs/go/main/App.js';

  const MODELS = [
    { id: 'claude-haiku-4-5-20251001', label: 'Haiku 4.5 (fast)' },
    { id: 'claude-sonnet-4-6',         label: 'Sonnet 4.6 (recommended)' },
    { id: 'claude-opus-4-6',           label: 'Opus 4.6 (most capable)' },
  ];

  let apiKey = $state('');
  let model = $state('claude-sonnet-4-6');
  let status = $state<'idle' | 'saving' | 'saved' | 'error'>('idle');
  let showKey = $state(false);

  onMount(async () => {
    const cfg = await GetConfig();
    apiKey = cfg.apiKey ?? '';
    model = cfg.model ?? 'claude-sonnet-4-6';
  });

  async function save() {
    status = 'saving';
    try {
      await SaveConfig(apiKey.trim(), model);
      status = 'saved';
      setTimeout(() => (status = 'idle'), 2000);
    } catch {
      status = 'error';
    }
  }
</script>

<div class="settings-panel">
  <h2>Settings</h2>

  <label class="field">
    <span class="field-label">Anthropic API Key</span>
    <div class="key-row">
      {#if showKey}
        <input
          type="text"
          class="field-input"
          bind:value={apiKey}
          placeholder="sk-ant-…"
          autocomplete="off"
          spellcheck={false}
        />
      {:else}
        <input
          type="password"
          class="field-input"
          bind:value={apiKey}
          placeholder="sk-ant-…"
          autocomplete="off"
        />
      {/if}
      <button class="toggle-btn" onclick={() => (showKey = !showKey)}>
        {showKey ? 'hide' : 'show'}
      </button>
    </div>
  </label>

  <label class="field">
    <span class="field-label">Model</span>
    <select class="field-select" bind:value={model}>
      {#each MODELS as m}
        <option value={m.id}>{m.label}</option>
      {/each}
    </select>
  </label>

  <div class="actions">
    <button class="save-btn" onclick={save} disabled={status === 'saving'}>
      {#if status === 'saving'}Saving…
      {:else if status === 'saved'}Saved ✓
      {:else if status === 'error'}Error — retry
      {:else}Save
      {/if}
    </button>
  </div>

  <p class="hint">
    Get your API key at <strong>console.anthropic.com</strong>.<br/>
    Keys are stored only in <code>~/.lay/config.json</code>.
  </p>
</div>

<style>
  .settings-panel {
    flex: 1;
    padding: 18px 20px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    overflow-y: auto;
  }

  h2 {
    margin: 0;
    font-size: 14px;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.6);
    letter-spacing: 0.05em;
    text-transform: uppercase;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .field-label {
    font-size: 12px;
    color: rgba(255, 255, 255, 0.5);
  }

  .key-row {
    display: flex;
    gap: 6px;
  }

  .field-input {
    flex: 1;
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 7px;
    color: rgba(255, 255, 255, 0.87);
    font-family: 'JetBrains Mono', monospace;
    font-size: 12px;
    padding: 7px 10px;
    outline: none;
    transition: border-color 0.15s;
  }

  .field-input:focus {
    border-color: rgba(124, 158, 245, 0.5);
  }

  .toggle-btn {
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 7px;
    color: rgba(255, 255, 255, 0.5);
    font-size: 11px;
    font-family: inherit;
    padding: 0 10px;
    cursor: pointer;
    transition: color 0.15s, background 0.15s;
  }

  .toggle-btn:hover {
    color: rgba(255, 255, 255, 0.8);
    background: rgba(255, 255, 255, 0.1);
  }

  .field-select {
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 7px;
    color: rgba(255, 255, 255, 0.87);
    font-family: inherit;
    font-size: 13px;
    padding: 7px 10px;
    outline: none;
    cursor: pointer;
    appearance: auto;
  }

  .actions {
    display: flex;
    justify-content: flex-start;
  }

  .save-btn {
    background: rgba(124, 158, 245, 0.2);
    border: 1px solid rgba(124, 158, 245, 0.35);
    border-radius: 7px;
    color: #7c9ef5;
    font-family: inherit;
    font-size: 13px;
    padding: 7px 20px;
    cursor: pointer;
    transition: background 0.15s, opacity 0.15s;
  }

  .save-btn:hover:not(:disabled) {
    background: rgba(124, 158, 245, 0.35);
  }

  .save-btn:disabled {
    opacity: 0.6;
    cursor: default;
  }

  .hint {
    font-size: 11px;
    color: rgba(255, 255, 255, 0.25);
    line-height: 1.6;
    margin: 4px 0 0;
  }

  .hint code {
    background: rgba(255, 255, 255, 0.07);
    border-radius: 3px;
    padding: 1px 4px;
    font-family: monospace;
  }
</style>
