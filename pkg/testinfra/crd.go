package testinfra

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var crdsInitialized = false

var (
	crdMutex sync.Mutex
)

func initCrds(projectRoot string) (dirSkr, dirKcp, dirGarden string, err error) {
	crdMutex.Lock()
	defer crdMutex.Unlock()

	if crdsInitialized {
		err = errors.New("the CRDs are already initialized")
		return
	}

	dirSkr = path.Join(projectRoot, "bin", "cloud-manager", "skr")
	dirKcp = path.Join(projectRoot, "bin", "cloud-manager", "kcp")
	dirGarden = path.Join(projectRoot, "bin", "cloud-manager", "garden")

	// recreate destination directories
	if err = createDir(dirSkr); err != nil {
		return
	}
	if err = createDir(dirKcp); err != nil {
		return
	}
	if err = createDir(dirGarden); err != nil {
		return
	}

	var list []os.DirEntry

	// copy cloud-manager CRDs to destination dirs
	{
		configCrdBases := filepath.Join(projectRoot, "config", "crd", "bases")
		list, err = os.ReadDir(configCrdBases)
		if err != nil {
			err = fmt.Errorf("error reading files in config/crd/bases dir: %w", err)
			return
		}

		prefixMap := map[string]string{
			"cloud-control.":   dirKcp,
			"cloud-resources.": dirSkr,
		}
		for _, x := range list {
			if x.IsDir() {
				continue
			}
			for prefix, dir := range prefixMap {
				if strings.HasPrefix(x.Name(), prefix) {
					err = copyFile(
						filepath.Join(configCrdBases, x.Name()),
						filepath.Join(dir, x.Name()),
					)
					if err != nil {
						return
					}
					break
				}
			}
		}
	}

	// copy gardener CRDs
	{
		gardenerCrdsDir := filepath.Join(projectRoot, "config", "crd", "gardener")
		list, err = os.ReadDir(gardenerCrdsDir)
		if err != nil {
			err = fmt.Errorf("error listing gardener crds: %w", err)
			return
		}
		for _, f := range list {
			err = copyFile(
				filepath.Join(gardenerCrdsDir, f.Name()),
				filepath.Join(dirGarden, f.Name()),
			)
			if err != nil {
				err = fmt.Errorf("error copying gardener crd %s: %w", f.Name(), err)
				return
			}
		}
	}

	// copy operator CRDs
	{
		operatorCrdsDir := filepath.Join(projectRoot, "config", "crd", "operator")
		list, err = os.ReadDir(operatorCrdsDir)
		if err != nil {
			err = fmt.Errorf("error listing operator crds: %w", err)
			return
		}
		for _, f := range list {
			err = copyFile(
				filepath.Join(operatorCrdsDir, f.Name()),
				filepath.Join(dirKcp, f.Name()),
			)
			if err != nil {
				err = fmt.Errorf("error copying operator crd %s: %w", f.Name(), err)
				return
			}
		}
	}

	crdsInitialized = true
	return
}

func createDir(dir string) error {
	_, err := os.Stat(dir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error getting dir stats: %w", err)
	}
	if err == nil {
		// dir exists, remove it first, so it gets created empty
		err = os.RemoveAll(dir)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			err = os.RemoveAll(dir)
		}
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			err = os.RemoveAll(dir)
		}
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			err = os.RemoveAll(dir)
		}
		if err != nil {
			return fmt.Errorf("error removing dir: %w", err)
		}
	}
	if err := os.MkdirAll(dir, 0777); err != nil {
		return fmt.Errorf("error creating dir: %w", err)
	}
	return nil
}

func copyFile(src, dest string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	return err
}
