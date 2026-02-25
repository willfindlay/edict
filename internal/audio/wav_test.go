package audio

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestEncodeWAV(t *testing.T) {
	samples := []int16{0, 1000, -1000, 32767, -32768}
	var buf bytes.Buffer

	err := EncodeWAV(&buf, samples, 16000, 1)
	if err != nil {
		t.Fatalf("EncodeWAV failed: %v", err)
	}

	data := buf.Bytes()

	// RIFF header
	if string(data[0:4]) != "RIFF" {
		t.Errorf("expected RIFF, got %q", string(data[0:4]))
	}
	if string(data[8:12]) != "WAVE" {
		t.Errorf("expected WAVE, got %q", string(data[8:12]))
	}

	// File size: 36 + data
	fileSize := binary.LittleEndian.Uint32(data[4:8])
	expectedFileSize := uint32(36 + len(samples)*2)
	if fileSize != expectedFileSize {
		t.Errorf("file size: got %d, want %d", fileSize, expectedFileSize)
	}

	// fmt chunk
	if string(data[12:16]) != "fmt " {
		t.Errorf("expected 'fmt ', got %q", string(data[12:16]))
	}
	fmtSize := binary.LittleEndian.Uint32(data[16:20])
	if fmtSize != 16 {
		t.Errorf("fmt size: got %d, want 16", fmtSize)
	}

	audioFormat := binary.LittleEndian.Uint16(data[20:22])
	if audioFormat != 1 {
		t.Errorf("audio format: got %d, want 1 (PCM)", audioFormat)
	}

	channels := binary.LittleEndian.Uint16(data[22:24])
	if channels != 1 {
		t.Errorf("channels: got %d, want 1", channels)
	}

	sampleRate := binary.LittleEndian.Uint32(data[24:28])
	if sampleRate != 16000 {
		t.Errorf("sample rate: got %d, want 16000", sampleRate)
	}

	byteRate := binary.LittleEndian.Uint32(data[28:32])
	if byteRate != 32000 { // 16000 * 1 * 2
		t.Errorf("byte rate: got %d, want 32000", byteRate)
	}

	blockAlign := binary.LittleEndian.Uint16(data[32:34])
	if blockAlign != 2 {
		t.Errorf("block align: got %d, want 2", blockAlign)
	}

	bitsPerSample := binary.LittleEndian.Uint16(data[34:36])
	if bitsPerSample != 16 {
		t.Errorf("bits per sample: got %d, want 16", bitsPerSample)
	}

	// data chunk
	if string(data[36:40]) != "data" {
		t.Errorf("expected 'data', got %q", string(data[36:40]))
	}

	dataSize := binary.LittleEndian.Uint32(data[40:44])
	expectedDataSize := uint32(len(samples) * 2)
	if dataSize != expectedDataSize {
		t.Errorf("data size: got %d, want %d", dataSize, expectedDataSize)
	}

	// Verify sample data roundtrips
	decoded := make([]int16, len(samples))
	if err := binary.Read(bytes.NewReader(data[44:]), binary.LittleEndian, &decoded); err != nil {
		t.Fatalf("failed to decode samples: %v", err)
	}
	for i, s := range samples {
		if decoded[i] != s {
			t.Errorf("sample[%d]: got %d, want %d", i, decoded[i], s)
		}
	}
}

func TestEncodeWAVEmpty(t *testing.T) {
	var buf bytes.Buffer
	err := EncodeWAV(&buf, nil, 16000, 1)
	if err != nil {
		t.Fatalf("EncodeWAV with empty samples failed: %v", err)
	}

	data := buf.Bytes()
	// Header is 44 bytes, no sample data
	if len(data) != 44 {
		t.Errorf("expected 44 bytes, got %d", len(data))
	}
}

func TestEncodeWAVStereo(t *testing.T) {
	// 2 channels, 4 samples total (2 frames)
	samples := []int16{100, -100, 200, -200}
	var buf bytes.Buffer

	err := EncodeWAV(&buf, samples, 16000, 2)
	if err != nil {
		t.Fatalf("EncodeWAV stereo failed: %v", err)
	}

	data := buf.Bytes()
	channels := binary.LittleEndian.Uint16(data[22:24])
	if channels != 2 {
		t.Errorf("channels: got %d, want 2", channels)
	}

	blockAlign := binary.LittleEndian.Uint16(data[32:34])
	if blockAlign != 4 { // 2 channels * 2 bytes
		t.Errorf("block align: got %d, want 4", blockAlign)
	}
}
