package breakpad

import "bp-server/internal/conf"

func init() {
	if conf.Xml.DumpPath == "" {
		panic("config file 'dump' is empty")
	}
	if conf.Xml.SymbolPath == "" {
		panic("config file 'symbol' is empty")
	}
}

func WalkStack(id string) ([]string, error) {
	return []string{}, nil
}
