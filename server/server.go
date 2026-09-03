package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/gen2brain/malgo"
	"github.com/zaf/g711"
)

func main() {
	fmt.Println("Starting UDP Audio Server (Speaker)...")
	listener, err := net.ResolveUDPAddr("udp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", listener)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	fmt.Println("Listening on", listener)

	// Channel to pass audio from network to soundcard
	audioChan := make(chan []byte, 100)

	// Initialize Malgo context
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG: %v\n", message)
	})
	if err != nil {
		fmt.Println("Error initializing context:", err)
		os.Exit(1)
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	// Configure Playback Device
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = 1
	deviceConfig.SampleRate = 44100

	// Audio Callback
	onRecvFrames := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		select {
		case audioData := <-audioChan:
			copy(pOutputSample, audioData)
		default:
			// Fill with silence if no data
			for i := range pOutputSample {
				pOutputSample[i] = 0
			}
		}
	}

	// Initialize and start device
	deviceCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		fmt.Println("Error initializing device:", err)
		os.Exit(1)
	}
	defer device.Uninit()

	err = device.Start()
	if err != nil {
		fmt.Println("Error starting device:", err)
		os.Exit(1)
	}

	// Network loop
	buffer := make([]byte, 4096)
	var lastReceivedSeq uint32 = 0
	
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Read error: %v", err)
			continue
		}
		
		// Ensure packet is large enough to contain our 4-byte header
		if n < 4 {
			continue
		}
		
		// Extract sequence number from the first 4 bytes
		incomingSeq := binary.LittleEndian.Uint32(buffer[0:4])
		
		// Jitter Fix: Drop out-of-order / late packets
		// If this is the first packet, or if it's newer than the last one we processed
		if incomingSeq > lastReceivedSeq || lastReceivedSeq == 0 {
			lastReceivedSeq = incomingSeq
			
			// Optional: Print to console to prove it works
			// fmt.Printf("Playing packet %d\n", incomingSeq)

			// Decode the 8-bit G.711 chunk (skipping the 4-byte header) back into 16-bit PCM
			decompressedAudio := g711.DecodeUlaw(buffer[4:n])
			
			// Non-blocking send
			select {
			case audioChan <- decompressedAudio:
			default:
				// Dropping packet because buffer is full
			}
		} else {
			// If we get here, the packet arrived late and we are throwing it away!
			// fmt.Printf("Dropped late packet %d (Current: %d)\n", incomingSeq, lastReceivedSeq)
		}
	}
}
