package device

import (
	"errors"
	"fmt"
)

type Pwm struct {
	pca        *PCA9685
	pin        int
	last_value float32
}

func (p *PCA9685) NewPwm(pin int) *Pwm {

	p.log.Info(fmt.Sprintf("Creating a new Pwm controller at pin %v from %v", pin, p.name))

	return &Pwm{
		pca:        p,
		pin:        pin,
		last_value: 0.0,
	}
}

func (pwm *Pwm) SetPercentage(percentage float32) error {

	if percentage < 0.0 || percentage > 100.0 {
		return errors.New(fmt.Sprintf("Percentage must be between 0.0 and 100.0. Got %v.", percentage))
	}

	pwm.pca.log.Info(fmt.Sprintf("Setting pwm #%v to %v%% at \"%v\" device.", pwm.pin, percentage, pwm.pca.name))

	pulseRange := pwm.pca.maxPulse - pwm.pca.minPulse

	value := ((float32(percentage) * float32(pulseRange)) / 100.0) + float32(pwm.pca.minPulse)


	pwm.pca.setServoPulse(pwm.pin, int(value))

	return nil
}

func (pwm *Pwm) SetPulse(on int, off int) error {
	if on < 0 || on > off || off > 4096 {
		return errors.New(fmt.Sprintf(
			"On/Off (%d/%d) must be between 0 and %d",
			on,
			off,
			STEP_COUNT,
		))
	}

	pwm.pca.setPwm(pwm.pin, on, off)

	return nil
}