package config

import "context"

// DefaultSources groups sources by the documented default precedence model.
// Order within each category is preserved.
type DefaultSources struct {
	// Defaults supplies the priority-10 lowest-precedence category.
	Defaults []Source
	// DiscoveredBase supplies priority-20 base documents, not automatic discovery.
	DiscoveredBase []Source
	// DiscoveredProfile supplies priority-30 profile documents.
	DiscoveredProfile []Source
	// ExplicitFiles supplies priority-40 explicitly chosen file sources.
	ExplicitFiles []Source
	// Dotenv supplies priority-50 textual sources.
	Dotenv []Source
	// Environment supplies priority-60 environment sources.
	Environment []Source
	// Overrides supplies the priority-70 final category; custom NewPlan priorities may differ.
	Overrides []Source
}

// NewDefaultPlan assigns category priorities without mutating Source.Info and
// returns a plan with stable caller order within each category. It borrows the
// source instances; the category slices are not retained. Duplicate names across
// categories and nil sources fail as in NewPlan.
func NewDefaultPlan(sources DefaultSources) (Plan, error) {
	var resolved []Source
	resolved = appendPriority(resolved, sources.Defaults, PriorityDefaults)
	resolved = appendPriority(resolved, sources.DiscoveredBase, PriorityDiscoveredBase)
	resolved = appendPriority(resolved, sources.DiscoveredProfile, PriorityDiscoveredProfile)
	resolved = appendPriority(resolved, sources.ExplicitFiles, PriorityExplicitFiles)
	resolved = appendPriority(resolved, sources.Dotenv, PriorityDotenv)
	resolved = appendPriority(resolved, sources.Environment, PriorityEnvironment)
	resolved = appendPriority(resolved, sources.Overrides, PriorityOverrides)
	return NewPlan(resolved...)
}

type prioritySource struct {
	source Source
	info   SourceInfo
}

func (s prioritySource) Info() SourceInfo { return s.info }

func (s prioritySource) Load(ctx context.Context) (Document, error) {
	return s.source.Load(ctx)
}

func appendPriority(destination, sources []Source, priority int) []Source {
	for _, source := range sources {
		if source == nil || isNilSource(source) {
			destination = append(destination, nil)
		} else {
			info := source.Info()
			info.Priority = priority
			destination = append(destination, prioritySource{source: source, info: info})
		}
	}
	return destination
}
