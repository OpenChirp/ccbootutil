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

	"strconv"

	"github.com/jacobsa/go-serial/serial"
	"github.com/openchirp/ccboot"
)

const (
	// This is the default speed that the SmartRf Flash Programmer 2 uses
	SmartRFFlashProgrammer2Speed = 460800
)

const (
	// timeout in ms
	readTimeout = 1000
)

var commands = []string{
	"sync",
	"ping",
	"download",
	"getstatus",
	"getchipid",
	"bankerase",
	"reset",
	"setccfg",
	"flash",
	"verify",
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
	flag.UintVar(&portSpeed, "speed", SmartRFFlashProgrammer2Speed, "The serial baud rate to use")
}

func main() {
	setupFlags()
	flag.Parse()

	if len(flag.Args()) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	args := flag.Args()[:]
	portName := args[0]
	cmd := args[1]

	args = args[2:]

	// Set up options.
	options := serial.OpenOptions{
		PortName:              portName,
		BaudRate:              portSpeed,
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
			log.Fatalf("Error synchronizing device: %v\n", err)
		}
		log.Println("Synchronization success")
	case "ping":
		// Ping Device Bootloader
		log.Println("Pinging")
		err = d.Ping()
		if err != nil {
			log.Fatalf("Error pinging device: %v\n", err)
		}
		log.Println("Ping success")
	case "download":
		// Send Download command
		log.Println("Downloading")
		if len(args) < 2 {
			log.Fatalf("Error - does not specify address and size")
		}
		addr, err := strconv.ParseUint(args[0], 16, 32)
		if err != nil {
			log.Fatalf("Error parsing address: %v\n", err)
		}
		size, err := strconv.ParseUint(args[1], 10, 32)
		if err != nil {
			log.Fatalf("Error parsing size: %v\n", err)
		}
		log.Printf("Sending Download command with 0x%x and %d\n", uint32(addr), uint32(size))
		err = d.Download(uint32(addr), uint32(size))
		if err != nil {
			log.Fatalf("Error sending download todevice: %v\n", err)
		}
		log.Println("Download success")
	case "getstatus":
		// Get Status
		log.Println("Getting status")
		status, err := d.GetStatus()
		if err != nil {
			fmt.Println() // maintain parsibility
			log.Fatalf("# Error - %v\n", err)
		}
		fmt.Println(status)
	case "getchipid":
		// Get Chip ID
		log.Println("Getting chip ID")
		id, err := d.GetChipID()
		if err != nil {
			log.Fatalf("Error reading chip id: %v\n", err)
		}
		fmt.Printf("0x%.8X\n", id)
	case "bankerase":
		// Bank Erase
		log.Println("Bank erasing")
		err = d.BankErase()
		if err != nil {
			log.Fatalf("Error - Could not bank erase device: %v\n", err)
		}
		log.Println("Bank erase success")
	case "reset":
		// Reset Device
		log.Println("Resetting device")
		err = d.Reset()
		if err != nil {
			log.Fatalf("Error - Could not reset chip: %v\n", err)
		}
		log.Println("Device reset")
	case "setccfg":
		// Set CCFGs
		if (len(args) < 2) || (len(args)%2 != 0) {
			log.Fatalf("Error - Parameters for CCFG should specify <CCFG_FIELD_ID> followed by <value>")
		}

		for i := 0; i < len(args)/2; i++ {
			fieldid, err := ccboot.ParseCCFGFieldID(args[i*2])
			if err != nil {
				log.Fatalf("Error - Could not parse Field ID %s\n", args[i*2])
			}
			// base=0 means it will try to auto detect base
			fieldvalue, err := strconv.ParseUint(args[i*2+1], 0, 32)
			if err != nil {
				log.Fatalf("Error - Could not parse Field Value %s\n", args[i*2+1])
			}
			log.Printf("Setting CCFG %v to 0x%X (%d)", fieldid, uint32(fieldvalue), uint32(fieldvalue))
			err = d.SetCCFG(fieldid, uint32(fieldvalue))
			if err != nil {
				log.Fatalf("Error - Could not set CCFG %v to 0x%X: %v\n", fieldid, uint32(fieldvalue), err)
			}
		}
		log.Println("Device CCFG set")
	case "flash":
		log.Println("Flashing device")
		if len(args) != 1 {
			fmt.Println("FAILURE")
			log.Fatalf("Error - No ELF binary specified")
		}
		if err := flash(d, args[0]); err != nil {
			fmt.Println("FAILURE")
			log.Fatalf("Error - Failed to flash: %v\n", err)
		}
		fmt.Println("SUCCESS")
	case "verify":
		rcount := uint64(0)
		log.Println("Verifying device image")
		if len(args) != 1 {
			fmt.Println("FAILURE")
			log.Fatalf("Error - No ELF binary specified")
		}
		// TODO: Fix rcount > 0
		// if len(args) == 2 {
		// 	rcount, err = strconv.ParseUint(args[1], 0, 32)
		// 	if err != nil {
		// 		log.Fatalf("Error - Failed to parse read cycle count %s: %v\n", args[1], err)
		// 	}
		// }
		pass, err := verify(d, args[0], uint32(rcount))
		if err != nil {
			fmt.Println("FAILURE")
			log.Fatalf("Error - Failed to verify: %v\n", err)
		}
		if pass {
			fmt.Println("SUCCESS")
		} else {
			fmt.Println("FAILURE")
		}
	default:
		log.Fatalf("Error - Invalid command given")
	}

	log.Printf("Exiting\n")
}
