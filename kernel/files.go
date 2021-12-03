package kernel

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func copyDir(dest string, source string) error {
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)
	objects, err := directory.Readdir(-1)
	if err != nil {
		return err
	}

	for _, obj := range objects {
		sourcefilepointer := filepath.Join(source, obj.Name())
		destinationfilepointer := filepath.Join(dest, obj.Name())

		if obj.IsDir() {
			// create sub-directories - recursively
			err = copyDir(destinationfilepointer, sourcefilepointer)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			// perform copy
			err = copyFile(destinationfilepointer, sourcefilepointer)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	return nil
}

func copyFile(dest string, source string) error {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
