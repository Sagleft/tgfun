package tgfun

// NewFunnel - funnel constructor
func NewFunnel(data FunnelData, script FunnelScript) (*Funnel, error) {
	f := Funnel{
		Data:   data,
		Script: script,
	}

	return &f, checkErrors(
		f.setupBot,
	)
}
