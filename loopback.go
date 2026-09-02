package main

import (
	"fmt"
	"os"

	"github.com/gen2brain/malgo"
)

func main() {
	// Step 1: Initialize Context
	// This connects your Go program to the underlying OS audio system (Windows Core Audio, ALSA on Linux, etc.)
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

	// Step 2: Configure the Device
	// We want a "Duplex" device, meaning it handles both Capture (mic) and Playback (speaker) simultaneously.
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Duplex)
	// Set the format to 16-bit PCM. This is standard raw audio.
	deviceConfig.Capture.Format = malgo.FormatS16
	// 1 channel means Mono sound. 2 would be Stereo.
	deviceConfig.Capture.Channels = 1
	
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = 1
	
	// Sample rate: 44100 Hz (44.1 kHz) is standard CD quality. 
	deviceConfig.SampleRate = 44100

	// Step 3: Define the Audio Callback
	// This function is triggered automatically by the sound card hundreds of times per second.
	onRecvFrames := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		
		copy(pOutputSample, pInputSamples)
	}

	// Step 4: Initialize the Device with our configuration and callback
	deviceCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}

	device, err := malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		fmt.Println("Error initializing device:", err)
		os.Exit(1)
	}
	defer device.Uninit()

	// Step 5: Start the hardware
	err = device.Start()
	if err != nil {
		fmt.Println("Error starting device:", err)
		os.Exit(1)
	}

	// Keep the program running until the user presses Enter
	fmt.Println("Mic loopback started! Speak into your microphone and you will hear yourself.")
	fmt.Println("Press ENTER to quit...")
	var input string
	fmt.Scanln(&input)
}