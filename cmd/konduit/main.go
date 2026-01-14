package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
)

type RootCmd struct {
	Globals

	CUE       CUECmd       `cmd:"" help:"Run Helm with CUE evaluation of Helm values and Kustomize patches."`
	Kustomize KustomizeCmd `cmd:"" hidden:"" help:"Run the Konduit-compatible Kustomize post-renderer."`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ctx.Done()
		stop()
	}()

	root := RootCmd{
		Globals: Globals{
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		},
	}

	cli := kong.Parse(&root,
		kong.Name("konduit"),
		kong.Description("Helm meets Kustomize, but without the YAML."),
		kong.Writers(root.Stdout, root.Stderr),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact:             true,
			FlagsLast:           true,
			NoExpandSubcommands: true,
		}),
		kong.Bind(&root.Globals),
		kong.BindTo(ctx, (*context.Context)(nil)),
	)

	cli.FatalIfErrorf(cli.Run())
}
