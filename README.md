# Rookout Go SDK
[Rookout](https://rookout.com)'s Go SDK is a Go module that supports on the fly debugging and data extraction from Go applications in production. 

## Compatibility
Go versions 1.15 and higher are supported. 

Linux and Mac operating systems are supported. 

## Installation

Add the Rookout SDK by running: 
```bash
go get -d github.com/Rookout/GoSDK
```

## Usage

Import Rookout in the file containing your `main` package:
```go
import (
    rookout "github.com/Rookout/GoSDK"
)
```

Call `rookout.Start` as early as possible in your `main` function: 
```go
rookout.Start(rookout.RookOptions{token=[Your Rookout Token]})
```

Build your application:
```bash
go build -gcflags="all=-dwarflocationlists=true"
```

## Advanced 

For more info see the complete setup guide [in our docs](https://docs.rookout.com/docs/go-setup/).