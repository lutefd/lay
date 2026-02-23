package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func (a *App) StartRecording() (string, error) {
	a.currentTranscript = ""
	dir, err := recordingsDir()
	if err != nil {
		return "", err
	}
	if err := StartCapture(dir); err != nil {
		return "", err
	}
	return dir, nil
}

func (a *App) StopRecording() error {
	StopCapture()
	return nil
}

// Transcribe converts mic.caf to 16 kHz mono, runs whisper-cli, saves the
// result to ~/.lay/transcripts/<session>.md and returns the transcript text.
func (a *App) Transcribe(recordingDir string) (string, error) {
	micCaf := filepath.Join(recordingDir, "mic.caf")
	mixedWav := filepath.Join(recordingDir, "mixed.wav")

	if err := afconvert(micCaf, mixedWav); err != nil {
		return "", fmt.Errorf("audio conversion: %w", err)
	}

	whisperBin, err := findWhisper()
	if err != nil {
		return "", err
	}
	modelPath, err := findModel()
	if err != nil {
		return "", err
	}

	transcript, err := runWhisper(whisperBin, modelPath, mixedWav)
	if err != nil {
		return "", err
	}

	if err := saveTranscript(recordingDir, transcript); err != nil {
		return "", fmt.Errorf("save transcript: %w", err)
	}

	a.currentTranscript = transcript
	os.RemoveAll(recordingDir)
	return transcript, nil
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

	_, err = fmt.Fprintf(f, "\n\n## Transcript — %s\n\n%s\n", session, stripTimestamps(string(data)))
	return err
}

// afconvert resamples src to 16 kHz mono signed-int16 WAV using the macOS
// built-in afconvert tool (no external dependencies required).
func afconvert(src, dst string) error {
	out, err := exec.Command(
		"afconvert", "-f", "WAVE", "-d", "LEI16@16000", "-c", "1", src, dst,
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
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
	transcript := strings.TrimSpace(string(out))
	if transcript == "" {
		return "", fmt.Errorf("whisper returned no output — check the audio file")
	}
	return transcript, nil
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

// stripTimestamps removes whisper's [HH:MM:SS.mmm --> HH:MM:SS.mmm] prefixes.
func stripTimestamps(s string) string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		if idx := strings.Index(line, "]"); idx >= 0 && strings.HasPrefix(strings.TrimSpace(line), "[") {
			line = strings.TrimSpace(line[idx+1:])
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func recordingsDir() (string, error) {
	ts := time.Now().Format("2006-01-02-15-04-05")
	dir := filepath.Join(layDir(), "recordings", ts)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}
