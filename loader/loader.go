package loader

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
)

type Loader struct {
	PluginsDir string
	ObjectsDir string
}

// InitLoader
func InitLoader() (*Loader, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("could not find current directory: %v", err)
	}
	//fmt.Println(dir)
	pluginsDir := filepath.Join(dir, "plugins")
	//fmt.Println(pluginsDir)
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, fmt.Errorf("could not create objects dir: %v", err)
	}

	return &Loader{PluginsDir: pluginsDir, ObjectsDir: tempDir}, nil
}

// Destroy
func (l *Loader) Destroy() {
	os.RemoveAll(l.PluginsDir)
}

// Plugins
func (l *Loader) Plugins() []string {
	f, err := os.Open(l.PluginsDir)
	if err != nil {
		log.Printf("open plugins dir err: %v", err)
		return nil
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	if err != nil {
		log.Printf("read dir names err: %v", err)
		return nil
	}

	var res []string
	for _, name := range names {
		if filepath.Ext(name) == ".go" {
			res = append(res, name)
		}
	}
	return res
}

// CompileAndRun
func (l *Loader) CompileAndRun(name string) error {
	obj, err := l.compile(name)

	if err != nil {
		return fmt.Errorf("could not compile %s: %v", name, err)
	}
	defer os.Remove(obj)

	if err := l.call(obj); err != nil {
		return fmt.Errorf("could not run %s: %v", obj, err)
	}

	return nil
}

// call
func (l *Loader) call(object string) error {
	p, err := plugin.Open(object)
	if err != nil {
		return fmt.Errorf("could not open %s: %v", object, err)
	}
	run, err := p.Lookup("Run")
	if err != nil {
		return fmt.Errorf("could not find Run function: %v", err)
	}
	runFunc, ok := run.(func() error)
	if !ok {
		return fmt.Errorf("found Run but type is %T instead of func() error", run)
	}
	if err := runFunc(); err != nil {
		return fmt.Errorf("plugin failed with error %v", err)
	}
	return nil
}

// compile
func (l *Loader) compile(name string) (string, error) {
	f, err := ioutil.ReadFile(filepath.Join(l.PluginsDir, name))
	if err != nil {
		return "", fmt.Errorf("could not read %s: %v", name, err)
	}

	name = fmt.Sprintf("%d.go", rand.Int())
	srcPath := filepath.Join(l.ObjectsDir, name)
	if err := ioutil.WriteFile(srcPath, f, 0666); err != nil {
		return "", fmt.Errorf("could not write %s: %v", name, err)
	}

	objectPath := srcPath[:len(srcPath)-3] + "so"

	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o="+objectPath, srcPath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("could not compile %s: %v", name, err)
	}
	return objectPath, nil
}
