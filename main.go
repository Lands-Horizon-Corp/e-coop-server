package main

import (
	_ "net/http/pprof"

	"github.com/Lands-Horizon-Corp/e-coop-server/cmd"
)

func main() {
	cmd.Execute()
}
