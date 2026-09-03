package main

import (
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
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Read error: %v", err)
			continue
		}
		// Decode the 8-bit G.711 chunk back into 16-bit PCM
		decompressedAudio := g711.DecodeUlaw(buffer[:n])
		
		// Non-blocking send (drop packets if channel is full to prevent memory leak)
		select {
		case audioChan <- decompressedAudio:
		default:
			// Dropping packet because buffer is full
		}
	}
}
