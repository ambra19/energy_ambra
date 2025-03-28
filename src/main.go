package main

import (
	"fmt"
	"os"
	"time"

	pb_outputs "github.com/VU-ASE/rovercom/packages/go/outputs"
	roverlib "github.com/VU-ASE/roverlib-go/src"

	"github.com/rs/zerolog/log"
)

// The main user space program
// this program has all you need from roverlib: service identity, reading, writing and configuration
func run(service roverlib.Service, configuration *roverlib.ServiceConfiguration) error {
	//
	// Get configuration values
	//
	if configuration == nil {
		return fmt.Errorf("Configuration cannot be accessed")
	}

	//
	// Access the service identity, who am I?
	//
	log.Info().Msgf("Hello world, a new Go service '%s' was born at version %s", *service.Name, *service.Version)

	//
	// Access the service configuration, to use runtime parameters
	//
	tunableSpeed, err := configuration.GetFloatSafe("speed")
	if err != nil {
		return fmt.Errorf("Failed to get configuration: %v", err)
	}
	log.Info().Msgf("Fetched runtime configuration example tunable number: %f", tunableSpeed)

	//
	// Reading from an input, to get data from other services (see service.yaml to understand the input name)
	//
	readStream := service.GetReadStream("imaging", "path")
	if readStream == nil {
		return fmt.Errorf("Failed to get read stream")
	}

	//
	// Writing to an output that other services can read (see service.yaml to understand the output name)
	//
	writeStream := service.GetWriteStream("decision")
	if writeStream == nil {
		return fmt.Errorf("Failed to create write stream 'decision'")
	}

	for {
		//
		// Reading one message from the stream
		//
		data, err := readStream.Read()
		if data == nil || err != nil {
			return fmt.Errorf("Failed to read from 'imaging' service")
		}

		// When did the imaging service create this message?
		createdAt := data.Timestamp
		log.Info().Msgf("Recieved message with timestamp: %d", createdAt)

		// Get the imaging data
		imagingData := data.GetCameraOutput()
		if imagingData == nil {
			return fmt.Errorf("Message does not contain camera output. What did imaging do??")
		}
		log.Info().Msgf("Imaging service captured a %d by %d image", imagingData.Trajectory.Width, imagingData.Trajectory.Height)

		// Print the X and Y coordinates of the middle point of the track that Imaging has detected
		if len(imagingData.Trajectory.Points) > 0 {
			log.Info().Msgf("The X: %d and Y: %d values of the middle point of the track", imagingData.Trajectory.Points[0].X, imagingData.Trajectory.Points[0].Y)
		} else {
			log.Info().Msgf("imaging could didn't detect track edges. Is the Rover on the track?")
		}

		// This value holds the steering position that we want to pass to the servo (-1 = left, 0 = center, 1 = right)
		steerPosition := float32(-0.5)

		// Initialize the message that we want to send to the actuator
		actuatorMsg := pb_outputs.SensorOutput{
			Timestamp: uint64(time.Now().UnixMilli()), // milliseconds since epoch
			Status:    0,                              // all is well
			SensorId:  1,                              
			SensorOutput: &pb_outputs.SensorOutput_ControllerOutput{
				ControllerOutput: &pb_outputs.ControllerOutput{
					SteeringAngle: steerPosition,
					LeftThrottle:  float32(tunableSpeed),
					RightThrottle: float32(tunableSpeed),
					FanSpeed:      0,
					FrontLights:   false,
				},
			},
		}

		// Send the message to the actuator
		err = writeStream.Write(&actuatorMsg)
		if err != nil {
			log.Warn().Err(err).Msg("Could not write to actuator")
		}

		//
		// Now do something else fun, see if our "tunable_speed" is updated
		//
		curr := tunableSpeed

		log.Info().Msg("Checking for tunable number update")

		// We are not using the safe version here, because using locks is boring
		// (this is perfectly fine if you are constantly polling the value)
		// nb: this is not a blocking call, it will return the last known value
		newVal, err := configuration.GetFloat("speed")
		if err != nil {
			return fmt.Errorf("Failed to get updated tunable number: %v", err)
		}

		if curr != newVal {
			log.Info().Msgf("Tunable number updated: %f -> %f", curr, newVal)
			curr = newVal
		}
		tunableSpeed = curr
	}
}

// This function gets called when roverd wants to terminate the service
func onTerminate(sig os.Signal) error {
	log.Info().Str("signal", sig.String()).Msg("Terminating service")

	//
	// ...
	// Any clean up logic here
	// ...
	//

	return nil
}

// This is just a wrapper to run the user program
// it is not recommended to put any other logic here
func main() {
	roverlib.Run(run, onTerminate)
}
