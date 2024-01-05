package registry

// Config is the JSON encoded file for a set of related projects
type Config struct {
	// Services is a map of name => relative file path for the project
	Services map[string]string `json:"services"`
}
