package epaper

import (
	"log"
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

// Pin GPIO assegnati
const (
	DC_PIN  = "GPIO25"
	RST_PIN = "GPIO17"
	CS_PIN  = "GPIO8"
	BSY_PIN = "GPIO24"
)

// Comandi e-paper V4
const (
	POWER_ON           = 0x04
	PANEL_SETTING      = 0x00
	BOOSTER_SOFT_START = 0x06
	DISPLAY_REFRESH    = 0x12
	DEEP_SLEEP         = 0x10
)

type EPD_2in13_V4 struct {
	spi  spi.Conn
	dc   gpio.PinOut
	cs   gpio.PinOut
	rst  gpio.PinOut
	busy gpio.PinIO
}

func NewEPD_2in13_V4() *EPD_2in13_V4 {
	if _, err := host.Init(); err != nil {
		log.Fatalf("Errore inizializzazione periph: %v", err)
	}

	// Inizializza SPI
	spiPort, err := spireg.Open("/dev/spidev0.0")
	if err != nil {
		log.Fatalf("Errore apertura SPI: %v", err)
	}
	conn, err := spiPort.Connect(4*physic.MegaHertz, spi.Mode0, 8)
	if err != nil {
		log.Fatalf("Errore connessione SPI: %v", err)
	}

	// Configura GPIO
	dc := gpioreg.ByName(DC_PIN)
	cs := gpioreg.ByName(CS_PIN)
	rst := gpioreg.ByName(RST_PIN)
	busy := gpioreg.ByName(BSY_PIN)

	if dc == nil || rst == nil || busy == nil {
		log.Fatal("Errore: Impossibile trovare i pin GPIO")
	}

	// Inizializza il display
	return &EPD_2in13_V4{spi: conn, dc: dc, cs: cs, rst: rst, busy: busy}
}

func (e *EPD_2in13_V4) Init() {
	e.Reset()

	e.WaitUntilIdle()
	e.SendCommand(0x12) //SWRESET
	e.WaitUntilIdle()

	e.SendCommand(0x01) //Driver output control
	e.SendData(0xF9)
	e.SendData(0x00)
	e.SendData(0x00)

	e.SendCommand(0x11) //data entry mode
	e.SendData(0x03)

	e.SetWindows(0, 0, EPD_WIDTH-1, EPD_HEIGHT-1)
	e.SetCursor(0, 0)

	e.SendCommand(0x3C) //BorderWavefrom
	e.SendData(0x05)

	e.SendCommand(0x21) //  Display update control
	e.SendData(0x00)
	e.SendData(0x80)

	e.SendCommand(0x18) //Read built-in temperature sensor
	e.SendData(0x80)
	e.WaitUntilIdle()
}

func (e *EPD_2in13_V4) TurnOnDisplay() {
	e.SendCommand(0x22) // Display Update Control
	e.SendData(0xf7)
	e.SendCommand(0x20) // Activate Display Update Sequence
	e.WaitUntilIdle()
}

var (
	BLACK byte = 0x00
	WHITE byte = 0xFF
)

func (e *EPD_2in13_V4) Paint(color byte) {
	// UBYTE *BlackImage;
	// UWORD Imagesize = ((EPD_2in13_V4_WIDTH % 8 == 0)? (EPD_2in13_V4_WIDTH / 8 ): (EPD_2in13_V4_WIDTH / 8 + 1)) * EPD_2in13_V4_HEIGHT;
	// if((BlackImage = (UBYTE *)malloc(Imagesize)) == NULL) {
	// 	Debug("Failed to apply for black memory...\r\n");
	// 	return -1;
	// }
	// Debug("Paint_NewImage\r\n");
	// Paint_NewImage(BlackImage, EPD_2in13_V4_WIDTH, EPD_2in13_V4_HEIGHT, 90, WHITE);
	// Paint_Clear(WHITE);

	// translate the code above in Go
	// blackImage := make([]byte, (EPD_WIDTH/8+1)*EPD_HEIGHT)
	// for i := range blackImage {
	// 	blackImage[i] = 0xFF
	// }

	// e.SendCommand(0x24)
	// for j := 0; j < EPD_HEIGHT; j++ {
	// 	for i := 0; i < EPD_WIDTH/8; i++ {
	// 		e.SendData(blackImage[i])
	// 	}
	// }

	// e.TurnOnDisplay()
	img := make([]byte, (EPD_WIDTH/8+1)*EPD_HEIGHT)
	e.PaintNewImage(img, EPD_WIDTH, EPD_HEIGHT, 90, color)
	// e.PaintClear(WHITE)
}

type paint struct {
	image []byte
	// WidthMemory  byte
	// HeightMemory byte
	color byte
	// Scale        uint8
	// WidthByte    byte
	// HeightByte   byte
	// rotate byte
	// Mirror       byte
	width  int
	height int
}

func (e *EPD_2in13_V4) PaintNewImage(img []byte, width, height, rotate int, color byte) {
	p := paint{
		image: img,
		color: color,
		// rotate: rotate,
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

	e.Display(p)
}

func (e *EPD_2in13_V4) Display(p paint) {
	w := (p.width/8 + 1)
	if p.width%8 == 0 {
		w = (p.width / 8)
	}
	h := p.height

	e.SendCommand(0x24)
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			e.SendData(p.image[i+j*w])
		}
	}

	e.TurnOnDisplay()
}

func (e *EPD_2in13_V4) Clear() {
	// w := (EPD_WIDTH/8 + 1)
	// if EPD_WIDTH%8 == 0 {
	// 	w = (EPD_WIDTH / 8)
	// }
	// h := EPD_HEIGHT

	// e.SendCommand(0x24)
	// for j := 0; j < h; j++ {
	// 	for i := 0; i < w; i++ {
	// 		e.SendData(WHITE)
	// 	}
	// }

	// e.TurnOnDisplay()

	// return nil
	e.Paint(WHITE)
}

// void Paint_Clear(UWORD Color)
// {
// 	if(Paint.Scale == 2) {
// 		for (UWORD Y = 0; Y < Paint.HeightByte; Y++) {
// 			for (UWORD X = 0; X < Paint.WidthByte; X++ ) {//8 pixel =  1 byte
// 				UDOUBLE Addr = X + Y*Paint.WidthByte;
// 				Paint.Image[Addr] = Color;
// 			}
// 		}
//     }else if(Paint.Scale == 4) {
//         for (UWORD Y = 0; Y < Paint.HeightByte; Y++) {
// 			for (UWORD X = 0; X < Paint.WidthByte; X++ ) {
// 				UDOUBLE Addr = X + Y*Paint.WidthByte;
// 				Paint.Image[Addr] = (Color<<6)|(Color<<4)|(Color<<2)|Color;
// 			}
// 		}
// 	}else if(Paint.Scale == 6 || Paint.Scale == 7 || Paint.Scale == 16) {
// 		for (UWORD Y = 0; Y < Paint.HeightByte; Y++) {
// 			for (UWORD X = 0; X < Paint.WidthByte; X++ ) {
// 				UDOUBLE Addr = X + Y*Paint.WidthByte;
// 				Paint.Image[Addr] = (Color<<4)|Color;
// 			}
// 		}
// 	}
// }

func (e *EPD_2in13_V4) Reset() {
	e.rst.Out(gpio.High)
	time.Sleep(20 * time.Millisecond)
	e.rst.Out(gpio.Low)
	time.Sleep(2 * time.Millisecond)
	e.rst.Out(gpio.High)
	time.Sleep(20 * time.Millisecond)
}

func (e *EPD_2in13_V4) SendCommand(cmd byte) {
	e.dc.Out(gpio.Low)
	e.cs.Out(gpio.Low)
	e.spi.Tx([]byte{cmd}, nil)
	e.cs.Out(gpio.High)
}

func (e *EPD_2in13_V4) SendData(data byte) {
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

func (e *EPD_2in13_V4) SetWindows(Xstart, Ystart, Xend, Yend byte) {
	e.SendCommand(0x44) // SET_RAM_X_ADDRESS_START_END_POSITION
	e.SendData((Xstart >> 3) & 0xFF)
	e.SendData((Xend >> 3) & 0xFF)

	e.SendCommand(0x45) // SET_RAM_Y_ADDRESS_START_END_POSITION
	e.SendData(Ystart & 0xFF)
	e.SendData((Ystart >> 8) & 0xFF)
	e.SendData(Yend & 0xFF)
	e.SendData((Yend >> 8) & 0xFF)
}

func (e *EPD_2in13_V4) SetCursor(Xstart, Ystart byte) {
	e.SendCommand(0x4E) // SET_RAM_X_ADDRESS_COUNTER
	e.SendData(Xstart & 0xFF)

	e.SendCommand(0x4F) // SET_RAM_Y_ADDRESS_COUNTER
	e.SendData(Ystart & 0xFF)
	e.SendData((Ystart >> 8) & 0xFF)
}
