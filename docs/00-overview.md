# Overview

This service is the software support for the INA226 current sensor. It can be mounted on the Rover and attached to the available I2C bus. The sensor's shunt resistor must be connected at the battery's connection terminals to measure the following in real-time:
* Current Draw (amps)
* Power Consumption (watts)
* Supply Voltage (vols)

The service publishes the following output message:
``` go
outputMsg := pb_outputs.SensorOutput{
    Timestamp: uint64(time.Now().UnixMilli()),
    Status:    0,
    SensorId:  1,
    SensorOutput: &pb_outputs.SensorOutput_EnergyOutput{
        EnergyOutput: &pb_outputs.EnergySensorOutput{
            CurrentAmps:   float32(data.CurrentAmps),
            SupplyVoltage: float32(data.SupplyVoltage),
            PowerWatts:    float32(data.PowerWatts),
        },
    },
}
```

By default, the service outputs 5 measurements each second, however this can be adjusted in the service.yaml under the configuration option `updates-per-second`.

