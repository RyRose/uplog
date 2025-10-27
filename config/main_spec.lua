describe("main", function()
	local main

	before_each(function()
		package.loaded["config.main"] = nil
		package.loaded["config.env"] = nil
		package.loaded["config.types"] = nil
	end)

	describe("structural snapshot", function()
		it("should match expected structure with defaults", function()
			--- @diagnostic disable-next-line: duplicate-set-field
			os.getenv = function()
				return nil
			end
			main = require("config.main")

			-- Version will be auto-generated with timestamp, so we need to check it exists
			assert.is_not_nil(main.version)
			assert.is_string(main.version)
			assert.is_truthy(main.version:match("^auto%-"))

			-- Swagger URL should include version query param
			local expected_swagger_prefix = "http://localhost:8080/docs/swagger.json?v=auto-"
			assert.is_truthy(main.swagger_url:sub(1, #expected_swagger_prefix) == expected_swagger_prefix)

			-- Check other fields with exact values
			assert.equal(false, main.debug)
			assert.equal("./tmp/db/data.db", main.database_path)
			assert.equal("8080", main.port)
		end)

		it("should match expected structure with all env vars set", function()
			--- @diagnostic disable-next-line: duplicate-set-field
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
				swagger_url = "http://localhost:3000/docs/swagger.json?v=1.0.0",
			}

			assert.same(expected, main)
		end)
	end)
end)
