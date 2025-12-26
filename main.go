package main

import (
	"os"
	"runtime/pprof"

	"github.com/Lands-Horizon-Corp/e-coop-server/cmd"
	"github.com/fatih/color"
)

func main() {
	f, _ := os.Create("startup-heap.pprof")
	pprof.WriteHeapProfile(f)
	f.Close()

	color.Blue("Starting E-Coop Server CLI...")
	cmd.Execute()
}
