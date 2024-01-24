package testinfra

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

var crdsInitialized = false

var (
	dirSkr    string
	dirKcp    string
	dirGarden string

	crdMutex sync.Mutex
)

func initCrds() error {
	crdMutex.Lock()
	defer crdMutex.Unlock()

	if crdsInitialized {
		return nil
	}

	localBin := os.Getenv("LOCALBIN")
	if len(localBin) == 0 {
		localBin = filepath.Join("..", "..", "bin")
	}

	dirSkr = path.Join(localBin, "cloud-manager", "skr")
	dirKcp = path.Join(localBin, "cloud-manager", "kcp")
	dirGarden = path.Join(localBin, "cloud-manager", "garden")

	// recreate destination directories
	if err := createDir(dirSkr); err != nil {
		return err
	}
	if err := createDir(dirKcp); err != nil {
		return err
	}
	if err := createDir(dirGarden); err != nil {
		return err
	}

	// copy CRDs to destination dirs
	configCrdBases := filepath.Join("..", "..", "config", "crd", "bases")
	list, err := os.ReadDir(configCrdBases)
	if err != nil {
		return fmt.Errorf("error reading files in config/crd/bases dir: %w", err)
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
				err := copyFile(
					filepath.Join(configCrdBases, x.Name()),
					filepath.Join(dir, x.Name()),
				)
				if err != nil {
					return err
				}
				break
			}
		}
	}

	// generate all gardener CRDs
	cmd := exec.Command(
		filepath.Join(localBin, "controller-gen"),
		"crd:allowDangerousTypes=true",
		fmt.Sprintf(`paths="%s/pkg/mod/github.com/gardener/gardener@v1.85.0/pkg/apis/core/v1beta1"`, os.Getenv("GOPATH")),
		fmt.Sprintf("output:crd:artifacts:config=%s", dirGarden),
	)
	err = cmd.Run()
	if err != nil {
		return err
	}

	// delete all garedener CRDs except few we want to keep
	gardenFilesToKeep := map[string]struct{}{
		"core.gardener.cloud_shoots.yaml":         {},
		"core.gardener.cloud_secretbindings.yaml": {},
	}
	list, err = os.ReadDir(dirGarden)
	if err != nil {
		return err
	}
	for _, f := range list {
		_, keep := gardenFilesToKeep[f.Name()]
		if keep {
			continue
		}
		err = os.Remove(filepath.Join(dirGarden, f.Name()))
		if err != nil {
			return err
		}
	}

	crdsInitialized = true
	return nil
}

func createDir(dir string) error {
	_, err := os.Stat(dir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		// dir exists, remove it first, so it gets created empty
		err = os.RemoveAll(dir)
		if err != nil {
			return err
		}
	}
	if err := os.MkdirAll(dir, 0777); err != nil {
		return err
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
