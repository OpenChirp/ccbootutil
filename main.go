// Copyright 2017 OpenChirp. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
//
// March 13, 2017
// Craig Hesling <craig@hesling.com>

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"os"

	"github.com/jacobsa/go-serial/serial"
	"github.com/openchirp/ccboot"
)

const (
	// timeout in ms
	readTimeout = 1000
)

var commands = []string{
	"sync",
	"ping",
	"getstatus",
	"getchipid",
	"bankerase",
	"reset",
}

var verbose bool
var portSpeed uint

func setupFlags() {
	flag.Usage = func() {
		fmt.Printf("Usage: ccbootutil [options] <portname> <command> [parameters]\n\n")
		fmt.Printf("Commands:\n")
		for _, cmd := range commands {
			fmt.Printf("\t%s\n", cmd)
		}
		fmt.Printf("\n")

		fmt.Printf("Options:\n")
		flag.PrintDefaults()
	}
	flag.BoolVar(&verbose, "verbose", false, "Toggles the verbose setting")
	flag.UintVar(&portSpeed, "speed", 115200, "The serial baud rate to use")
}

func main() {
	setupFlags()
	flag.Parse()

	if len(flag.Args()) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	portname := flag.Args()[0]
	cmd := flag.Args()[1]

	// Set up options.
	options := serial.OpenOptions{
		PortName:              portname,
		BaudRate:              115200,
		DataBits:              8,
		StopBits:              1,
		MinimumReadSize:       1,
		InterCharacterTimeout: readTimeout,
	}

	log.SetOutput(ioutil.Discard)
	if verbose {
		log.SetOutput(os.Stderr)
	}

	// Open the port.
	log.Println("Opening Serial")
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("Failed to open serial port: %v", err)
	}
	// Make sure to close it later.
	defer port.Close()

	d := ccboot.NewDevice(port)

	switch cmd {
	case "sync":
		// Synchronize
		log.Println("Synchronizing")
		if err = d.Sync(); err != nil {
			log.Printf("Error synchronizing device: %v\n", err)
			os.Exit(1)
		}
		log.Println("Synchronization success")
	case "ping":
		// Ping Device Bootloader
		log.Println("Pinging")
		err = d.Ping()
		if err != nil {
			log.Printf("Error pinging device: %s\n", err.Error())
			os.Exit(1)
		}
		log.Println("Ping success")
	case "getstatus":
		// Get Status
		log.Println("Getting status")
		status, err := d.GetStatus()
		if err != nil {
			fmt.Println() // maintain parsibility
			log.Printf("# Error - %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Println(status.GetString())
	case "getchipid":
		// Get Chip ID
		log.Println("Getting chip ID")
		id, err := d.GetChipID()
		if err != nil {
			log.Printf("Error reading chip id: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Printf("0x%.8X\n", id)
	case "bankerase":
		// Bank Erase
		log.Println("Bank erasing")
		err = d.BankErase()
		if err != nil {
			log.Printf("Error bank erasing device: %s\n", err.Error())
			os.Exit(1)
		}
		log.Println("Bank erase success")
	case "reset":
		// Reset Device
		log.Println("Resetting device")
		err = d.Reset()
		if err != nil {
			log.Printf("Error resetting chip: %s\n", err.Error())
			os.Exit(1)
		}
		log.Println("Device reset")
	}

	log.Printf("Exiting\n")
}
