# AsesemblyAi Go

This is a Go client of the AsesemblyAi API.

## Installation

```bash
go get github.com/assemblyai/assembly-ai-go
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/assemblyai/assembly-ai-go"

    func main() {
        client := assemblyai.New("https://api.AssemblyAI.com/v2", "my-api-key")
        resp, err := client.Transcript("https://www.youtube.com/watch?v=QH2-TGUlwu4")
        if err != nil {
            log.Fatal(err)
        }
        log.Println(resp)
    }
)
```

## License

MIT


