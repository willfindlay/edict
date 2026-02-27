package transcribe

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"
)

// ServerConfig holds configuration for the whisper-server subprocess.
type ServerConfig struct {
	ServerPath string
	Host       string
	Port       int
	ModelPath  string
	GPULayers  int
	Threads    int
}

// Server manages the whisper-server subprocess lifecycle.
type Server struct {
	cfg  ServerConfig
	cmd  *exec.Cmd
	addr string
}

// NewServer creates a new whisper-server manager.
func NewServer(cfg ServerConfig) *Server {
	return &Server{
		cfg:  cfg,
		addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	}
}

// Addr returns the host:port address of the server.
func (s *Server) Addr() string {
	return s.addr
}

// Start launches the whisper-server subprocess and waits for it to be ready.
func (s *Server) Start(ctx context.Context) error {
	args := []string{
		"--host", s.cfg.Host,
		"--port", fmt.Sprintf("%d", s.cfg.Port),
		"--threads", fmt.Sprintf("%d", s.cfg.Threads),
		"--gpu-layers", fmt.Sprintf("%d", s.cfg.GPULayers),
	}

	if s.cfg.ModelPath != "" {
		args = append(args, "--model", s.cfg.ModelPath)
	}

	s.cmd = exec.CommandContext(ctx, s.cfg.ServerPath, args...)
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("start whisper-server: %w", err)
	}

	if err := s.waitReady(ctx, 30*time.Second); err != nil {
		s.Stop()
		return fmt.Errorf("whisper-server not ready: %w", err)
	}

	return nil
}

// Stop terminates the whisper-server subprocess.
func (s *Server) Stop() {
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		_ = s.cmd.Wait()
		s.cmd = nil
	}
}

// waitReady polls the TCP port until the server accepts connections.
func (s *Server) waitReady(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		conn, err := net.DialTimeout("tcp", s.addr, 500*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for %s after %v", s.addr, timeout)
}
