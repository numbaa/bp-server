/*
 * BSD 3-Clause License
 *
 * Copyright (c) 2024 Zhennan Tu <zhennan.tu@gmail.com>
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice, this
 *    list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 *    this list of conditions and the following disclaimer in the documentation
 *    and/or other materials provided with the distribution.
 *
 * 3. Neither the name of the copyright holder nor the names of its
 *    contributors may be used to endorse or promote products derived from
 *    this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
 * SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
 * CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
 * OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package conf

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"
)

const defaultXmlPath = "bp-server.xml"
const defaultXmlConfig = `
<?xml version="1.0" encoding="UTF-8" ?>
<bp-server>

    <log>
        <path>log</path>
        <prefix>bp-server</prefix>
        <level>info</level>
        <maxsize>10</maxsize>
        <maxage>30</maxage>
    </log>

    <db>./dumps.db</db>
    <dump>./dumps/</dump>
    <symbol>./symbols</symbol>
    <exe>minidump_stackwalk</exe>

    <net>
        <mode>release</mode>
        <prefix></prefix>
        <view_ip>0.0.0.0</view_ip>
        <view_port>17000</view_port>
        <upload_ip>0.0.0.0</upload_ip>
        <upload_port>17001</upload_port>
    </net>

</bp-server>
`

var Xml relayConf

type relayConf struct {
	Log        logConf `xml:"log"`
	Net        netConf `xml:"net"`
	DB         string  `xml:"db"`
	DumpPath   string  `xml:"dump"`
	SymbolPath string  `xml:"symbol"`
	ExePath    string  `xml:"exe"`
}

type logConf struct {
	Path    string `xml:"path"`
	Prefix  string `xml:"prefix"`
	Level   string `xml:"level"`
	MaxSize int    `xml:"maxsize"`
	MaxAge  int    `xml:"maxage"`
}

type netConf struct {
	Mode       string `xml:"mode"`
	Prefix     string `xml:"prefix"`
	ViewPort   uint16 `xml:"view_port"`
	ViewIP     string `xml:"view_ip"`
	UploadPort uint16 `xml:"upload_port"`
	UploadIP   string `xml:"upload_ip"`
}

func init() {
	xmlPath := flag.String("c", defaultXmlPath, "config file path")
	flag.Parse()
	if err := loadConfig(*xmlPath); err != nil {
		panic(err)
	}
}

func loadConfig(xmlPath string) error {
	content, err := os.ReadFile(xmlPath)
	if err != nil {
		fmt.Printf("Read config from '%s' failed, using default config.\n\n", xmlPath)
		content = []byte(defaultXmlConfig)
	}
	cfg := relayConf{}
	err = xml.Unmarshal(content, &cfg)
	if err != nil {
		return err
	}
	Xml = cfg
	return nil
}
