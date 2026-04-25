package sensors

import (
	"fmt"
	"math/rand"
	"time"
)

// SensorReading represents a single data point from a sensor.
type SensorReading struct {
	SensorID  string    `json:"sensor_id"`
	Type      string    `json:"type"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
}

// SensorDriver provides an interface for reading sensor data.
type SensorDriver interface {
	Read() SensorReading
	Type() string
}

// TemperatureSensor reads ambient temperature.
type TemperatureSensor struct {
	ID       string
	baseTemp float64
}

func NewTemperatureSensor(id string) *TemperatureSensor {
	return &TemperatureSensor{ID: id, baseTemp: 22.0}
}

func (s *TemperatureSensor) Read() SensorReading {
	return SensorReading{
		SensorID:  s.ID,
		Type:      "TEMPERATURE",
		Value:     s.baseTemp + rand.Float64()*8.0 - 4.0,
		Unit:      "C",
		Timestamp: time.Now().UTC(),
	}
}

func (s *TemperatureSensor) Type() string { return "TEMPERATURE" }

// HumiditySensor reads relative humidity.
type HumiditySensor struct {
	ID        string
	baseLevel float64
}

func NewHumiditySensor(id string) *HumiditySensor {
	return &HumiditySensor{ID: id, baseLevel: 55.0}
}

func (s *HumiditySensor) Read() SensorReading {
	return SensorReading{
		SensorID:  s.ID,
		Type:      "HUMIDITY",
		Value:     s.baseLevel + rand.Float64()*20.0 - 10.0,
		Unit:      "%",
		Timestamp: time.Now().UTC(),
	}
}

func (s *HumiditySensor) Type() string { return "HUMIDITY" }

// MotionSensor detects motion events.
type MotionSensor struct {
	ID       string
	IsActive bool
}

func NewMotionSensor(id string) *MotionSensor {
	return &MotionSensor{ID: id}
}

func (s *MotionSensor) Read() SensorReading {
	detected := rand.Float64() > 0.8
	return SensorReading{
		SensorID:  s.ID,
		Type:      "MOTION",
		Value:     boolToFloat(detected),
		Unit:      "binary",
		Timestamp: time.Now().UTC(),
	}
}

func (s *MotionSensor) Type() string { return "MOTION" }

// DoorSensor reads door open/close state.
type DoorSensor struct {
	ID       string
	IsOpen   bool
}

func NewDoorSensor(id string) *DoorSensor {
	return &DoorSensor{ID: id}
}

func (s *DoorSensor) Read() SensorReading {
	open := rand.Float64() > 0.7
	return SensorReading{
		SensorID:  s.ID,
		Type:      "DOOR",
		Value:     boolToFloat(open),
		Unit:      "binary",
		Timestamp: time.Now().UTC(),
	}
}

func (s *DoorSensor) Type() string { return "DOOR" }

// SensorHub manages multiple sensor drivers.
type SensorHub struct {
	sensors []SensorDriver
}

func NewSensorHub(sensors ...SensorDriver) *SensorHub {
	return &SensorHub{sensors: sensors}
}

func (h *SensorHub) ReadAll() []SensorReading {
	readings := make([]SensorReading, 0, len(h.sensors))
	for _, s := range h.sensors {
		readings = append(readings, s.Read())
	}
	return readings
}

func (h *SensorHub) AddSensor(s SensorDriver) {
	h.sensors = append(h.sensors, s)
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

// Ensure fmt is used
var _ = fmt.Sprintf
