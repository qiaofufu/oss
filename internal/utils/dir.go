package utils

import "os"

// if dir not exists, create it
func CreateDirIfNotExists(dir string) error {
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
