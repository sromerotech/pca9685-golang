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
	SLEEP         byte = 0x10
	BYTE          byte = 0xFF
)

type PCA9685 struct {
	i2cBus    *i2c.Device
	name      string
	initiated bool
	minPulse  int
	maxPulse  int
	log       *logging.Logger
	frequency float32
}

type Pwm struct {
	pca        *PCA9685
	pin        int
	last_value float32
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
		frequency: 1000.0,
	}
}

func (p *PCA9685) NewPwm(pin int) *Pwm {

	p.log.Info(fmt.Sprintf("Creating a new Pwm controler at pin %v from %v", pin, p.name))

	return &Pwm{
		pca:        p,
		pin:        pin,
		last_value: 0.0,
	}
}

func (p *PCA9685) Init() {

	if p.initiated {

		p.log.Warning(fmt.Sprintf("Device \"%v\" already initiated!", p.name))

	} else {

		p.log.Info(fmt.Sprintf("Initiating \"%v\" PCA9685 device", p.name))

		p.set_all_pwm(0, 0)
		p.i2cBus.WriteReg(MODE2, []byte{OUTDRV})

		time.Sleep(5 * time.Millisecond)

		var mode1 byte
		err := p.i2cBus.ReadReg(MODE1, []byte{mode1})

		if err != nil {

			p.log.Error("Can't read!")
			return
		}

		mode1 &= BYTE
		mode1 = mode1 & ^SLEEP

		p.i2cBus.WriteReg(MODE1, []byte{mode1 & 0xFF})

		time.Sleep(5 * time.Millisecond)

		p.set_pwm_freq(p.frequency)

		p.initiated = true
	}
}

func (p *PCA9685) SwichOn(pwm []int) error {

	if !p.initiated {

		return errors.New(fmt.Sprintf("Device \"%v\"is not initiated!", p.name))
	}

	for i := 0; i < len(pwm); i++ {

		p.log.Info(fmt.Sprintf("Swiching on pwm #%v", pwm[i]))

		p.set_servo_pulse(pwm[i], p.maxPulse)
	}

	return nil
}

func (p *PCA9685) SwichOff(pwm []int) error {

	if !p.initiated {

		return errors.New(fmt.Sprintf("Device \"%v\"is not initiated!", p.name))
	}

	for i := 0; i < len(pwm); i++ {

		p.log.Info(fmt.Sprintf("Swiching off pwm #%v", pwm[i]))

		p.set_servo_pulse(pwm[i], p.minPulse)
	}

	return nil
}

func (p *PCA9685) FadeInOut(pwmNumber int) error {

	if !p.initiated {

		return errors.New(fmt.Sprintf("Device \"%v\"is not initiated!", p.name))
	}

	p.log.Info(fmt.Sprintf("Fading pwm #%v from %v to %v pulse...", pwmNumber, p.minPulse, p.maxPulse))

	for i := p.minPulse; i < p.maxPulse; i++ {

		p.set_servo_pulse(pwmNumber, i)
	}

	for i := p.maxPulse; i > p.minPulse; i-- {

		p.set_servo_pulse(pwmNumber, i)
	}

	return nil
}

func (p *PCA9685) Wink(pwm []int, times int, speed int) {

	p.log.Info(fmt.Sprintf("Winking pwm's: %v, %v times at speed %vms", pwm, times, speed))

	for i := 0; i < times; i++ {

		p.SwichOn(pwm)

		var halfSpeed time.Duration = time.Duration(speed/2) * time.Millisecond

		time.Sleep(halfSpeed)

		p.SwichOff(pwm)

		time.Sleep(halfSpeed)
	}
}

func (p *PCA9685) Demo(pwm []int) {

	for i := 0; i < len(pwm); i++ {

		p.FadeInOut(pwm[i])
	}

	p.Wink(pwm, 4, 800)
}

func (p *PCA9685) set_pwm_freq(freqHz float32) {

	var prescaleValue float32 = 25000000.0 // 25MHz
	prescaleValue /= 4096.0
	prescaleValue /= freqHz
	prescaleValue -= 1.0

	p.log.Debug(fmt.Sprintf("Setting PWM frequency to %v Hz", freqHz))
	p.log.Debug(fmt.Sprintf("Esimated pre-scale: %v", prescaleValue))

	prescale := int(math.Floor(float64(prescaleValue + 0.5)))

	p.log.Debug(fmt.Sprintf("Final pre.scale: %v", prescale))

	var old_mode byte
	err := p.i2cBus.ReadReg(MODE1, []byte{old_mode})

	if err != nil {

		p.log.Error("Can't read!")
	}

	old_mode &= BYTE

	new_mode := (old_mode & 0x7F) | 0x10

	p.i2cBus.WriteReg(MODE1, []byte{new_mode & BYTE})
	p.i2cBus.WriteReg(PRESCALE, []byte{byte(prescale) & BYTE})
	p.i2cBus.WriteReg(MODE1, []byte{old_mode & BYTE})

	time.Sleep(5 * time.Millisecond)

	p.i2cBus.WriteReg(MODE1, []byte{old_mode&BYTE | 0x80})
}

func (p *PCA9685) set_servo_pulse(pwmNumber int, pulse int) {

	var pulse_length float32 = 1000000

	pulse_length /= float32(60)

	//p.log.Debug(fmt.Sprintf("%vus per period", pulse_length))

	pulse_length /= float32(4096)

	//p.log.Debug(fmt.Sprintf("%vus per bit", pulse_length))

	var pulseF float32 = float32(pulse)

	pulseF /= pulse_length

	p.log.Debug(fmt.Sprintf("Setting pwm #%v to pulse %v", pwmNumber, pulseF))

	p.set_pwm(pwmNumber, 0, int(pulseF))
}

func (p *PCA9685) set_all_pwm(on int, off int) {

	onB := byte(on) & BYTE
	offB := byte(off) & BYTE

	p.i2cBus.WriteReg(ALL_LED_ON_L, []byte{onB & BYTE})
	p.i2cBus.WriteReg(ALL_LED_ON_H, []byte{onB & BYTE})
	p.i2cBus.WriteReg(ALL_LED_OFF_L, []byte{offB & BYTE})
	p.i2cBus.WriteReg(ALL_LED_OFF_H, []byte{offB & BYTE})
}

func (p *PCA9685) set_pwm(pwm int, on int, off int) {

	onB := byte(on) & BYTE
	offB := byte(off) & BYTE

	p.i2cBus.WriteReg(LED0_ON_L+byte(4)*byte(pwm), []byte{onB & BYTE})
	p.i2cBus.WriteReg(LED0_ON_H+byte(4)*byte(pwm), []byte{onB >> 8})
	p.i2cBus.WriteReg(LED0_OFF_L+byte(4)*byte(pwm), []byte{offB & BYTE})
	p.i2cBus.WriteReg(LED0_OFF_H+byte(4)*byte(pwm), []byte{offB >> 8})
}

func (pwm *Pwm) SetPercentage(percentage float32) error {

	if percentage < 0.0 || percentage > 100.0 {

		return errors.New(fmt.Sprintf("Percetage must be between 0.0 and 100.0. Got %v.", percentage))
	}

	pwm.pca.log.Info(fmt.Sprintf("Setting pwm #%v to %v%% at \"%v\" device.", pwm.pin, percentage, pwm.pca.name))

	max := pwm.pca.maxPulse - pwm.pca.minPulse

	value := ((float32(percentage) * float32(max)) / 100.0) + float32(pwm.pca.minPulse)

	pwm.pca.set_servo_pulse(pwm.pin, int(value))

	return nil
}
