package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	logr "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spiffe/spire/cmd/spire-agent/cli/run"
	"github.com/spiffe/spire/pkg/agent"
	"github.com/spiffe/spire/pkg/common/log"
	"github.com/vishnusomank/spire-agent/internal/helper"
)

var LocalConfig struct {
	Config     string
	Token      string
	ServerAddr string
}

var agentConf *agent.Config
var err error

func main() {

	cmd := &cobra.Command{
		Use:   "spire-agent",
		Short: "SPIFFE SPIRE Agent",
		Long:  `SPIFFE SPIRE Agent Service`,
		Run: func(cmd *cobra.Command, args []string) {
			initConfig()
			if secret := helper.GetK8sSecrets(agentConf); secret.Name == "" {
				logr.Warn("no SVID found.")
				go startAgent()
				time.Sleep(5 * time.Second)
				if err := helper.CreateK8sSecrets(agentConf); err != nil {
					logr.WithError(err).Error("Could not create k8s secrets:")
					return
				}
			} else {
				err := helper.WriteSVIDKey(agentConf)
				if err != nil {
					return
				}
				startAgent()
			}

		},
	}

	cmd.PersistentFlags().StringVarP(&LocalConfig.Config, "config", "c", "", "path to configuration file")
	cmd.PersistentFlags().StringVarP(&LocalConfig.ServerAddr, "server", "s", "", "Server address [ip:port]")
	cmd.PersistentFlags().StringVarP(&LocalConfig.Token, "token", "t", "", "Join Token for Spire Agent")

	if err := cmd.Execute(); err != nil {
		logr.WithError(err).Error("error while running spire-agent:")
		os.Exit(1)
	}

}

func initConfig() {
	var args []string
	if LocalConfig.Config != "" {
		args = []string{"-config=" + LocalConfig.Config, "-insecureBootstrap=true"}
	} else {
		args = []string{"-insecureBootstrap=true"}
	}

	agentConf, err = run.LoadConfig("Agent Config", args, []log.Option{}, &io.PipeWriter{}, true)
	if err != nil {
		logr.WithError(err).Error("err loading config file:")
		return
	}

}

func startAgent() {

	if LocalConfig.ServerAddr != "" {

		agentConf.ServerAddress = "dns:///" + LocalConfig.ServerAddr
	}
	if LocalConfig.Token != "" {
		agentConf.JoinToken = LocalConfig.Token
	}

	a := agent.New(agentConf)

	ctx := context.Background()

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	err = a.Run(ctx)
	if err != nil {
		logr.WithError(err).Error("Agent crashed: ")
		helper.DeleteSVIDSecret()
		return
	}

	logr.Warn("Agent stopped gracefully")

}
