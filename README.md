# lay

Always-on-top meeting assistant built with Wails and Svelte. It stays off screen capture/sharing, making it ideal for presentations, meeting notes, and quick AI answers without showing the app to others.

**Highlights**
- Frameless, translucent window with native rounded corners on Windows 11
- Always-on-top utility window that opens at the top-right
- Notes editor and AI chat with model selection, tuned for meeting workflows
- Global hotkeys for window toggle and opacity (macOS)

**Keybinds (macOS)**
- Toggle window: `Cmd+Shift+L`
- Opacity 100%: `Cmd+Option+1`
- Opacity 75%: `Cmd+Option+2`
- Opacity 50%: `Cmd+Option+3`
- Opacity 25%: `Cmd+Option+4`
- Move window (focused): `Cmd+Shift+Arrow` (left/right/up/down)

**Data & Config**
- Notes: `~/.lay/notes.md`
- Config: `~/.lay/config.json`
- Default model: `claude-sonnet-4-6`

**Gateway**

You can route all AI requests through a custom gateway (e.g. a corporate proxy that handles auth and model routing). The gateway must expose a Chat Completions–compatible endpoint.

There are two ways to configure a gateway:

*Embed at build time* — place a `gateway.json` in `internal/app/defaults/` before building. It gets baked into the binary so every user gets it out of the box. This file is `.gitignore`d.

```bash
cp gateway.json internal/app/defaults/gateway.json
wails build
```

*Override per user* — drop a `gateway.json` at `~/.lay/gateway.json`. This takes priority over any embedded config.

**`gateway.json` format:**

```json
{
  "name": "My Gateway",
  "url": "https://gateway.example.com/v1/chat/completions",
  "models": [
    { "value": "gpt-4o", "label": "GPT-4o" },
    { "value": "claude-sonnet-4-6", "label": "Claude Sonnet 4.6" },
    { "value": "gemini-2.5-flash", "label": "Gemini 2.5 Flash" }
  ]
}
```

| Field | Description |
|-------|-------------|
| `name` | Label shown in Settings (e.g. "Corp Gateway") |
| `url` | Full endpoint URL — no path is appended by the app |
| `models` | List of models the gateway supports; each needs a `value` (sent to the API) and a `label` (shown in the UI) |

When a gateway is configured, Settings shows a toggle and a model group with the gateway's name. Enabling the toggle routes all requests through the gateway URL. Disabling it reverts to direct Anthropic/OpenAI calls.

**Behavior**
- Initial size: `520x360`
- Minimum size: `520x360`
- Initial position: top-right with a small margin

**Tech Stack**
- Go + Wails v2
- Svelte 5 + Vite
- Anthropic and OpenAI SDKs

**Development**
1. Install the Wails CLI and Go toolchain.
2. Install frontend deps:
   ```bash
   cd frontend
   npm install
   ```
3. Run the app in dev mode:
   ```bash
   wails dev
   ```

**Build**
```bash
wails build
```

**Project Layout**
- `main.go` app options, startup wiring, window positioning
- `app.go` Wails binding wrapper and service interface
- `internal/app/` core logic: config, chat, export, transcription
- `internal/ai/` AI client: Anthropic, OpenAI, and gateway routing
- `internal/platform/` macOS hotkeys, stealth window, audio capture
- `internal/app/defaults/` build-time embedded files (gateway config)
- `frontend/src/` Svelte UI
