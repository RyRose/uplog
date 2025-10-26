local env = require("config.env")
local types = require("config.types")

local port = env.Or("PORT", "8080")

---@type Data
local M = {
	debug = env.Or("DEBUG", false),
	version = env.Get("VERSION", types.String),
	database_path = env.Or("DATABASE_PATH", "./tmp/db/data.db"),
	port = port,
	swagger_url = "http://localhost:" .. port .. "/docs/swagger.json",
}

return M
