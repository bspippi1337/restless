package version

import (
	"fmt"
	"runtime"
)

var (
	Version = "0.4.0"
	Commit  = "dev"
	Date    = "unknown"
)

func Short() string {
	return Version
}

func String() string {
	return fmt.Sprintf("restless %s (%s %s)", Version, Commit, Date)
}

func Details() string {
	return fmt.Sprintf("restless %s\ncommit: %s\nbuilt: %s\ngo: %s\nplatform: %s/%s\n", Version, Commit, Date, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
