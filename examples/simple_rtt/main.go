package main

import (
	"device/arm"
	"fmt"
	"github.com/m-chichikalov/segger/rtt"
	"machine"
	"time"
)

func main() {

	rtt.InitRtt(64, 8)

	log := rtt.NewTerminal(0)
	error := rtt.NewTerminal(1)

	i := 0

	for {
		i += 200

		s := fmt.Sprintf("%dms after start\n", i)
		log.WriteString(s)

		time.Sleep(200 * time.Millisecond)

		if i >= 10000 {
			error.Write([]byte("Err: 10s limit reached.\n"))
			arm.Asm("bkpt")
		}
	}
}
