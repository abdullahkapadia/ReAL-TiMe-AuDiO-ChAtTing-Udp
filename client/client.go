package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/gen2brain/malgo"
	"github.com/zaf/g711"
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
	fmt.Println("Connected to", addr)

	// Global mute state
	var isMuted bool = true
	var mu sync.Mutex

	// HTTP Server for UI
	http.HandleFunc("/api/toggle", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			mu.Lock()
			isMuted = !isMuted
			mu.Unlock()
			w.WriteHeader(http.StatusOK)
		}
	})
	
	// Serve static files from "public" directory
	http.Handle("/", http.FileServer(http.Dir("./public")))

	go func() {
		fmt.Println("UI Dashboard running at http://localhost:3000")
		if err := http.ListenAndServe(":3000", nil); err != nil {
			log.Fatal(err)
		}
	}()

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
	var sequenceNumber uint32 = 0
	onRecvFrames := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		mu.Lock()
		muted := isMuted
		mu.Unlock()

		if muted {
			return // Microphone is off
		}

		// Compress 16-bit PCM to 8-bit G.711 u-law
		compressedAudio := g711.EncodeUlaw(pInputSamples)
		
		// Create a custom packet: [4-byte Sequence] + [Compressed Audio]
		packet := make([]byte, 4+len(compressedAudio))
		binary.LittleEndian.PutUint32(packet[0:4], sequenceNumber)
		copy(packet[4:], compressedAudio)
		
		// Send custom packet directly over UDP
		_, err := conn.Write(packet)
		if err != nil {
			fmt.Println("Error sending audio:", err)
		}
		
		sequenceNumber++
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
