<script lang="ts">
  import { tick } from 'svelte';
  import { SendMessage } from '../../wailsjs/go/main/App.js';
  import Markdown from './Markdown.svelte';
  import type { ChatMessage } from './types.js';

  interface Props {
    messages: ChatMessage[];
  }

  let { messages = $bindable([]) }: Props = $props();

  let input = $state('');
  let loading = $state(false);
  let error = $state('');
  let messagesEl = $state<HTMLElement | null>(null);
  let pendingImages = $state<string[]>([]);

  function fileToBase64(file: File): Promise<string> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => {
        const result = reader.result as string;
        // strip the data:image/...;base64, prefix
        resolve(result.replace(/^data:image\/[^;]+;base64,/, ''));
      };
      reader.onerror = reject;
      reader.readAsDataURL(file);
    });
  }

  function onPaste(e: ClipboardEvent) {
    const items = e.clipboardData?.items;
    if (!items) return;
    for (const item of items) {
      if (item.type.startsWith('image/')) {
        e.preventDefault();
        const file = item.getAsFile();
        if (!file) continue;
        fileToBase64(file).then((b64) => {
          pendingImages = [...pendingImages, b64];
        });
      }
    }
  }

  function removePendingImage(index: number) {
    pendingImages = pendingImages.filter((_, i) => i !== index);
  }

  async function send() {
    const text = input.trim();
    if (!text && pendingImages.length === 0) return;
    if (loading) return;

    error = '';
    const imgs = pendingImages.length > 0 ? [...pendingImages] : undefined;
    input = '';
    pendingImages = [];
    messages = [...messages, { role: 'user', content: text, images: imgs }];
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

  function copyMessage(content: string) {
    navigator.clipboard.writeText(content);
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
        <div class="msg-header">
          <span class="role-label">{msg.role === 'user' ? 'you' : 'ai'}</span>
          {#if msg.role === 'assistant'}
            <button class="copy-btn" onclick={() => copyMessage(msg.content)} title="Copy raw markdown">
              copy
            </button>
          {/if}
        </div>
        {#if msg.role === 'assistant'}
          <!-- copyRaw=true: Cmd+C anywhere on this bubble gives raw markdown -->
          <div class="bubble assistant-bubble">
            <Markdown raw={msg.content} copyRaw={true} />
          </div>
        {:else}
          <div class="bubble user-bubble">
            {#if msg.images && msg.images.length > 0}
              <div class="msg-images">
                {#each msg.images as img}
                  <img src="data:image/png;base64,{img}" alt="Attached image" class="msg-img" />
                {/each}
              </div>
            {/if}
            {#if msg.content}{msg.content}{/if}
          </div>
        {/if}
      </div>
    {/each}

    {#if loading}
      <div class="message assistant">
        <div class="msg-header">
          <span class="role-label">ai</span>
        </div>
        <div class="bubble assistant-bubble loading">
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

  {#if pendingImages.length > 0}
    <div class="pending-images">
      {#each pendingImages as img, i}
        <div class="pending-thumb">
          <img src="data:image/png;base64,{img}" alt="Pasted image" />
          <button class="remove-img" onclick={() => removePendingImage(i)}>×</button>
        </div>
      {/each}
    </div>
  {/if}

  <div class="chat-input-row">
    <textarea
      class="chat-input"
      bind:value={input}
      onkeydown={onKeydown}
      onpaste={onPaste}
      placeholder="Message… (Enter to send, Ctrl+V to paste image)"
      rows={2}
      disabled={loading}
    ></textarea>
    <button class="send-btn" onclick={send} disabled={loading || (!input.trim() && pendingImages.length === 0)}>
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
    min-width: 0;
    overflow: hidden;
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
    overflow-x: hidden;
    padding: 8px 12px;
    display: flex;
    flex-direction: column;
    gap: 14px;
    scroll-behavior: smooth;
    min-width: 0;
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
    gap: 4px;
    min-width: 0;
  }

  .msg-header {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .role-label {
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: rgba(255, 255, 255, 0.3);
  }

  .message.user .role-label   { color: #7c9ef5aa; }
  .message.assistant .role-label { color: #4caf82aa; }

  .copy-btn {
    background: none;
    border: none;
    color: rgba(255, 255, 255, 0.2);
    font-size: 10px;
    font-family: inherit;
    padding: 1px 5px;
    border-radius: 3px;

    transition: color 0.15s, background 0.15s;
    margin-left: auto;
  }

  .copy-btn:hover {
    color: rgba(255, 255, 255, 0.6);
    background: rgba(255, 255, 255, 0.06);
  }

  .bubble {
    border-radius: 8px;
    padding: 8px 12px;
    min-width: 0;
    overflow-wrap: break-word;
  }

  .assistant-bubble {
    background: rgba(255, 255, 255, 0.05);
  }

  .user-bubble {
    background: rgba(124, 158, 245, 0.12);
    align-self: flex-end;
    font-size: 13px;
    line-height: 1.6;
    color: rgba(255, 255, 255, 0.87);
    white-space: pre-wrap;
    word-break: break-word;
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

    transition: background 0.15s, opacity 0.15s;
    flex-shrink: 0;
    align-self: flex-end;
  }

  .send-btn:hover:not(:disabled) {
    background: rgba(124, 158, 245, 0.35);
  }

  .send-btn:disabled {
    opacity: 0.35;
  }

  .pending-images {
    display: flex;
    gap: 6px;
    padding: 4px 10px 0;
    flex-wrap: wrap;
    flex-shrink: 0;
  }

  .pending-thumb {
    position: relative;
    width: 48px;
    height: 48px;
    border-radius: 6px;
    overflow: hidden;
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  .pending-thumb img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .remove-img {
    position: absolute;
    top: -1px;
    right: -1px;
    background: rgba(0, 0, 0, 0.7);
    border: none;
    color: rgba(255, 255, 255, 0.7);
    font-size: 12px;
    width: 16px;
    height: 16px;
    line-height: 16px;
    text-align: center;
    padding: 0;
    border-radius: 0 5px 0 4px;
  }

  .msg-images {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
    margin-bottom: 4px;
  }

  .msg-img {
    max-width: 200px;
    max-height: 150px;
    border-radius: 6px;
    object-fit: contain;
  }
</style>
