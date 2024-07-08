package flags

import "flag"

var Flags Flag

func init() {
	Flags.Cfg = flag.String("f", "config.toml", "the config file")

	flag.Parse()
}

type Flag struct {
	Cfg *string
}
