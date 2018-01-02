package main

import (
	"github.com/op/go-logging"
	"golang.org/x/exp/io/i2c"
	"log"
	"github.com/ataboo/pca9685-golang/device"
	"time"
)

const (
	I2C_ADDR           = "/dev/i2c-1"
	ADDR_01            = 0x40

	SERVO_CHANNEL      = 15
	MIN_PULSE	   = 150
	MAX_PULSE	   = 650
)

func main() {
	logger := logging.Logger{}
	dev, err := i2c.Open(&i2c.Devfs{Dev: I2C_ADDR}, ADDR_01)
	if err != nil {
		log.Fatal(err)
	}

	pca := device.NewPCA9685(dev, "Servo Controller", MIN_PULSE, MAX_PULSE, &logger)
	pca.Frequency = 60.0
	pca.Init()

	servo := pca.NewPwm(SERVO_CHANNEL)

	setPercentage(servo, 100.0)
	time.Sleep(2 * time.Second)
	setPercentage(servo, 0.0)
	time.Sleep(2 * time.Second)

	for i:=0; i<1000; i++ {
		setPercentage(servo, float32(i) / 10)
		time.Sleep(time.Millisecond * 10)
	}

	pca.SetAllPwm(0, 0)
}

func setPercentage(p *device.Pwm, percent float32) {
	pulseLength := int((MAX_PULSE - MIN_PULSE) * percent / 100 + MIN_PULSE)

	p.SetPulse(0, pulseLength)
}

