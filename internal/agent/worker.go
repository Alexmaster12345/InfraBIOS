package agent

import (
	"encoding/json"
	"time"

	"go.uber.org/zap"
)

const agentVersion = "1.0.0"

// Worker runs the agent polling loop.
type Worker struct {
	client   *Client
	serverID string
	vendor   string
	interval time.Duration
	log      *zap.Logger
}

func NewWorker(client *Client, serverID, vendor string, interval time.Duration, log *zap.Logger) *Worker {
	return &Worker{
		client:   client,
		serverID: serverID,
		vendor:   vendor,
		interval: interval,
		log:      log,
	}
}

// Run starts the collection loop. Blocks until the channel is closed.
func (w *Worker) Run(stop <-chan struct{}) {
	w.log.Info("agent worker started",
		zap.String("server_id", w.serverID),
		zap.Duration("interval", w.interval),
	)
	w.collect() // immediate first collection
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.collect()
		case <-stop:
			w.log.Info("agent worker stopped")
			return
		}
	}
}

func (w *Worker) collect() {
	settings, err := CollectBIOSSettings(w.vendor)
	if err != nil {
		w.log.Error("collect BIOS settings", zap.Error(err))
		return
	}

	firmware := buildFirmwarePayload()

	if err := w.client.SubmitInventory(settings, firmware, agentVersion); err != nil {
		w.log.Error("submit inventory", zap.Error(err))
		return
	}
	w.log.Info("inventory submitted")
}

func buildFirmwarePayload() json.RawMessage {
	hw := CollectHardware()
	payload, _ := json.Marshal(map[string]string{
		"bios": hw.BIOSVersion,
		"bmc":  hw.BMCVersion,
	})
	return payload
}
