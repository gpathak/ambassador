package agent

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/datawire/ambassador/cmd/entrypoint"
	"github.com/datawire/ambassador/pkg/agent"
	"github.com/datawire/ambassador/pkg/busy"
	"github.com/datawire/ambassador/pkg/logutil"
	"github.com/datawire/dlib/dlog"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

// internal k8s service
const DefaultSnapshotURLFmt = "http://ambassador-admin:%d/snapshot-external"

func run(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	ambAgent := agent.NewAgent(nil)

	// all log things need to happen here because we still allow the agent to run in amb-sidecar
	// and amb-sidecar should control all the logging if it's kicking off the agent.
	// this codepath is only hit when the agent is running on its own
	logLevel := os.Getenv("AES_LOG_LEVEL")
	// by default, suppress everything except fatal things
	// the watcher in the agent will spit out a lot of errors because we don't give it rbac to
	// list secrets initially.
	klogLevel := 3
	if logLevel != "" {
		logrusLevel, err := logrus.ParseLevel(logLevel)
		if err != nil {
			dlog.Errorf(ctx, "error parsing log level, running with default level: %+v", err)
		} else {
			busy.SetLogLevel(logrusLevel)
		}
		klogLevel = logutil.LogrusToKLogLevel(logrusLevel)
	}
	klogFlags := flag.NewFlagSet(os.Args[0], flag.PanicOnError)
	klog.InitFlags(klogFlags)
	klogFlags.Parse([]string{fmt.Sprintf("-stderrthreshold=%d", klogLevel), "-v=2", "-logtostderr=false"})
	snapshotURL := os.Getenv("AES_SNAPSHOT_URL")
	if snapshotURL == "" {
		snapshotURL = fmt.Sprintf(DefaultSnapshotURLFmt, entrypoint.ExternalSnapshotPort)
	}

	ambAgent.Watch(ctx, snapshotURL)

	return nil
}

func Main(ctx context.Context, version string, args ...string) error {
	argparser := &cobra.Command{
		Use:           os.Args[0],
		Version:       version,
		RunE:          run,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	argparser.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		if err == nil {
			return nil
		}
		dlog.Errorf(ctx, "%s\nSee '%s --help'.\n", err, cmd.CommandPath())
		return nil
	})
	argparser.SetArgs(args)
	return argparser.ExecuteContext(ctx)
}
