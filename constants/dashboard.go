package constants

const (
	DashboardServerDefaultPort = 9194
	// TODO [rerports] derive from steampipe version
	DashboardAssetsVersion  = "0.13.0-alpha.13"
	DashboardAssetsImageRef = "us-docker.pkg.dev/steampipe/steampipe/assets:" + DashboardAssetsVersion
)
