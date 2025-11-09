package plugin

// Note: Tests for Remove() function require InstallDir to be set up,
// which is complex to mock in unit tests. Remove() is better tested
// through integration tests or by testing the CLI commands that use it.
//
// The Remove() function itself is simple and delegates most work to:
// - os.Stat / os.RemoveAll (standard library, well-tested)
// - versionfile.LoadPluginVersionFile / Save (tested in pipe-fittings)
// - filepaths.EnsurePluginDir (tested in pipe-fittings)
//
// Edge cases tested through existing actions_test.go tests for detectLocalPlugin()
