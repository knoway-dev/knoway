package openai

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseImageGenerationsSizeString(t *testing.T) {
	size, err := parseImageGenerationsSizeString(nil)
	require.NoError(t, err)
	require.Nil(t, size)

	size, err = parseImageGenerationsSizeString(lo.ToPtr(""))
	require.Error(t, err)
	require.Nil(t, size)
	require.EqualError(t, err, "empty size string")

	size, err = parseImageGenerationsSizeString(lo.ToPtr("1024x1024"))
	require.NoError(t, err)
	require.NotNil(t, size)
	assert.Equal(t, uint64(1024), size.Width)
	assert.Equal(t, uint64(1024), size.Height)

	size, err = parseImageGenerationsSizeString(lo.ToPtr("1024x"))
	require.Error(t, err)
	require.Nil(t, size)
	require.EqualError(t, err, "invalid height `` in \"size\" value `1024x`")

	size, err = parseImageGenerationsSizeString(lo.ToPtr("x1024"))
	require.Error(t, err)
	require.Nil(t, size)
	require.EqualError(t, err, "invalid width `` in \"size\" value `x1024`")

	size, err = parseImageGenerationsSizeString(lo.ToPtr("1024x1024x1024"))
	require.Error(t, err)
	require.Nil(t, size)
	require.EqualError(t, err, "invalid `1024x1024x1024` in \"size\" value: too many parts")

	size, err = parseImageGenerationsSizeString(lo.ToPtr("1024"))
	require.Error(t, err)
	require.Nil(t, size)
	require.EqualError(t, err, "invalid `1024` in \"size\" value")

	size, err = parseImageGenerationsSizeString(lo.ToPtr("1024x1024x"))
	require.Error(t, err)
	require.Nil(t, size)
	require.EqualError(t, err, "invalid `1024x1024x` in \"size\" value: too many parts")

	size, err = parseImageGenerationsSizeString(lo.ToPtr("testx1024"))
	require.Error(t, err)
	require.Nil(t, size)
	require.EqualError(t, err, "invalid width `test` in \"size\" value `testx1024`")
}
