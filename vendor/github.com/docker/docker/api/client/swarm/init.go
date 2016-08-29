package swarm

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/client"
	"github.com/docker/docker/cli"
	"github.com/docker/engine-api/types/swarm"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	generatedSecretEntropyBytes = 16
	generatedSecretBase         = 36
	// floor(log(2^128-1, 36)) + 1
	maxGeneratedSecretLength = 25
)

type initOptions struct {
	swarmOptions
	listenAddr NodeAddrOption
	// Not a NodeAddrOption because it has no default port.
	advertiseAddr   string
	forceNewCluster bool
}

func newInitCommand(dockerCli *client.DockerCli) *cobra.Command {
	opts := initOptions{
		listenAddr: NewListenAddrOption(),
	}

	cmd := &cobra.Command{
		Use:   "init [OPTIONS]",
		Short: "Initialize a swarm",
		Args:  cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(dockerCli, cmd.Flags(), opts)
		},
	}

	flags := cmd.Flags()
	flags.Var(&opts.listenAddr, flagListenAddr, "Listen address (format: <ip|interface>[:port])")
	flags.StringVar(&opts.advertiseAddr, flagAdvertiseAddr, "", "Advertised address (format: <ip|interface>[:port])")
	flags.BoolVar(&opts.forceNewCluster, "force-new-cluster", false, "Force create a new cluster from current state.")
	addSwarmFlags(flags, &opts.swarmOptions)
	return cmd
}

func runInit(dockerCli *client.DockerCli, flags *pflag.FlagSet, opts initOptions) error {
	client := dockerCli.Client()
	ctx := context.Background()

	req := swarm.InitRequest{
		ListenAddr:      opts.listenAddr.String(),
		AdvertiseAddr:   opts.advertiseAddr,
		ForceNewCluster: opts.forceNewCluster,
		Spec:            opts.swarmOptions.ToSpec(),
	}

	nodeID, err := client.SwarmInit(ctx, req)
	if err != nil {
		if strings.Contains(err.Error(), "could not choose an IP address to advertise") || strings.Contains(err.Error(), "could not find the system's IP address") {
			return errors.New(err.Error() + " - specify one with --advertise-addr")
		}
		return err
	}

	fmt.Fprintf(dockerCli.Out(), "Swarm initialized: current node (%s) is now a manager.\n\n", nodeID)

	if err := printJoinCommand(ctx, dockerCli, nodeID, true, false); err != nil {
		return err
	}

	fmt.Fprint(dockerCli.Out(), "To add a manager to this swarm, run 'docker swarm join-token manager' and follow the instructions.\n\n")
	return nil
}
