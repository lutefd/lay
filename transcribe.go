package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const liveChunkInterval = 30 * time.Second

func (a *App) StartRecording() (string, error) {
	a.currentTranscript = ""
	a.liveMu.Lock()
	a.liveSegments = nil
	a.liveChunkSeq = 0
	a.liveMu.Unlock()

	dir, err := recordingsDir()
	if err != nil {
		return "", err
	}
	if err := StartCapture(dir); err != nil {
		return "", err
	}

	liveCtx, cancel := context.WithCancel(a.ctx)
	a.liveCancel = cancel
	go a.liveTranscribeLoop(liveCtx, dir)

	return dir, nil
}

func (a *App) StopRecording() error {
	if a.liveCancel != nil {
		a.liveCancel()
		a.liveCancel = nil
	}
	StopCapture()
	return nil
}

// Transcribe runs whisper separately on mic.caf and system.caf, then merges
// the results by timestamp, saves the transcript, and returns the text.
func (a *App) Transcribe(recordingDir string) (string, error) {
	whisperBin, err := findWhisper()
	if err != nil {
		return "", err
	}
	modelPath, err := findModel()
	if err != nil {
		return "", err
	}

	micPath := filepath.Join(recordingDir, "mic.caf")
	sysPath := filepath.Join(recordingDir, "system.caf")

	transcript, err := transcribeDual(whisperBin, modelPath, micPath, sysPath)
	if err != nil {
		return "", err
	}
	if transcript == "" {
		return "", fmt.Errorf("no transcript produced — check whisper setup and audio")
	}

	if err := saveTranscript(recordingDir, transcript); err != nil {
		return "", fmt.Errorf("save transcript: %w", err)
	}

	a.currentTranscript = transcript
	os.RemoveAll(recordingDir)
	return transcript, nil
}

// liveTranscribeLoop rotates the live mic chunk every liveChunkInterval,
// converts each completed chunk and runs whisper for the live preview.
func (a *App) liveTranscribeLoop(ctx context.Context, dir string) {
	ticker := time.NewTicker(liveChunkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.liveMu.Lock()
			seq := a.liveChunkSeq
			a.liveChunkSeq++
			a.liveMu.Unlock()

			newMic := filepath.Join(dir, fmt.Sprintf("chunk-%d.caf", seq+1))
			if err := RotateChunk(newMic); err != nil {
				continue
			}
			oldMic := filepath.Join(dir, fmt.Sprintf("chunk-%d.caf", seq))
			go a.processChunk(oldMic)
		}
	}
}

// processChunk transcribes mic and system audio chunks separately, merges
// them by timestamp, and emits the result for the live preview.
// The sys chunk path is derived by convention: chunk-N.caf → chunk-sys-N.caf.
func (a *App) processChunk(micCaf string) {
	whisperBin, err := findWhisper()
	if err != nil {
		return
	}
	modelPath, err := findModel()
	if err != nil {
		return
	}

	sysCaf := filepath.Join(
		filepath.Dir(micCaf),
		strings.Replace(filepath.Base(micCaf), "chunk-", "chunk-sys-", 1),
	)

	text, err := transcribeDual(whisperBin, modelPath, micCaf, sysCaf)
	os.Remove(micCaf)
	os.Remove(sysCaf)
	if err != nil || text == "" {
		return
	}

	runtime.EventsEmit(a.ctx, "transcribe:segment", text)
}

// AppendTranscriptToNotes reads the saved transcript for recordingDir and
// appends it as a dated section to ~/.lay/notes.md.
func (a *App) AppendTranscriptToNotes(recordingDir string) error {
	session := filepath.Base(recordingDir)
	src := filepath.Join(layDir(), "transcripts", session+".md")
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("transcript not found: %w", err)
	}

	f, err := os.OpenFile(filepath.Join(layDir(), "notes.md"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "\n\n## Transcript — %s\n\n%s\n", session, string(data))
	return err
}

// minWavBytes is the minimum WAV size worth passing to whisper-cli.
// A 16 kHz mono int16 WAV needs at least ~0.5 s of audio to avoid crashing.
const minWavBytes = 16000 * 2 / 2 // 0.5 s × 16000 Hz × 2 bytes, ÷2 safety margin

// transcribeDual converts micCaf and sysCaf to WAV separately, runs whisper
// on each, merges the results chronologically, and returns clean text.
// System audio is optional — mic-only output is returned when sysCaf is absent.
// Whisper failures are treated as empty output so one bad source never blocks the other.
func transcribeDual(whisperBin, modelPath, micCaf, sysCaf string) (string, error) {
	micRaw := whisperOnCaf(whisperBin, modelPath, micCaf)
	sysRaw := whisperOnCaf(whisperBin, modelPath, sysCaf)
	if micRaw == "" && sysRaw == "" {
		return "", nil
	}
	return mergeTranscripts(micRaw, sysRaw), nil
}

// whisperOnCaf converts a CAF file to WAV and runs whisper, returning the raw
// timestamped output. Returns "" on any error (missing file, too short, crash).
func whisperOnCaf(whisperBin, modelPath, cafPath string) string {
	if fi, err := os.Stat(cafPath); err != nil || fi.Size() == 0 {
		return ""
	}
	wavPath := cafPath + ".wav"
	if err := afconvert(cafPath, wavPath); err != nil {
		return ""
	}
	defer os.Remove(wavPath)
	if fi, err := os.Stat(wavPath); err != nil || fi.Size() < minWavBytes {
		return ""
	}
	out, _ := runWhisper(whisperBin, modelPath, wavPath)
	return out
}

type tsSegment struct {
	start float64
	text  string
	label string // "you" or "them"
}

func parseSegments(raw, label string) []tsSegment {
	var segs []tsSegment
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "[") {
			continue
		}
		idx := strings.Index(line, "]")
		if idx < 0 {
			continue
		}
		meta := line[1:idx] // "00:00:01.000 --> 00:00:04.000"
		text := strings.TrimSpace(line[idx+1:])
		if text == "" {
			continue
		}
		parts := strings.SplitN(meta, " --> ", 2)
		if len(parts) != 2 {
			continue
		}
		start := parseWhisperTS(strings.TrimSpace(parts[0]))
		if start < 0 {
			continue
		}
		segs = append(segs, tsSegment{start: start, text: text, label: label})
	}
	return segs
}

// mergeTranscripts interleaves mic and system whisper output by timestamp.
// Mic segments are prefixed with "[You]" so speakers are distinguishable.
func mergeTranscripts(micRaw, sysRaw string) string {
	segs := parseSegments(micRaw, "you")
	segs = append(segs, parseSegments(sysRaw, "them")...)
	sort.Slice(segs, func(i, j int) bool {
		return segs[i].start < segs[j].start
	})
	var sb strings.Builder
	for _, s := range segs {
		if s.label == "you" {
			sb.WriteString("[You] ")
		} else {
			sb.WriteString("[Them] ")
		}
		sb.WriteString(s.text)
		sb.WriteByte('\n')
	}
	return strings.TrimSpace(sb.String())
}

// afconvert converts src (any CoreAudio-readable format) to a 16 kHz mono
// little-endian int16 WAV at dst using the system afconvert utility.
func afconvert(src, dst string) error {
	cmd := exec.Command("afconvert", "-f", "WAVE", "-d", "LEI16@16000", "-c", "1", src, dst)
	cmd.Stderr = io.Discard
	return cmd.Run()
}

// findWhisper locates the whisper-cli binary: app bundle → ~/.lay/ → $PATH.
func findWhisper() (string, error) {
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "..", "Resources", "whisper-cli")
		if _, err := os.Stat(candidate); err == nil {
			return filepath.Clean(candidate), nil
		}
	}
	local := filepath.Join(layDir(), "whisper-cli")
	if _, err := os.Stat(local); err == nil {
		return local, nil
	}
	for _, name := range []string{"whisper-cli", "main"} {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf(
		"whisper-cli not found — place it at ~/.lay/whisper-cli or run: brew install whisper-cpp",
	)
}

// findModel locates ggml-small.bin: app bundle → ~/.lay/models/.
func findModel() (string, error) {
	const model = "ggml-small.bin"
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "..", "Resources", "models", model)
		if _, err := os.Stat(candidate); err == nil {
			return filepath.Clean(candidate), nil
		}
	}
	local := filepath.Join(layDir(), "models", model)
	if _, err := os.Stat(local); err == nil {
		return local, nil
	}
	return "", fmt.Errorf(
		"model not found — download ggml-small.bin to ~/.lay/models/ from huggingface.co/ggerganov/whisper.cpp",
	)
}

// runWhisper runs whisper-cli and returns the timestamped transcript from stdout.
func runWhisper(bin, model, audio string) (string, error) {
	cmd := exec.Command(bin, "-m", model, "-f", audio, "-l", "auto")
	cmd.Stderr = io.Discard
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("whisper failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// saveTranscript writes the transcript to ~/.lay/transcripts/<session>.md.
func saveTranscript(recordingDir, transcript string) error {
	dir := filepath.Join(layDir(), "transcripts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	session := filepath.Base(recordingDir)
	content := fmt.Sprintf("# Transcript — %s\n\n%s\n", session, transcript)
	return os.WriteFile(filepath.Join(dir, session+".md"), []byte(content), 0o644)
}

func recordingsDir() (string, error) {
	ts := time.Now().Format("2006-01-02-15-04-05")
	dir := filepath.Join(layDir(), "recordings", ts)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func parseWhisperTS(s string) float64 {
	s = strings.TrimSpace(s)
	parts := strings.SplitN(s, ":", 3)
	if len(parts) != 3 {
		return -1
	}
	h, e1 := strconv.Atoi(parts[0])
	m, e2 := strconv.Atoi(parts[1])
	secParts := strings.SplitN(parts[2], ".", 2)
	if len(secParts) != 2 || e1 != nil || e2 != nil {
		return -1
	}
	sec, e3 := strconv.Atoi(secParts[0])
	ms, e4 := strconv.Atoi(secParts[1])
	if e3 != nil || e4 != nil {
		return -1
	}
	return float64(h*3600+m*60+sec) + float64(ms)/1000.0
}
