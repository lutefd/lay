<script lang="ts">
  import { tick } from 'svelte';
  import { SendMessage } from '../../wailsjs/go/main/App.js';

  interface ChatMessage {
    role: 'user' | 'assistant';
    content: string;
  }

  let messages = $state<ChatMessage[]>([]);
  let input = $state('');
  let loading = $state(false);
  let error = $state('');
  let messagesEl = $state<HTMLElement | null>(null);

  async function send() {
    const text = input.trim();
    if (!text || loading) return;

    error = '';
    input = '';
    messages = [...messages, { role: 'user', content: text }];
    loading = true;

    await tick();
    scrollToBottom();

    try {
      const reply = await SendMessage(JSON.stringify(messages));
      messages = [...messages, { role: 'assistant', content: reply }];
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : String(e);
    } finally {
      loading = false;
      await tick();
      scrollToBottom();
    }
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      send();
    }
  }

  function scrollToBottom() {
    if (messagesEl) messagesEl.scrollTop = messagesEl.scrollHeight;
  }

  function clearChat() {
    messages = [];
    error = '';
  }
</script>

<div class="chat-panel">
  <div class="chat-toolbar">
    {#if messages.length > 0}
      <button class="clear-btn" onclick={clearChat}>clear</button>
    {/if}
  </div>

  <div class="messages" bind:this={messagesEl}>
    {#if messages.length === 0 && !loading}
      <p class="empty-hint">Ask anything about your meeting…</p>
    {/if}

    {#each messages as msg}
      <div class="message {msg.role}">
        <span class="role-label">{msg.role === 'user' ? 'you' : 'ai'}</span>
        <div class="bubble">{msg.content}</div>
      </div>
    {/each}

    {#if loading}
      <div class="message assistant">
        <span class="role-label">ai</span>
        <div class="bubble loading">
          <span class="dot-bounce">●</span>
          <span class="dot-bounce delay1">●</span>
          <span class="dot-bounce delay2">●</span>
        </div>
      </div>
    {/if}

    {#if error}
      <div class="error-banner">{error}</div>
    {/if}
  </div>

  <div class="chat-input-row">
    <textarea
      class="chat-input"
      bind:value={input}
      onkeydown={onKeydown}
      placeholder="Message… (Enter to send, Shift+Enter for new line)"
      rows={2}
      disabled={loading}
    ></textarea>
    <button class="send-btn" onclick={send} disabled={loading || !input.trim()}>
      {loading ? '…' : '↑'}
    </button>
  </div>
</div>

<style>
  .chat-panel {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-height: 0;
  }

  .chat-toolbar {
    display: flex;
    justify-content: flex-end;
    padding: 4px 10px;
    min-height: 24px;
    flex-shrink: 0;
  }

  .clear-btn {
    background: none;
    border: none;
    color: rgba(255, 255, 255, 0.3);
    font-size: 11px;
    font-family: inherit;
    cursor: pointer;
    padding: 2px 6px;
    border-radius: 4px;
    transition: color 0.15s, background 0.15s;
  }

  .clear-btn:hover {
    color: rgba(255, 255, 255, 0.7);
    background: rgba(255, 255, 255, 0.06);
  }

  .messages {
    flex: 1;
    overflow-y: auto;
    padding: 8px 12px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    scroll-behavior: smooth;
  }

  .messages::-webkit-scrollbar { width: 4px; }
  .messages::-webkit-scrollbar-track { background: transparent; }
  .messages::-webkit-scrollbar-thumb { background: rgba(255, 255, 255, 0.12); border-radius: 2px; }

  .empty-hint {
    color: rgba(255, 255, 255, 0.2);
    font-size: 13px;
    text-align: center;
    margin: auto;
  }

  .message {
    display: flex;
    flex-direction: column;
    gap: 3px;
  }

  .role-label {
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: rgba(255, 255, 255, 0.3);
  }

  .message.user .role-label  { color: #7c9ef5aa; }
  .message.assistant .role-label { color: #4caf82aa; }

  .bubble {
    background: rgba(255, 255, 255, 0.05);
    border-radius: 8px;
    padding: 8px 12px;
    font-size: 13px;
    line-height: 1.6;
    color: rgba(255, 255, 255, 0.87);
    white-space: pre-wrap;
    word-break: break-word;
  }

  .message.user .bubble {
    background: rgba(124, 158, 245, 0.12);
    align-self: flex-end;
  }

  .bubble.loading {
    display: flex;
    gap: 4px;
    align-items: center;
    color: rgba(255, 255, 255, 0.4);
    font-size: 10px;
  }

  .dot-bounce {
    animation: bounce 1.2s infinite;
  }
  .dot-bounce.delay1 { animation-delay: 0.2s; }
  .dot-bounce.delay2 { animation-delay: 0.4s; }

  @keyframes bounce {
    0%, 80%, 100% { transform: translateY(0); opacity: 0.4; }
    40%           { transform: translateY(-4px); opacity: 1; }
  }

  .error-banner {
    background: rgba(220, 60, 60, 0.15);
    border: 1px solid rgba(220, 60, 60, 0.3);
    border-radius: 6px;
    color: #ff8a8a;
    font-size: 12px;
    padding: 8px 12px;
  }

  .chat-input-row {
    display: flex;
    gap: 6px;
    padding: 8px 10px;
    border-top: 1px solid rgba(255, 255, 255, 0.07);
    flex-shrink: 0;
  }

  .chat-input {
    flex: 1;
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    color: rgba(255, 255, 255, 0.87);
    font-family: inherit;
    font-size: 13px;
    line-height: 1.5;
    padding: 7px 10px;
    resize: none;
    outline: none;
    transition: border-color 0.15s;
  }

  .chat-input:focus {
    border-color: rgba(124, 158, 245, 0.5);
  }

  .chat-input::placeholder {
    color: rgba(255, 255, 255, 0.2);
  }

  .chat-input:disabled {
    opacity: 0.5;
  }

  .send-btn {
    background: rgba(124, 158, 245, 0.2);
    border: 1px solid rgba(124, 158, 245, 0.3);
    border-radius: 8px;
    color: #7c9ef5;
    font-size: 18px;
    width: 38px;
    cursor: pointer;
    transition: background 0.15s, opacity 0.15s;
    flex-shrink: 0;
    align-self: flex-end;
  }

  .send-btn:hover:not(:disabled) {
    background: rgba(124, 158, 245, 0.35);
  }

  .send-btn:disabled {
    opacity: 0.35;
    cursor: not-allowed;
  }
</style>
