# What is yet to be done

 - [ ] [fix flakyness][#flakyness]
 - [ ] [caching store][#caching-store]

## <a href=#flakyness>Fix flakyness of integration tests</a>

From time to time the integration tests for the postgres-backed store
implementation fail, seemingly returning items out of the expected order. Test
if this is a bug of the implementation or of the tests.

## <a href=#caching-store>Implement a caching store</a>

Every operation on the store is cacheable, thanks to the univocity of the input.
`Add`, `Update` and `Remove` are all univocal. To cache `List` and `Page` one
can leverage the `Hash` method of each `Condition`.
