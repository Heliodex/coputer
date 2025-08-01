local layout = require "./layout"
local e, c = layout.elements, layout.config

return function(): layout.Tags
	return {
		e.h(1, { "System" }),
		e.box { padding = { top = 4, bottom = 6 } } {
			e.p {
				"This page provides some explanation for how the Coputer network works by showing how this page was generated and served to you. The network is split up into a number of separate systems, so the function of each will be outlined here.",
			},
		},

		e.h(2, { "Gateway" }),
		e.box { padding = { top = 4, bottom = 6 } } {
			e.p {
				"The process begins when your browser sends a request is sent to a ",
				e.bold "gateway",
				", the entrypoint for the Coputer network for any programs written to produce web requests. ",
				e.small "(aka all of them, since web programs are the only available program type at the moment)",
			},
			e.p {
				"The gateway, such as the one hosted at ",
				e.outlink "https://coputer.heliodex.cf" { "coputer.heliodex.cf" },
				", looks at the parameters of the request and determines which program to execute. Programs are identified by a name and the public key of the publisher. The gateway requests the nearest ",
				e.bold "communication system",
				" for the output of the program, providing it with its name, public key, and request parameters.",
			},
			e.box { padding = { top = 4 } } {
				e.outlink "https://github.com/Heliodex/coputer/tree/main/gateway" {
					"Source code for the gateway reference implementation",
				},
			},
		},

		e.h(2, { "Communication system" }),
		e.box { padding = { top = 4, bottom = 6 } } {
			e.p {
				"The communication system is a service responsible for key management and connecting and distributing Coputer programs and their outputs. The program's publisher will have used it to store the program and distribute it across the network.",
			},
			e.p {
				"The communication system receives the request from the gateway and asks its ",
				e.bold "execution system",
				" to process the request. If it fails (can't connect, or the program is not stored, cannot be executed, returns an error etc), the communication system will send out a request to its peers for the program output. If the output is found it is sent back to the gateway, otherwise an error is returned.",
			},
			e.box { padding = { top = 4 } } {
				e.outlink "https://github.com/Heliodex/coputer/tree/main/wallflower" {
					"Source code for the communication system reference implementation",
				},
			},
		},

		e.h(2, { "Execution system" }),
		e.box { padding = { top = 4, bottom = 6 } } {
			e.p {
				"The execution system is responsible for storing, executing, and caching responses of Coputer programs. Upon receiving the request to process the program, it checks if the program is stored, returning an error if it isn't. Then it will check if the program output is already cached, returning the cached response if it is.",
			},
			e.p {
				"If we have the program and its response is not cached, it's now time to run it. The entrypoint to the program is compiled then executed by the virtual machine, the most complex part of the Coputer system. The VM dynamically compiles any dependencies and executes the program in a sandboxed, deterministic environment. Once a response is obtained, it is sent back through the communication system to the gateway, completing the request.",
			},
			e.box { padding = { top = 4 } } {
				e.outlink "https://github.com/Heliodex/coputer/tree/main/litecode" {
					"Source code for the execution system reference implementation",
				},
			},
		},
	}
end
