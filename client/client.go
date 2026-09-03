package main

import (
	"encoding/binary"
	"encoding/json"
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

	// Global mute state and Room ID
	var isMuted bool = true
	var currentRoomID uint32 = 123
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

	http.HandleFunc("/api/room", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var req struct {
				RoomID uint32 `json:"roomId"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
				mu.Lock()
				currentRoomID = req.RoomID
				mu.Unlock()
				w.WriteHeader(http.StatusOK)
			} else {
				http.Error(w, "Bad request", http.StatusBadRequest)
			}
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

	// Configure Duplex Device (Capture & Playback)
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Duplex)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = 1
	deviceConfig.SampleRate = 44100

	// Channel to pass audio from network to soundcard
	audioChan := make(chan []byte, 100)

	// Audio Callback
	var sequenceNumber uint32 = 0
	onRecvFrames := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		// --- SPEAKER PLAYBACK LOGIC ---
		select {
		case audioData := <-audioChan:
			copy(pOutputSample, audioData)
		default:
			// Fill with silence if no data
			for i := range pOutputSample {
				pOutputSample[i] = 0
			}
		}

		// --- MICROPHONE CAPTURE LOGIC ---
		mu.Lock()
		muted := isMuted
		roomID := currentRoomID
		mu.Unlock()

		if muted {
			return // Microphone is off
		}

		// Compress 16-bit PCM to 8-bit G.711 u-law
		compressedAudio := g711.EncodeUlaw(pInputSamples)

		// Create a custom packet: [4-byte RoomID] + [4-byte Sequence] + [Compressed Audio]
		packet := make([]byte, 8+len(compressedAudio))
		binary.LittleEndian.PutUint32(packet[0:4], roomID)
		binary.LittleEndian.PutUint32(packet[4:8], sequenceNumber)
		copy(packet[8:], compressedAudio)
		
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

	// UDP Receiver Goroutine
	go func() {
		buffer := make([]byte, 4096)
		var lastReceivedSeq uint32 = 0
		
		for {
			n, err := conn.Read(buffer)
			if err != nil {
				continue
			}
			if n < 8 {
				continue
			}
			
			// buffer[0:4] is RoomID (skip it since we are already in the room)
			incomingSeq := binary.LittleEndian.Uint32(buffer[4:8])
			
			// Drop out-of-order packets (Jitter Fix)
			if incomingSeq > lastReceivedSeq || lastReceivedSeq == 0 {
				lastReceivedSeq = incomingSeq
				
				// Decode the 8-bit G.711 back into 16-bit PCM
				decompressedAudio := g711.DecodeUlaw(buffer[8:n])
				
				select {
				case audioChan <- decompressedAudio:
				default:
					// Drop if speaker buffer is full
				}
			}
		}
	}()

	fmt.Println("Recording and streaming... Press ENTER to stop.")
	var input string
	fmt.Scanln(&input)
}
