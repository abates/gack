package pkg

import (
	"fmt"
	"os"

	"github.com/abates/gack"
	"github.com/abates/gack/build"
	"github.com/xor-gate/debpkg"
)

type debPackager struct {
	config *gack.Config
}

func NewDebPackager(mux *gack.Mux) *debPackager {
	return &debPackager{
		config: mux.Config,
	}
}

func (p *debPackager) clean(context *gack.Context) error {
	fmt.Printf("Cleaning pkg/deb/*\n")
	return os.RemoveAll("pkg/deb/")
}

func (p *debPackager) Execute(context *gack.Context) error {
	fmt.Printf("Packaging %s\n", context.Target.Subject())
	deb := debpkg.New()
	defer deb.Close()

	deb.SetName(p.config.PackageName)
	deb.SetVersion(p.config.Version)
	deb.SetArchitecture(context.Param("architecture"))
	deb.SetMaintainer(p.config.Maintainer)
	deb.SetMaintainerEmail(p.config.MaintainerEmail)
	deb.SetHomepage(p.config.Homepage)

	deb.SetShortDescription(p.config.ShortDescription)
	deb.SetDescription(p.config.Description)

	err := os.MkdirAll("pkg/deb/", 0755)
	if err == nil {
		deb.AddFile(fmt.Sprintf("build/%s", build.Name(p.config.PackageName, "linux", context.Param("architecture"))))
		err = deb.Write(fmt.Sprintf("pkg/deb/%s_%s_%s.deb", p.config.PackageName, p.config.Version, context.Param("architecture")))
	}
	return err
}

func (p *debPackager) Register(mux *gack.Mux) error {
	pkg := mux.Config.PackageName
	for _, dependency := range build.Dependencies(mux) {
		if dependency.Platform == "linux" {
			mux.AddDependency("pkg/deb", fmt.Sprintf("pkg/deb/%s_%s_%s.deb", pkg, mux.Config.Version, dependency.Arch), nil)
		}
	}

	mux.Register("pkg/deb/:package_:version_:architecture.deb", p, "build/:package_linux_:architecture")
	mux.AddDependency("clean/pkg", "clean/pkg/deb", gack.ExecuteFunc(p.clean))
	return nil
}

func Register(mux *gack.Mux) {
	NewDebPackager(mux).Register(mux)
	mux.AddDependency("clean", "clean/pkg", gack.ExecuteFunc(func(*gack.Context) error {
		fmt.Printf("Cleaning pkg/*\n")
		return os.RemoveAll("pkg/")
	}))
}
