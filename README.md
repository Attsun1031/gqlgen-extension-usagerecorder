# gqlgen-extension-usagerecorder
This is an extension of gqlgen.

This extension records graphql query information to log or other sources.

## How to use
```golang
emitter := usagerecorder.NewLogEmitter(logger, 1.0)
gqlServer.Use(usagerecorder.New(
	usagerecorder.WithLogger(logger),
	usagerecorder.WithEmitter(emitter),
)
```
