<script lang="ts">
  import { marked } from 'marked';

  interface Props {
    raw: string;
    /** When true, Cmd/Ctrl+C anywhere on the block copies the raw markdown. */
    copyRaw?: boolean;
    class?: string;
  }

  let { raw, copyRaw = false, class: className = '' }: Props = $props();

  // marked.parse is sync when no async extensions are configured.
  let html = $derived(marked.parse(raw) as string);

  function handleCopy(e: ClipboardEvent) {
    if (!copyRaw) return;
    e.preventDefault();
    e.clipboardData?.setData('text/plain', raw);
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="md {className}" oncopy={handleCopy}>
  {@html html}
</div>

<style>
  .md {
    font-size: 13px;
    line-height: 1.65;
    color: rgba(255, 255, 255, 0.87);
    word-break: break-word;
  }

  /* --- Block elements --- */
  .md :global(p) {
    margin: 0 0 0.7em;
  }
  .md :global(p:last-child) {
    margin-bottom: 0;
  }

  .md :global(h1),
  .md :global(h2),
  .md :global(h3),
  .md :global(h4),
  .md :global(h5),
  .md :global(h6) {
    font-weight: 600;
    line-height: 1.3;
    margin: 0.9em 0 0.4em;
    color: #fff;
  }
  .md :global(h1) { font-size: 1.25em; }
  .md :global(h2) { font-size: 1.1em; }
  .md :global(h3),
  .md :global(h4),
  .md :global(h5),
  .md :global(h6) { font-size: 1em; }

  .md :global(ul),
  .md :global(ol) {
    margin: 0.4em 0 0.7em;
    padding-left: 1.4em;
  }
  .md :global(li) {
    margin-bottom: 0.25em;
  }

  .md :global(blockquote) {
    border-left: 3px solid rgba(124, 158, 245, 0.4);
    margin: 0.6em 0;
    padding: 0.3em 0.8em;
    color: rgba(255, 255, 255, 0.55);
    font-style: italic;
  }

  .md :global(hr) {
    border: none;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    margin: 0.8em 0;
  }

  /* --- Inline code --- */
  .md :global(code) {
    background: rgba(255, 255, 255, 0.08);
    border-radius: 4px;
    padding: 0.1em 0.35em;
    font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace;
    font-size: 0.88em;
    color: #b5d0ff;
  }

  /* --- Code blocks --- */
  .md :global(pre) {
    background: rgba(0, 0, 0, 0.35);
    border: 1px solid rgba(255, 255, 255, 0.07);
    border-radius: 7px;
    padding: 10px 14px;
    overflow-x: auto;
    margin: 0.6em 0;
  }
  .md :global(pre code) {
    background: none;
    padding: 0;
    border-radius: 0;
    color: rgba(255, 255, 255, 0.82);
    font-size: 12px;
    line-height: 1.6;
  }

  /* --- Tables --- */
  .md :global(table) {
    border-collapse: collapse;
    width: 100%;
    margin: 0.6em 0;
    font-size: 12px;
  }
  .md :global(th),
  .md :global(td) {
    border: 1px solid rgba(255, 255, 255, 0.1);
    padding: 5px 10px;
    text-align: left;
  }
  .md :global(th) {
    background: rgba(255, 255, 255, 0.05);
    color: rgba(255, 255, 255, 0.7);
    font-weight: 600;
  }

  /* --- Links --- */
  .md :global(a) {
    color: #7c9ef5;
    text-decoration: none;
  }
  .md :global(a:hover) {
    text-decoration: underline;
  }

  /* --- Strong / em --- */
  .md :global(strong) { color: #fff; font-weight: 600; }
  .md :global(em) { color: rgba(255, 255, 255, 0.75); }
</style>
