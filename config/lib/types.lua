-- types contains the Lua type definitions for Uplog configuration.

local M = {}

M.Boolean = false
M.String = ""
M.Number = 0
M.Nil = nil

M.BooleanS = type(M.Boolean)
M.StringS = type(M.String)
M.NumberS = type(M.Number)
M.NilS = type(M.Nil)

return M
