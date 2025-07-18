local it = 69

for i, v in coroutine.create(next), it do
	print(i, v)
end
