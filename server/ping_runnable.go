package main

import (
	"os/exec"

	"github.com/suborbital/grav/grav"
	"github.com/suborbital/reactr/rt"
)

type ping struct{}

func (r ping) Run(job rt.Job, ctx *rt.Ctx) (interface{}, error) {
	host := job.String()
	out, err := exec.Command("ping", "-q", "-c", "1", host).Output()

	if err != nil {
		return nil, err
	}
	return grav.NewMsg(msgTypePingRes, out), nil
}

func (r ping) OnChange(change rt.ChangeEvent) error {
	return nil
}
