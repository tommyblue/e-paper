package main

import (
	"log"
	"time"

	epaper "github.com/tommyblue/e-paper"
)

func main() {
	cfg := epaper.Config{
		Model: epaper.EPD_2IN13_V4,
	}
	epd, err := epaper.New(cfg)
	if err != nil {
		log.Fatalf("cannot create e-paper: %v", err)
	}

	log.Println("Init")
	epd.Init()

	log.Println("wait 5 seconds")
	time.Sleep(5 * time.Second)

	log.Println("Clear")
	epd.Clear()

	log.Println("wait 5 seconds")
	time.Sleep(5 * time.Second)

	log.Println("White")
	epd.Paint(epaper.WHITE)

	log.Println("wait 5 seconds")
	time.Sleep(5 * time.Second)

	log.Println("Black")
	epd.Paint(epaper.BLACK)

	log.Println("wait 5 seconds")
	time.Sleep(5 * time.Second)

	log.Println("Clear")
	epd.Clear()
}
