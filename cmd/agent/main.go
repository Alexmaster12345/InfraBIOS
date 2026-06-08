package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Alexmaster12345/infrabios/internal/agent"
)

func main() {
	_ = godotenv.Load()

	log := buildLogger(getEnv("INFRABIOS_LOG_LEVEL", "info"))
	defer log.Sync() //nolint:errcheck

	serverURL := getEnv("INFRABIOS_SERVER_URL", "http://localhost:8080")
	agentToken := getEnv("INFRABIOS_AGENT_TOKEN", "agent-changeme")
	serverID := getEnv("INFRABIOS_SERVER_ID", "")
	intervalStr := getEnv("INFRABIOS_COLLECT_INTERVAL", "60s")
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		interval = 60 * time.Second
	}

	hw := agent.CollectHardware()
	log.Info("hardware detected",
		zap.String("vendor", hw.Vendor),
		zap.String("model", hw.Model),
		zap.String("bios", hw.BIOSVersion),
	)

	client := agent.NewClient(serverURL, agentToken, serverID)

	if serverID == "" {
		if err := client.Register(hw); err != nil {
			log.Error("failed to register server", zap.Error(err))
		}
	}

	worker := agent.NewWorker(client, serverID, hw.Vendor, interval, log)

	stop := make(chan struct{})
	go worker.Run(stop)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	close(stop)
	log.Info("agent stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func buildLogger(level string) *zap.Logger {
	lvl := zapcore.InfoLevel
	_ = lvl.UnmarshalText([]byte(level))
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	l, _ := cfg.Build()
	return l
}
