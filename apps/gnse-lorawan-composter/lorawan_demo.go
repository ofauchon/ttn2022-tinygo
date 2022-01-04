package main

// In this example, we will join TTN Network and send a sample packet every 180s

import (
	"device/stm32"
	"machine"
	"runtime/interrupt"
	"time"

	"encoding/hex"

	extra "./extra"
	rfswitch "./rfswitch"
	cayennelpp "github.com/TheThingsNetwork/go-cayenne-lib"
	"github.com/ofauchon/go-lorawan-stack"
	"tinygo.org/x/drivers/shtc3"
	"tinygo.org/x/drivers/sx126x"
)

var (
	loraRadio *sx126x.Device
	loraStack lorawan.LoraWanStack
)

// Handle sx126x interrupts
func radioIntHandler(intr interrupt.Interrupt) {
	loraRadio.HandleInterrupt()

}

func main() {
	println("# TinyGo GNSE Composter Demo ")
	println("")

	// Init LED
	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.LED.High()

	// Onboard shtc3 Init (Power GPIO, I2C0 Bus, Driver)
	machine.SENSOR_EN.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.SENSOR_EN.High()
	machine.I2C0.Configure(machine.I2CConfig{})
	sensor := shtc3.New(machine.I2C0)

	// Define Node OOTA settings
	switch provider := "orange"; provider {
	case "chirpstack":
		loraStack.SetOtaa(
			[8]uint8{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			[8]uint8{0xA8, 0x40, 0x41, 0x00, 0x01, 0x81, 0xB3, 0x65},
			[16]uint8{0x2C, 0x44, 0xFC, 0xF8, 0x6C, 0x7B, 0x76, 0x7B, 0x8F, 0xD3, 0x12, 0x4F, 0xCE, 0x7A, 0x32, 0x16},
		)
	case "ttn":
		loraStack.SetOtaa(
			[8]uint8{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			[8]uint8{0x70, 0xB3, 0xD5, 0x7E, 0xD0, 0x04, 0xA9, 0x12},
			[16]uint8{0x67, 0x57, 0xBB, 0x98, 0x1D, 0x0E, 0x26, 0x71, 0xF4, 0x0F, 0x53, 0x4F, 0x6E, 0x4C, 0xD8, 0x7F},
		)
	case "orange":
		loraStack.SetOtaa(
			[8]uint8{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			[8]uint8{0x71, 0x33, 0x17, 0x88, 0x0C, 0x10, 0x88, 0x01},
			[16]uint8{0x61, 0x52, 0xB4, 0x33, 0x17, 0x12, 0x33, 0x44, 0xBE, 0xAF, 0xF0, 0x0F, 0x01, 0x02, 0x03, 0x01},
		)

	}
	println("APPEUI:", hex.EncodeToString(loraStack.Otaa.AppEUI[:]))
	println("APPEUI:", hex.EncodeToString(loraStack.Otaa.DevEUI[:]))
	println("APPKEY", hex.EncodeToString(loraStack.Otaa.AppKey[:]))

	// Initialize DevNonce with TRNG
	rnd := extra.GetRand16()
	loraStack.Otaa.DevNonce[0] = rnd[0]
	loraStack.Otaa.DevNonce[1] = rnd[1]

	// Driver for SX126x device on SubGhzSPI (SPI3)
	loraRadio = sx126x.New(machine.SPI3)
	loraRadio.SetDeviceType(sx126x.DEVICE_TYPE_SX1262)

	// Create RF Switch
	var radioSwitch rfswitch.CustomSwitch
	loraRadio.SetRfSwitch(radioSwitch)

	// Check the radio is ready
	state := loraRadio.DetectDevice()
	if !state {
		println("sx126x not detected... Aborting")
		for {
		}
	}

	// Attach the Lora Radio to LoraStack
	loraStack.AttachLoraRadio(loraRadio)

	// Add interrupt handler for Radio IRQs (DIO)
	intr := interrupt.New(stm32.IRQ_Radio_IRQ_Busy, radioIntHandler)
	intr.Enable()

	// Prepare for Lora Operation
	loraConf := sx126x.LoraConfig{
		Freq:           868100000,
		Bw:             sx126x.SX126X_LORA_BW_125_0,
		Sf:             sx126x.SX126X_LORA_SF9,
		Cr:             sx126x.SX126X_LORA_CR_4_7,
		HeaderType:     sx126x.SX126X_LORA_HEADER_EXPLICIT,
		Preamble:       12,
		Ldr:            sx126x.SX126X_LORA_LOW_DATA_RATE_OPTIMIZE_OFF,
		Iq:             sx126x.SX126X_LORA_IQ_STANDARD,
		Crc:            sx126x.SX126X_LORA_CRC_ON,
		SyncWord:       sx126x.SX126X_LORA_MAC_PUBLIC_SYNCWORD,
		LoraTxPowerDBm: 20,
	}
	loraRadio.LoraConfig(loraConf)

	loraConnected := false

	// Go routine for joining Lorawan
	go func() {
		for {
			for !loraConnected {
				err := loraStack.LoraWanJoin()
				if err != nil {
					println("Error joining: ", err, ",wait 300 sec")
					time.Sleep(time.Second * 300)
				} else {
					loraConnected = true
				}
			}
			// We are connected
			println("Lorawan connection state:", loraConnected)
			if loraConnected {
				machine.LED.Set(!machine.LED.Get())
			}
			time.Sleep(time.Second * 3)
		}
	}()

	// Wait 10 sec to give a chance to Lorawan connexion
	time.Sleep(time.Second * 10)

	encoder := cayennelpp.NewEncoder()
	// main loop
	for {

		// Get temperature from sensor
		sensor.WakeUp()

		// Read internal sensor
		temp, humidity, _ := sensor.ReadTemperatureHumidity()

		// Encode payload
		encoder.Reset()
		encoder.AddTemperature(1, float64(temp)/1000)
		encoder.AddRelativeHumidity(2, float64(humidity)/100)
		cayBytes := encoder.Bytes()

		// Send payload if connected
		if loraConnected {
			println("lorawan: Sending Cayenne: ", hex.EncodeToString(cayBytes))
			err := loraStack.LoraSendUplink(cayBytes)
			if err != nil {
				println(err)
			}
		} else {
			println("main: Waiting for Lorawan connectivity")
		}
		// Go to sleep
		println("Sleep 180s")
		time.Sleep(180 * time.Second)
	}

}
