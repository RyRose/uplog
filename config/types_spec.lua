describe("types", function()
	local types

	before_each(function()
		types = require("config.types")
	end)

	describe("type constants", function()
		it("should define Boolean constant", function()
			assert.is_false(types.Boolean)
		end)

		it("should define String constant", function()
			assert.equal("", types.String)
		end)

		it("should define Number constant", function()
			assert.equal(0, types.Number)
		end)

		it("should define Nil constant", function()
			assert.is_nil(types.Nil)
		end)
	end)

	describe("type string constants", function()
		it("should define BooleanS as 'boolean'", function()
			assert.equal("boolean", types.BooleanS)
		end)

		it("should define StringS as 'string'", function()
			assert.equal("string", types.StringS)
		end)

		it("should define NumberS as 'number'", function()
			assert.equal("number", types.NumberS)
		end)

		it("should define NilS as 'nil'", function()
			assert.equal("nil", types.NilS)
		end)
	end)
end)
