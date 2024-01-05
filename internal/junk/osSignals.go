package junk

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"os/signal"
)

// ShutdownOnOSSignals will install OS signal handles, calling onDone when a shutdown signal is invoked.
func ShutdownOnOSSignals(onDone func()) {
	processSignals := make(chan os.Signal, 8)
	signal.Notify(processSignals, unix.SIGTERM, unix.SIGINT)
	go func() {
		for sig := range processSignals {
			switch sig {
			case unix.SIGTERM:
				fmt.Printf("<<proc>> Received SIGTERM, shutting down.\n")
				onDone()
			case unix.SIGINT:
				fmt.Printf("<<proc>> Received SIGINT, shutting down.\n")
				onDone()
			}
		}
	}()
}
