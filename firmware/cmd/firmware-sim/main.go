package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sentinel-device-manager/firmware/internal/hal"
	"github.com/sentinel-device-manager/firmware/internal/network"
	"github.com/sentinel-device-manager/firmware/internal/sensors"
)

func main() {
	logger := log.New(os.Stdout, "[firmware-sim] ", log.LstdFlags|log.Lshortfile)

	serial := envOrDefault("DEVICE_SERIAL", "VKD-CAM-SIM-001")
	mac := envOrDefault("DEVICE_MAC", "AA:BB:CC:DD:EE:FF")
	model := envOrDefault("DEVICE_MODEL", "D30-SIM")
	apiURL := envOrDefault("API_URL", "http://localhost:8080")

	// Initialize HAL
	h := hal.NewLinuxHAL(serial, mac, model)

	// Initialize sensors
	hub := sensors.NewSensorHub(
		sensors.NewTemperatureSensor("temp-01"),
		sensors.NewHumiditySensor("hum-01"),
		sensors.NewMotionSensor("mot-01"),
		sensors.NewDoorSensor("door-01"),
	)

	// Initialize network client
	client := network.NewClient(apiURL, "")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Register device
	logger.Printf("Registering device %s...", serial)
	regResp, err := client.Register(ctx, map[string]interface{}{
		"serial_number":   h.GetSerialNumber(),
		"device_type":     "CAMERA",
		"model":           h.GetModelName(),
		"site_id":         "site-simulation",
		"organization_id": "org-simulation",
		"ip_address":      "192.168.1.200",
		"mac_address":     h.GetMACAddress(),
		"config":          map[string]string{"resolution": "4K", "retention_days": "30"},
	})
	if err != nil {
		logger.Printf("Registration failed (will continue in standalone mode): %v", err)
	} else {
		if deviceID, ok := regResp["device_id"].(string); ok {
			client = network.NewClient(apiURL, deviceID)
			logger.Printf("Registered with device_id: %s", deviceID)
		}
	}

	// Main loop: send heartbeats and sensor readings
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Printf("Firmware simulator running. Ctrl+C to stop.")

	for {
		select {
		case <-ticker.C:
			// Read sensors
			readings := hub.ReadAll()
			for _, r := range readings {
				logger.Printf("Sensor %s (%s): %.2f %s", r.SensorID, r.Type, r.Value, r.Unit)
			}

			// Send heartbeat
			health := map[string]interface{}{
				"device_id":          client.DeviceID(),
				"cpu_usage":          h.ReadCPUUsage(),
				"memory_usage":       h.ReadMemoryUsage(),
				"temperature_c":      h.ReadTemperature(),
				"uptime_seconds":     h.GetUptimeSeconds(),
				"network_latency_ms": h.ReadNetworkLatency(),
			}
			if err := client.SendHeartbeat(ctx, health); err != nil {
				logger.Printf("Heartbeat failed: %v", err)
			}

			// Randomly generate events
			if rand.Float64() > 0.7 {
				eventTypes := []string{"MOTION_DETECTED", "DOOR_OPENED", "TEMPERATURE_THRESHOLD", "TAMPER_DETECTED"}
				evt := eventTypes[rand.Intn(len(eventTypes))]
				_, err := client.SendEvent(ctx, map[string]interface{}{
					"device_id":  "sim",
					"event_type": evt,
					"severity":   randomSeverity(),
					"payload":    map[string]string{"source": "firmware-sim"},
				})
				if err != nil {
					logger.Printf("Event send failed: %v", err)
				} else {
					logger.Printf("Sent event: %s", evt)
				}
			}

		case sig := <-sigChan:
			logger.Printf("Received signal %v, shutting down...", sig)
			fmt.Println("\nFirmware simulator stopped.")
			return
		}
	}
}

func randomSeverity() string {
	severities := []string{"INFO", "INFO", "INFO", "WARNING", "WARNING", "CRITICAL"}
	return severities[rand.Intn(len(severities))]
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
