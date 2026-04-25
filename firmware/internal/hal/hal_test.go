package hal

import (
	"testing"
)

func TestLinuxHAL_ReadTemperature(t *testing.T) {
	h := NewLinuxHAL("VKD-TEST-001", "AA:BB:CC:DD:EE:FF", "D30")
	for i := 0; i < 100; i++ {
		temp := h.ReadTemperature()
		if temp < 45.0 || temp > 60.0 {
			t.Errorf("temperature %.2f out of expected range [45, 60]", temp)
		}
	}
}

func TestLinuxHAL_ReadCPUUsage(t *testing.T) {
	h := NewLinuxHAL("VKD-TEST-001", "AA:BB:CC:DD:EE:FF", "D30")
	for i := 0; i < 100; i++ {
		cpu := h.ReadCPUUsage()
		if cpu < 10.0 || cpu > 90.0 {
			t.Errorf("cpu usage %.2f out of expected range [10, 90]", cpu)
		}
	}
}

func TestLinuxHAL_ReadMemoryUsage(t *testing.T) {
	h := NewLinuxHAL("VKD-TEST-001", "AA:BB:CC:DD:EE:FF", "D30")
	for i := 0; i < 100; i++ {
		mem := h.ReadMemoryUsage()
		if mem < 30.0 || mem > 80.0 {
			t.Errorf("memory usage %.2f out of expected range [30, 80]", mem)
		}
	}
}

func TestLinuxHAL_ReadNetworkLatency(t *testing.T) {
	h := NewLinuxHAL("VKD-TEST-001", "AA:BB:CC:DD:EE:FF", "D30")
	for i := 0; i < 100; i++ {
		lat := h.ReadNetworkLatency()
		if lat < 1.0 || lat > 21.0 {
			t.Errorf("network latency %.2f out of expected range [1, 21]", lat)
		}
	}
}

func TestLinuxHAL_GetSerialNumber(t *testing.T) {
	h := NewLinuxHAL("VKD-TEST-001", "AA:BB:CC:DD:EE:FF", "D30")
	if h.GetSerialNumber() != "VKD-TEST-001" {
		t.Errorf("expected serial VKD-TEST-001, got %s", h.GetSerialNumber())
	}
}

func TestLinuxHAL_GetMACAddress(t *testing.T) {
	h := NewLinuxHAL("VKD-TEST-001", "AA:BB:CC:DD:EE:FF", "D30")
	if h.GetMACAddress() != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("expected MAC AA:BB:CC:DD:EE:FF, got %s", h.GetMACAddress())
	}
}

func TestLinuxHAL_GetModelName(t *testing.T) {
	h := NewLinuxHAL("VKD-TEST-001", "AA:BB:CC:DD:EE:FF", "D30")
	if h.GetModelName() != "D30" {
		t.Errorf("expected model D30, got %s", h.GetModelName())
	}
}

func TestLinuxHAL_GetUptimeSeconds(t *testing.T) {
	h := NewLinuxHAL("VKD-TEST-001", "AA:BB:CC:DD:EE:FF", "D30")
	uptime := h.GetUptimeSeconds()
	if uptime < 86400 {
		t.Errorf("uptime %d should be at least 86400", uptime)
	}
}

func TestHAL_Interface(t *testing.T) {
	var _ HAL = NewLinuxHAL("s", "m", "mod")
}
