package cloud

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"

	"github.com/elliotchance/pie/v2"
)

//go:embed tf.tf
var fs embed.FS

type TFWorkspace interface {
	Create() error
	Init() error
	Plan() error
	Apply() error
	Outputs() map[string]any
	Destroy() error

	Out() string
}

type tfWorkspace struct {
	rootDir string
	name    string
	tfCmd   string
	env     map[string]string

	data TfTemplateData

	out string

	outputs map[string]interface{}

	_dir string
}

type TfTemplateData struct {
	Providers []TfTemplateProvider
	Module    TfTemplateModule
}

type TfTemplateProvider struct {
	Name    string
	Source  string
	Version string
}

type TfTemplateModule struct {
	Source    string
	Version   string
	Variables map[string]string
}

func (w *tfWorkspace) pluginCacheDir() string {
	return path.Join(w.rootDir, ".plugin_cache_dir")
}

func (w *tfWorkspace) xdgBaseDir() string {
	return path.Join(w.rootDir, ".config")
}

func (w *tfWorkspace) dir() string {
	if w._dir == "" {
		w._dir = path.Join(w.rootDir, w.name)
	}
	return w._dir
}

func (w *tfWorkspace) Out() string {
	return w.out
}

func (w *tfWorkspace) Create() error {
	txt, err := fs.ReadFile("tf.tf")
	if err != nil {
		return fmt.Errorf("could not read embedded .tf file: %w", err)
	}

	tpl := template.Must(template.New("tf.tf").Parse(string(txt)))
	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, w.data)
	if err != nil {
		return fmt.Errorf("could not execute embedded .tf file: %w", err)
	}

	err = os.MkdirAll(w.dir(), 0755)
	if err != nil {
		return fmt.Errorf("could not create workspace directory: %w", err)
	}

	err = os.WriteFile(path.Join(w.dir(), "main.tf"), buf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("could not write workspace main.tf file: %w", err)
	}

	if err := os.MkdirAll(w.pluginCacheDir(), 0755); err != nil {
		return fmt.Errorf("could not create plugin cache directory: %w", err)
	}
	configDir := path.Join(w.xdgBaseDir(), "opentofu")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("could not create xdg opentofu directory: %w", err)
	}

	rcTpl := `
plugin_cache_dir = "%s"
`
	rc := fmt.Sprintf(rcTpl, w.pluginCacheDir())

	err = os.WriteFile(path.Join(configDir, "tofurc"), []byte(rc), 0644)
	if err != nil {
		return fmt.Errorf("could not write workspace tofurc file: %w", err)
	}
	err = os.WriteFile(path.Join(configDir, "terraformrc"), []byte(rc), 0644)
	if err != nil {
		return fmt.Errorf("could not write workspace terraformrc file: %w", err)
	}

	return nil
}

func (w *tfWorkspace) getEnv() []string {
	result := append([]string{}, os.Environ()...)
	result = append(result, fmt.Sprintf("XDG_CONFIG_HOME=%s", w.xdgBaseDir()))
	for k, v := range w.env {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

func (w *tfWorkspace) credsToOut() {
	w.out = fmt.Sprintf("%s\n\nenv credentials:\n%s\n", w.out, pie.Join(w.getEnv(), "\n"))
}

func (w *tfWorkspace) Init() error {
	cmd := exec.Command(w.tfCmd, "init")
	cmd.Dir = w.dir()
	cmd.Env = w.getEnv()
	txt, err := cmd.CombinedOutput()
	w.out = fmt.Sprintf("%s\n\nINIT:\n%s", w.out, string(txt))
	if err != nil {
		return fmt.Errorf("could not initialize tf: %w\n%s", err, string(txt))
	}
	return nil
}

func (w *tfWorkspace) Plan() error {
	w.credsToOut()
	cmd := exec.Command(w.tfCmd, "plan")
	cmd.Dir = w.dir()
	cmd.Env = w.getEnv()
	txt, err := cmd.CombinedOutput()
	w.out = fmt.Sprintf("%s\n\nPLAN:\n%s", w.out, string(txt))
	if err != nil {
		return fmt.Errorf("could not initialize tf: %w\n%s", err, string(txt))
	}
	return nil
}

type TfOutput struct {
	Values struct {
		Value map[string]interface{} `json:"value"`
	} `json:"values"`
}

func (w *tfWorkspace) Apply() error {
	cmd := exec.Command(w.tfCmd, "apply", "-auto-approve")
	cmd.Dir = w.dir()
	cmd.Env = w.getEnv()
	txt, err := cmd.CombinedOutput()
	w.out = fmt.Sprintf("%s\n\nAPPLY:\n%s", w.out, string(txt))
	if err != nil {
		return fmt.Errorf("could not apply tf: %w\n%s", err, string(txt))
	}

	cmd = exec.Command(w.tfCmd, "output", "-json")
	cmd.Dir = w.dir()
	cmd.Env = w.getEnv()

	txt, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not get tf outputs: %w\n%s", err, string(txt))
	}

	data := &TfOutput{}
	err = json.Unmarshal(txt, &data)
	if err != nil {
		return fmt.Errorf("could not unmarshal tf outputs: %w\n%s", err, string(txt))
	}

	w.outputs = data.Values.Value

	return nil
}

func (w *tfWorkspace) Outputs() map[string]any {
	return w.outputs
}

func (w *tfWorkspace) Destroy() error {
	if _, err := os.Stat(w.dir()); os.IsNotExist(err) {
		return nil
	}

	cmd := exec.Command(w.tfCmd, "destroy", "-auto-approve")
	cmd.Dir = w.dir()
	cmd.Env = w.getEnv()
	txt, err := cmd.CombinedOutput()
	w.out = fmt.Sprintf("%s\n\nDESTROY:\n%s", w.out, string(txt))

	ignoreErrors := []string{
		"Required plugins are not installed",
	}

	if err != nil {
		shouldIgnore := false
		for _, ig := range ignoreErrors {
			if strings.Contains(string(txt), ig) {
				shouldIgnore = true
			}
		}
		if !shouldIgnore {
			return fmt.Errorf("could not destroy tf: %w\n%s", err, string(txt))
		}
	}

	return os.RemoveAll(w.dir())
}
