package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/gen2brain/malgo"
)

func main() {
	fmt.Println("Starting UDP Audio Client (Microphone)...")
	
	// Connect to server
	addr, err := net.ResolveUDPAddr("udp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

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

	// Configure Capture Device
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = 44100

	// Audio Callback
	onRecvFrames := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		// Send captured audio directly over UDP
		_, err := conn.Write(pInputSamples)
		if err != nil {
			fmt.Println("Error sending audio:", err)
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

	fmt.Println("Recording and streaming... Press ENTER to stop.")
	var input string
	fmt.Scanln(&input)
}
