package usagerecorder

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LogEmitter_Sampling(t *testing.T) {
	logger := slog.Default()
	t.Run("False if sampling rate is 0", func(t *testing.T) {
		sut := NewLogEmitter(logger, 0)
		assert.False(t, sut.Sampling())
	})

	t.Run("True if sampling rate is 100", func(t *testing.T) {
		sut := NewLogEmitter(logger, 100)
		assert.True(t, sut.Sampling())
	})

	t.Run("True if sampling rate is over random value", func(t *testing.T) {
		sut := LogEmitter{Logger: logger, SamplingRate: 20, RandomValueGenerator: func() int { return 10 }}
		assert.True(t, sut.Sampling())
	})

	t.Run("False if sampling rate is over random value", func(t *testing.T) {
		sut := LogEmitter{Logger: logger, SamplingRate: 50, RandomValueGenerator: func() int { return 51 }}
		assert.False(t, sut.Sampling())
	})
}
