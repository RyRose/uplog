-- Utility functions for configuration.

local types = require("config.lib.types")

local M = {}

--- Returns the value of an environment variable or a default value with typing.
---
--- The environment variable is trimmed of leading and trailing whitespace.
--- If nil or empty after trimming, the default value is returned, which may be nil.
--- If not nil or empty, the value is interpreted based on the type of the `typ` value.
--- The semantics are:
---
--- * If typ is a boolean, the value is interpreted as a boolean. "1", "true", "yes",
---	  and "on" (case insensitive) are true; everything else is false.
--- * If typ is a number, the value is parsed as a number.
--- * If typ is a string or nil, the value is returned.
---
---@generic T
---@param key string
---@param default T
---@param typ T
---@return T
local function envorT(key, default, typ)
	local value = os.getenv(key)
	if value == nil then
		return default
	end
	value = value:gsub("^%s+", ""):gsub("%s+$", "")
	if value == "" then
		return default
	end

	local t = type(typ)
	if t == types.BooleanS then
		value = value:lower()
		if value == "1" or value == "true" or value == "yes" or value == "on" then
			return true
		end
		return false
	elseif t == types.NumberS then
		local num = tonumber(value)
		if num == nil then
			error(("invalid numeric value for %s: %q"):format(key, value))
		end
		return num
	elseif t == types.StringS or t == types.NilS then
		return value
	end

	error(("unsupported type %s for env var %s: %q"):format(t, key, value))
end

--- Returns the value of an environment variable or a default value.
---
--- The environment variable is trimmed of leading and trailing whitespace.
--- If nil or empty after trimming, the default value is returned, which may be nil.
--- If not nil or empty, the value is interpreted based on the type of the default value.
--- The semantics are:
---
--- * If default is a boolean, the value is interpreted as a boolean. "1", "true", "yes",
---	  and "on" (case insensitive) are true; everything else is false.
--- * If default is a number, the value is parsed as a number.
--- * If default is a string or nil, the value is returned.
--- * An error is thrown if default's type is not handled.
---
---@generic T
---@param key string
---@param default T
---@return T
function M.Or(key, default)
	return envorT(key, default, default)
end

--- Returns the value of an environment variable or nil.
---
--- The environment variable is trimmed of leading and trailing whitespace.
--- If nil or empty after trimming, nil is returned, which may be nil.
--- If not nil or empty, the value is interpreted based on the type of the second parameter.
--- The semantics are:
---
--- * If type is a boolean, the value is interpreted as a boolean. "1", "true", "yes",
---	  and "on" (case insensitive) are true; everything else is false.
--- * If type is a number, the value is parsed as a number.
--- * If type is a string or nil, the value is returned.
--- * An error is thrown if default's type is not handled.
---
---@generic T
---@param key string
---@param type T
---@return T?
function M.Get(key, type)
	return envorT(key, nil, type)
end

return M
