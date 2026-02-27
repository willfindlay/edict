package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/willfindlay/edict/internal/audio"
	"github.com/willfindlay/edict/internal/config"
	"github.com/willfindlay/edict/internal/hotkey"
	"github.com/willfindlay/edict/internal/input"
	"github.com/willfindlay/edict/internal/output"
	"github.com/willfindlay/edict/internal/overlay"
	"github.com/willfindlay/edict/internal/transcribe"
	"github.com/willfindlay/edict/internal/whisper"
)

func main() {
	configPath := flag.String("config", "", "path to config.toml")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signalNotify(sigCh)
	go func() {
		<-sigCh
		log.Println("shutting down...")
		cancel()
	}()

	// Check output backend availability
	typer := output.NewTyper(output.Backend(cfg.Output.Backend), cfg.Output.KeystrokeDelayUs)
	if err := typer.CheckAvailable(); err != nil {
		log.Printf("warning: %v", err)
	}

	// Start whisper server
	whisperSrv := transcribe.NewServer(transcribe.ServerConfig{
		ServerPath: cfg.Whisper.ServerPath,
		Host:       cfg.Whisper.Host,
		Port:       cfg.Whisper.Port,
		ModelPath:  cfg.Whisper.ModelPath,
		GPULayers:  cfg.Whisper.GPULayers,
		Threads:    cfg.Whisper.Threads,
	})

	if cfg.Whisper.ModelPath != "" {
		log.Println("starting whisper-server...")
		if err := whisperSrv.Start(ctx); err != nil {
			log.Fatalf("whisper-server: %v", err) //nolint:gocritic // defers are cleanup; process is exiting
		}
		defer whisperSrv.Stop()
		log.Println("whisper-server ready")
	} else {
		log.Printf("using existing whisper-server at %s", whisperSrv.Addr())
	}

	whisperClient := transcribe.NewClient(whisperSrv.Addr())

	if err := whisperClient.Ping(); err != nil {
		log.Fatalf("whisper-server at %s is not reachable: %v\nStart whisper-server with: docker compose up -d", whisperSrv.Addr(), err)
	}

	// Audio buffer
	ringBuf := audio.NewSampleBuffer()

	// Input mode
	mode := createMode(cfg)

	// Overlay
	done := make(chan struct{})
	win := overlay.NewWindow(overlay.WindowConfig{
		Width:          cfg.Overlay.Width,
		Height:         cfg.Overlay.Height,
		FPS:            cfg.Overlay.FPS,
		Enabled:        cfg.Overlay.Enabled,
		PreviewEnabled: cfg.Preview.Enabled,
		FontSize:       cfg.Preview.FontSize,
		FontPath:       cfg.Preview.FontPath,
		MaxLines:       cfg.Preview.MaxLines,
		Padding:        cfg.Preview.Padding,
	})

	// Hotkey listener goroutine
	listener := hotkey.NewListener(cfg.Hotkey.Modifier, cfg.Hotkey.Key)
	go listener.Start()

	// Pipeline goroutine
	go runPipeline(ctx, listener, mode, ringBuf, whisperClient, typer, win, cfg)

	// Wait for shutdown in main goroutine, running overlay event loop
	go func() {
		<-ctx.Done()
		listener.Stop()
		close(done)
	}()

	hotkey := cfg.Hotkey.Modifier
	if cfg.Hotkey.Key != "" {
		hotkey += "+" + cfg.Hotkey.Key
	}
	log.Printf("edict ready (hotkey: %s, mode: %s)", hotkey, cfg.Input.Mode)

	// Overlay run blocks on main goroutine (required for OpenGL)
	win.Run(done)
}

func loadConfig(path string) (config.Config, error) {
	if path == "" {
		for _, c := range configCandidates() {
			if _, err := os.Stat(c); err == nil {
				return config.Load(c)
			}
		}
		cfg := config.Default()
		if err := cfg.Validate(); err != nil {
			return cfg, fmt.Errorf("validate default config: %w", err)
		}
		return cfg, nil
	}
	return config.Load(path)
}

func createMode(cfg config.Config) input.Mode {
	switch cfg.Input.Mode {
	case "toggle":
		return input.NewToggle()
	case "auto":
		return input.NewVAD(cfg.Input.VADThreshold, cfg.Input.VADSilenceMs)
	default:
		return input.NewHold()
	}
}

func runPipeline(
	ctx context.Context,
	listener *hotkey.Listener,
	mode input.Mode,
	ringBuf *audio.SampleBuffer,
	client *transcribe.Client,
	typer output.Typer,
	win *overlay.Window,
	cfg config.Config,
) {
	// Pin this goroutine to a single OS thread. WASAPI (Windows) uses COM for
	// audio device access, and COM must be initialized on the thread that uses
	// it. Without this, Go's scheduler can migrate the goroutine between OS
	// threads, causing COM calls in miniaudio to fail.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var recorder *audio.Recorder
	var recorderErr error

	// VAD auto-stop signal channel
	vadStop := make(chan struct{}, 1)

	// Preview transcription cancellation
	var previewCancel context.CancelFunc

	for {
		select {
		case <-ctx.Done():
			if previewCancel != nil {
				previewCancel()
			}
			if recorder != nil {
				recorder.Close()
			}
			return

		case <-vadStop:
			if previewCancel != nil {
				previewCancel()
				previewCancel = nil
			}
			handleStop(recorder, ringBuf, client, typer, win, cfg)

		case ev := <-listener.Events():
			action := mode.HandleEvent(ev)

			switch action {
			case input.Start:
				ringBuf.Reset()

				if recorder == nil {
					recorder, recorderErr = audio.NewRecorder(
						audio.RecorderConfig{
							SampleRate: cfg.Audio.SampleRate,
							Channels:   cfg.Audio.Channels,
							Backend:    cfg.Audio.Backend,
						},
						func(samples []int16) {
							ringBuf.Write(samples)
							win.AddSamples(samples)

							// For VAD mode, check audio on each callback
							if a := mode.HandleAudio(samples); a == input.Stop {
								select {
								case vadStop <- struct{}{}:
								default:
								}
							}
						},
					)
					if recorderErr != nil {
						log.Printf("recorder init failed: %v", recorderErr)
						continue
					}
				}

				if err := recorder.Start(); err != nil {
					log.Printf("recorder start failed: %v", err)
					continue
				}

				win.Commands() <- overlay.Show
				log.Println("recording...")

				// Spawn preview transcription goroutine
				if cfg.Preview.Enabled {
					var previewCtx context.Context
					previewCtx, previewCancel = context.WithCancel(ctx)
					go runPreviewTranscription(previewCtx, ringBuf, client, win, cfg)
				}

			case input.Stop:
				if previewCancel != nil {
					previewCancel()
					previewCancel = nil
				}
				handleStop(recorder, ringBuf, client, typer, win, cfg)
			}
		}
	}
}

func runPreviewTranscription(
	ctx context.Context,
	ringBuf *audio.SampleBuffer,
	client *transcribe.Client,
	win *overlay.Window,
	cfg config.Config,
) {
	ticker := time.NewTicker(time.Duration(cfg.Preview.IntervalMs) * time.Millisecond)
	defer ticker.Stop()

	minSamples := cfg.Audio.SampleRate / 2 // 0.5 second minimum

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			samples := ringBuf.Snapshot()
			if len(samples) < minSamples {
				continue
			}

			var wavBuf bytes.Buffer
			if err := audio.EncodeWAV(&wavBuf, samples, cfg.Audio.SampleRate, cfg.Audio.Channels); err != nil {
				log.Printf("preview WAV encode failed: %v", err)
				continue
			}

			text, err := client.Transcribe(wavBuf.Bytes(), whisper.Prompt)
			if err != nil {
				log.Printf("preview transcription failed: %v", err)
				continue
			}

			if text != "" {
				log.Printf("preview: %q", text)
			}
			win.SetPreviewText(text)
		}
	}
}

func handleStop(
	recorder *audio.Recorder,
	ringBuf *audio.SampleBuffer,
	client *transcribe.Client,
	typer output.Typer,
	win *overlay.Window,
	cfg config.Config,
) {
	if recorder != nil {
		recorder.Stop()
	}
	win.Commands() <- overlay.Hide

	samples := ringBuf.Drain()
	if len(samples) == 0 {
		log.Println("no audio captured")
		return
	}

	// Encode WAV
	var wavBuf bytes.Buffer
	if err := audio.EncodeWAV(&wavBuf, samples, cfg.Audio.SampleRate, cfg.Audio.Channels); err != nil {
		log.Printf("WAV encode failed: %v", err)
		return
	}

	// Transcribe
	log.Println("transcribing...")
	text, err := client.Transcribe(wavBuf.Bytes(), whisper.Prompt)
	if err != nil {
		log.Printf("transcription failed: %v", err)
		return
	}

	if text == "" {
		log.Println("empty transcription")
		return
	}

	log.Printf("transcribed: %q", text)

	// Type result
	if err := typer.Type(text); err != nil {
		log.Printf("typing failed: %v", err)
	}
}
