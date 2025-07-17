local frozen = table.freeze({ "one", "two", "three" }) -- doesn't work for math in ref imp 4 sum reason

table.remove(frozen)
