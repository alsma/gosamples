package utils

import (
	"log"
	"os"
)

func CheckReadableDirectoryExists(directory string) {
	info, err := os.Stat(directory)

	if os.IsNotExist(err) || (info.Mode()&(1<<2) == 0) {
		log.Fatal("Data directory does not exist")
	}
}

func EnsureDirectoryExists(directory string, perm os.FileMode) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		// trying to create
		err := os.Mkdir(directory, perm)

		if nil != err {
			log.Fatalf("Unable to create directory, error: %s", err.Error())
		}
	}
}
