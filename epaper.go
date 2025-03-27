package epaper

import (
	"fmt"

	"github.com/tommyblue/e-paper/internal/epd_2in13_v4"
)

type epaper interface {
	Init()
	Clear()
	PaintColor(byte)
	PaintImage([]byte)
}

// list of supported models
const (
	EPD_2IN13_V4 = "EPD_2IN13_V4"
)

type Config struct {
	Model string
}

var (
	BLACK byte = 0x00
	WHITE byte = 0xFF
)

func New(cfg Config) (epaper, error) {
	switch cfg.Model {
	case EPD_2IN13_V4:

		return epd_2in13_v4.New()
	default:
		return nil, fmt.Errorf("unsupported model: %s", cfg.Model)
	}
}
