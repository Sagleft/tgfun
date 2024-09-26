package tgfun

import (
	"testing"

	"github.com/test-go/testify/assert"
	"github.com/test-go/testify/require"
)

func TestFilterUserPayloadSimple(t *testing.T) {
	// given
	payloadRaw := "dzen_org"

	// when
	payload, err := filterUserPayload(payloadRaw)

	// then
	require.NoError(t, err)
	assert.Equal(t, "dzen", payload.UTMSource)
	assert.Equal(t, "org", payload.UTMCampaign)
}

func TestFilterUserPayloadYclid(t *testing.T) {
	// given
	payloadRaw := "yandex_search_100"

	// when
	payload, err := filterUserPayload(payloadRaw)

	// then
	require.NoError(t, err)
	assert.Equal(t, "yandex", payload.UTMSource)
	assert.Equal(t, "search", payload.UTMCampaign)
	assert.Equal(t, "100", payload.Yclid)
}

func TestFilterUserPayloadNotSet(t *testing.T) {
	// given
	payloadRaw := ""

	// when
	payload, err := filterUserPayload(payloadRaw)

	// then
	require.NoError(t, err)
	assert.Equal(t, "", payload.UTMSource)
	assert.Equal(t, "", payload.UTMCampaign)
	assert.Equal(t, "", payload.Yclid)
}

func TestFilterUserPayloadNotSetSome(t *testing.T) {
	// given
	payloadRaw := "dzen"

	// when
	payload, err := filterUserPayload(payloadRaw)

	// then
	require.NoError(t, err)
	assert.Equal(t, "", payload.UTMSource)
	assert.Equal(t, "", payload.UTMCampaign)
	assert.Equal(t, "", payload.Yclid)
}

func TestFilterUserPayloadBacklink(t *testing.T) {
	// given
	payloadRaw := "howItWorks_back"

	// when
	payload, err := filterUserPayload(payloadRaw)

	// then
	require.NoError(t, err)
	assert.Equal(t, "howItWorks", payload.UTMSource)
	assert.Equal(t, "back", payload.UTMCampaign)
	assert.Equal(t, "", payload.Yclid)
	assert.Equal(t, "howItWorks", payload.BackLinkEventID)
}

func TestFilterUserPayloadBase64EventID(t *testing.T) {
	// given
	payloadRaw := "cz1kemVuJmM9b3JnJmI9ZXZlbnRJRA"

	// when
	payload, err := filterUserPayload(payloadRaw)

	// then
	require.NoError(t, err)
	assert.Equal(t, "dzen", payload.UTMSource)
	assert.Equal(t, "org", payload.UTMCampaign)
	assert.Equal(t, "", payload.Yclid)
	assert.Equal(t, "eventID", payload.BackLinkEventID)
}

func TestFilterUserPayloadBase64(t *testing.T) {
	// given
	payloadRaw := "cz1kemVuJmM9b3Jn"

	// when
	payload, err := filterUserPayload(payloadRaw)

	// then
	require.NoError(t, err)
	assert.Equal(t, "dzen", payload.UTMSource)
	assert.Equal(t, "org", payload.UTMCampaign)
	assert.Equal(t, "", payload.Yclid)
	assert.Equal(t, "", payload.BackLinkEventID)
}

func TestFilterUserPayloadBase64Yclid(t *testing.T) {
	// given
	payloadRaw := "cz15YW5kZXgmYz1zZWFyY2gmeT0xMDA"

	// when
	payload, err := filterUserPayload(payloadRaw)

	// then
	require.NoError(t, err)
	assert.Equal(t, "yandex", payload.UTMSource)
	assert.Equal(t, "search", payload.UTMCampaign)
	assert.Equal(t, "100", payload.Yclid)
	assert.Equal(t, "", payload.BackLinkEventID)
}

func TestFilterUserPayloadBase64OnlyEventID(t *testing.T) {
	// given
	payloadRaw := "Yj1ldmVudElE"

	// when
	payload, err := filterUserPayload(payloadRaw)

	// then
	require.NoError(t, err)
	assert.Equal(t, "", payload.UTMSource)
	assert.Equal(t, "", payload.UTMCampaign)
	assert.Equal(t, "", payload.Yclid)
	assert.Equal(t, "eventID", payload.BackLinkEventID)
}

func TestIsBase64Success(t *testing.T) {
	// given
	encoded := "cz1kemVuJmM9b3JnJmI9ZXZlbnRJRA"

	// when
	result := isBase64(encoded)

	// then
	assert.True(t, result)
}

func TestIsBase64WithPaddingSuccess(t *testing.T) {
	// given
	encoded := "cz1kemVuJmM9b3JnJmI9ZXZlbnRJRA=="

	// when
	result := isBase64(encoded)

	// then
	assert.True(t, result)
}

func TestLimitStringLen(t *testing.T) {
	// given
	str := "abracadabra"
	maxLength := 4

	// when
	newStr := LimitStringLen(str, maxLength)

	// then
	assert.Equal(t, "abra", newStr)
}

func TestLimitStringLen2(t *testing.T) {
	// given
	str := "abracadabra"
	maxLength := 50

	// when
	newStr := LimitStringLen(str, maxLength)

	// then
	assert.Equal(t, str, newStr)
}
