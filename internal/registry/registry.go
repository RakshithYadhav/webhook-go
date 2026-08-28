package registry


type Registry struct {
	endpoints map[string][]string
}

func (r *Registry) Seed(endpoints map[string][]string) {
	r.endpoints = endpoints
}

func (r *Registry) Lookup(eventType string) []string {
	v, ok := r.endpoints[eventType]
	if !ok {
		return []string{}
	}

	return v
}
