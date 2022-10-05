/*
	Config file containing functions for the progressbar
*/

package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type progressBar struct {
	current int // current value of the progressbar
	total   int // total value of the progresssbar
}

func NewProgressBar(current int, total int) *progressBar {
	return &progressBar{
		current: current,
		total:   total,
	}
}

func (CDM *CDMConfig) initProgress(termSig chan struct{}) {
	// start a timer to rewrite the progress bar every 1 second
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		// rewrite the progress bar each second
		case <-ticker.C:
			for routine := 0; routine < len(CDM.progress); routine++ {

				// update the progress for this routine
				CDM.mutex.RLock()
				prog := *CDM.progress[routine]
				CDM.mutex.RUnlock()

				showProgress(routine, prog)
			}

			// bring the cursor back to the start of the line
			for routine := 0; routine < len(CDM.progress); routine++ {
				fmt.Print("\033[F")
			}

		// listen for signals from the terminal
		case <-termSig:
			for routine := 0; routine < len(CDM.progress); routine++ {
				CDM.mutex.RLock()
				prog := *CDM.progress[routine]
				CDM.mutex.RUnlock()
				showProgress(routine, prog)
			}
			return
		}
	}
}

func showProgress(routine int, prog progressBar) {
	// create a string that acts as the progress bar
	bar := strings.Builder{}

	// calculate percentage value from the routine's current and total progress
	percent := math.Round((float64(prog.current) / float64(prog.total)) * 100)

	// convert number of characters to display
	reached := int((percent / 100) * PROGRESS_SIZE)

	bar.WriteString("[")

	// write to bar until we reach counted chracters
	for i := 0; i < PROGRESS_SIZE; i++ {
		if i <= reached {
			bar.WriteString("=")
		} else {
			bar.WriteString(" ")
		}
	}

	bar.WriteString("]")
	bar.WriteString(fmt.Sprintf(" %v%%", percent))
	fmt.Printf("Goroutine %d  :  %s\n", routine+1, bar.String())
}
