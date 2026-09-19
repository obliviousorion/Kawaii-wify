package ipc

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"

	"github.com/obliviousorion/kawaii-wify/internal/engine"
)

var (
	StatePaused = "Paused"
)

const (
	defaultConnectTimeout = 10 * time.Second
)

// Controller defines the engine capabilities exposed over the IPC layer.

type Controller interface {
	State() engine.State
	Username() string
	SessionToken() string
	Uptime() time.Duration
	LastProbe() time.Time
	IsPaused() bool

	Connect(timeout time.Duration) error
	Disconnect()
}


type DaemonService struct {
	controller Controller
}

func NewDaemonService(controller Controller) *DaemonService {
	return &DaemonService{
		controller: controller,
	}
}

// net/rpc requires you follow the func signature 
// func (s *DaemonService) MethodName(req RequestType, resp *ResponseType) error

func (s *DaemonService) GetStatus(req StatusRequest, resp *StatusResponse) error {
	state := s.controller.State().String()
	if s.controller.IsPaused() {
		state = StatePaused
	}

	resp.State = state
	resp.Username = s.controller.Username()
	resp.SessionToken = s.controller.SessionToken()
	resp.Paused = s.controller.IsPaused()
	resp.LastProbe = s.controller.LastProbe()
	resp.Uptime = s.controller.Uptime().Truncate(time.Second)

	return nil
}

// Connect triggers a synchronous network probe and authentication attempt.

func (s *DaemonService) Connect(req ActionRequest, resp *ActionResponse) error {
	err := s.controller.Connect(defaultConnectTimeout)

	if err != nil {
		resp.Success = false
		resp.Message = fmt.Sprintf("Authentication failed: %v", err)
		return nil
	}

	resp.Success = true
	resp.Message = "Connected successfully."
	return nil
}


func (s *DaemonService) Disconnect(req ActionRequest, resp *ActionResponse) error {
	s.controller.Disconnect()
	resp.Success = true
	resp.Message = "Session disconnected and Daemon Paused"
	return nil
}

// Serve starts the RPC server on the listener and blocks until ctx is canceled.
func Serve(ctx context.Context, listener net.Listener, controller Controller) error {
    server := rpc.NewServer()
    service := NewDaemonService(controller)

    if err := server.RegisterName("Daemon", service); err != nil {
        return fmt.Errorf("failed to register daemon rpc service: %w", err)
    }

    go func() {
        <-ctx.Done()
        _ = listener.Close()
    }()

    for {
        conn, err := listener.Accept()
        if err != nil {
            select {
            case <-ctx.Done():
                return nil
            default:
                log.Printf("[WARN] IPC accept error: %v", err)
                continue
            }
        }
        go server.ServeConn(conn)
    }
}