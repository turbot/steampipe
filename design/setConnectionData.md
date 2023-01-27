    SetAllConnectionConfigs
            for each connection
            if !aggregator 
                setConnectionData
                    parse connection config
                    getConnectionSchema
                        initialiseTables
                            TableMapFunc
                        buildSchema
                    // add to connection map
                    // update the watch paths for the connection file watcher

    UpdateConnectionConfigs
        // remove deleted connections