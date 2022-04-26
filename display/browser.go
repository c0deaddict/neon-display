package display

import (
	"bufio"
	"io"
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"
)

func (d *Display) startBrowser(url string) (*os.Process, error) {
	cmd := exec.Command(d.config.FirefoxBin, "-kiosk", url)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	go browserLogger(stdout, "stdout")
	go browserLogger(stderr, "stderr")

	return cmd.Process, nil
}

func browserLogger(out io.ReadCloser, name string) {
	in := bufio.NewScanner(out)
	for in.Scan() {
		log.Info().Str("line", in.Text()).Msgf("browser %s", name)
	}

	if err := in.Err(); err != nil {
		log.Error().Err(err).Msgf("browser %s", name)
	}
}

// func start() {
// 	// Start browser process.
// 	url := fmt.Sprintf("http://localhost:%d", d.config.WebPort)
// 	p, err := d.startBrowser(url)
// 	if err != nil {
// 		// NOTE: deferred's aren't run on Fatal..
// 		log.Fatal().Err(err).Msg("start browser")
// 	}
// 	go func() {
// 		state, err := p.Wait()
// 		if err != nil {
// 			log.Fatal().Err(err).Msg("browser process wait")
// 		}
// 		log.Fatal().Msgf("browser exited with %v", state.ExitCode())
// 	}()
// 	// Stop the browser at exit.
// 	defer p.Kill()
// }
