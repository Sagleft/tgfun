package tgfun

import (
	"fmt"
	"regexp"

	"github.com/microcosm-cc/bluemonday"
)

// NewFunnel - funnel constructor
func NewFunnel(data FunnelData, script FunnelScript) *Funnel {
	return &Funnel{
		Data:      data,
		Script:    script,
		sanitizer: bluemonday.StrictPolicy(),
	}
}

// EnableUsersFeature !
func (f *Funnel) EnableUsersFeature(feature UsersFeature) {
	f.features.Users = &feature
}

type UTMTags struct {
	Source   string `json:"source"`
	Campaign string `json:"campaign"`
	Content  string `json:"content"`
}

type UTMTagsFeature struct {
	GetUserUTMTags GetUTMTagsCallback
}

type GetUTMTagsCallback func(telegramUserID int64) UTMTags

func (f *Funnel) EnableUTMTagsFeauture(feature UTMTagsFeature) {
	f.features.UTM = &feature
}

func (f *funnelFeatures) IsUTMTagsFeatureActive() bool {
	return f.UTM != nil
}

type UserInputFeature struct {
	Regexp               string
	InputVerifiedEventID string
	InvalidFormatEventID string
	OnEventVerified      func(telegramUserID int64, input string)

	compiledRegexp *regexp.Regexp
}

func (f *Funnel) EnableUserInputFeature(feature UserInputFeature) error {
	var err error
	f.features.UserInput = &feature
	f.features.UserInput.compiledRegexp, err = regexp.Compile(feature.Regexp)
	if err != nil {
		return fmt.Errorf("compile regexp: %w", err)
	}
	return nil
}

func (f *funnelFeatures) IsUserInputFeatureActive() bool {
	return f.UserInput != nil
}
