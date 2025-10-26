package config

import "log/slog"

type Data struct {
	Debug        bool
	Version      *string
	DatabasePath string
	Port         string
	SwaggerURL   string
}

func (u Data) LogValue() slog.Value {
	// Custom representation for logging
	groups := []slog.Attr{
		slog.Bool("debug", u.Debug),
		slog.String("database_path", u.DatabasePath),
		slog.String("port", u.Port),
		slog.String("swagger_url", u.SwaggerURL),
	}
	if u.Version != nil {
		groups = append(groups, slog.String("version", *u.Version))
	}
	return slog.GroupValue(groups...)
}

// LuaTypes returns a slice of config types for Lua integration.
func LuaTypes() []any {
	return []any{
		Data{},
	}
}
