package sensors

import (
	"testing"
)

func TestTemperatureSensor_Read(t *testing.T) {
	s := NewTemperatureSensor("temp-01")
	for i := 0; i < 100; i++ {
		r := s.Read()
		if r.SensorID != "temp-01" {
			t.Errorf("expected sensor_id temp-01, got %s", r.SensorID)
		}
		if r.Type != "TEMPERATURE" {
			t.Errorf("expected type TEMPERATURE, got %s", r.Type)
		}
		if r.Unit != "C" {
			t.Errorf("expected unit C, got %s", r.Unit)
		}
		if r.Value < 18.0 || r.Value > 26.0 {
			t.Errorf("temperature %.2f out of range [18, 26]", r.Value)
		}
	}
}

func TestTemperatureSensor_Type(t *testing.T) {
	s := NewTemperatureSensor("temp-01")
	if s.Type() != "TEMPERATURE" {
		t.Errorf("expected TEMPERATURE, got %s", s.Type())
	}
}

func TestHumiditySensor_Read(t *testing.T) {
	s := NewHumiditySensor("hum-01")
	for i := 0; i < 100; i++ {
		r := s.Read()
		if r.Type != "HUMIDITY" {
			t.Errorf("expected HUMIDITY, got %s", r.Type)
		}
		if r.Unit != "%" {
			t.Errorf("expected unit %%, got %s", r.Unit)
		}
		if r.Value < 45.0 || r.Value > 65.0 {
			t.Errorf("humidity %.2f out of range [45, 65]", r.Value)
		}
	}
}

func TestMotionSensor_Read(t *testing.T) {
	s := NewMotionSensor("mot-01")
	r := s.Read()
	if r.Type != "MOTION" {
		t.Errorf("expected MOTION, got %s", r.Type)
	}
	if r.Value != 0.0 && r.Value != 1.0 {
		t.Errorf("motion value should be 0 or 1, got %.2f", r.Value)
	}
}

func TestDoorSensor_Read(t *testing.T) {
	s := NewDoorSensor("door-01")
	r := s.Read()
	if r.Type != "DOOR" {
		t.Errorf("expected DOOR, got %s", r.Type)
	}
	if r.Value != 0.0 && r.Value != 1.0 {
		t.Errorf("door value should be 0 or 1, got %.2f", r.Value)
	}
}

func TestSensorHub_ReadAll(t *testing.T) {
	hub := NewSensorHub(
		NewTemperatureSensor("temp-01"),
		NewHumiditySensor("hum-01"),
		NewMotionSensor("mot-01"),
		NewDoorSensor("door-01"),
	)

	readings := hub.ReadAll()
	if len(readings) != 4 {
		t.Fatalf("expected 4 readings, got %d", len(readings))
	}

	types := map[string]bool{}
	for _, r := range readings {
		types[r.Type] = true
	}
	if len(types) != 4 {
		t.Errorf("expected 4 distinct types, got %d", len(types))
	}
}

func TestSensorHub_AddSensor(t *testing.T) {
	hub := NewSensorHub()
	if len(hub.ReadAll()) != 0 {
		t.Error("expected 0 readings from empty hub")
	}
	hub.AddSensor(NewTemperatureSensor("temp-01"))
	if len(hub.ReadAll()) != 1 {
		t.Error("expected 1 reading after add")
	}
}

func TestSensorDriver_Interface(t *testing.T) {
	var _ SensorDriver = NewTemperatureSensor("t")
	var _ SensorDriver = NewHumiditySensor("h")
	var _ SensorDriver = NewMotionSensor("m")
	var _ SensorDriver = NewDoorSensor("d")
}

func TestSensorReadings_HaveTimestamp(t *testing.T) {
	sensors := []SensorDriver{
		NewTemperatureSensor("temp-01"),
		NewHumiditySensor("hum-01"),
		NewMotionSensor("mot-01"),
		NewDoorSensor("door-01"),
	}
	for _, s := range sensors {
		r := s.Read()
		if r.Timestamp.IsZero() {
			t.Errorf("sensor %s reading has zero timestamp", r.SensorID)
		}
	}
}

func TestSensorReadings_HaveSensorID(t *testing.T) {
	ids := []string{"temp-01", "hum-01", "mot-01", "door-01"}
	drivers := []SensorDriver{
		NewTemperatureSensor(ids[0]),
		NewHumiditySensor(ids[1]),
		NewMotionSensor(ids[2]),
		NewDoorSensor(ids[3]),
	}
	for i, s := range drivers {
		r := s.Read()
		if r.SensorID != ids[i] {
			t.Errorf("expected sensor_id %s, got %s", ids[i], r.SensorID)
		}
	}
}
