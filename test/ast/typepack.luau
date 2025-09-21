type Signal1<T, U...> = { data: T, f: (T, U...) -> () }
type Signal2<T, U> = { data: T, f: (T, ...U) -> () }
