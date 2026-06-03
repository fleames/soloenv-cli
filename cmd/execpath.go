package cmd

import "os"

func executablePath() (string, error) {
	return os.Executable()
}
