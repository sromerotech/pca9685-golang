package servo

import (
	"github.com/op/go-logging"
	"golang.org/x/exp/io/i2c"
	"log"
	"github.com/sergiorb/pca9685-golang/device"
	"time"
)

const (
	I2C_ADDR       = "/dev/i2c-1"
	ADDR_01    	   = 0x40

	FIRST_SERVO    = 0
	SECOND_SERVO   = 1
	MIN_PULSE 	   = 150
	MAX_PULSE 	   = 650
)

func main() {
	logger := logging.Logger{}
	dev, err := i2c.Open(&i2c.Devfs{Dev: I2C_ADDR}, ADDR_01)
	if err != nil {
		log.Fatal(err)
	}

	pca := device.NewPCA9685(dev, "Servo Controller", MIN_PULSE, MAX_PULSE, &logger)
	pca.Frequency = 50.0
	pca.Init()

	firstServo := pca.NewPwm(FIRST_SERVO)
	//secondServo := pca.NewPwm(SECOND_SERVO)

	for i:=0; i<100; i++ {
		firstServo.SetPercentage(float32(i))
		time.Sleep(time.Millisecond * 10)
	}
}