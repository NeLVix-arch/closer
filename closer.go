package closer

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type closer struct {
	mu    sync.Mutex
	funcs []func() error
}

var instance *closer

func init() {
	instance = &closer{}
	go waitForSignal()
}

func Add(f func() error) {
	instance.mu.Lock()
	defer instance.mu.Unlock()
	instance.funcs = append(instance.funcs, f)
}

func close() error {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	var errs []error
	for _, f := range instance.funcs {
		if err := f(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func waitForSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Received termination signal. Closing resources...")
	if err := close(); err != nil {
		log.Printf("Error closing resources: %v\n", err)
	}
	log.Println("Resources closed. Exiting.")
	os.Exit(0)
}
