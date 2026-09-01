package main

import(
	//"fmt"
	"net"
	"log"
)


func main(){
	addr , err := net.ResolveUDPAddr("udp","localhost:8080")

	if err != nil {
		log.Fatal(err)
	}

	conn , err := net.DialUDP("udp",nil,addr)

	defer conn.Close()
	if err != nil {
		log.Fatal(err)
	}

	message := []byte("Hello World")
	_, err = conn.Write(message)
	if err != nil {
        log.Printf("Send failed: %v", err)
        return
    }

}