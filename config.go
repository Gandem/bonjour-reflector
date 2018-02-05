package main

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type brconfig struct {
	NetInterface string `toml:"net_interface"`
}

func readConfig(path string) (cfg brconfig, err error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return brconfig{}, err
	}
	_, err = toml.Decode(string(content), &cfg)
	return cfg, err
}
