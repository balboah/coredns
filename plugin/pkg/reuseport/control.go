package reuseport

import "syscall"

func chainControls(funcs ...func(network, address string, c syscall.RawConn) error) func(network, address string, c syscall.RawConn) error {
	filtered := make([]func(network, address string, c syscall.RawConn) error, 0, len(funcs))
	for _, fn := range funcs {
		if fn != nil {
			filtered = append(filtered, fn)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return func(network, address string, c syscall.RawConn) error {
		for _, fn := range filtered {
			if err := fn(network, address, c); err != nil {
				return err
			}
		}
		return nil
	}
}
