<script lang="ts">
  import { WindowMinimise, Quit } from '../../wailsjs/runtime/runtime.js';

  interface Props {
    activeTab: 'notes' | 'chat' | 'transcribe' | 'settings';
  }

  let { activeTab = $bindable('notes') }: Props = $props();
</script>

<!-- Entire header is the drag region; only the buttons themselves opt out -->
<header style="--wails-draggable: drag">
  <span class="app-title">lay</span>

  <nav>
    <button
      class="tab-btn"
      class:active={activeTab === 'notes'}
      onclick={() => (activeTab = 'notes')}
      style="--wails-draggable: no-drag"
    >Notes</button>
    <button
      class="tab-btn"
      class:active={activeTab === 'chat'}
      onclick={() => (activeTab = 'chat')}
      style="--wails-draggable: no-drag"
    >Chat</button>
    <button
      class="tab-btn"
      class:active={activeTab === 'transcribe'}
      onclick={() => (activeTab = 'transcribe')}
      style="--wails-draggable: no-drag"
    >Transcribe</button>
    <button
      class="tab-btn"
      class:active={activeTab === 'settings'}
      onclick={() => (activeTab = 'settings')}
      style="--wails-draggable: no-drag"
    >Settings</button>
  </nav>

  <div class="window-controls">
    <button class="ctrl-btn" onclick={WindowMinimise} title="Minimise" style="--wails-draggable: no-drag">−</button>
    <button class="ctrl-btn close" onclick={Quit} title="Quit" style="--wails-draggable: no-drag">×</button>
  </div>
</header>

<style>
  header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 8px;
    height: 38px;
    background: rgba(12, 12, 18, 0.85);
    border-bottom: 1px solid rgba(255, 255, 255, 0.07);
    flex-shrink: 0;
    cursor: move;
  }

  .app-title {
    font-size: 13px;
    font-weight: 600;
    letter-spacing: 0.06em;
    color: rgba(255, 255, 255, 0.5);
    user-select: none;
    padding: 0 6px;
    flex-shrink: 0;
  }

  nav {
    display: flex;
    gap: 2px;
    flex: 1;
  }

  .tab-btn {
    background: none;
    border: none;
    color: rgba(255, 255, 255, 0.45);
    font-size: 12px;
    font-family: inherit;
    padding: 4px 10px;
    border-radius: 5px;
    cursor: pointer;
    transition: color 0.15s, background 0.15s;
  }

  .tab-btn:hover {
    color: rgba(255, 255, 255, 0.8);
    background: rgba(255, 255, 255, 0.06);
  }

  .tab-btn.active {
    color: #fff;
    background: rgba(255, 255, 255, 0.1);
  }

  .window-controls {
    display: flex;
    gap: 4px;
    margin-left: auto;
  }

  .ctrl-btn {
    background: none;
    border: none;
    color: rgba(255, 255, 255, 0.35);
    font-size: 16px;
    line-height: 1;
    width: 26px;
    height: 26px;
    border-radius: 5px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: color 0.15s, background 0.15s;
  }

  .ctrl-btn:hover {
    color: rgba(255, 255, 255, 0.9);
    background: rgba(255, 255, 255, 0.08);
  }

  .ctrl-btn.close:hover {
    background: rgba(220, 60, 60, 0.5);
    color: #fff;
  }
</style>
