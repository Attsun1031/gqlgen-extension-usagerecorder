package usagerecorder

import (
	"encoding/json"
	"log/slog"
	"math/rand"
	"time"

	"github.com/Attsun1031/gqlgen-extension-usagerecorder/usagerecorder/model"
)

type LogEmitter struct {
	Logger               *slog.Logger
	SamplingRate         int
	RandomValueGenerator func() int
}

func defaultRandomValueGenerator() func() int {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // #nosec G404
	return func() int {
		return r.Intn(100)
	}
}

func NewLogEmitter(logger *slog.Logger, samplingRate int) *LogEmitter {
	return &LogEmitter{
		Logger:               logger,
		SamplingRate:         samplingRate,
		RandomValueGenerator: defaultRandomValueGenerator(),
	}
}

func (x *LogEmitter) Emit(usage *model.GraphqlUsage) error {
	if !x.Sampling() {
		return nil
	}

	// To convert the GraphqlUsage struct to map[string]interface{}, we marshal and unmarshal it.
	data, err := json.Marshal(usage)
	if err != nil {
		return err
	}
	var jsonObject map[string]interface{}
	err = json.Unmarshal(data, &jsonObject)
	if err != nil {
		return err
	}
	x.Logger.With("graphqlusage", jsonObject).Info("graphql usage recorded")
	return nil
}

// Sampling returns true if the sampling rate is 100 or the random value is less than the sampling rate.
func (x *LogEmitter) Sampling() bool {
	if x.SamplingRate == 0 {
		// Always false if sampling rate is 0
		return false
	}
	if x.SamplingRate == 100 {
		// Always true if sampling rate is 100
		return true
	}

	return x.RandomValueGenerator() <= x.SamplingRate
}
