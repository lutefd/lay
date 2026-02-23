<script lang="ts">
  import { StartRecording, StopRecording, Transcribe } from '../../wailsjs/go/main/App.js';
  import Transcript from './Transcript.svelte';

  type State = 'idle' | 'recording' | 'stopping' | 'transcribing' | 'done';

  let state = $state<State>('idle');
  let recordingDir = $state('');
  let transcript = $state('');
  let elapsed = $state(0);
  let error = $state('');
  let intervalId: ReturnType<typeof setInterval> | null = null;

  async function start() {
    error = '';
    transcript = '';
    state = 'recording';
    elapsed = 0;
    intervalId = setInterval(() => elapsed++, 1000);
    try {
      recordingDir = await StartRecording();
    } catch (e: unknown) {
      clearInterval(intervalId!);
      intervalId = null;
      error = e instanceof Error ? e.message : String(e);
      state = 'idle';
    }
  }

  async function stop() {
    if (intervalId) { clearInterval(intervalId); intervalId = null; }
    state = 'stopping';
    try {
      await StopRecording();
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : String(e);
      state = 'idle';
      return;
    }
    state = 'transcribing';
    try {
      transcript = await Transcribe(recordingDir);
      state = 'done';
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : String(e);
      state = 'idle';
    }
  }

  function reset() {
    state = 'idle';
    transcript = '';
    error = '';
  }

  function formatElapsed(s: number): string {
    const m = Math.floor(s / 60).toString().padStart(2, '0');
    const sec = (s % 60).toString().padStart(2, '0');
    return `${m}:${sec}`;
  }
</script>

{#if state === 'done'}
  <Transcript text={transcript} {recordingDir} onNew={reset} />
{:else}
  <div class="transcribe">
    {#if state === 'idle'}
      <div class="status-row">
        <span class="dot idle"></span>
        <span class="status-label">Ready</span>
      </div>

      {#if error}
        <p class="error">{error}</p>
      {/if}

      <button class="record-btn start" onclick={start}>Start Recording</button>
      <p class="hint">Captures mic + system audio · transcribes with Whisper</p>

    {:else if state === 'recording'}
      <div class="status-row">
        <span class="dot recording"></span>
        <span class="status-label">Recording — {formatElapsed(elapsed)}</span>
      </div>
      <button class="record-btn stop" onclick={stop}>Stop</button>

    {:else}
      <div class="status-row">
        <span class="dot busy"></span>
        <span class="status-label">
          {state === 'stopping' ? 'Stopping…' : 'Transcribing…'}
        </span>
      </div>
      {#if state === 'transcribing'}
        <p class="hint">This may take a moment</p>
      {/if}
    {/if}
  </div>
{/if}

<style>
  .transcribe {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 16px;
    flex: 1;
    padding: 24px 20px;
    color: rgba(255, 255, 255, 0.85);
    font-size: 13px;
  }

  .status-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .dot.idle { background: rgba(255, 255, 255, 0.2); }
  .dot.busy { background: rgba(255, 255, 255, 0.2); }

  .dot.recording {
    background: #e05252;
    box-shadow: 0 0 6px #e05252;
    animation: pulse 1.2s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50%       { opacity: 0.4; }
  }

  .status-label {
    font-size: 13px;
    color: rgba(255, 255, 255, 0.7);
    font-variant-numeric: tabular-nums;
  }

  .record-btn {
    border: none;
    border-radius: 7px;
    font-size: 13px;
    font-family: inherit;
    padding: 8px 22px;
    cursor: pointer;
    transition: opacity 0.15s;
  }

  .record-btn:hover { opacity: 0.85; }
  .record-btn.start { background: rgba(255,255,255,0.12); color: #fff; }
  .record-btn.stop  { background: rgba(224,82,82,0.25); color: #e05252; }

  .hint {
    font-size: 11px;
    color: rgba(255, 255, 255, 0.3);
    text-align: center;
    margin: 0;
  }

  .error {
    font-size: 11px;
    color: #e05252;
    text-align: center;
    margin: 0;
    max-width: 340px;
  }
</style>
