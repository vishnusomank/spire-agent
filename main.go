package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	logr "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spiffe/spire/cmd/spire-agent/cli/run"
	"github.com/spiffe/spire/pkg/agent"
	"github.com/spiffe/spire/pkg/common/log"
	"github.com/vishnusomank/spire-agent/internal/constants"
)

var LocalConfig struct {
	Config     string
	Token      string
	ServerAddr string
}

var agentConf *agent.Config
var err error
var wg sync.WaitGroup

func main() {

	cmd := &cobra.Command{
		Use:   "spire-agent",
		Short: "SPIFFE SPIRE Agent",
		Long:  `SPIFFE SPIRE Agent Service`,
		Run: func(cmd *cobra.Command, args []string) {
			initConfig()

			wg.Add(1)
			go startAgent()
			wg.Wait()

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
		args = []string{"-config=" + constants.DEFAULT_AGENT_CONFIG_PATH, "-insecureBootstrap=true"}
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
	if os.Getenv("JOIN_TOKEN") != "" {
		agentConf.JoinToken = os.Getenv("JOIN_TOKEN")
	} else if LocalConfig.Token != "" {
		agentConf.JoinToken = LocalConfig.Token
	}

	logr.Infof("Starting spire-agent with token: %v", agentConf.JoinToken)

	a := agent.New(agentConf)

	ctx := context.Background()

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	err = a.Run(ctx)
	if err != nil {
		defer wg.Done()
		logr.WithError(err).Error("Agent crashed: ")
		//helper.DeleteSVIDSecret()
		return
	}
	defer wg.Done()

	logr.Warn("Agent stopped gracefully")

}
