package main

import (
	"golang.org/x/sys/unix"
)

func init() {
	targetOSPrivilegeDropper = openBSDPrivilegeDropper{}
}

type openBSDPrivilegeDropper struct{}

func (d openBSDPrivilegeDropper) Drop() {
	unix.Pledge("stdio", "")
}
