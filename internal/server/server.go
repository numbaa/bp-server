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
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Server struct {
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
	return &Server{
		router:     gin.Default(),
		stopedChan: make(chan struct{}, 2),
	}
}

func (svr *Server) Start() {
	svr.router.GET("/list/:page", svr.list)
	svr.router.GET("/view/:id", svr.view)
	svr.router.POST("/upload", svr.upload)
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
	//TODO: render dumps to html
	ctx.String(http.StatusOK, "Success ", dumps)
}

func (svr *Server) view(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		logrus.Warnf("/view/:id: GET parameter 'id' is empty")
		ctx.String(http.StatusOK, "GET parameter 'id' is empty")
		return
	}
	frames, err := breakpad.WalkStack(id)
	if err != nil {
		ctx.String(http.StatusOK, "Get crash info failed")
		return
	}
	//TODO: render frames to html
	ctx.String(http.StatusOK, "Success ", frames)
}

func (svr *Server) upload(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		msg := fmt.Sprintf("Upload file failed: %v", err)
		logrus.Warn(msg)
		ctx.String(http.StatusOK, msg)
	} else {
		msg := fmt.Sprintf("Upload file: %s, size: %d", file.Filename, file.Size)
		logrus.Print(msg)
		//ctx.SaveUploadedFile(file, "/path/to/minidump")
		ctx.String(http.StatusOK, msg)
	}
}
