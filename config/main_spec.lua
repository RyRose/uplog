describe("main", function()
	local main

	before_each(function()
		package.loaded["config.main"] = nil
		package.loaded["config.env"] = nil
		package.loaded["config.types"] = nil
	end)

	describe("structural snapshot", function()
		it("should match expected structure with defaults", function()
			os.getenv = function()
				return nil
			end
			main = require("config.main")

			local expected = {
				debug = false,
				version = nil,
				database_path = "./tmp/db/data.db",
				port = "8080",
				swagger_url = "http://localhost:8080/docs/swagger.json",
			}

			assert.same(expected, main)
		end)

		it("should match expected structure with all env vars set", function()
			os.getenv = function(key)
				if key == "DEBUG" then
					return "true"
				elseif key == "VERSION" then
					return "1.0.0"
				elseif key == "DATABASE_PATH" then
					return "/var/db/data.db"
				elseif key == "PORT" then
					return "3000"
				end
			end
			main = require("config.main")

			local expected = {
				debug = true,
				version = "1.0.0",
				database_path = "/var/db/data.db",
				port = "3000",
				swagger_url = "http://localhost:3000/docs/swagger.json",
			}

			assert.same(expected, main)
		end)
	end)
end)
