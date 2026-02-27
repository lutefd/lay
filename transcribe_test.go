package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestChunkSysPath(t *testing.T) {
	got := chunkSysPath("/tmp/session/chunk-12.caf")
	want := "/tmp/session/chunk-sys-12.caf"
	if got != want {
		t.Fatalf("chunkSysPath() = %q, want %q", got, want)
	}
}

func TestParseWhisperTS(t *testing.T) {
	got := parseWhisperTS("01:02:03.456")
	want := 3723.456
	if got != want {
		t.Fatalf("parseWhisperTS() = %v, want %v", got, want)
	}
	if parseWhisperTS("invalid") != -1 {
		t.Fatalf("parseWhisperTS() should return -1 for invalid input")
	}
}

func TestMergeTranscriptsSortedWithOffset(t *testing.T) {
	micRaw := "[00:00:01.000 --> 00:00:02.000] hi from mic"
	sysRaw := "[00:00:00.500 --> 00:00:01.000] hi from system"

	got := mergeTranscripts(micRaw, sysRaw, 60)
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 merged lines, got %d: %q", len(lines), got)
	}
	if !strings.Contains(lines[0], "[00:01:00.500] [Them] hi from system") {
		t.Fatalf("unexpected first line: %q", lines[0])
	}
	if !strings.Contains(lines[1], "[00:01:01.000] [You] hi from mic") {
		t.Fatalf("unexpected second line: %q", lines[1])
	}
}

func TestDeduplicateSegmentsAcrossSpeakers(t *testing.T) {
	segs := []tsSegment{
		{start: 10, text: "Same text", label: "them"},
		{start: 10.2, text: "same text", label: "you"},
		{start: 75, text: "same text", label: "you"},
	}
	got := deduplicateSegments(segs)
	if len(got) != 2 {
		t.Fatalf("expected 2 segments after dedupe, got %d", len(got))
	}
	if got[0].label != "them" || got[1].start != 75 {
		t.Fatalf("unexpected dedupe result: %+v", got)
	}
}

func TestIsWhisperHallucination(t *testing.T) {
	if !isWhisperHallucination("[Music]") {
		t.Fatalf("expected bracketed non-speech token to be treated as hallucination")
	}
	if !isWhisperHallucination("♪♪ ...") {
		t.Fatalf("expected symbol-only token to be treated as hallucination")
	}
	if isWhisperHallucination("Hello team") {
		t.Fatalf("expected real speech to not be treated as hallucination")
	}
}

func TestAppendLiveSegmentCapsCount(t *testing.T) {
	a := &App{}
	for i := 0; i < maxLiveSegments+25; i++ {
		a.appendLiveSegment(fmt.Sprintf("segment-%03d", i))
	}
	if len(a.liveSegments) != maxLiveSegments {
		t.Fatalf("expected %d segments, got %d", maxLiveSegments, len(a.liveSegments))
	}
	if a.liveSegments[0] != "segment-025" {
		t.Fatalf("expected oldest kept segment to be segment-025, got %q", a.liveSegments[0])
	}
}

func TestAppendLiveSegmentCapsChars(t *testing.T) {
	a := &App{}
	block := strings.Repeat("a", 50000)
	for i := 0; i < 4; i++ {
		a.appendLiveSegment(fmt.Sprintf("%d-%s", i, block))
	}

	total := 0
	for _, s := range a.liveSegments {
		total += len(s)
	}
	if total > maxLiveChars {
		t.Fatalf("expected capped chars <= %d, got %d", maxLiveChars, total)
	}

	joined := strings.Join(a.liveSegments, "\n")
	if !strings.Contains(joined, "3-") {
		t.Fatalf("expected newest segment to be retained")
	}
}

