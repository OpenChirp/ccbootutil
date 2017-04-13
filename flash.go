// This file holds the second order functions

package main

import (
	"debug/elf"
	"fmt"
	"log"

	"errors"

	"hash/crc32"
	"io/ioutil"

	"os"

	"github.com/openchirp/ccboot"
)

var ErrDownload = errors.New("Error sending download command")
var ErrSendData = errors.New("Error sending senddata command")

func flash(d *ccboot.Device, filepath string) error {
	log.Printf("Parsing %s\n", filepath)
	file, err := elf.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// log.Printf("Reading from memory 0x400910\n")
	// buf, err := d.MemoryRead(0x40091090, ccboot.ReadType8Bit, 4)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// log.Println(hex.EncodeToString(buf))

	log.Printf("Mass erasing chip\n")
	if err := d.BankErase(); err != nil {
		// communication error
		return err
	}
	log.Printf("Sending GetStatus\n")
	status, err := d.GetStatus()
	if err != nil {
		// communication error
		return err
	}
	if status != ccboot.COMMAND_RET_SUCCESS {
		fmt.Printf("Error sending bank erase: %v\n", status)
		return ErrDownload
	}

	// Print out a useful table of all regions
	// fmt.Println("# Program Regions #")
	// for _, p := range file.Progs {
	// 	fmt.Printf("0x%X (%d) aligned %d: %v\n", p.Paddr, p.Memsz, p.Align, p)
	// }
	// fmt.Println()

	for _, p := range file.Progs {
		addr := uint32(p.Paddr)
		size := uint32(p.Memsz)
		bytestream := p.Open()

		log.Printf("Flashing region at 0x%X of size %d and alignment %d: %v\n", p.Paddr, p.Memsz, p.Align, *p)

		if p.Paddr > 0x10000000 {
			log.Printf("Skipping - Region starts after flash area")
			continue
		}

		// ASSERT: Program block being written must be a multiple of the memory
		// word size -- must be aligned p.Align
		if (uint64(p.Memsz) % uint64(p.Align)) != 0 {
			log.Printf("# Error Detected: Program block being written, total size of %d, is not aligned to %d\n", p.Memsz, p.Align)
		}

		log.Printf("Sending Download for address 0x%X for %d bytes\n", addr, size)
		if err := d.Download(addr, size); err != nil {
			// communication error
			return err
		}
		log.Printf("Sending GetStatus\n")
		status, err := d.GetStatus()
		if err != nil {
			// communication error
			return err
		}
		if status != ccboot.COMMAND_RET_SUCCESS {
			fmt.Printf("Error sending download address and size: %v\n", status)
			return ErrDownload
		}
		log.Printf("Status: %v\n", status)

		log.Printf("Sending Data Blocks - Start")
		blockbuf := make([]byte, ccboot.SendDataMaxSize)
		for n, _ := bytestream.Read(blockbuf); n > 0; n, _ = bytestream.Read(blockbuf) {
			// time.Sleep(time.Second * time.Duration(2))
			block := blockbuf[0:n]

			// log.Printf("Sending SendData: [%d]=%s\n", len(block), hex.EncodeToString(block))
			// log.Printf("Sending SendData: [%d]=%s\n", len(block), "")
			if err := d.SendData(block); err != nil {
				// communication error
				return err
			}
			// log.Printf("Sending GetStatus\n")
			status, err = d.GetStatus()
			if err != nil {
				// communication error
				return err
			}
			if status != ccboot.COMMAND_RET_SUCCESS {
				fmt.Printf("Error sending data: %v\n", status)
				return ErrSendData
			}
			// log.Printf("Status: %v\n", status)

			// print dot for each block successfully flashed
			if verbose {
				fmt.Fprint(os.Stderr, ".")
			}
		}
		// clean up flashing dots
		if verbose {
			fmt.Fprintln(os.Stderr)
		}
		log.Printf("Sending Data Blocks - Complete")

	}
	log.Println("Flash done!")
	return nil
}

// TODO: Fix rcount greater than 0 case.
func verify(d *ccboot.Device, filepath string, rcount uint32) (bool, error) {
	log.Printf("Parsing %s\n", filepath)
	file, err := elf.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// log.Printf("Reading from memory 0x400910\n")
	// buf, err := d.MemoryRead(0x40091090, ccboot.ReadType8Bit, 4)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// log.Println(hex.EncodeToString(buf))

	for _, p := range file.Progs {
		addr := uint32(p.Paddr)
		size := uint32(p.Memsz)
		bytestream := p.Open()

		log.Printf("Checking region at 0x%X of size %d and alignment %d: %v\n", p.Paddr, p.Memsz, p.Align, *p)

		if p.Paddr > 0x10000000 {
			log.Printf("Skipping - Region starts after flash area")
			continue
		}

		log.Printf("Sending CRC32 for address 0x%X for %d bytes with %d reads cycles\n", addr, size, rcount)
		targetcrc32, err := d.CRC32(addr, size, rcount)
		if err != nil {
			// communication error
			return false, err
		}

		hostdata, err := ioutil.ReadAll(bytestream)
		if err != nil {
			return false, err
		}

		hostdatacoppied := make([]byte, len(hostdata), len(hostdata)*(int(rcount)+1))
		copy(hostdatacoppied, hostdata)

		for i := 0; i < int(rcount); i++ {
			// duplicate at end rcount times
			hostdatacoppied = append(hostdatacoppied, hostdata...)
		}

		hostcrc32 := crc32.ChecksumIEEE(hostdatacoppied)

		log.Printf("Target CRC32 (%d): 0x%.8X\n", rcount, targetcrc32)
		log.Printf("Host CRC32 (%d): 0x%.8X\n", rcount, hostcrc32)

		if targetcrc32 != hostcrc32 {
			log.Println("Verification failed!")
			return false, nil
		}
	}
	log.Println("Verification succeeded!")
	return true, nil
}
