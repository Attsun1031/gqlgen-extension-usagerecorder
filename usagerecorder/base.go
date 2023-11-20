package usagerecorder

import "github.com/Attsun1031/gqlgen-extension-usagerecorder/usagerecorder/model"

type UsageEmitter interface {
	Emit(usage *model.GraphqlUsage) error
}
