package keepalive

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

type mockDBPinger struct {
	pingCount int32
}

func (m *mockDBPinger) Ping(ctx context.Context) error {
	atomic.AddInt32(&m.pingCount, 1)
	return nil
}

func TestKeepAlive_ExternalAndDBPing(t *testing.T) {
	var httpPingCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			atomic.AddInt32(&httpPingCount, 1)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	mockDB := &mockDBPinger{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	Start(ctx, Config{
		ExternalURL: server.URL,
		Interval:    50 * time.Millisecond,
		DB:          mockDB,
		Client:      server.Client(),
	})

	// Wait for at least 2 ticks
	time.Sleep(140 * time.Millisecond)
	cancel()

	if atomic.LoadInt32(&httpPingCount) < 1 {
		t.Errorf("expected at least 1 HTTP ping, got %d", atomic.LoadInt32(&httpPingCount))
	}
	if atomic.LoadInt32(&mockDB.pingCount) < 1 {
		t.Errorf("expected at least 1 DB ping, got %d", atomic.LoadInt32(&mockDB.pingCount))
	}
}

func TestKeepAlive_NoTargetRunsGracefully(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	Start(ctx, Config{
		Interval: 20 * time.Millisecond,
	})
	time.Sleep(50 * time.Millisecond)
	cancel()
}
