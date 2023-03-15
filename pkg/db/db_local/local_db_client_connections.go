//
//// before finishing - be sure to save connection state if
//// 	- it was modified in the loading process (indicating it contained non-existent connections)
////  - connections have been updated
//defer func() {
//	if res.Error == nil && connectionUpdates.ConnectionStateModified || res.UpdatedConnections {
//		serialiseConnectionState(res, connectionUpdates)
//	}
//}()
//
//func serialiseConnectionState(res *steampipeconfig.RefreshConnectionResult, connectionUpdates *steampipeconfig.ConnectionUpdates) {
//// now serialise the connection state
//connectionState := make(steampipeconfig.ConnectionDataMap, len(connectionUpdates.RequiredConnectionState))
//for k, v := range connectionUpdates.RequiredConnectionState {
//connectionState[k] = v
//}
//// NOTE: add any connection which failed
//for c := range res.FailedConnections {
//connectionState[c].Loaded = false
//connectionState[c].Error = "plugin failed to start"
//}
//for pluginName, connections := range connectionUpdates.MissingPlugins {
//// add in missing connections
//for _, c := range connections {
//connectionData := steampipeconfig.NewConnectionData(pluginName, &c, time.Now())
//connectionData.Loaded = false
//connectionData.Error = "plugin not installed"
//connectionState[c.Name] = connectionData
//}
//}
//
//// update connection state and write the missing and failed plugin connections
//if err := connectionState.Save(); err != nil {
//res.Error = err
//}
//}
