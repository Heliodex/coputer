type Point = { x: number, y: number }
type Array<T> = { T }
type Callback = (number, string) -> boolean
type Optional = number?
type Union = number | string | boolean
type Both = (number | string)?
type Intersection = { a: number } & { b: string }
type WithAccess = { read x: number, write y: string }
type Indexer = { [string]: number }
type Variadic = (...number) -> ...string
type Typeof = typeof(x)
type Singleton = "hello" | true
export type Exported<T = number> = T
