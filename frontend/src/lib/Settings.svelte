<script lang="ts">
  import { onMount } from 'svelte';
  import { GetConfig, GetGatewayConfig, SaveConfig } from '../../wailsjs/go/main/App.js';
  import type { app } from '../../wailsjs/go/models';

  const defaultModel = 'claude-sonnet-4-6';
  const baseModelGroups: { label: string; options: { value: string; label: string }[] }[] = [
    {
      label: 'Anthropic',
      options: [
        { value: 'claude-haiku-4-5-20251001', label: 'Haiku 4.5 — fast' },
        { value: 'claude-sonnet-4-6', label: 'Sonnet 4.6 — recommended' },
        { value: 'claude-opus-4-6', label: 'Opus 4.6 — most capable' },
      ],
    },
    {
      label: 'OpenAI',
      options: [
        { value: 'gpt-5-nano', label: 'GPT-5 nano — fastest' },
        { value: 'gpt-5-mini', label: 'GPT-5 mini — fast' },
        { value: 'gpt-5.1', label: 'GPT-5.1' },
        { value: 'gpt-5.2', label: 'GPT-5.2' },
        { value: 'gpt-5.2-chat-latest', label: 'GPT-5.2 chat latest' },
      ],
    },
  ];

  const transcribeLangs = [
    { value: '',   label: 'Auto-detect' },
    { value: 'pt', label: 'Portuguese' },
    { value: 'es', label: 'Spanish' },
    { value: 'en', label: 'English' },
    { value: 'fr', label: 'French' },
    { value: 'de', label: 'German' },
    { value: 'it', label: 'Italian' },
    { value: 'zh', label: 'Chinese' },
    { value: 'ja', label: 'Japanese' },
  ] as const;

  let anthropicKey = $state('');
  let openaiKey = $state('');
  let model = $state(defaultModel);
  let gatewayURL = $state('');
  let transcribeLang = $state('');
  let showAnthropic = $state(false);
  let showOpenAI = $state(false);
  let gwConfig = $state<app.GatewayConfig | null>(null);

  let modelGroups = $derived([
    ...baseModelGroups,
    ...(gwConfig ? [{ label: gwConfig.name, options: gwConfig.models }] : []),
  ]);
  let supportedModels = $derived(new Set(modelGroups.flatMap((group) => group.options.map((option) => option.value))));

  function normalizeModel(value: string | undefined): string {
    if (!value || !supportedModels.has(value)) {
      return defaultModel;
    }
    return value;
  }

  onMount(async () => {
    gwConfig = await GetGatewayConfig();
    const cfg = await GetConfig();
    anthropicKey = cfg.anthropicKey ?? '';
    openaiKey = cfg.openaiKey ?? '';
    model = normalizeModel(cfg.model);
    gatewayURL = cfg.gatewayURL ?? '';
    transcribeLang = cfg.transcribeLang ?? '';
  });

  async function save() {
    try {
      await SaveConfig(anthropicKey.trim(), openaiKey.trim(), normalizeModel(model), gatewayURL, transcribeLang);
    } catch {
      // ignore
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
        <input type="text"     class="field-input" bind:value={anthropicKey} placeholder="sk-ant-…" autocomplete="off" spellcheck={false} onblur={save} />
      {:else}
        <input type="password" class="field-input" bind:value={anthropicKey} placeholder="sk-ant-…" autocomplete="off" onblur={save} />
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
        <input type="text"     class="field-input" bind:value={openaiKey} placeholder="sk-proj-…" autocomplete="off" spellcheck={false} onblur={save} />
      {:else}
        <input type="password" class="field-input" bind:value={openaiKey} placeholder="sk-proj-…" autocomplete="off" onblur={save} />
      {/if}
      <button class="toggle-btn" onclick={() => (showOpenAI = !showOpenAI)}>
        {showOpenAI ? 'hide' : 'show'}
      </button>
    </div>
  </label>

  <!-- Gateway -->
  {#if gwConfig}
  <div class="field">
    <span class="field-label">{gwConfig.name}</span>
    <div class="gateway-row">
      <button
        type="button"
        class="gateway-toggle"
        class:active={gatewayURL !== ''}
        onclick={() => { gatewayURL = gatewayURL !== '' ? '' : gwConfig!.url; save(); }}
      >
        {gatewayURL !== '' ? 'Enabled' : 'Disabled'}
      </button>
    </div>
    <p class="gateway-hint">Route all requests through {gwConfig.name}. No API key required.</p>
  </div>
  {/if}

  <!-- Model -->
  <label class="field">
    <span class="field-label">Model</span>
    <div class="model-picker">
      {#each modelGroups as group}
        <div class="model-group">
          <p class="model-group-label">{group.label}</p>
          <div class="model-options">
            {#each group.options as option}
              <button
                type="button"
                class="model-option"
                class:selected={model === option.value}
                onclick={() => { model = option.value; save(); }}
              >
                {option.label}
              </button>
            {/each}
          </div>
        </div>
      {/each}
    </div>
  </label>

  <!-- Transcription language -->
  <label class="field">
    <span class="field-label">Transcription Language</span>
    <div class="model-options">
      {#each transcribeLangs as lang}
        <button
          type="button"
          class="model-option"
          class:selected={transcribeLang === lang.value}
          onclick={() => { transcribeLang = lang.value; save(); }}
        >
          {lang.label}
        </button>
      {/each}
    </div>
    <p class="gateway-hint">Force Whisper to a specific language to avoid misdetection between similar languages (e.g. Portuguese vs Spanish).</p>
  </label>

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

  .model-picker {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .model-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .model-group-label {
    margin: 0;
    font-size: 10px;
    color: rgba(255, 255, 255, 0.35);
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }

  .model-options {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .model-option {
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 7px;
    color: rgba(255, 255, 255, 0.75);
    font-family: inherit;
    font-size: 12px;
    padding: 6px 10px;
    cursor: pointer;
    transition: border-color 0.15s, color 0.15s, background 0.15s;
  }

  .model-option:hover {
    border-color: rgba(124, 158, 245, 0.45);
    color: rgba(255, 255, 255, 0.9);
  }

  .model-option.selected {
    background: rgba(124, 158, 245, 0.2);
    border-color: rgba(124, 158, 245, 0.45);
    color: #8cabff;
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

  .gateway-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .gateway-toggle {
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 7px;
    color: rgba(255, 255, 255, 0.45);
    font-size: 12px;
    font-family: inherit;
    padding: 5px 12px;
    cursor: pointer;
    transition: color 0.15s, background 0.15s, border-color 0.15s;
  }

  .gateway-toggle.active {
    background: rgba(52, 199, 89, 0.15);
    border-color: rgba(52, 199, 89, 0.4);
    color: #34c759;
  }

  .gateway-toggle:hover {
    color: rgba(255, 255, 255, 0.8);
    background: rgba(255, 255, 255, 0.1);
  }

  .gateway-toggle.active:hover {
    background: rgba(52, 199, 89, 0.25);
    color: #4cd964;
  }

  .gateway-hint {
    margin: 0;
    font-size: 10px;
    color: rgba(255, 255, 255, 0.28);
    line-height: 1.5;
  }
</style>
