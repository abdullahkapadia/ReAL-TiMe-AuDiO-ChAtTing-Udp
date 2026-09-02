package main

import(
	"fmt"
	"net"
	"log"
	"os"
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

	if len(os.Args) < 2 {
		log.Fatal("Please provide a message as an argument.")
	}
	b := []byte(os.Args[1])

	cc, wrerr := conn.Write(b)
	if wrerr != nil {
		fmt.Printf("conn.Write() error: %s\n", wrerr)
	} else {
		fmt.Printf("Wrote %d bytes to socket\n", cc)
		c := make([]byte, cc+10)
		cc, rderr := conn.Read(c)
		if rderr != nil {
			fmt.Printf("conn.Read() error: %s\n", rderr)
		} else {
			fmt.Printf("Read %d bytes from socket\n", cc)
			fmt.Printf("Bytes: %q\n", string(c[0:cc]))
		}
	}

	if err = conn.Close(); err != nil {
		log.Fatal(err)
	}
}
