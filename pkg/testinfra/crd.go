package testinfra

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strconv"
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

	prefix := ""
	if inPipeline, err := strconv.ParseBool(os.Getenv("PIPELINE")); err == nil && inPipeline {
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		prefix = fmt.Sprintf("-%d", rnd.Uint32())
	}

	dirSkr = path.Join(projectRoot, "bin", "cloud-manager", "skr"+prefix)
	dirKcp = path.Join(projectRoot, "bin", "cloud-manager", "kcp"+prefix)
	dirGarden = path.Join(projectRoot, "bin", "cloud-manager", "garden"+prefix)

	// recreate destination directories
	if err = createDir(dirSkr); err != nil {
		err = fmt.Errorf("error creating SKR dir: %w", err)
		return
	}
	if err = createDir(dirKcp); err != nil {
		err = fmt.Errorf("error creating KCP dir: %w", err)
		return
	}
	if err = createDir(dirGarden); err != nil {
		err = fmt.Errorf("error creating Garden dir: %w", err)
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
		return nil
	}
	if err := os.MkdirAll(dir, 0777); err != nil {
		return fmt.Errorf("error creating dir: %w", err)
	}
	return nil
}

func copyFile(src, dest string) error {
	s, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", src, err)
	}
	defer s.Close()

	dest1 := dest + ".tmp"
	d, err := os.Create(dest1)
	if err != nil {
		return fmt.Errorf("error creating tmp destination %s: %w", dest1, err)
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	if err != nil {
		return fmt.Errorf("error copying to file %s: %w", dest1, err)
	}

	err = os.Rename(dest1, dest)
	if err != nil {
		return fmt.Errorf("error renaming destination %s: %w", dest, err)
	}

	return nil
}
