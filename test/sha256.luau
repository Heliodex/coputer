-- Not using boatbomber and howmanysmaII's implementation because I DESPISE howmanysmaII

local W = {}

local sha2Khi = {
	0x428a2f98,
	0x71374491,
	0xb5c0fbcf,
	0xe9b5dba5,
	0x3956c25b,
	0x59f111f1,
	0x923f82a4,
	0xab1c5ed5,
	0xd807aa98,
	0x12835b01,
	0x243185be,
	0x550c7dc3,
	0x72be5d74,
	0x80deb1fe,
	0x9bdc06a7,
	0xc19bf174,
	0xe49b69c1,
	0xefbe4786,
	0x0fc19dc6,
	0x240ca1cc,
	0x2de92c6f,
	0x4a7484aa,
	0x5cb0a9dc,
	0x76f988da,
	0x983e5152,
	0xa831c66d,
	0xb00327c8,
	0xbf597fc7,
	0xc6e00bf3,
	0xd5a79147,
	0x06ca6351,
	0x14292967,
	0x27b70a85,
	0x2e1b2138,
	0x4d2c6dfc,
	0x53380d13,
	0x650a7354,
	0x766a0abb,
	0x81c2c92e,
	0x92722c85,
	0xa2bfe8a1,
	0xa81a664b,
	0xc24b8b70,
	0xc76c51a3,
	0xd192e819,
	0xd6990624,
	0xf40e3585,
	0x106aa070,
	0x19a4c116,
	0x1e376c08,
	0x2748774c,
	0x34b0bcb5,
	0x391c0cb3,
	0x4ed8aa4a,
	0x5b9cca4f,
	0x682e6ff3,
	0x748f82ee,
	0x78a5636f,
	0x84c87814,
	0x8cc70208,
	0x90befffa,
	0xa4506ceb,
	0xbef9a3f7,
	0xc67178f2,
	0xca273ece,
	0xd186b8c7,
	0xeada7dd6,
	0xf57d4f7f,
	0x06f067aa,
	0x0a637dc5,
	0x113f9804,
	0x1b710b35,
	0x28db77f5,
	0x32caab7b,
	0x3c9ebe0a,
	0x431d67c4,
	0x4cc5d4be,
	0x597f299c,
	0x5fcb6fab,
	0x6c44198c,
}

local function sha256feed64(
	H: { number },
	str: string,
	offs: number,
	size: number
)
	-- offs >= 0, size >= 0, size is multiple of 64
	local h1, h2, h3, h4, h5, h6, h7, h8 =
		H[1], H[2], H[3], H[4], H[5], H[6], H[7], H[8]
	for pos = offs, offs + size - 1, 64 do
		for j = 1, 16 do
			pos += 4
			local a, b, c, d = string.byte(str, pos - 3, pos)
			W[j] = ((a * 256 + b) * 256 + c) * 256 + d
		end
		for j = 17, 64 do
			local a, b = W[j - 15], W[j - 2]
			local a7, a18, b17, b19 =
				a / 2 ^ 7, a / 2 ^ 18, b / 2 ^ 17, b / 2 ^ 19
			W[j] = (
				bit32.bxor(
					a7 % 1 * (2 ^ 32 - 1) + a7,
					a18 % 1 * (2 ^ 32 - 1) + a18,
					(a - a % 2 ^ 3) / 2 ^ 3
				)
				+ W[j - 16]
				+ W[j - 7]
				+ bit32.bxor(
					b17 % 1 * (2 ^ 32 - 1) + b17,
					b19 % 1 * (2 ^ 32 - 1) + b19,
					(b - b % 2 ^ 10) / 2 ^ 10
				)
			) % 2 ^ 32
		end
		local a, b, c, d, e, f, g, h = h1, h2, h3, h4, h5, h6, h7, h8
		for j = 1, 64 do
			e %= 2 ^ 32
			local e6, e11, e7 = e / 2 ^ 6, e / 2 ^ 11, e * 2 ^ 7
			local e7_lo = e7 % 2 ^ 32
			local z = bit32.band(e, f)
				+ bit32.band(-1 - e, g)
				+ h
				+ sha2Khi[j]
				+ W[j]
				+ bit32.bxor(
					e6 % 1 * (2 ^ 32 - 1) + e6,
					e11 % 1 * (2 ^ 32 - 1) + e11,
					e7_lo + (e7 - e7_lo) / 2 ^ 32
				)
			h = g
			g = f
			f = e
			e = z + d
			d = c
			c = b
			b = a % 2 ^ 32
			local b2, b13, b10 = b / 2 ^ 2, b / 2 ^ 13, b * 2 ^ 10
			local b10_lo = b10 % 2 ^ 32
			a = z
				+ bit32.band(d, c)
				+ bit32.band(b, bit32.bxor(d, c))
				+ bit32.bxor(
					b2 % 1 * (2 ^ 32 - 1) + b2,
					b13 % 1 * (2 ^ 32 - 1) + b13,
					b10_lo + (b10 - b10_lo) / 2 ^ 32
				)
		end
		h1, h2, h3, h4, h5, h6, h7, h8 =
			(a + h1) % 2 ^ 32,
			(b + h2) % 2 ^ 32,
			(c + h3) % 2 ^ 32,
			(d + h4) % 2 ^ 32,
			(e + h5) % 2 ^ 32,
			(f + h6) % 2 ^ 32,
			(g + h7) % 2 ^ 32,
			(h + h8) % 2 ^ 32
	end
	H[1], H[2], H[3], H[4], H[5], H[6], H[7], H[8] =
		h1, h2, h3, h4, h5, h6, h7, h8
end

local function byteify(chunks: { number }): { number }
	-- chunks is an array of 4-byte numbers
	local result = {}
	for i = 1, #chunks do
		local chunk = chunks[i]

		table.insert(result, chunk // 2 ^ 24)
		table.insert(result, chunk // 2 ^ 16 % 256)
		table.insert(result, chunk // 2 ^ 8 % 256)
		table.insert(result, chunk % 256)
	end
	return result
end

local hexs = {
	"0",
	"1",
	"2",
	"3",
	"4",
	"5",
	"6",
	"7",
	"8",
	"9",
	"a",
	"b",
	"c",
	"d",
	"e",
	"f",
}

local function bytesToHex(bytes: { number }): string
	local hex = ""
	for _, v in bytes do
		local v1, v2 = v // 16, v % 16
		hex ..= hexs[v1 + 1] .. hexs[v2 + 1]
	end
	return hex
end

local function sha(message: string): string
	-- Create an instance (private objects for current calculation)
	local H = {
		0x6a09e667,
		0xbb67ae85,
		0x3c6ef372,
		0xa54ff53a,
		0x510e527f,
		0x9b05688c,
		0x1f83d9ab,
		0x5be0cd19,
	}
	local length = #message
	local offs = 0
	local size = length - offs
	local size_tail = size % 64
	sha256feed64(H, message, offs, size - size_tail)

	local final_blocks = {
		string.sub(message, length + 1 - size_tail),
		"\128",
		string.rep("\0", (-9 - length) % 64 + 1),
	}

	length *= (8 / 256 ^ 7) -- convert "byte-counter" to "bit-counter" and move decimal point to t1 left
	for j = 4, 10 do
		length = (length % 1) * 256
		final_blocks[j] = string.char(length // 1)
	end
	local blockstr = table.concat(final_blocks)
	sha256feed64(H, blockstr, 0, #blockstr)
	return bytesToHex(byteify(H))
end

print(sha "Hello, world!")
print(sha "What's up?")
print(sha "")
print(sha(sha ""))
print(sha(sha "" .. sha ""))
print(
	sha "TWFuIGlzIGRpc3Rpbmd1aXNoZWQsIG5vdCBvbmx5IGJ5IGhpcyByZWFzb24sIGJ1dCBieSB0aGlzIHNpbmd1bGFyIHBhc3Npb24gZnJvbSBvdGhlciBhbmltYWxzLCB3aGljaCBpcyBhIGx1c3Qgb2YgdGhlIG1pbmQsIHRoYXQgYnkgYSBwZXJzZXZlcmFuY2Ugb2YgZGVsaWdodCBpbiB0aGUgY29udGludWVkIGFuZCBpbmRlZmF0aWdhYmxlIGdlbmVyYXRpb24gb2Yga25vd2xlZGdlLCBleGNlZWRzIHRoZSBzaG9ydCB2ZWhlbWVuY2Ugb2YgYW55IGNhcm5hbCBwbGVhc3VyZS4"
)
