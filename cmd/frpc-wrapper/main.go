package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fatedier/frp/pkg/wrapper"
)

func main() {
	mode := flag.String("mode", wrapper.DefaultMode(), "wrapper mode: gui or tui")
	flag.Parse()

	store, err := wrapper.NewStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "init store: %v\n", err)
		os.Exit(1)
	}
	runner := wrapper.NewRunner()

	switch *mode {
	case "gui":
		gui := wrapper.NewGUI(store, runner)
		if err := gui.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "gui error: %v\n", err)
			os.Exit(1)
		}
	case "tui":
		if err := wrapper.RunTUI(store, runner); err != nil {
			fmt.Fprintf(os.Stderr, "tui error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown mode %q\n", *mode)
		os.Exit(1)
	}
}
