package usagerecorder

type UsageEmitter interface {
	Emit(usage *GraphqlUsage) error
}
