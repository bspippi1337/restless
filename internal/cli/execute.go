package cli

import "os"

func MustExecute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
