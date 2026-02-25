package audio

import (
	"encoding/binary"
	"io"
)

// EncodeWAV writes a RIFF/WAV file with PCM S16LE mono data.
func EncodeWAV(w io.Writer, samples []int16, sampleRate, channels int) error {
	dataSize := len(samples) * 2 // 2 bytes per int16
	bitsPerSample := 16
	blockAlign := channels * bitsPerSample / 8
	byteRate := sampleRate * blockAlign

	// RIFF header
	if _, err := w.Write([]byte("RIFF")); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(36+dataSize)); err != nil {
		return err
	}
	if _, err := w.Write([]byte("WAVE")); err != nil {
		return err
	}

	// fmt sub-chunk
	if _, err := w.Write([]byte("fmt ")); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(16)); err != nil { // sub-chunk size
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint16(1)); err != nil { // PCM format
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint16(channels)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(sampleRate)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(byteRate)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint16(blockAlign)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint16(bitsPerSample)); err != nil {
		return err
	}

	// data sub-chunk
	if _, err := w.Write([]byte("data")); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(dataSize)); err != nil {
		return err
	}
	return binary.Write(w, binary.LittleEndian, samples)
}
