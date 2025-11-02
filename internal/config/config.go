package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/RyRose/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

func Load(ctx context.Context, configPath string) (*Data, error) {
	L := lua.NewState()
	L.SetContext(ctx)
	L.OpenLibs()
	defer L.Close()

	if err := L.DoFile(configPath); err != nil {
		return nil, fmt.Errorf("failed to load config file %s: %w", configPath, err)
	}

	lv := L.Get(-1)
	if lv.Type() != lua.LTTable {
		return nil, fmt.Errorf("config file must return a table, got %s", lv.Type())
	}

	var data Data
	mapper := gluamapper.NewMapper(gluamapper.Option{NameFunc: ToUpperCamelCase})
	if err := mapper.Map(lv.(*lua.LTable), &data); err != nil {
		return nil, fmt.Errorf("failed to map lua table to config: %w", err)
	}

	return &data, nil
}

func GenerateLuaTypesFile() (string, error) {
	lines := []string{strings.Join([]string{
		"---@meta",
		"",
		"-- This file is auto-generated. Do not edit manually.",
		"--",
		"-- Lua type definitions for Uplog configuration. These are",
		"-- automatically generated from types in internal/config/data.go.",
		"-- Run `make types` to regenerate.",
	}, "\n")}
	for _, t := range LuaTypes() {
		s, err := GenerateLuaType(t)
		if err != nil {
			return "", fmt.Errorf("file so far:\n%s\ngenerating lua type for %T: %w", strings.Join(lines, "\n\n"), t, err)
		}
		lines = append(lines, s)
	}
	return strings.Join(lines, "\n\n"), nil
}
