# Chio

Chio (Bancho Input/Output) is a bancho packet serialization library written in go, that aims to support as many clients as possible.

> [!WARNING]
> This project is a work in progress and is not ready for production usage yet.
> For a full implementation, please view [chio.py](https://github.com/Lekuruu/chio.py)!

## Example Usage

```go
import (
    "io"
    "fmt"

    "github.com/lekuruu/chio"
)

// Assuming you have some kind of tcp server
func HandleConnection(stream io.ReadWriteCloser, version int) {
    defer stream.Close()

    io := chio.GetClientInterface(stream, version)
    io.WriteLoginReply(2)
    io.WriteUserStats(chio.UserInfo{ ... })
    io.WriteAnnouncement("Hello, World!")

    for {
        packet, err := client.IO.ReadPacket()
        if err != nil {
            fmt.Println("Error reading packet:", err.Error())
            break
        }

        fmt.Printf("Received packet: %d, %v\n", packet.Id, packet.Data)
    }
}
```
