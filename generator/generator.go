package generator

import (
	"os"

	"github.com/abates/gack"
)

type generator gack.Mux

func (g generator) Execute(ctx *gack.Context) (err error) {
	config := gack.DefaultConfig(gack.Mux(g))
	file, err := os.OpenFile("gack.yml", os.O_RDWR|os.O_CREATE, 0644)
	if err == nil {
		err = gack.WriteConfig(file, config)
	}
	return err
}

func Register(mux *gack.Mux) {
	mux.Register("generate", generator(*mux))
}
