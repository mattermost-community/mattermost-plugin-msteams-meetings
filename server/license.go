package main

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	e20          = "E20"
	professional = "professional"
	enterprise   = "enterprise"
)

func HasEnterpriseFeatures(config *model.Config, license *model.License) bool {
	if license != nil && (license.SkuShortName == e20 || license.SkuShortName == enterprise || license.SkuShortName == professional) {
		return true
	}

	return pluginapi.IsE20LicensedOrDevelopment(config, license)
}
