package ui

import "errors"

var (
	ErrRestartRequested = errors.New("restart requested")
	ErrExitRequested    = errors.New("exit requested")
)
