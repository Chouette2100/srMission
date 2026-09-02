package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func installSIGTERMTracebackHandler() {
	installSignalTracebackHandlers()
}

func installSignalTracebackHandlers() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

	go func() {
		for {
			sig := <-sigCh
			log.Printf("received signal: %s\n", sig.String())
			dumpAllGoroutineStacks()

			switch sig {
			case syscall.SIGINT:
				log.Printf("terminating after SIGINT traceback dump\n")
				os.Exit(128 + 2)
			case syscall.SIGTERM:
				log.Printf("terminating after SIGTERM traceback dump\n")
				// Keep SIGTERM semantics: terminate process after dumping trace.
				os.Exit(128 + 15)
			case syscall.SIGUSR1:
				log.Printf("continuing after SIGUSR1 traceback dump\n")
			}
		}
	}()
}

func dumpAllGoroutineStacks() {
	buf := make([]byte, 1<<20) // 1 MiB
	n := runtime.Stack(buf, true)
	for n == len(buf) {
		buf = make([]byte, len(buf)*2)
		n = runtime.Stack(buf, true)
		if len(buf) >= (16 << 20) {
			break
		}
	}

	log.Printf("========== goroutine traceback (%s) =========="+"\n", time.Now().Format(time.RFC3339Nano))
	log.Printf("%s\n", string(buf[:n]))
	log.Printf("========== end goroutine traceback ==========\n")
}
