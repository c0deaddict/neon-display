// SPDX-FileCopyrightText: 2020 Kent Gibson <warthog618@gmail.com>
//
// SPDX-License-Identifier: MIT

//go:build linux
// +build linux

// A simple example that watches an input pin and reports edge events.
package gpio

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"
)

const (
	redButton      = 22
	yellowButton   = 27
	pir            = 10
	debounceDelay  = time.Duration(50 * time.Millisecond)
	edgeDetectPoll = time.Duration(10 * time.Millisecond)
)

type Event uint8

const (
	RedButtonLow Event = iota
	RedButtonHigh
	YellowButtonLow
	YellowButtonHigh
	PirLow
	PirHigh
)

func eventHandler(evt gpiod.LineEvent) {
	t := time.Now()
	edge := "rising"
	if evt.Type == gpiod.LineEventFallingEdge {
		edge = "falling"
	}
	if evt.Seqno != 0 {
		// only uAPI v2 populates the sequence numbers
		fmt.Printf("event: #%d(%d)%3d %-7s %s (%s)\n",
			evt.Seqno,
			evt.LineSeqno,
			evt.Offset,
			edge,
			t.Format(time.RFC3339Nano),
			evt.Timestamp)
	} else {
		fmt.Printf("event:%3d %-7s %s (%s)\n",
			evt.Offset,
			edge,
			t.Format(time.RFC3339Nano),
			evt.Timestamp)
	}
}

// Watches GPIO 22 (Raspberry Pi J8-15) and reports when it changes state.
func WatchButtons() {
	// Pin numbering: https://elinux.org/RPi_BCM2835_GPIOs
	offset := rpi.J8p15
	period := 10 * time.Millisecond
	l, err := gpiod.RequestLine("gpiochip0", offset,
		gpiod.WithPullUp,
		gpiod.WithBothEdges,
		gpiod.WithDebounce(period),
		gpiod.WithEventHandler(eventHandler))
	if err != nil {
		fmt.Printf("RequestLine returned error: %s\n", err)
		if err == syscall.Errno(22) {
			fmt.Println("Note that the WithPullUp option requires kernel V5.5 or later - check your kernel version.")
		}
		os.Exit(1)
	}
	defer l.Close()

	// In a real application the main thread would do something useful.
	// But we'll just run for a minute then exit.
	fmt.Printf("Watching Pin %d...\n", offset)
	time.Sleep(time.Minute)
	fmt.Println("exiting...")
}
