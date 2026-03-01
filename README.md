# ExploreNYC-backend



## Running Tests
To run tests for a certain package (such as edges), use the following command. the '-v' flag enables verbose mode, which prints the test output
```
go test ./integrations/edges -v

```
since main_test.go is in the root or main package, you use a period in place
```
go test .
```

## Running Main
You use a period as main
```
go run .
```