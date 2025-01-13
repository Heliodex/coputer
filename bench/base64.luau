local base64 = {}

function base64.makeencoder()
	local encoder = {}
	for b64code, char in
		{
			[0] = "A",
			"B",
			"C",
			"D",
			"E",
			"F",
			"G",
			"H",
			"I",
			"J",
			"K",
			"L",
			"M",
			"N",
			"O",
			"P",
			"Q",
			"R",
			"S",
			"T",
			"U",
			"V",
			"W",
			"X",
			"Y",
			"Z",
			"a",
			"b",
			"c",
			"d",
			"e",
			"f",
			"g",
			"h",
			"i",
			"j",
			"k",
			"l",
			"m",
			"n",
			"o",
			"p",
			"q",
			"r",
			"s",
			"t",
			"u",
			"v",
			"w",
			"x",
			"y",
			"z",
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
			"+",
			"/",
			"=",
		}
	do
		encoder[b64code] = string.byte(char)
	end
	return encoder
end

function base64.makedecoder()
	local decoder = {}
	for b64code, charcode in pairs(base64.makeencoder()) do
		decoder[charcode] = b64code
	end
	return decoder
end

local extract, char, concat, byte =
	bit32.extract, string.char, table.concat, string.byte

function base64.decode(b64)
	local decoder = base64.makedecoder()
	local t, k = {}, 1
	local n = #b64
	local padding = string.sub(b64, -2) == "==" and 2
		or string.sub(b64, -1) == "=" and 1
		or 0
	for i = 1, padding > 0 and n - 4 or n, 4 do
		local a, b, c, d = byte(b64, i, i + 3)
		local s
		local v = decoder[a] * 0x40000
			+ decoder[b] * 0x1000
			+ decoder[c] * 0x40
			+ decoder[d]

		local b1 = extract(v, 16, 8)
		local b2 = extract(v, 8, 8)
		local b3 = extract(v, 0, 8)
		s = char(b1, b2, b3)
		t[k] = s
		k += 1
	end
	if padding == 1 then
		local a, b, c = byte(b64, n - 3, n - 1)
		local v = decoder[a] * 0x40000 + decoder[b] * 0x1000 + decoder[c] * 0x40

		local b1 = extract(v, 16, 8)
		local b2 = extract(v, 8, 8)
		t[k] = char(b1, b2)
	elseif padding == 2 then
		local a, b = byte(b64, n - 3, n - 2)
		local v = decoder[a] * 0x40000 + decoder[b] * 0x1000

		local b1 = extract(v, 16, 8)
		t[k] = char(b1)
	end
	return concat(t)
end

local b
for i = 1, 200 do
	b =
		base64.decode "TWFuIGlzIGRpc3Rpbmd1aXNoZWQsIG5vdCBvbmx5IGJ5IGhpcyByZWFzb24sIGJ1dCBieSB0aGlzIHNpbmd1bGFyIHBhc3Npb24gZnJvbSBvdGhlciBhbmltYWxzLCB3aGljaCBpcyBhIGx1c3Qgb2YgdGhlIG1pbmQsIHRoYXQgYnkgYSBwZXJzZXZlcmFuY2Ugb2YgZGVsaWdodCBpbiB0aGUgY29udGludWVkIGFuZCBpbmRlZmF0aWdhYmxlIGdlbmVyYXRpb24gb2Yga25vd2xlZGdlLCBleGNlZWRzIHRoZSBzaG9ydCB2ZWhlbWVuY2Ugb2YgYW55IGNhcm5hbCBwbGVhc3VyZS4="
end

print(b)
