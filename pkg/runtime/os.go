package runtime

import (
	"os"
)

func OsArgs() []string {
	return os.Args
}

func OsGetenv(key string) string {
	return os.Getenv(key)
}

func OsExit(code int) {
	os.Exit(code)
}
