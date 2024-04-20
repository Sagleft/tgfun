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
