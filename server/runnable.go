package main

import (
	"os/exec"

	"github.com/suborbital/reactr/rt"
)

type ping struct{}

func (r ping) Run(job rt.Job, ctx *rt.Ctx) (interface{}, error) {
	host := job.String()
	out, err := exec.Command("ping", host).Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r ping) OnChange(change rt.ChangeEvent) error {
	return nil
}
