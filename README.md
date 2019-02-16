# httpbulk-go

httpbulk-go is a small wrapper lib, intended to ease the parallel loading of multiple http resources.

This is particularly useful, if you are aggregating an object from multiple resources.

Download:

```
go get github.com/kernle32dll/httpbulk-go
```

### Usage

First, you have to instantiate a `bulk.Executor`. This is done via `bulk.NewExecutor` (which takes option style parameters).

```go
bulk.NewExecutor(bulk.Client(http.DefaultClient), bulk.ConcurrencyLimit(10))
```

With the executor instantiated, you can issue asynchronous requests via the `AddRequests` method, which returns a
`chan Result` slice. Read from these channels, to retrieve the results.

```go
urls := []string{
    "https://www.google.com",
    "https://www.bing.com",
    "https://www.yahoo.com",
    "https://www.tarent.de",
}

results := executor.AddRequests(context.Background(), urls...)
```

The `Result` object has several methods to introspect the response. The most useful being `Res()` for getting the
original `*http.Response`, and `Err()` for getting the error, if any ocured while fetching.

**Implementation note**: If the context used for the requests is canceled, or exceeds its deadline, the corresponding
error is propagated in the `Result` object.

For more control, you can use the `AddRequestsWithInterceptor` method, which allows you to modify the request prior to sending.
This might be useful for setting headers and/or changing the request type.

```go
executor.AddRequestsWithInterceptor(context.Background(), func(r *http.Request) error {
    // Change the request method to HEAD
    r.Method = "HEAD"
    return nil
}, urls...)
```

### Thanks

This lib has been derived from the following code gist. All kudos to Montana Flynn (montanaflynn)

https://gist.github.com/montanaflynn/ea4b92ed640f790c4b9cee36046a5383

