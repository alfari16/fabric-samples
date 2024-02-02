package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudflare/tableflip"
	"github.com/hyperledger-labs/fabric-smart-client/pkg/api"
	"github.com/hyperledger-labs/fabric-smart-client/pkg/node"
	fabric "github.com/hyperledger-labs/fabric-smart-client/platform/fabric/sdk"
	viewregistry "github.com/hyperledger-labs/fabric-smart-client/platform/view"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/services/flogging"
	tokensdk "github.com/hyperledger-labs/fabric-token-sdk/token/sdk"
	"github.com/hyperledger/fabric-samples/token-sdk/owner/routes"
	"github.com/hyperledger/fabric-samples/token-sdk/owner/service"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var logger = flogging.MustGetLogger("main")

func main() {
	dir := getEnv("CONF_DIR", "./conf/owner1")
	port := getEnv("PORT", "9200")

	upg, _ := tableflip.New(tableflip.Options{})
	defer upg.Stop()

	// Do an upgrade on SIGHUP
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP)
		for range sig {
			_ = upg.Upgrade()
		}
	}()

	var server http.Server
	ln, _ := upg.Fds.Listen("tcp", fmt.Sprintf(":%s", port))

	// watch if the config file changes
	changes, coreConfig := make(chan int), fmt.Sprintf("%s/core.yaml", dir)
	var conf string
	go func() {
		for {
			newConf, err := fetchConfig(coreConfig)
			if err != nil {
				logger.Errorf("failed fetching config: %w", err)
				os.Exit(1)
			}
			if newConf != conf {
				conf = newConf
				changes <- 1
			}
			time.Sleep(time.Second)
		}
	}()

	// waiting for changes
	go func() {
		for range changes {
			logger.Infof("config changes detected, restarting web server")
			fsc := startFabricSmartClient(dir)
			// Tell the service how to respond to other nodes when they initiate an action
			registry := viewregistry.GetRegistry(fsc)
			succeedOrPanic(registry.RegisterResponder(&service.AcceptCashView{}, "github.com/hyperledger/fabric-samples/token-sdk/issuer/service/IssueCashView"))
			succeedOrPanic(registry.RegisterResponder(&service.AcceptCashView{}, &service.TransferView{}))

			controller := routes.Controller{Service: service.TokenService{FSC: fsc}}
			ns := routes.StartWebServer(controller, logger)
			go serve(&ns, ln)
			err := server.Shutdown(context.TODO())
			if err != nil {
				logger.Errorf("error shutting down old config server: %w", err)
			}
			server = ns
		}
	}()

	_ = upg.Ready()
	<-upg.Exit()

	server.Shutdown(context.TODO())
}

type Node interface {
	api.ServiceProvider
	Stop()
}

func startFabricSmartClient(confDir string) Node {
	logger.Infof("Initializing Fabric Smart Client and Token SDK...")
	fsc := node.NewFromConfPath(confDir)
	succeedOrPanic(fsc.InstallSDK(fabric.NewSDK(fsc)))
	succeedOrPanic(fsc.InstallSDK(tokensdk.NewSDK(fsc)))
	succeedOrPanic(fsc.Start())

	// Stop gracefully
	go handleSignals((map[os.Signal]func(){
		syscall.SIGINT: func() {
			logger.Info("Stopping FSC node...")
			fsc.Stop()
			os.Exit(130)
		},
		syscall.SIGTERM: func() {
			logger.Info("Stopping FSC node...")
			fsc.Stop()
			os.Exit(143)
		},
		syscall.SIGSTOP: func() {
			logger.Info("Stopping FSC node...")
			fsc.Stop()
			os.Exit(145)
		},
		syscall.SIGHUP: func() {
			logger.Info("Stopping FSC node...")
			fsc.Stop()
			os.Exit(129)
		},
	}))
	logger.Infof("FSC node is ready!")

	return fsc
}

// getEnv returns an environment variable or the fallback
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func succeedOrPanic(err error) {
	if err != nil {
		logger.Fatalf("Failed initializing Token SDK - %s", err.Error())
		os.Exit(1)
	}
}

func handleSignals(handlers map[os.Signal]func()) {
	var signals []os.Signal
	for sig := range handlers {
		signals = append(signals, sig)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, signals...)

	for sig := range signalChan {
		logger.Infof("Received signal: %d (%s)", sig, sig)
		handlers[sig]()
	}
}

func fetchConfig(dir string) (string, error) {
	f, err := os.Open(dir)
	if err != nil {
		return "", err
	}

	b, err := io.ReadAll(f)

	return string(b), err
}

func serve(server *http.Server, ln net.Listener) {
	err := server.Serve(ln)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			logger.Infof("Webserver closing, exiting...", err.Error())
			//fsc.Stop()
		} else {
			logger.Fatalf("echo error - %s", err.Error())
			//fsc.Stop()
			os.Exit(1)
		}
	}
}
