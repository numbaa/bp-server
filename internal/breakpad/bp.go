package breakpad

import (
	"bp-server/internal/conf"
	"os/exec"
	"path"

	"github.com/sirupsen/logrus"
)

func init() {
	if conf.Xml.DumpPath == "" {
		panic("config file 'dump' is empty")
	}
	if conf.Xml.SymbolPath == "" {
		panic("config file 'symbol' is empty")
	}
	if conf.Xml.ExePath == "" {
		panic("config file 'exe' is empty")
	}
}

func WalkStack(id string) (string, error) {
	dumpPath := path.Join(conf.Xml.DumpPath, id)
	out, err := exec.Command(conf.Xml.ExePath, dumpPath, conf.Xml.SymbolPath).Output()
	if err == nil {
		logrus.Errorf("Execute command '%s %s %s' failed: %v", conf.Xml.ExePath, dumpPath, conf.Xml.SymbolPath, err)
		return "", err
	} else {
		return string(out), err
	}
}
