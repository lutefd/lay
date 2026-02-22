<script lang="ts">
  import Header from './lib/Header.svelte';
  import Notes from './lib/Notes.svelte';
  import Chat from './lib/Chat.svelte';
  import Settings from './lib/Settings.svelte';
  import type { ChatMessage } from './lib/types.js';

  let activeTab = $state<'notes' | 'chat' | 'settings'>('notes');
  let chatMessages = $state<ChatMessage[]>([]);
</script>

<div class="app">
  <Header bind:activeTab />

  <main class="content">
    {#if activeTab === 'notes'}
      <Notes />
    {:else if activeTab === 'chat'}
      <Chat bind:messages={chatMessages} />
    {:else}
      <Settings />
    {/if}
  </main>
</div>

<style>
  .app {
    display: flex;
    flex-direction: column;
    height: 100vh;
    overflow: hidden;
    border-radius: 10px;
    border: 1px solid rgba(255, 255, 255, 0.08);
  }

  .content {
    display: flex;
    flex: 1;
    min-height: 0;
    overflow: hidden;
  }
</style>
