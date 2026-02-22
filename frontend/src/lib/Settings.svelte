<script lang="ts">
  import { onMount } from 'svelte';
  import { GetConfig, SaveConfig } from '../../wailsjs/go/main/App.js';

  let anthropicKey = $state('');
  let openaiKey = $state('');
  let model = $state('claude-sonnet-4-6');
  let status = $state<'idle' | 'saving' | 'saved' | 'error'>('idle');
  let showAnthropic = $state(false);
  let showOpenAI = $state(false);

  onMount(async () => {
    const cfg = await GetConfig();
    anthropicKey = cfg.anthropicKey ?? '';
    openaiKey = cfg.openaiKey ?? '';
    model = cfg.model ?? 'claude-sonnet-4-6';
  });

  async function save() {
    status = 'saving';
    try {
      await SaveConfig(anthropicKey.trim(), openaiKey.trim(), model);
      status = 'saved';
      setTimeout(() => (status = 'idle'), 2000);
    } catch {
      status = 'error';
    }
  }
</script>

<div class="settings-panel">
  <h2>Settings</h2>

  <!-- Anthropic key -->
  <label class="field">
    <span class="field-label">Anthropic API Key</span>
    <div class="key-row">
      {#if showAnthropic}
        <input type="text"     class="field-input" bind:value={anthropicKey} placeholder="sk-ant-…" autocomplete="off" spellcheck={false} />
      {:else}
        <input type="password" class="field-input" bind:value={anthropicKey} placeholder="sk-ant-…" autocomplete="off" />
      {/if}
      <button class="toggle-btn" onclick={() => (showAnthropic = !showAnthropic)}>
        {showAnthropic ? 'hide' : 'show'}
      </button>
    </div>
  </label>

  <!-- OpenAI key -->
  <label class="field">
    <span class="field-label">OpenAI API Key</span>
    <div class="key-row">
      {#if showOpenAI}
        <input type="text"     class="field-input" bind:value={openaiKey} placeholder="sk-proj-…" autocomplete="off" spellcheck={false} />
      {:else}
        <input type="password" class="field-input" bind:value={openaiKey} placeholder="sk-proj-…" autocomplete="off" />
      {/if}
      <button class="toggle-btn" onclick={() => (showOpenAI = !showOpenAI)}>
        {showOpenAI ? 'hide' : 'show'}
      </button>
    </div>
  </label>

  <!-- Model -->
  <label class="field">
    <span class="field-label">Model</span>
    <select class="field-select" bind:value={model}>
      <optgroup label="Anthropic">
        <option value="claude-haiku-4-5-20251001">Haiku 4.5 — fast</option>
        <option value="claude-sonnet-4-6">Sonnet 4.6 — recommended</option>
        <option value="claude-opus-4-6">Opus 4.6 — most capable</option>
      </optgroup>
      <optgroup label="OpenAI">
        <option value="gpt-4o-mini">GPT-4o mini — fast</option>
        <option value="gpt-4o">GPT-4o</option>
        <option value="gpt-4-turbo">GPT-4 Turbo</option>
        <option value="o1-mini">o1 mini</option>
        <option value="o1">o1</option>
        <option value="o3-mini">o3 mini</option>
      </optgroup>
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
    Anthropic: <strong>console.anthropic.com</strong><br/>
    OpenAI: <strong>platform.openai.com/api-keys</strong><br/>
    Keys stored in <code>~/.lay/config.json</code>.
  </p>
</div>

<style>
  .settings-panel {
    flex: 1;
    padding: 18px 20px;
    display: flex;
    flex-direction: column;
    gap: 14px;
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
    gap: 5px;
  }

  .field-label {
    font-size: 11px;
    color: rgba(255, 255, 255, 0.45);
    font-weight: 500;
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
    padding: 6px 10px;
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
    color: rgba(255, 255, 255, 0.45);
    font-size: 11px;
    font-family: inherit;
    padding: 0 10px;
    cursor: pointer;
    transition: color 0.15s, background 0.15s;
    flex-shrink: 0;
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
    padding: 6px 10px;
    outline: none;
    cursor: pointer;
    appearance: auto;
  }

  .field-select :global(optgroup) {
    color: rgba(255, 255, 255, 0.5);
    font-size: 11px;
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
    color: rgba(255, 255, 255, 0.22);
    line-height: 1.7;
    margin: 0;
  }

  .hint code {
    background: rgba(255, 255, 255, 0.07);
    border-radius: 3px;
    padding: 1px 4px;
    font-family: monospace;
  }
</style>
