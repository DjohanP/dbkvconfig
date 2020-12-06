package dbkvconfig

// Const Default Config
const (
	DefaultApplication = "Default-Application"
	KeyRedisDefault    = "dbkvconfiglistener"
)

// Const For Error
const (
	ConfigNotFoundMsg = "ConfigNotFound"
	UpdateTooFast     = "UpdateConfigIsTooFast"
)

// Const Field Redis Without PubSub
const (
	fieldLastChange      = "field"
	fieldLastChangesTime = "time_last_changes"
)
