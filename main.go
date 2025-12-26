package main

import (
	"log"
	"os"
	"runtime/pprof"

	"net/http"
	_ "net/http/pprof"

	"github.com/Lands-Horizon-Corp/e-coop-server/cmd"
	"github.com/fatih/color"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	color.Blue("Starting E-Coop Server CLI...")
	cmd.Execute()

	f, err := os.Create("startup-heap.pprof")
	if err != nil {
		log.Println("Could not create heap profile:", err)
		return
	}
	defer f.Close()

	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Println("Could not write heap profile:", err)
		return
	}

	color.Green("Startup heap profile written to startup-heap.pprof")
}
