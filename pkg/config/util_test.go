package config

import (
	"os"
)

func copyFile(src, dest string) error {
	buf, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	err = os.WriteFile(dest, buf, 0644)
	if err != nil {
		return err
	}
	return nil
}
