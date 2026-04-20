//go:build !linux

package wrapper

import "fmt"

func RunTUI(_ *Store, _ *Runner) error {
	return fmt.Errorf("tui mode is only available on linux")
}
