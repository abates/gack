package gack

import (
	"fmt"
	"sort"
)

var configFile = "gack.yml"

type Context struct {
	Target *Match
	Config *Config
}

func (c *Context) Param(name string) string {
	return c.Target.Param(name)
}

type DefaultConfigurable interface {
	DefaultConfig() (string, interface{})
}

type ExecuteFunc func(*Context) error

func (f ExecuteFunc) Execute(context *Context) error {
	return f(context)
}

type Executable interface {
	Execute(*Context) error
}

type Target struct {
	Executable
	pattern      Pattern
	dependencies []string
}

type Mux struct {
	Config      *Config
	targets     map[string]*Target
	targetNames []string
}

func NewMux() (mux *Mux, err error) {
	mux = &Mux{
		targets: make(map[string]*Target),
	}

	mux.Config, err = ReadConfigFile(configFile)
	return mux, err
}

func (mux *Mux) Lookup(input string) (*Target, []string, *Match) {
	for _, targetName := range mux.targetNames {
		target := mux.targets[targetName]
		if match := target.pattern.Match(input); match.Matches() {
			return target, target.dependencies, &match
		}
	}
	return nil, nil, nil
}

func (mux *Mux) AddDependency(pattern, dependency string, executable Executable, dependencies ...string) error {
	target := mux.targets[pattern]
	if target == nil {
		target, _ = mux.Register(pattern, nil)
	}

	if executable != nil {
		mux.Register(dependency, executable, dependencies...)
	}
	target.dependencies = append(target.dependencies, dependency)
	return nil
}

func (mux *Mux) Register(pattern string, executable Executable, dependencies ...string) (target *Target, err error) {
	var found bool
	if target, found = mux.targets[pattern]; found {
		target.Executable = executable
		target.dependencies = append(target.dependencies, dependencies...)
	} else {
		target = &Target{executable, NewPattern(pattern), dependencies}
		mux.targets[pattern] = target
		mux.targetNames = append(mux.targetNames, pattern)
		sort.Sort(sort.Reverse(sort.StringSlice(mux.targetNames)))
	}
	return target, nil
}

func (mux *Mux) Execute(subject string) (err error) {
	target, dependencies, match := mux.Lookup(subject)
	if target == nil {
		err = fmt.Errorf("No targets match %v", subject)
	} else {
		for _, dependency := range dependencies {
			str := match.Interpolate(dependency)
			err = mux.Execute(str)
			if err != nil {
				break
			}
		}

		if err == nil && target.Executable != nil {
			err = target.Execute(&Context{
				Target: match,
			})
		}
	}
	return err
}

func (mux *Mux) TargetNames() []string {
	return mux.targetNames
}
