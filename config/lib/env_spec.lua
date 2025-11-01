describe("env", function()
	local env
	local types

	before_each(function()
		package.loaded["config.env"] = nil
		package.loaded["config.types"] = nil
		types = require("config.lib.types")
		env = require("config.lib.env")
	end)

	describe("Or", function()
		describe("with string defaults", function()
			it("should return default when env var is not set", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function()
					return nil
				end
				assert.equal("default", env.Or("TEST_VAR", "default"))
			end)

			it("should return env var value when set", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return "value"
					end
				end
				assert.equal("value", env.Or("TEST_VAR", "default"))
			end)

			it("should trim whitespace from env var", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return "  value  "
					end
				end
				assert.equal("value", env.Or("TEST_VAR", "default"))
			end)

			it("should return default when env var is empty string", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return ""
					end
				end
				assert.equal("default", env.Or("TEST_VAR", "default"))
			end)

			it("should return default when env var is whitespace only", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return "   "
					end
				end
				assert.equal("default", env.Or("TEST_VAR", "default"))
			end)
		end)

		describe("with boolean defaults", function()
			it("should return default when env var is not set", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function()
					return nil
				end
				assert.is_false(env.Or("TEST_VAR", false))
			end)

			it("should return true for '1'", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return "1"
					end
				end
				assert.is_true(env.Or("TEST_VAR", false))
			end)

			it("should return true for 'true'", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return "true"
					end
				end
				assert.is_true(env.Or("TEST_VAR", false))
			end)

			it("should return true for 'TRUE' (case insensitive)", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return "TRUE"
					end
				end
				assert.is_true(env.Or("TEST_VAR", false))
			end)

			it("should return true for 'yes'", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return "yes"
					end
				end
				assert.is_true(env.Or("TEST_VAR", false))
			end)

			it("should return true for 'on'", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return "on"
					end
				end
				assert.is_true(env.Or("TEST_VAR", false))
			end)

			it("should return false for other values", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return "0"
					end
				end
				assert.is_false(env.Or("TEST_VAR", true))
			end)

			it("should return false for 'false'", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return "false"
					end
				end
				assert.is_false(env.Or("TEST_VAR", true))
			end)
		end)

		describe("with number defaults", function()
			it("should return default when env var is not set", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function()
					return nil
				end
				assert.equal(42, env.Or("TEST_VAR", 42))
			end)

			it("should parse integer values", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return "123"
					end
				end
				assert.equal(123, env.Or("TEST_VAR", 0))
			end)

			it("should parse float values", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return "3.14"
					end
				end
				assert.equal(3.14, env.Or("TEST_VAR", 0))
			end)

			it("should parse negative numbers", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return "-42"
					end
				end
				assert.equal(-42, env.Or("TEST_VAR", 0))
			end)

			it("should throw error for invalid numeric values", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return "not a number"
					end
				end
				assert.has_error(function()
					env.Or("TEST_VAR", 0)
				end, 'invalid numeric value for TEST_VAR: "not a number"')
			end)
		end)
	end)

	describe("Get", function()
		describe("with string type", function()
			it("should return nil when env var is not set", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function()
					return nil
				end
				assert.is_nil(env.Get("TEST_VAR", types.String))
			end)

			it("should return env var value when set", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return "value"
					end
				end
				assert.equal("value", env.Get("TEST_VAR", types.String))
			end)

			it("should trim whitespace from env var", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return "  value  "
					end
				end
				assert.equal("value", env.Get("TEST_VAR", types.String))
			end)

			it("should return nil when env var is empty string", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return ""
					end
				end
				assert.is_nil(env.Get("TEST_VAR", types.String))
			end)
		end)

		describe("with boolean type", function()
			it("should return nil when env var is not set", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function()
					return nil
				end
				assert.is_nil(env.Get("TEST_VAR", types.Boolean))
			end)

			it("should return true for 'true'", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return "true"
					end
				end
				assert.is_true(env.Get("TEST_VAR", types.Boolean))
			end)

			it("should return false for other values", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return "false"
					end
				end
				assert.is_false(env.Get("TEST_VAR", types.Boolean))
			end)
		end)

		describe("with number type", function()
			it("should return nil when env var is not set", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function()
					return nil
				end
				assert.is_nil(env.Get("TEST_VAR", types.Number))
			end)

			it("should parse integer values", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return "456"
					end
				end
				assert.equal(456, env.Get("TEST_VAR", types.Number))
			end)

			it("should throw error for invalid numeric values", function()
				--- @diagnostic disable-next-line: duplicate-set-field
				os.getenv = function(key)
					if key == "TEST_VAR" then
						return "invalid"
					end
				end
				assert.has_error(function()
					env.Get("TEST_VAR", types.Number)
				end, 'invalid numeric value for TEST_VAR: "invalid"')
			end)
		end)
	end)
end)
