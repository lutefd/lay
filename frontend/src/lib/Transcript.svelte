<script lang="ts">
  import { AppendTranscriptToNotes } from '../../wailsjs/go/main/App.js';

  interface Props {
    text: string;
    recordingDir: string;
    onNew: () => void;
  }

  let { text, recordingDir, onNew }: Props = $props();

  let appended = $state(false);
  let error = $state('');

  async function appendToNotes() {
    try {
      await AppendTranscriptToNotes(recordingDir);
      appended = true;
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : String(e);
    }
  }
</script>

<div class="transcript">
  <div class="header">
    <span class="label">Transcript</span>
    <div class="actions">
      <button
        class="btn"
        class:success={appended}
        onclick={appendToNotes}
        disabled={appended}
      >
        {appended ? 'Added to Notes' : 'Add to Notes'}
      </button>
      <button class="btn secondary" onclick={onNew}>New</button>
    </div>
  </div>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <div class="scroll">
    <pre class="text">{text}</pre>
  </div>
</div>

<style>
  .transcript {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-height: 0;
    padding: 12px 14px;
    gap: 8px;
  }

  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-shrink: 0;
  }

  .label {
    font-size: 11px;
    color: rgba(255, 255, 255, 0.4);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .actions {
    display: flex;
    gap: 6px;
  }

  .btn {
    background: rgba(255, 255, 255, 0.1);
    border: none;
    border-radius: 5px;
    color: #fff;
    font-size: 11px;
    font-family: inherit;
    padding: 4px 10px;
    cursor: pointer;
    transition: opacity 0.15s, background 0.15s;
  }

  .btn:hover:not(:disabled) { opacity: 0.8; }
  .btn:disabled { opacity: 0.5; cursor: default; }

  .btn.success {
    background: rgba(80, 200, 120, 0.2);
    color: #50c878;
  }

  .btn.secondary {
    background: rgba(255, 255, 255, 0.06);
    color: rgba(255, 255, 255, 0.45);
  }

  .error {
    font-size: 11px;
    color: #e05252;
    margin: 0;
  }

  .scroll {
    flex: 1;
    overflow-y: auto;
    min-height: 0;
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.04);
  }

  .text {
    margin: 0;
    padding: 10px;
    font-family: 'SF Mono', 'Menlo', monospace;
    font-size: 11px;
    line-height: 1.6;
    color: rgba(255, 255, 255, 0.75);
    white-space: pre-wrap;
    word-break: break-word;
  }
</style>
