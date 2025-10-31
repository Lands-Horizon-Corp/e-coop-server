package main

import (
	_ "net/http/pprof" // #nosec G108 -- profiling endpoint intentionally imported for debugging

	"github.com/Lands-Horizon-Corp/e-coop-server/cmd"
)

func main() {
	cmd.Execute()
}
