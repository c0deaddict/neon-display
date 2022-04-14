package display

import "os/exec"

func setDisplayPower(state bool) error {
	s := "0"
	if state {
		s = "1"
	}

	return exec.Command("vcgencmd", "display_power", s).Run()
}
