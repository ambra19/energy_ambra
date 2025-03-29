package main

import (
	"fmt"
	"os"
	"time"

	roverlib "github.com/VU-ASE/roverlib-go/src"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"

	"github.com/rs/zerolog/log"
)

const (
	// Device address
	ina226Address = 0x40

	// Register addresses
	configReg      = 0x00
	shuntVoltReg   = 0x01
	busVoltReg     = 0x02
	powerReg       = 0x03
	currentReg     = 0x04
	calibrationReg = 0x05

	// Configuration values
	configValue = 0x4127 // Default configuration

	// Conversion factors
	busVoltageConversion = 1.25 / 1000.0 // 1.25 mV/bit
	currentLSB           = 0.001         // 1 mA/bit (adjust based on your calibration)
	powerLSB             = 25.0 * 0.001  // 25 * currentLSB (25 mW/bit)
)

type INA226 struct {
	dev i2c.Dev
}

func NewINA226(bus i2c.BusCloser) (*INA226, error) {
	ina := &INA226{
		dev: i2c.Dev{Bus: bus, Addr: ina226Address},
	}

	// Initialize device
	if err := ina.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize INA226: %v", err)
	}

	return ina, nil
}

func (ina *INA226) initialize() error {
	// Set configuration register
	if err := ina.writeRegister(configReg, configValue); err != nil {
		return err
	}

	// Set calibration register (2560 or 0xA00 for a 2mΩ shunt resistor)
	// This value should be calculated based on your specific shunt resistor
	if err := ina.writeRegister(calibrationReg, 2560); err != nil {
		return err
	}

	return nil
}

func (ina *INA226) writeRegister(reg uint8, value uint16) error {
	// Convert value to big-endian bytes
	data := []byte{reg, byte(value >> 8), byte(value & 0xFF)}
	return ina.dev.Tx(data, nil)
}

func (ina *INA226) readRegister(reg uint8) (uint16, error) {
	// Write register address
	if err := ina.dev.Tx([]byte{reg}, nil); err != nil {
		return 0, err
	}

	// Read register value (2 bytes)
	data := make([]byte, 2)
	if err := ina.dev.Tx(nil, data); err != nil {
		return 0, err
	}

	// Convert from big-endian
	return uint16(data[0])<<8 | uint16(data[1]), nil
}

func (ina *INA226) ReadBusVoltage() (float64, error) {
	raw, err := ina.readRegister(busVoltReg)
	if err != nil {
		return 0, err
	}
	return float64(raw) * busVoltageConversion, nil
}

func (ina *INA226) ReadCurrent() (float64, error) {
	raw, err := ina.readRegister(currentReg)
	if err != nil {
		return 0, err
	}
	// Check if value is negative (two's complement)
	value := int16(raw)
	return float64(value) * currentLSB, nil
}

func (ina *INA226) ReadPower() (float64, error) {
	raw, err := ina.readRegister(powerReg)
	if err != nil {
		return 0, err
	}
	return float64(raw) * powerLSB, nil
}

type CurrentSensorOutput struct {
	SupplyVoltage float64
	CurrentAmps   float64
	PowerWatts    float64
}

func (ina *INA226) ReadSensorData() (*CurrentSensorOutput, error) {
	// Read bus voltage
	voltage, err := ina.ReadBusVoltage()
	if err != nil {
		return nil, fmt.Errorf("failed to read bus voltage: %v", err)
	}

	// Read current
	current, err := ina.ReadCurrent()
	if err != nil {
		return nil, fmt.Errorf("failed to read current: %v", err)
	}

	// Read power
	power, err := ina.ReadPower()
	if err != nil {
		return nil, fmt.Errorf("failed to read power: %v", err)
	}

	return &CurrentSensorOutput{
		SupplyVoltage: voltage,
		CurrentAmps:   current,
		PowerWatts:    power,
	}, nil
}

func run(service roverlib.Service, configuration *roverlib.ServiceConfiguration) error {

	// From the service.yaml, read the configuration value for the update-frequency
	// of the service.
	if configuration == nil {
		return fmt.Errorf("configuration cannot be accessed")
	}
	updateFrequency, err := configuration.GetFloat("updates-per-second")
	if err == nil {
		return fmt.Errorf("unable to read configuration")
	}


	// Initialize periph.io
	if _, err := host.Init(); err != nil {
		log.Error().Msgf("failed to initialize periph: %v", err)
	}

	// Open I2C bus
	bus, err := i2creg.Open("5")
	if err != nil {
		log.Error().Msgf("failed to open I2C bus: %v", err)
	}
	defer bus.Close()

	// Create a new INA226 instance
	ina226, err := NewINA226(bus)
	if err != nil {
		log.Error().Msgf("%v", err)
	}

	for {
		time.Sleep((1.0 * time.Second) / time.Duration(updateFrequency))

		// Refetch to make it possible to tune
		updateFrequency, err = configuration.GetFloat("updates-per-second")
		if err == nil {
			return fmt.Errorf("unable to read configuration")
		}
		
		// Read sensor data
		data, err := ina226.ReadSensorData()
		if err != nil {
			log.Error().Msgf("Failed to read sensor data: %v", err)
		}
		// Currently just printing the results.
		// TODO: output current sensor definition here
		log.Info().Msgf("Bus Voltage: %.3f V\n", data.SupplyVoltage)
		log.Info().Msgf("Current: %.3f A\n", data.CurrentAmps)
		log.Info().Msgf("Power: %.3f W\n", data.PowerWatts)
	}
}

// When the service is stopped externally, this function is called.
// Currently, there are no clean up routines.
func onTerminate(sig os.Signal) error {
	log.Info().Str("signal", sig.String()).Msg("Terminating service")
	return nil
}

// Entry point of the program
func main() {
	roverlib.Run(run, onTerminate)
}
