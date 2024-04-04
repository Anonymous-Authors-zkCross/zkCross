package experiment

import (
	"encoding/json"
	"fmt"
	kitlog "github.com/go-kit/kit/log"
	"io"
	"os"
	"time"
)

var syncWriter io.Writer

func InitOutputFile(outputFile string) {

	if outputFile == "" {
		_, _ = fmt.Fprintln(os.Stderr, "You must specify the log file for the experiment, e.g. \"geth --experiment.output=/path/to/file.txt\"")
		os.Exit(1)
	}

	var err error
	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Failed to create experiment log file.")
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	syncWriter = kitlog.NewSyncWriter(file)
	_, _ = fmt.Fprintln(os.Stderr, "Experiment log initialized.")
	err = Record(map[string]interface{}{"Message": "Experiment log initialized.", "Type": "Message"})

	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Failed to write experiment log file.")
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func Record(experimentLog map[string]interface{}) error {
	experimentLog["Timestamp"] = time.Now().UnixNano()

	b, err := json.Marshal(experimentLog)
	if err != nil {
		return err
	}

	b = append(b, '\n')

	if syncWriter == nil {
		_, _ = fmt.Fprintln(os.Stderr, "Warning! Experiment log has NOT been initialized yet.")
		_, _ = fmt.Fprintln(os.Stderr, string(b))
	} else {
		_, err = syncWriter.Write(b)
		if err != nil {
			return err
		}
	}

	return nil
}
