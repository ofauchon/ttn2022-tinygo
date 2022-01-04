//go:build nucleowl55jc
// +build nucleowl55jc

/*
 Nucleo WL55JC1
 RFSwitch

				  Tx_HP   Tx_LP   RX
 FE_CTRL1 (PC4)    OFF     ON     ON
 FE_CTRL2 (PC5)    ON      ON     OFF
 FE_CTRL3 (PC3)    ON      ON     ON
*/
package radio

import (
	"machine"

	"tinygo.org/x/drivers/sx126x"
)

type CustomSwitch struct {
}

func (s CustomSwitch) InitRFSwitch() {
	machine.PC4.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.PC5.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.PC3.Configure(machine.PinConfig{Mode: machine.PinOutput})
}

func (s CustomSwitch) SetRfSwitchMode(mode int) error {
	switch mode {

	case sx126x.RFSWITCH_RX:
		machine.PC4.Set(true)
		machine.PC5.Set(false)
		machine.PC3.Set(true)

	case sx126x.RFSWITCH_TX_LP:
		machine.PC4.Set(true)
		machine.PC5.Set(true)
		machine.PC3.Set(true)

	case sx126x.RFSWITCH_TX_HP:
		machine.PC4.Set(false)
		machine.PC5.Set(true)
		machine.PC3.Set(true)
	}
	return nil

}
