package tgfun

// NewFunnel - funnel constructor
func NewFunnel(data FunnelData, script FunnelScript) *Funnel {
	return &Funnel{
		Data:   data,
		Script: script,
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
