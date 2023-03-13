package db_local

//
//// LocalDbConnection wraps over DbConneciont, adding service lifecycle management
//type LocalDbConnection struct {
//	Connection *pgx.Conn
//	invoker    constants.Invoker
//}
//
//// GetLocalConn starts service if needed and creates a new LocalDbConnection
//func GetLocalConn(ctx context.Context, invoker constants.Invoker, connectionOpts *CreateDbOptions, opts ...LocalClientOption) (_ *LocalDbConnection, err error) {
//	utils.LogTime("db.GetLocalConn start")
//	defer utils.LogTime("db.GetLocalConn end")
//
//	config := &LocalClientConfiguration{
//		// default to checking for db installation
//		ensureDBInstalled: true,
//	}
//	for _, o := range opts {
//		o(config)
//	}
//
//	if config.ensureDBInstalled {
//		// start db if necessary
//		if err := EnsureDBInstalled(ctx); err != nil {
//			return nil, err
//		}
//
//		startResult := StartServices(ctx, viper.GetInt(constants.ArgDatabasePort), ListenTypeLocal, invoker)
//		if startResult.Error != nil {
//			return nil, startResult.Error
//		}
//		defer func() {
//			if err != nil {
//				ShutdownService(ctx, invoker)
//			}
//		}()
//	}
//
//	conn, err := CreateLocalDbConnection(ctx, connectionOpts)
//	if err != nil {
//		return nil, err
//	}
//	return &LocalDbConnection{
//		Connection: conn,
//		invoker:    invoker,
//	}, nil
//}
//
//// Close implements Client
//// close the connection to the database and shuts down the backend if we are the last connection
//func (c *LocalDbConnection) Close(ctx context.Context) error {
//	if err := c.Connection.Close(ctx); err != nil {
//		return err
//	}
//	log.Printf("[TRACE] local client close complete")
//
//	log.Printf("[TRACE] shutdown local service %v", c.invoker)
//	ShutdownService(ctx, c.invoker)
//	return nil
//}
