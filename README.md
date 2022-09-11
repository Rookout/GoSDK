# Rookout Go SDK
[Rookout](https://rookout.com)'s Go SDK is a Go module that supports on the fly debugging and data extraction from Go applications in production. 

## Installation
Run the following from inside your Go application's directory to add the Rookout SDK to the `go.mod` file. 
```bash
go get -d github.com/Rookout/GoSDK
```

## Setup
Go to the file containing your `main` package and import Rookout:
```javascript
import (
    rookout "github.com/Rookout/GoSDK"
)
```
Then in your `main` function, call `rookout.Start` as early as possible: 
```javascript
err := rookout.Start(rookout.RookOptions{token=[your-rookout-token]})
fmt.Println(err)
```

For more info see [Rookout's docs](https://docs.rookout.com/docs/go-setup/).