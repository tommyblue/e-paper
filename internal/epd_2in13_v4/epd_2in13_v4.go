package epd_2in13_v4

import (
	"fmt"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

const (
	EPD_WIDTH  = 122
	EPD_HEIGHT = 250
)

const (
	EPD_DC_PIN  = "GPIO25"
	EPD_RST_PIN = "GPIO17"
	EPD_CS_PIN  = "GPIO8"
	EPD_BSY_PIN = "GPIO24"
)

type EPD_2in13_V4 struct {
	spi  spi.Conn
	dc   gpio.PinOut
	cs   gpio.PinOut
	rst  gpio.PinOut
	busy gpio.PinIO
}

func New() (*EPD_2in13_V4, error) {
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("cannot initialize host library: %v", err)
	}

	spiPort, err := spireg.Open("/dev/spidev0.0")
	if err != nil {
		return nil, fmt.Errorf("cannot open SPI device: %v", err)
	}
	conn, err := spiPort.Connect(4*physic.MegaHertz, spi.Mode0, 8)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to SPI: %v", err)
	}

	dc := gpioreg.ByName(EPD_DC_PIN)
	cs := gpioreg.ByName(EPD_CS_PIN)
	rst := gpioreg.ByName(EPD_RST_PIN)
	busy := gpioreg.ByName(EPD_BSY_PIN)

	if dc == nil || rst == nil || busy == nil {
		return nil, fmt.Errorf("cannot find GPIO pins")
	}

	epd := &EPD_2in13_V4{
		spi:  conn,
		dc:   dc,
		cs:   cs,
		rst:  rst,
		busy: busy,
	}

	return epd, nil
}

// Init the e-paper display
func (e *EPD_2in13_V4) Init() {
	e.reset()

	e.WaitUntilIdle()
	e.sendCommand(0x12) // SWRESET
	e.WaitUntilIdle()

	e.sendCommand(0x01) // Driver output control
	e.sendData(0xF9)
	e.sendData(0x00)
	e.sendData(0x00)

	e.sendCommand(0x11) // data entry mode
	e.sendData(0x03)

	e.setWindows(0, 0, EPD_WIDTH-1, EPD_HEIGHT-1)
	e.setCursor(0, 0)

	e.sendCommand(0x3C) // BorderWavefrom
	e.sendData(0x05)

	e.sendCommand(0x21) // Display update control
	e.sendData(0x00)
	e.sendData(0x80)

	e.sendCommand(0x18) // Read built-in temperature sensor
	e.sendData(0x80)
	e.WaitUntilIdle()
}

func (e *EPD_2in13_V4) Paint(color byte) {
	img := make([]byte, (EPD_WIDTH/8+1)*EPD_HEIGHT)
	e.paintNewImage(img, EPD_WIDTH, EPD_HEIGHT, 0, color)
}

// Clear the image painting to white
func (e *EPD_2in13_V4) Clear() {
	e.Paint(0xFF)
}

// Internal

func (e *EPD_2in13_V4) turnOnDisplay() {
	e.sendCommand(0x22) // Display Update Control
	e.sendData(0xf7)
	e.sendCommand(0x20) // Activate Display Update Sequence
	e.WaitUntilIdle()
}

type paint struct {
	image  []byte
	color  byte
	width  int
	height int
}

func (e *EPD_2in13_V4) paintNewImage(img []byte, width, height, rotate int, color byte) {
	p := paint{
		image:  img,
		color:  color,
		width:  width,
		height: height,
	}

	// paint color
	w := (p.width/8 + 1)
	if p.width%8 == 0 {
		w = (p.width / 8)
	}
	h := p.height
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			p.image[i+j*w] = color
		}
	}

	if rotate == 0 || rotate == 180 {
		p.width = width
		p.height = height
	} else {
		p.width = height
		p.height = width
	}

	e.display(p)
}

func (e *EPD_2in13_V4) display(p paint) {
	w := (p.width/8 + 1)
	if p.width%8 == 0 {
		w = (p.width / 8)
	}
	h := p.height

	e.sendCommand(0x24)
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			e.sendData(p.image[i+j*w])
		}
	}

	e.turnOnDisplay()
}

func (e *EPD_2in13_V4) reset() {
	e.rst.Out(gpio.High)
	time.Sleep(20 * time.Millisecond)
	e.rst.Out(gpio.Low)
	time.Sleep(2 * time.Millisecond)
	e.rst.Out(gpio.High)
	time.Sleep(20 * time.Millisecond)
}

func (e *EPD_2in13_V4) sendCommand(cmd byte) {
	e.dc.Out(gpio.Low)
	e.cs.Out(gpio.Low)
	e.spi.Tx([]byte{cmd}, nil)
	e.cs.Out(gpio.High)
}

func (e *EPD_2in13_V4) sendData(data byte) {
	e.dc.Out(gpio.High)
	e.cs.Out(gpio.Low)
	e.spi.Tx([]byte{data}, nil)
	e.cs.Out(gpio.High)
}

func (e *EPD_2in13_V4) WaitUntilIdle() {
	for e.busy.Read() == gpio.High {
		time.Sleep(10 * time.Millisecond)
	}
}

func (e *EPD_2in13_V4) setWindows(Xstart, Ystart, Xend, Yend byte) {
	e.sendCommand(0x44) // SET_RAM_X_ADDRESS_START_END_POSITION
	e.sendData((Xstart >> 3) & 0xFF)
	e.sendData((Xend >> 3) & 0xFF)

	e.sendCommand(0x45) // SET_RAM_Y_ADDRESS_START_END_POSITION
	e.sendData(Ystart & 0xFF)
	e.sendData((Ystart >> 8) & 0xFF)
	e.sendData(Yend & 0xFF)
	e.sendData((Yend >> 8) & 0xFF)
}

func (e *EPD_2in13_V4) setCursor(Xstart, Ystart byte) {
	e.sendCommand(0x4E) // SET_RAM_X_ADDRESS_COUNTER
	e.sendData(Xstart & 0xFF)

	e.sendCommand(0x4F) // SET_RAM_Y_ADDRESS_COUNTER
	e.sendData(Ystart & 0xFF)
	e.sendData((Ystart >> 8) & 0xFF)
}
