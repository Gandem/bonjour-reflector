package main

var targetOSPrivilegeDropper privilegeDropper = nil

func dropPrivilegesIfSupported() {
	if targetOSPrivilegeDropper != nil {
		targetOSPrivilegeDropper.Drop()
	}
}

type privilegeDropper interface {
	Drop()
}
