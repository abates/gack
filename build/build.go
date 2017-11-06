package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/abates/gack"
)

type builder struct {
	hasDependencies bool
}

type Executor interface {
	CombinedOutput() ([]byte, error)
}

func execCommand(name string, args ...string) Executor {
	return exec.Command(name, args...)
}

var getExecCommand = execCommand

type Config struct {
	Platforms map[string][]string
}

func GetConfig(mux *gack.Mux) *Config {
	_, config := DefaultConfig()
	mux.Config.Get("build", config)
	return config
}

func DefaultConfig() (string, *Config) {
	return "build", &Config{
		Platforms: map[string][]string{
			"linux":       []string{"386", "amd64", "arm-5", "arm-6", "arm-7", "arm64", "mips", "mips64", "mips64le", "mipsle"},
			"windows-10":  []string{"386", "amd64"},
			"darwin-10.6": []string{"386", "amd64"},
		},
	}
}

func (b *builder) DefaultConfig() (string, interface{}) {
	return DefaultConfig()
}

func (b *builder) clean(*gack.Context) error {
	fmt.Printf("Cleaning build/*\n")
	return os.RemoveAll("build/")
}

func (b *builder) dependencies(*gack.Context) (err error) {
	if !b.hasDependencies {
		fmt.Printf("Installing build dependencies\n")
		_, err = b.execute("docker", "pull", "karalabe/xgo-latest")
		b.hasDependencies = true
	}
	return err
}

func ContextName(ctx *gack.Context) string {
	return Name(ctx.Param("package"), ctx.Param("platform"), ctx.Param("architecture"))
}

func Name(pkg, platform, arch string) string {
	return fmt.Sprintf("%s_%s_%s", pkg, platform, arch)
}

func (b *builder) Execute(ctx *gack.Context) error {
	fmt.Printf("Building %s\n", ctx.Target.Subject())
	target := fmt.Sprintf("--targets=%s/%s", ctx.Param("platform"), ctx.Param("architecture"))

	packageName := fmt.Sprintf("build/tmp_%s", ctx.Param("package"))
	_, err := b.execute("xgo", target, "-out", packageName, "./")
	var files []string
	if files, err = filepath.Glob(fmt.Sprintf("build/tmp_%s-%s*", ctx.Param("package"), ctx.Param("platform"))); err == nil {
		if len(files) == 1 {
			err = os.Rename(files[0], fmt.Sprintf("build/%s", ContextName(ctx)))
		} else {
			err = fmt.Errorf("Expected xgo to return only one file, but got %v", files)
		}
	}
	return err
}

func (b builder) execute(name string, args ...string) (string, error) {
	executor := getExecCommand(name, args...)
	output, err := executor.CombinedOutput()
	return string(output), err
}

type Dependency struct {
	Platform string
	Arch     string
}

func Dependencies(mux *gack.Mux) []Dependency {
	dependencies := []Dependency{}
	config := GetConfig(mux)

	for platform, archs := range config.Platforms {
		for _, arch := range archs {
			dependencies = append(dependencies, Dependency{platform, arch})
		}
	}

	return dependencies
}

func Register(mux *gack.Mux) {
	b := &builder{}

	pkg := mux.Config.PackageName

	for _, dependency := range Dependencies(mux) {
		mux.AddDependency("build", fmt.Sprintf("build/%s_%s_%s", pkg, dependency.Platform, dependency.Arch), nil)
	}

	mux.Register("build/:package_:platform_:architecture", b, "dependencies/build")
	mux.AddDependency("dependencies", "dependencies/build", gack.ExecuteFunc(b.dependencies))
	mux.AddDependency("clean", "clean/build", gack.ExecuteFunc(b.clean))
}
