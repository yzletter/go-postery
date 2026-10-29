package utils

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func ListenTermSignal(f func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	sig := <-c
	log.Println("receive term signal " + sig.String() + ", going to exit")
	f()
	os.Exit(0)
}
