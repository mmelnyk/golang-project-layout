package config

import "os"

var (
	otherDirs = []string{os.Getenv("PROGRAMDATA")}
)
