package audio

import (
	"encoding/binary"
	"fmt"
	"log"
	"sync"

	"github.com/gen2brain/malgo"
)

// Recorder captures audio from the default microphone using miniaudio.
type Recorder struct {
	mu        sync.Mutex
	ctx       *malgo.AllocatedContext
	device    *malgo.Device
	onSamples func([]int16)
	running   bool
	config    RecorderConfig
}

type RecorderConfig struct {
	SampleRate int
	Channels   int
	Backend    string // empty = auto-detect
}

var audioBackendMap = map[string]malgo.Backend{
	"wasapi":     malgo.BackendWasapi,
	"dsound":     malgo.BackendDsound,
	"winmm":      malgo.BackendWinmm,
	"pulseaudio": malgo.BackendPulseaudio,
	"alsa":       malgo.BackendAlsa,
	"jack":       malgo.BackendJack,
}

func resolveBackends(name string) []malgo.Backend {
	if b, ok := audioBackendMap[name]; ok {
		return []malgo.Backend{b}
	}
	return nil
}

// NewRecorder creates a recorder (does not start capturing).
func NewRecorder(cfg RecorderConfig, onSamples func([]int16)) (*Recorder, error) {
	backends := resolveBackends(cfg.Backend)
	if cfg.Backend != "" {
		log.Printf("audio: using backend %q", cfg.Backend)
	}

	ctxCfg := malgo.ContextConfig{}
	ctx, err := malgo.InitContext(backends, ctxCfg, func(msg string) {
		log.Printf("miniaudio: %s", msg)
	})
	if err != nil {
		return nil, fmt.Errorf("init audio context: %w", err)
	}

	return &Recorder{
		ctx:       ctx,
		onSamples: onSamples,
		config:    cfg,
	}, nil
}

// Start begins capturing audio from the default input device.
func (r *Recorder) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.running {
		return nil
	}

	devCfg := malgo.DefaultDeviceConfig(malgo.Capture)
	devCfg.Capture.Format = malgo.FormatS16
	devCfg.Capture.Channels = uint32(r.config.Channels)
	devCfg.SampleRate = uint32(r.config.SampleRate)

	onRecvFrames := func(outputSamples, inputSamples []byte, frameCount uint32) {
		if r.onSamples == nil || len(inputSamples) == 0 {
			return
		}

		samples := make([]int16, len(inputSamples)/2)
		for i := range samples {
			samples[i] = int16(binary.LittleEndian.Uint16(inputSamples[i*2:]))
		}
		r.onSamples(samples)
	}

	devCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}

	dev, err := malgo.InitDevice(r.ctx.Context, devCfg, devCallbacks)
	if err != nil {
		r.logCaptureDevices()
		return fmt.Errorf("init capture device: %w", err)
	}

	if err := dev.Start(); err != nil {
		dev.Uninit()
		return fmt.Errorf("start capture: %w", err)
	}

	r.device = dev
	r.running = true
	return nil
}

// Stop halts audio capture.
func (r *Recorder) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running {
		return
	}

	_ = r.device.Stop()
	r.device.Uninit()
	r.device = nil
	r.running = false
}

// Close releases all resources.
func (r *Recorder) Close() {
	r.Stop()
	if r.ctx != nil {
		_ = r.ctx.Uninit()
		r.ctx.Free()
		r.ctx = nil
	}
}

func (r *Recorder) logCaptureDevices() {
	devices, err := r.ctx.Devices(malgo.Capture)
	if err != nil {
		log.Printf("audio: failed to enumerate capture devices: %v", err)
		return
	}
	if len(devices) == 0 {
		log.Printf("audio: no capture devices found")
		return
	}
	log.Printf("audio: %d capture device(s) available:", len(devices))
	for i, d := range devices {
		suffix := ""
		if d.IsDefault != 0 {
			suffix = " (default)"
		}
		log.Printf("audio:   %d: %s%s", i, d.Name(), suffix)
	}
}

// SetOnSamples changes the callback for incoming audio data.
func (r *Recorder) SetOnSamples(fn func([]int16)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onSamples = fn
}
