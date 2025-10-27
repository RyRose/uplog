local env = require("config.env")
local types = require("config.types")

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
}

return M
