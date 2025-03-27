package main

import (
	"io"
	"log"
	"os"
	"time"

	epaper "github.com/tommyblue/e-paper"
)

func main() {
	// open the bmp file
	f, err := os.Open("./2in13_1.bmp")
	if err != nil {
		log.Fatalf("cannot open bmp file: %v", err)
	}
	defer f.Close()
	bmp, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("cannot open bmp file: %v", err)
	}

	cfg := epaper.Config{
		Model: epaper.EPD_2IN13_V4,
	}
	epd, err := epaper.New(cfg)
	if err != nil {
		log.Fatalf("cannot create e-paper: %v", err)
	}

	log.Println("Init")
	epd.Init()

	log.Println("wait 2 seconds")
	time.Sleep(2 * time.Second)

	log.Println("Clear")
	epd.Clear()

	log.Println("wait 2 seconds")
	time.Sleep(2 * time.Second)

	// log.Println("White")
	// epd.PaintColor(epaper.WHITE)

	log.Println("Image")
	epd.PaintImage(bmp)

	log.Println("wait 2 seconds")
	time.Sleep(2 * time.Second)

	// log.Println("Black")
	// epd.PaintColor(epaper.BLACK)

	// log.Println("wait 2 seconds")
	// time.Sleep(2 * time.Second)

	log.Println("Clear")
	epd.Clear()
}
