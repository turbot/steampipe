package ociinstaller

import "github.com/turbot/pipe-fittings/ociinstaller"

type assetsImage struct {
	ReportUI string
}

func (s *assetsImage) Type() ociinstaller.ImageType {
	return ImageTypeAssets
}

// empty config for assets image
type assetsImageConfig struct {
	ociinstaller.OciConfigBase
}
