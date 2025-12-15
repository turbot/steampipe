package initialisation

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/steampipeconfig"
	"github.com/turbot/steampipe-plugin-sdk/v5/telemetry"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_client"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/db/db_local"
	"github.com/turbot/steampipe/v2/pkg/export"
	"github.com/turbot/steampipe/v2/pkg/statushooks"
)

type InitData struct {
	Client        db_common.Client
	Result        *db_common.InitResult
	PipesMetadata *steampipeconfig.PipesMetadata

	ShutdownTelemetry func()
	ExportManager     *export.Manager
}

func NewErrorInitData(err error) *InitData {
	return &InitData{
		Result: &db_common.InitResult{Error: err},
	}
}

func NewInitData() *InitData {
	i := &InitData{
		Result:        &db_common.InitResult{},
		ExportManager: export.NewManager(),
	}

	return i
}

func (i *InitData) RegisterExporters(exporters ...export.Exporter) *InitData {
	for _, e := range exporters {
		// Skip nil exporters to prevent nil pointer panic
		if e == nil {
			continue
		}
		if err := i.ExportManager.Register(e); err != nil {
			// short circuit if there is an error
			i.Result.Error = err
			return i
		}
	}
	return i
}

func (i *InitData) Init(ctx context.Context, invoker constants.Invoker, opts ...db_client.ClientOption) {
	defer func() {
		if r := recover(); r != nil {
			i.Result.Error = helpers.ToError(r)
		}
		// if there is no error, return context cancellation error (if any)
		if i.Result.Error == nil {
			i.Result.Error = ctx.Err()
		}
	}()

	log.Printf("[INFO] Initializing...")

	statushooks.SetStatus(ctx, "Initializing")

	// initialise telemetry
	shutdownTelemetry, err := telemetry.Init(app_specific.AppName)
	if err != nil {
		i.Result.AddWarnings(err.Error())
	} else {
		i.ShutdownTelemetry = shutdownTelemetry
	}

	// retrieve cloud metadata
	pipesMetadata, err := getPipesMetadata(ctx)
	if err != nil {
		i.Result.Error = err
		return
	}

	// set cloud metadata (may be nil)
	i.PipesMetadata = pipesMetadata

	// get a client
	// add a message rendering function to the context - this is used for the fdw update message and
	// allows us to render it as a standard initialisation message
	getClientCtx := statushooks.AddMessageRendererToContext(ctx, func(format string, a ...any) {
		i.Result.AddMessage(fmt.Sprintf(format, a...))
	})

	statushooks.SetStatus(ctx, "Connecting to steampipe database")
	log.Printf("[INFO] Connecting to steampipe database")
	client, errorsAndWarnings := GetDbClient(getClientCtx, invoker, opts...)
	if errorsAndWarnings.Error != nil {
		i.Result.Error = errorsAndWarnings.Error
		return
	}

	i.Result.AddWarnings(errorsAndWarnings.Warnings...)

	log.Printf("[INFO] ValidateClientCacheSettings")
	errorsAndWarnings = db_common.ValidateClientCacheSettings(client)
	if errorsAndWarnings.GetError() != nil {
		i.Result.Error = errorsAndWarnings.GetError()
	}
	i.Result.AddWarnings(errorsAndWarnings.Warnings...)

	i.Client = client
}

// GetDbClient either creates a DB client using the configured connection string (if present) or creates a LocalDbClient
func GetDbClient(ctx context.Context, invoker constants.Invoker, opts ...db_client.ClientOption) (db_common.Client, error_helpers.ErrorAndWarnings) {
	if connectionString := viper.GetString(pconstants.ArgConnectionString); connectionString != "" {
		statushooks.SetStatus(ctx, "Connecting to remote Steampipe database")
		client, err := db_client.NewDbClient(ctx, connectionString, opts...)
		if err != nil {
			return nil, error_helpers.NewErrorsAndWarning(err)
		}
		return client, error_helpers.NewErrorsAndWarning(err)
	}

	statushooks.SetStatus(ctx, "Starting local Steampipe database")
	log.Printf("[INFO] Starting local Steampipe database")

	return db_local.GetLocalClient(ctx, invoker, opts...)
}

func (i *InitData) Cleanup(ctx context.Context) {
	if i.Client != nil {
		i.Client.Close(ctx)
	}
	if i.ShutdownTelemetry != nil {
		i.ShutdownTelemetry()
	}
}
