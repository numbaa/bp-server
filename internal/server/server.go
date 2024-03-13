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

package server

import (
	"bp-server/internal/breakpad"
	"bp-server/internal/conf"
	"bp-server/internal/db"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const listTemplate = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Dumps</title>
		<style>
			th, td {
				padding: 20px;
			}
		</style>
	</head>
	<body>
		<table>
			<thead>
				<tr>
					<th>Program</th>
					<th>Version</th>
					<th>Build Time</th>
					<th>Crash Time</th>
					<th>Dump</th>
				</tr>
			</thead>
		<tbody>
		{{range . }}
			<tr>
				<td>{{ .Program }}</td>
				<td>{{ .Version }}</td>
				<td>{{ .Build }}</td>
				<td>{{ .CreatedAt }}</td>
				<td><a href="%s/view/ {{- .ID -}} "> {{ .Filename }} </a></td>
			</tr>
		{{end}}
		</tbody>
		</table>
	</body>
</html>`

type Server struct {
	tpl        *template.Template
	router     *gin.Engine
	stopedChan chan struct{}
	httpView   *http.Server
	httpUpload *http.Server
}

func toGinMode(mode string) string {
	mode = strings.ToLower(mode)
	if mode == gin.ReleaseMode || mode == gin.DebugMode || mode == gin.TestMode {
		return mode
	} else {
		logrus.Warnf("Unknown gin mode(%s), default to release mode", mode)
		return gin.ReleaseMode
	}
}

func New() *Server {
	gin.SetMode(toGinMode(conf.Xml.Net.Mode))
	template2 := fmt.Sprintf(listTemplate, conf.Xml.Net.Prefix)
	tpl, err := template.New("list").Parse(template2)
	if err != nil {
		panic(err)
	}
	return &Server{
		tpl:        tpl,
		router:     gin.Default(),
		stopedChan: make(chan struct{}, 2),
	}
}

func (svr *Server) Start() {
	svr.router.GET("/list/:page", svr.list)
	svr.router.GET("/view/:id", svr.view)
	svr.router.POST("/updump", svr.uploadDump)
	svr.router.POST("/upsym", svr.uploadSymbol)
	svr.httpUpload = &http.Server{
		Addr:    conf.Xml.Net.UploadIP + ":" + fmt.Sprint(conf.Xml.Net.UploadPort),
		Handler: svr.router,
	}
	svr.httpView = &http.Server{
		Addr:    conf.Xml.Net.ViewIP + ":" + fmt.Sprint(conf.Xml.Net.ViewPort),
		Handler: svr.router,
	}
	go func() {
		if err := svr.httpUpload.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Errorf("Upload HTTP listen error: %s", err)
		}
		svr.stopedChan <- struct{}{}
	}()
	go func() {
		if err := svr.httpView.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Errorf("View HTTP listen error: %s", err)
		}
		svr.stopedChan <- struct{}{}
	}()
}

func (svr *Server) Stop() {
	// ???
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	if err := svr.httpUpload.Shutdown(ctx); err != nil {
		logrus.Errorf("Shutdown upload http server error: %s", err)
	}
	if err := svr.httpView.Shutdown(ctx); err != nil {
		logrus.Errorf("Shutdown view http server error: %s", err)
	}
	cancel()
	logrus.Info("HTTP server stoped.")
	svr.stopedChan <- struct{}{}
}

func (svr *Server) StopedChan() chan struct{} {
	return svr.stopedChan
}

func (svr *Server) PrintStats() {
	//
}

func (svr *Server) list(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.Param("page"))
	if err != nil {
		logrus.Warnf("/list/:page: parse 'page' failed: %v", err)
		ctx.String(http.StatusOK, "Parse GET parameter 'page' as integer failed")
		return
	}
	dumps, err := db.QueryDumpList(page)
	if err != nil {
		logrus.Error("QueryDumpList failed ", err)
		ctx.String(http.StatusOK, "Query dump list internal error")
		return
	}
	ctx.Status(http.StatusOK)
	svr.tpl.Execute(ctx.Writer, dumps)
}

func (svr *Server) view(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logrus.Warnf("/view/:id: parse 'id' failed: %v", err)
		ctx.String(http.StatusOK, "Parse GET parameter 'id' as integer failed")
		return
	}
	if id < 0 {
		id = 0
	}
	dump, err := db.QueryDump(uint(id))
	if err != nil {
		msg := fmt.Sprintf("Query dump with id '%d' failed", id)
		logrus.Warn(msg)
		ctx.String(http.StatusOK, msg)
		return
	}
	fullpath := path.Join(conf.Xml.DumpPath, dump.Program, dump.Version, dump.Filename)
	content, err := breakpad.WalkStack(fullpath)
	if err != nil {
		ctx.String(http.StatusOK, "Get crash info failed")
		return
	}
	ctx.String(http.StatusOK, content)
}

func (svr *Server) uploadDump(ctx *gin.Context) {
	buildTime := ctx.PostForm("build")
	programName := ctx.PostForm("program")
	version := ctx.PostForm("version")
	if buildTime == "" || programName == "" || version == "" {
		logrus.Warn("Upload dump failed: invalid parameters")
		ctx.String(http.StatusOK, "Upload dump failed: invalid parameters")
		return
	}
	file, err := ctx.FormFile("file")
	if err != nil {
		msg := fmt.Sprintf("Upload dump failed: %v", err)
		logrus.Warn(msg)
		ctx.String(http.StatusOK, msg)
		return
	}
	fullpath := path.Join(conf.Xml.DumpPath, programName, version, file.Filename)
	err = ctx.SaveUploadedFile(file, fullpath)
	if err != nil {
		logrus.Warnf("Save dump file to disk failed: %v", err)
		ctx.String(http.StatusOK, "Save dump file to disk failed")
		return
	}
	err = db.AddDump(programName, version, file.Filename, buildTime)
	if err != nil {
		ctx.String(http.StatusOK, "Add meta info to database failed")
		return
	}
	logrus.Printf("Upload dump: %s, size: %d, program:%s, version:%s, build time:%s", file.Filename, file.Size, programName, version, buildTime)
	ctx.String(http.StatusOK, "Success")
}

func (svr *Server) uploadSymbol(ctx *gin.Context) {
	entry := ctx.PostForm("entry")
	id := ctx.PostForm("id")
	if entry == "" || id == "" {
		logrus.Errorf("Upload symbol file failed: invalid parameters")
		ctx.String(http.StatusOK, "Upload file symbol failed: invalid parameters")
		return
	}
	file, err := ctx.FormFile("file")
	if err != nil {
		msg := fmt.Sprintf("Upload symbol file failed: %v", err)
		logrus.Errorf(msg)
		ctx.String(http.StatusOK, msg)
		return
	}
	fullpath := path.Join(conf.Xml.SymbolPath, entry, id, file.Filename)
	err = ctx.SaveUploadedFile(file, fullpath)
	if err != nil {
		logrus.Errorf("Save uploaded symbol file '%s' to '%s' failed with: %v", file.Filename, fullpath, err)
		ctx.String(http.StatusOK, "Save file failed")
	} else {
		logrus.Infof("Saved uploaded symbol file '%s' to '%s'", file.Filename, fullpath)
		ctx.String(http.StatusOK, "Success")
	}
}
