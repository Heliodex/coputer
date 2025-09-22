local layout = require "./layout"
local e = layout.elements

return {
	e.h(1, { "Programs" }),
	e.box { padding = { top = 4, bottom = 6 } } {
		e.p {
			"This page shows how Coputer programs are written, stored, and executed, and some of the rationale behind the design decisions.",
		},
	},

	e.h(2, { "Language" }, "language"),
	e.box { padding = { top = 4 } } {
		e.p {
			"All Coputer programs are written in ",
			e.outlink "https://luau.org" { "Luau" },
			", a gradually typed scripting language derived from Lua. Luau is ",
			e.outlink "https://create.roblox.com/docs/luau" {
				"maintained and developed by Roblox",
			},
			".",
		},
		e.box {
			padding = { top = 2, bottom = 4, left = 4 },
			list = {
				direction = "horizontal",
				gap = 4,
			},
		} {
			e.image(
				"https://raw.githubusercontent.com/luau-lang/site/2d33fe820e91481dee7e128f42b385489e193905/logo.svg",
				"Luau logo",
				{ "The Luau logo" },
				35
			),
			e.image(
				"https://rfcs.luau.org/mascot.png",
				"Luau seal mascot",
				{ "Hina the Seal, the mascot for Luau" },
				50
			),
		},
		e.p {
			"Many options were considered before Luau, including Lua and loads of other scripting languages, a custom bytecode/VM, and WebAssembly. If an existing solution had been chosen, the project would probably be 4-6 months ahead of where it is now. The idea of using Luau was mainly out of love, since it's my favourite programming language and I learned a ton about it by integrating it into Coputer.",
		},
	},

	e.box { padding = { top = 4, bottom = 6 } } {
		e.p {
			"Programs are executed entirely within the ",
			e.link "/system#execution-system" { "Execution system" },
			", using a custom Luau VM implementation. The custom VM supports almost all of the core language and most of the standard library, with any exclusions being deliberate for determinism or security reasons.",
		},
		e.p {
			"Each program is bundled as source code in a compressed format (files are individually gzipped) and distributed by ",
			e.link "/system#communication-system" { "Communication systems" },
			", before finally being stored by Execution systems. The ",
			e.bold "luau-compile",
			" tool compiles the entrypoint of the program into Luau bytecode, which is executed by the VM. Any other files required by the entrypoint are loaded and compiled on demand.",
		},
	},

	e.h(2, { "Inputs and outputs" }, "inputs-and-outputs"),
	e.box { padding = { top = 4, bottom = 6 } } {
		e.p {
			"Every Coputer program has a type, which defines the shape of its inputs and outputs. The program you're currently viewing is a ",
			e.bold "Web",
			" program, meaning it takes HTTP request parameters as input and returns a response as output. The Coputer execution environment provides an ",
			e.bold "args",
			" package, allowing programs to access their input parameters. For a web program, this includes the HTTP method, headers, query parameters, and body. The program must also return a value in a specific format depending on its type.",
		},
	},

	e.h(2, { "Development" }, "development"),
	e.box { padding = { top = 4, bottom = 6 } } {
		e.p {
			"A Communication system can also implement a development mode, allowing for a program to be re-uploaded and re-executed automatically whenever its source code changes.",
		},
		e.p {
			"A development extension for VSCode, providing help to write Coputer programs, is currently a work in progress. It currently provides syntax highlighting and code formatting.",
		},
	},
} :: layout.Tags
