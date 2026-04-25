package hal

import "math/rand"

// HAL provides hardware abstraction for the embedded device runtime.
// In production, these functions interact with device-specific drivers.
type HAL interface {
	ReadTemperature() float64
	ReadCPUUsage() float64
	ReadMemoryUsage() float64
	ReadNetworkLatency() float64
	GetUptimeSeconds() int64
	GetSerialNumber() string
	GetMACAddress() string
	GetModelName() string
}

// LinuxHAL simulates hardware reads for a Linux-embedded device.
type LinuxHAL struct {
	SerialNumber string
	MACAddress   string
	Model        string
	baseTemp     float64
}

func NewLinuxHAL(serial, mac, model string) *LinuxHAL {
	return &LinuxHAL{
		SerialNumber: serial,
		MACAddress:   mac,
		Model:        model,
		baseTemp:     45.0,
	}
}

func (h *LinuxHAL) ReadTemperature() float64 {
	return h.baseTemp + rand.Float64()*15.0
}

func (h *LinuxHAL) ReadCPUUsage() float64 {
	return 10.0 + rand.Float64()*80.0
}

func (h *LinuxHAL) ReadMemoryUsage() float64 {
	return 30.0 + rand.Float64()*50.0
}

func (h *LinuxHAL) ReadNetworkLatency() float64 {
	return 1.0 + rand.Float64()*20.0
}

func (h *LinuxHAL) GetUptimeSeconds() int64 {
	return int64(86400 + rand.Intn(8640000))
}

func (h *LinuxHAL) GetSerialNumber() string {
	return h.SerialNumber
}

func (h *LinuxHAL) GetMACAddress() string {
	return h.MACAddress
}

func (h *LinuxHAL) GetModelName() string {
	return h.Model
}
