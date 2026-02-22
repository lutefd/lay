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

**Data & Config**
- Notes: `~/.lay/notes.md`
- Config: `~/.lay/config.json`
- Default model: `claude-sonnet-4-6`

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
- `main.go` app options, startup wiring, window behavior
- `app.go` backend methods and config persistence
- `macos_darwin.go` macOS hotkeys and window behavior
- `frontend/src` Svelte UI
