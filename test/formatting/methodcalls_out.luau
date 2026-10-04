local x = {}
x:func(whatever)
x:func()
x:func "str"
x:func {
	a = 1,
}
self:method(1, 2)
x:func(1)
x:func<<number>>(1)
x.func(y, whatever)
x.func(x.y, whatever)
x.y.func(x.y, whatever)
f().func(f(), whatever)
x.func(x :: T, 1)
x:func(1)
instance:func(1)
