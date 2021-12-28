# omise-go-challenge
Go-Challenge of Omise

# Points
- Distribute API call across multiple api accounts to run as fast as possible
- Use simple http client to reduce memory allocation
- Validate expiration date of cards in local to omit wasteful api calls

# Getting Start
```sh
# exports
$ export GOPATH=$(go env GOPATH)
$ export PATH="$GOPATH/bin:$PATH"

# install
$ make install

# run
$ omise-go-challenge ./data/fng.1000.csv.rot128

performing donations...
done.

           total received: THB             2,682,036,460
     successfully donated: THB             1,119,602,468
          faulty donation: THB             1,562,433,992

       average per person: THB              2,687,411.28
               top donors: THB Mrs. Mimosa R Tûk
                           THB Mr. Falco S Bracegirdle
                           THB Ms. Primrose N Fairbairn
```

# Client Bench Result
```
cpu: Intel(R) Core(TM) i5-8257U CPU @ 1.40GHz
BenchmarkClient-8        	    1316	   9316712 ns/op	  377457 B/op	    4903 allocs/op
BenchmarkOmiseClient-8   	    1142	  10581657 ns/op	  589926 B/op	    7307 allocs/op
```
