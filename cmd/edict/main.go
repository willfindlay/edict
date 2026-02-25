package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/willfindlay/edict/internal/audio"
	"github.com/willfindlay/edict/internal/config"
	edictctx "github.com/willfindlay/edict/internal/context"
	"github.com/willfindlay/edict/internal/hotkey"
	"github.com/willfindlay/edict/internal/input"
	"github.com/willfindlay/edict/internal/output"
	"github.com/willfindlay/edict/internal/overlay"
	"github.com/willfindlay/edict/internal/transcribe"

	"context"
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
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
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
			log.Fatalf("whisper-server: %v", err)
		}
		defer whisperSrv.Stop()
		log.Println("whisper-server ready")
	} else {
		log.Printf("using existing whisper-server at %s", whisperSrv.Addr())
	}

	whisperClient := transcribe.NewClient(whisperSrv.Addr())

	// Audio buffer
	ringBuf := audio.NewRingBuffer(cfg.Audio.SampleRate * 30) // 30 seconds max

	// Input mode
	mode := createMode(cfg)

	// Context scanner state
	var promptMu sync.RWMutex
	var whisperPrompt string

	// Overlay
	done := make(chan struct{})
	win := overlay.NewWindow(overlay.WindowConfig{
		Width:   cfg.Overlay.Width,
		Height:  cfg.Overlay.Height,
		FPS:     cfg.Overlay.FPS,
		Enabled: cfg.Overlay.Enabled,
	})

	// Hotkey listener goroutine
	listener := hotkey.NewListener(cfg.Hotkey.Modifier, cfg.Hotkey.Key)
	go listener.Start()

	// Context scanner goroutine
	go runContextScanner(ctx, &promptMu, &whisperPrompt)

	// Pipeline goroutine
	go runPipeline(ctx, listener, mode, ringBuf, whisperClient, typer, win, &promptMu, &whisperPrompt, cfg)

	// Wait for shutdown in main goroutine, running overlay event loop
	go func() {
		<-ctx.Done()
		listener.Stop()
		close(done)
	}()

	log.Printf("edict ready (hotkey: %s+%s, mode: %s)", cfg.Hotkey.Modifier, cfg.Hotkey.Key, cfg.Input.Mode)

	// Overlay run blocks on main goroutine (required for OpenGL)
	win.Run(done)
}

func loadConfig(path string) (config.Config, error) {
	if path == "" {
		candidates := []string{
			"config.toml",
			os.ExpandEnv("$HOME/.config/edict/config.toml"),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return config.Load(c)
			}
		}
		return config.Default(), nil
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

func runContextScanner(ctx context.Context, mu *sync.RWMutex, prompt *string) {
	scan := func() {
		procs := edictctx.ScanClaudeProcesses()
		if len(procs) == 0 {
			return
		}

		// Use the first detected project
		proc := procs[0]
		projectName := edictctx.ProjectName(proc.CWD)
		terms := edictctx.ExtractClaudeMDTerms(proc.CWD)
		memTerms := edictctx.ExtractMemoryTerms(proc.CWD)
		terms = append(terms, memTerms...)
		skillNames := edictctx.DiscoverSkills(proc.CWD)

		built := edictctx.BuildPrompt(projectName, terms, skillNames)

		mu.Lock()
		*prompt = built
		mu.Unlock()

		log.Printf("context updated: project=%s, terms=%d, skills=%d", projectName, len(terms), len(skillNames))
	}

	// Initial scan
	scan()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			scan()
		}
	}
}

func runPipeline(
	ctx context.Context,
	listener *hotkey.Listener,
	mode input.Mode,
	ringBuf *audio.RingBuffer,
	client *transcribe.Client,
	typer output.Typer,
	win *overlay.Window,
	promptMu *sync.RWMutex,
	whisperPrompt *string,
	cfg config.Config,
) {
	var recorder *audio.Recorder
	var recorderErr error

	// VAD auto-stop signal channel
	vadStop := make(chan struct{}, 1)

	for {
		select {
		case <-ctx.Done():
			if recorder != nil {
				recorder.Close()
			}
			return

		case <-vadStop:
			handleStop(recorder, ringBuf, client, typer, win, promptMu, whisperPrompt, cfg)

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

			case input.Stop:
				handleStop(recorder, ringBuf, client, typer, win, promptMu, whisperPrompt, cfg)
			}
		}
	}
}

func handleStop(
	recorder *audio.Recorder,
	ringBuf *audio.RingBuffer,
	client *transcribe.Client,
	typer output.Typer,
	win *overlay.Window,
	promptMu *sync.RWMutex,
	whisperPrompt *string,
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

	// Get current prompt
	promptMu.RLock()
	prompt := *whisperPrompt
	promptMu.RUnlock()

	// Transcribe
	log.Println("transcribing...")
	text, err := client.Transcribe(wavBuf.Bytes(), prompt)
	if err != nil {
		log.Printf("transcription failed: %v", err)
		return
	}

	if text == "" {
		log.Println("empty transcription")
		return
	}

	log.Printf("transcribed: %s", text)

	// Type result
	if err := typer.Type(text); err != nil {
		log.Printf("typing failed: %v", err)
	}
}
