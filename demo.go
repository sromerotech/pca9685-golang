package main

import (
	"github.com/op/go-logging"
	"github.com/sergiorb/pcm9685-golang/device"
	"golang.org/x/exp/io/i2c"
	"os"
	"time"
)

const (
	I2C_ADDR  = "/dev/i2c-1"
	ADDR_01   = 0x40
	MIN_PULSE = 0
	MAX_PULSE = 1000
)

func init() {

	stderrorLog := logging.NewLogBackend(os.Stderr, "", 0)

	stderrorLogLeveled := logging.AddModuleLevel(stderrorLog)
	stderrorLogLeveled.SetLevel(logging.INFO, "")

	logging.SetBackend(stderrorLogLeveled)
}

func main() {

	var mainLog = logging.MustGetLogger("PCA9685 Demo")

	i2cDevice, err := i2c.Open(&i2c.Devfs{Dev: I2C_ADDR}, ADDR_01)
	defer i2cDevice.Close()

	if err != nil {

		mainLog.Error(err)

	} else {

		var deviceLog = logging.MustGetLogger("PCA9685")

		pcm9685 := device.NewPCA9685(i2cDevice, "PWM Controller", MIN_PULSE, MAX_PULSE, deviceLog)

		pcm9685.Init()

		pcm9685.Demo([]int{0, 1, 2})

		pwm00 := pcm9685.NewPwm(0)
		pwm01 := pcm9685.NewPwm(1)
		pwm02 := pcm9685.NewPwm(2)

		_ = pwm00.SetPercentage(15.0)

		_ = pwm01.SetPercentage(50.0)

		_ = pwm02.SetPercentage(100.0)

		time.Sleep(2 * time.Second)

		pcm9685.SwichOff([]int{0, 1, 2})
	}
}
