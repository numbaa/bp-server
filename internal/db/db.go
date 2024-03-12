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

package db

import (
	"bp-server/internal/conf"
	"fmt"

	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var dbConn *gorm.DB

type Dump struct {
	gorm.Model
	Program  string
	Version  string
	Filename string
	Build    string
}

func init() {
	db, err := gorm.Open(sqlite.Open(conf.Xml.DB), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Failed to open sqlite database(%s): %v", conf.Xml.DB, err))
	}
	db.AutoMigrate(&Dump{})
	dbConn = db
}

func QueryDumpList(page int) ([]Dump, error) {
	const kLimit int = 20
	index := kLimit * (page - 1)
	if index < 0 {
		index = 0
	}
	var dumps []Dump
	result := dbConn.Limit(kLimit).Offset(index).Find(&dumps)
	if result.Error != nil {
		logrus.Errorf("Query table 'dumps' with limit(%d) offset(%d) failed with: %v", kLimit, index, result.Error)
		return nil, result.Error
	}
	return dumps, nil
}

func AddDump(program string, version string, filename string, buildTime string) error {
	dump := Dump{
		Program:  program,
		Version:  version,
		Filename: filename,
		Build:    buildTime,
	}
	result := dbConn.Create(&dump)
	if result.Error != nil {
		logrus.Error("Insert record to table 'dumps' failed")
		return result.Error
	}
	return nil
}
