<script lang="ts">
  import Header from './lib/Header.svelte';
  import Notes from './lib/Notes.svelte';
  import Chat from './lib/Chat.svelte';
  import Transcribe from './lib/Transcribe.svelte';
  import Settings from './lib/Settings.svelte';
  import type { ChatMessage } from './lib/types.js';

  let activeTab = $state<'notes' | 'chat' | 'transcribe' | 'settings'>('notes');
  let chatMessages = $state<ChatMessage[]>([]);
  let isRecording = $state(false);
</script>

<div class="app">
  <Header bind:activeTab {isRecording} />

  <main class="content">
    {#if activeTab === 'notes'}
      <Notes />
    {:else if activeTab === 'chat'}
      <Chat bind:messages={chatMessages} />
    {:else if activeTab === 'settings'}
      <Settings />
    {/if}
    <!-- Transcribe is always mounted so recording survives tab switches -->
    <div class="transcribe-slot" class:hidden={activeTab !== 'transcribe'}>
      <Transcribe onRecordingChange={(v) => (isRecording = v)} />
    </div>
  </main>
</div>

<style>
  .app {
    display: flex;
    flex-direction: column;
    height: 100vh;
    overflow: hidden;
    border-radius: var(--window-radius);
    border: 1px solid rgba(255, 255, 255, 0.08);
  }

  .content {
    display: flex;
    flex: 1;
    min-height: 0;
    overflow: hidden;
  }

  .transcribe-slot {
    display: contents;
  }

  .transcribe-slot.hidden {
    display: none;
  }
</style>
