package cli

import (
	"errors"
	"fmt"
	"strings"

	"pxcli/internal/daemon"
)

func formatDaemonError(err error) error {
	if err == nil {
		return nil
	}
	var daemonErr daemon.Error
	if errors.As(err, &daemonErr) {
		if strings.TrimSpace(daemonErr.Message) == "" {
			return fmt.Errorf("err %s", daemonErr.Code)
		}
		return fmt.Errorf("err %s %s", daemonErr.Code, daemonErr.Message)
	}
	return err
}
