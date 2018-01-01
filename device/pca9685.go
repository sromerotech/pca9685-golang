package device

import (
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"golang.org/x/exp/io/i2c"
	"math"
	"time"
)

const (
	MODE1         byte = 0x00
	MODE2         byte = 0x01
	PRESCALE      byte = 0xFE
	LED0_ON_L     byte = 0x06
	LED0_ON_H     byte = 0x07
	LED0_OFF_L    byte = 0x08
	LED0_OFF_H    byte = 0x09
	ALL_LED_ON_L  byte = 0xFA
	ALL_LED_ON_H  byte = 0xFB
	ALL_LED_OFF_L byte = 0xFC
	ALL_LED_OFF_H byte = 0xFD
	OUTDRV        byte = 0x04
	ALLCALL		  byte = 0x01
	SLEEP         byte = 0x10
	BYTE          byte = 0xFF

	DEFAULT_FREQ  float32 = 1000.0
	OSC_FREQ	  float32 = 25000000.0
	STEP_COUNT    float32 = 4096.0
)

type PCA9685 struct {
	i2cBus    *i2c.Device
	name      string
	initiated bool
	minPulse  int
	maxPulse  int
	log       *logging.Logger
	Frequency float32
}

func NewPCA9685(i2cDevice *i2c.Device, name string, minPulse int, maxPulse int, log *logging.Logger) *PCA9685 {

	log.Info(fmt.Sprintf("Creating a new PCA9685 device. Alias: %v", name))

	return &PCA9685{
		i2cBus:    i2cDevice,
		name:      name,
		initiated: false,
		minPulse:  minPulse,
		maxPulse:  maxPulse,
		log:       log,
		Frequency: DEFAULT_FREQ,
	}
}

func (p *PCA9685) Init() {

	if p.initiated {

		p.log.Warning(fmt.Sprintf("Device \"%v\" already initiated!", p.name))

	} else {
		p.log.Info(fmt.Sprintf("Initiating \"%v\" PCA9685 device", p.name))

		p.SetAllPwm(0, 0)
		p.writeByte(MODE2, OUTDRV)
		p.writeByte(MODE1, ALLCALL)

		time.Sleep(5 * time.Millisecond)

		mode := p.readByte(MODE1)
		mode = mode & ^SLEEP
		p.writeByte(MODE1, mode)

		time.Sleep(5 * time.Millisecond)

		p.setPwmFrequency(p.Frequency)

		p.initiated = true
	}
}

func (p *PCA9685) SwitchOn(pwm []int) error {

	if !p.initiated {

		return errors.New(fmt.Sprintf("Device \"%v\"is not initiated!", p.name))
	}

	for i := 0; i < len(pwm); i++ {

		p.log.Info(fmt.Sprintf("Switching on pwm #%v", pwm[i]))

		p.setServoPulse(pwm[i], p.maxPulse)
	}

	return nil
}

func (p *PCA9685) SwitchOff(pwm []int) error {

	if !p.initiated {

		return errors.New(fmt.Sprintf("Device \"%v\"is not initiated!", p.name))
	}

	for i := 0; i < len(pwm); i++ {

		p.log.Info(fmt.Sprintf("Switching off pwm #%v", pwm[i]))

		p.setServoPulse(pwm[i], p.minPulse)
	}

	return nil
}

func (p *PCA9685) FadeInOut(pwmNumber int) error {

	if !p.initiated {

		return errors.New(fmt.Sprintf("Device \"%v\"is not initiated!", p.name))
	}

	p.log.Info(fmt.Sprintf("Fading pwm #%v from %v to %v pulse...", pwmNumber, p.minPulse, p.maxPulse))

	for i := p.minPulse; i < p.maxPulse; i++ {

		p.setServoPulse(pwmNumber, i)
	}

	for i := p.maxPulse; i > p.minPulse; i-- {

		p.setServoPulse(pwmNumber, i)
	}

	return nil
}

func (p *PCA9685) Wink(pwm []int, times int, speed int) {

	p.log.Info(fmt.Sprintf("Winking pwm's: %v, %v times at speed %vms", pwm, times, speed))

	for i := 0; i < times; i++ {

		p.SwitchOn(pwm)

		var halfSpeed time.Duration = time.Duration(speed/2) * time.Millisecond

		time.Sleep(halfSpeed)

		p.SwitchOff(pwm)

		time.Sleep(halfSpeed)
	}
}

func (p *PCA9685) Demo(pwm []int) {

	for i := 0; i < len(pwm); i++ {

		p.FadeInOut(pwm[i])
	}

	p.Wink(pwm, 4, 800)
}

func (p *PCA9685) setPwmFrequency(freqHz float32) {

	preScaleValue := OSC_FREQ // 25MHz
	preScaleValue /= STEP_COUNT
	preScaleValue /= freqHz
	preScaleValue -= 1.0

	p.log.Debug(fmt.Sprintf("Setting PWM frequency to %v Hz", freqHz))
	p.log.Debug(fmt.Sprintf("Estimated pre-scale: %v", preScaleValue))

	preScale := int(math.Floor(float64(preScaleValue + 0.5)))

	p.log.Debug(fmt.Sprintf("Final pre-scale: %v", preScale))

	oldMode := p.read8(MODE1)

	newMode := (oldMode & 0x7F) | 0x10

	p.write8(MODE1, newMode)
	p.write8(PRESCALE, preScale)
	p.write8(MODE1, oldMode)

	time.Sleep(5 * time.Millisecond)

	p.write8(MODE1, oldMode| 0x80)
}

func (p *PCA9685) setServoPulse(pwmNumber int, pulse int) {

	var pulseLength float32 = 1000000

	pulseLength /= float32(60)

	//p.log.Debug(fmt.Sprintf("%vus per period", pulseLength))

	pulseLength /= STEP_COUNT

	//p.log.Debug(fmt.Sprintf("%vus per bit", pulseLength))

	pulseF := float32(pulse)

	pulseF /= pulseLength

	p.log.Debug(fmt.Sprintf("Setting pwm #%v to pulse %v", pwmNumber, pulseF))

	p.setPwm(pwmNumber, 0, int(pulseF))
}

func (p *PCA9685) SetAllPwm(on int, off int) {
	p.write8(ALL_LED_ON_L, on)
	p.write8(ALL_LED_ON_H, on >> 8)
	p.write8(ALL_LED_OFF_L, off)
	p.write8(ALL_LED_OFF_H, off >> 8)
}

func (p *PCA9685) setPwm(pwm int, on int, off int) {
	p.write8(LED0_ON_L+byte(4)*byte(pwm), on)
	p.write8(LED0_ON_H+byte(4)*byte(pwm), on >> 8)
	p.write8(LED0_OFF_L+byte(4)*byte(pwm), off)
	p.write8(LED0_OFF_H+byte(4)*byte(pwm), off >> 8)
}

func (p *PCA9685) write8(reg byte, intVal int) {
	byteVal := byte(intVal) & BYTE

	p.writeByte(reg, byteVal)
}

func (p *PCA9685) writeByte(reg byte, byteVal byte) {
	err := p.i2cBus.WriteReg(reg, []byte{byteVal})
	if err != nil {
		p.log.Error(fmt.Sprintf("Failed to read from register %#x.", reg))
		p.log.Error(err.Error())
	}

	p.log.Debug(fmt.Sprintf("Wrote %#x to register %#x.", byteVal, reg))
}

func (p *PCA9685) read8(reg byte) int {
	byteVal := p.readByte(reg)

	return int(byteVal)
}

func (p *PCA9685) readByte(reg byte) byte {
	buf := make([]byte, 1)
	err := p.i2cBus.ReadReg(reg, buf)
	if err != nil {
		p.log.Error(fmt.Sprintf("Failed to read from register %#x.", reg))
		p.log.Error(err.Error())
	}

	p.log.Debug(fmt.Sprintf("Read %#x from register %#x.", buf[0], reg))

	return buf[0]
}
