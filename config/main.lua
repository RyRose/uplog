local env = require("config.lib.env")
local types = require("config.lib.types")

local port = env.Or("PORT", "8080")
local curtime = os.time()
local version = env.Or("VERSION", "auto-" .. tostring(curtime))

---@type Data
local M = {
	debug = env.Or("DEBUG", false),
	version = version,
	database_path = env.Or("DATABASE_PATH", "./tmp/db/data.db"),
	port = port,
	swagger_url = env.Or("SWAGGER_URL", "http://localhost:" .. port .. "/docs/swagger.json") .. "?v=" .. version,
	first_day_of_week = 0, -- 0 = Sunday, 1 = Monday
}

return M
