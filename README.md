# ExploreNYC-backend

This codebase handles half of the process for generating a user trip for ExploreNYC. This codebase handles acquiring geolocation and transit data from multiple sources such as Google Maps, Mapbox and TomTom. This codebase is responsible for sending this data to the CP-SAT python codebase to generate a route, then polishing the result to output to the frontend.


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
