package gack

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

type configurableTestTarget struct {
	execute   func(*Context) error
	configure func(*Mux) error
}

func (c configurableTestTarget) Execute(ctx *Context) error { return c.execute(ctx) }

func (c configurableTestTarget) Configure(mux *Mux) error             { return c.configure(mux) }
func (c configurableTestTarget) DefaultConfig() (string, interface{}) { return "", nil }

type testTarget struct {
	name               string
	err                string
	configure          func(*testTarget, *Mux) error
	execute            func(*testTarget, *Context) error
	expectedConfigured []string
	expectedCalled     []string
	dependencies       []string
}

func newTestTarget(name, err string, execute func(*testTarget, *Context) error, expectedCalled, dependencies []string) *testTarget {
	return &testTarget{
		name:           name,
		err:            err,
		execute:        execute,
		dependencies:   dependencies,
		expectedCalled: expectedCalled,
	}
}

func (t *testTarget) target(called *[]string) Executable {
	execute := func(callback func(*testTarget, *Context) error) func(context *Context) error {
		return func(context *Context) error {
			*called = append(*called, t.name)
			if callback != nil {
				return callback(t, context)
			}
			return nil
		}
	}

	return ExecuteFunc(execute(t.execute))
}

func register(tests map[string]*testTarget, called *[]string, mux *Mux, name string, tt *testTarget) *Target {
	if tt == nil {
		tt = &testTarget{
			name: name,
		}
	}

	for _, dependency := range tt.dependencies {
		if tests[dependency] == nil {
			mux.AddDependency(name, dependency, nil)
		} else {
			mux.AddDependency(name, dependency, tests[dependency].target(called))
		}
	}
	target, _ := mux.Register(name, tt.target(called))
	return target
}

func TestGack(t *testing.T) {
	goodExecute := func(target *testTarget, context *Context) error {
		if context.Target.Subject() != target.name {
			t.Errorf("Expected target name %s but got %v", target.name, context.Target.Subject())
		}

		if context.Param("*") != target.name {
			t.Errorf("Expected param %s but got %v", target.name, context.Param("*"))
		}

		return nil
	}

	badExecute := func(target *testTarget, context *Context) error {
		return fmt.Errorf("Execute Fail")
	}

	tests := map[string]*testTarget{
		// name, error, execute, expectedCalled, dependencies
		"t1": newTestTarget("t1", "", goodExecute, []string{"t1"}, nil),
		"t2": newTestTarget("t2", "", goodExecute, []string{"t2"}, nil),
		"t3": newTestTarget("t3", "", goodExecute, []string{"t2", "t3"}, []string{"t2"}),
		"t4": newTestTarget("t5", "No targets match t5", goodExecute, nil, nil),
		"t6": newTestTarget("t6", "Execute Fail", badExecute, []string{"t6"}, nil),
		"t7": newTestTarget("t7", "Execute Fail", goodExecute, []string{"t6"}, []string{"t6"}),
		"t8": newTestTarget("t8", "", goodExecute, []string{"t1", "t8"}, []string{"t1", "t9"}),
	}

	// create an empty config file
	file, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err.Error())
	}
	configFile = file.Name()
	defer func() { configFile = "gack.yml" }()
	defer os.Remove(file.Name())
	fmt.Fprintf(file, "---\n")
	file.Close()

	for i, test := range tests {
		var called []string
		mux, _ := NewMux()
		register(tests, &called, mux, i, test)

		err := mux.Execute(test.name)
		// make sure error condition is correct
		if err != nil && test.err != "" && err.Error() != test.err {
			t.Errorf("Test %s: Expected error %s but got %v", i, test.err, err)
		} else if err != nil && test.err == "" {
			t.Errorf("Test %s: Expected no error but got %v", i, err)
		} else if err == nil && test.err != "" {
			t.Errorf("Test %s: Expected error %s but got nothing", i, test.err)
		}

		// make sure targets were called in the correct order
		if !reflect.DeepEqual(test.expectedCalled, called) {
			t.Errorf("Test %s: Expected called targets to be %v but got %v", i, test.expectedCalled, called)
		}
	}
}

func TestTargetNames(t *testing.T) {
	mux, _ := NewMux()
	mux.Register("t1", nil)
	mux.Register("t2", nil)
	mux.Register("t3", nil)
	mux.Register("t4", nil)

	expected := []string{"t4", "t3", "t2", "t1"}
	if !reflect.DeepEqual(expected, mux.TargetNames()) {
		t.Errorf("Expected %v but got %v", expected, mux.TargetNames())
	}
}
