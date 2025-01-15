package BaPs

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gucooing/BaPs/common/enter"
	"github.com/gucooing/BaPs/config"
	"github.com/gucooing/BaPs/db"
	"github.com/gucooing/BaPs/gateway"
	"github.com/gucooing/BaPs/gdconf"
	"github.com/gucooing/BaPs/pkg/logger"
	"github.com/gucooing/BaPs/sdk"
)

func NewBaPs() {
	err := config.LoadConfig()
	if err != nil {
		if err == config.FileNotExist {
			fmt.Printf("找不到配置文件准备生成默认配置文件\n")
			p, _ := json.MarshalIndent(config.DefaultConfig, "", "  ")
			cf, _ := os.Create("./config.json")
			_, err := cf.Write(p)
			cf.Close()
			if err != nil {
				fmt.Printf("生成默认配置文件失败 %s \n请检查是否有权限\n", err.Error())
				return
			} else {
				fmt.Printf("生成默认配置文件成功 \n请修改后重新启动")
				return
			}
		} else {
			panic(err)
		}
	}
	cfg := config.GetConfig()
	logger.InitLogger("BaPs", strings.ToUpper(cfg.LogLevel))
	logger.Info("BaPs")
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	// 初始化数据库
	db.NewPE(cfg.DB)
	// 尝试保存硬盘中的玩家数据
	if !enter.TaskUpDiskPlayerData() {
		logger.Info("请检查硬盘中的玩家数据是否正确再启动")
		logger.CloseLogger()
		return
	}
	// 初始化gin
	router, server := newGin(cfg.HttpNet)
	// 初始化sdk
	sdk.New(router)
	// 初始化gateWay
	gateway.NewGateWay(router)
	// 初始化资源文件
	gdconf.LoadGameConfig(cfg.DataPath, cfg.ResourcesPath)
	// 启动服务器
	go func() {
		if err = Run(cfg.HttpNet, server); err != nil {
			if !errors.Is(http.ErrServerClosed, err) {
				logger.Error("服务器错误:%s", err.Error())
				done <- syscall.SIGTERM
			}
		}
	}()

	// close
	clo := func() {
		_, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		logger.Info("Max Close Time 5 Minute")
		server.Close()
		enter.UpAllDate()
		logger.Info("BaPs Close")
		logger.CloseLogger()
		os.Exit(0)
	}

	go func() {
		select {
		case call := <-done:
			switch call {
			case syscall.SIGINT, syscall.SIGTERM:
				clo()
				return
			}
		}
	}()

	select {}
}

func newGin(appNet *config.HttpNet) (*gin.Engine, *http.Server) {
	gin.SetMode(gin.ReleaseMode)
	var router *gin.Engine
	if logger.GetLogLevel() == logger.DEBUG {
		router = gin.Default()
	} else {
		router = gin.New()
	}
	router.Use(gin.Recovery())
	addr := fmt.Sprintf("%s:%s", appNet.InnerAddr, appNet.InnerPort)
	if appNet.Tls {
		logger.Info("监听地址: https://%s", addr)
		logger.Info("对外地址: https://%s", fmt.Sprintf("%s:%s", appNet.OuterAddr, appNet.OuterPort))
		server := &http.Server{Addr: addr, Handler: router, TLSConfig: &tls.Config{InsecureSkipVerify: true}}
		return router, server
	}
	logger.Info("监听地址: http://%s", addr)
	logger.Info("对外地址: http://%s", fmt.Sprintf("%s:%s", appNet.OuterAddr, appNet.OuterPort))
	server := &http.Server{Addr: addr, Handler: router}
	return router, server
}

func Run(appNet *config.HttpNet, server *http.Server) error {
	if appNet.Tls {
		return server.ListenAndServeTLS(appNet.CertFile, appNet.KeyFile)
	}
	return server.ListenAndServe()
}
