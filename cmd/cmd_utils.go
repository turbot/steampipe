package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/turbot/steampipe-plugin-sdk/instrument"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func startCmdSpan(cmd *cobra.Command) (context.Context, trace.Span) {
	tr := instrument.GetTracer()
	tracingContext, span := tr.Start(cmd.Context(), cmd.CalledAs())
	span.SetAttributes(attribute.Key("cmd").String(cmd.CalledAs()))
	flags := cmd.Flags()
	flags.Visit(func(f *pflag.Flag) {
		span.SetAttributes(attribute.Key(fmt.Sprintf("flag-%s", f.Name)).String(f.Value.String()))
	})
	span.SetAttributes(attribute.Key("args").StringSlice(flags.Args()))

	return tracingContext, span
}
