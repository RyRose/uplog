package config

// Data holds the configuration data for the application typically loaded from a Lua file.
// To log this struct, to use slog's JSON logging to ensure certain fields are expunged.
// Sensitive fields should be marked with `json:"-"` struct tags.
type Data struct {
	// Debug enables debug mode. When enabled, the application will log additional
	// debug information.
	Debug bool
	// Version specifies the application version.
	Version *string
	// DatabasePath specifies the path to the database file.
	DatabasePath string
	// Port specifies the port on which the application will run.
	Port string
	// SwaggerURL specifies the URL for the Swagger documentation.
	SwaggerURL string
	// FirstDayOfWeek specifies the first day of the week (0 = Sunday, 1 = Monday, etc.).
	FirstDayOfWeek int
}

// LuaTypes returns a slice of config types for Lua integration.
func LuaTypes() []any {
	return []any{
		Data{},
	}
}
