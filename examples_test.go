package kmip

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
	"github.com/gsealy/kmip-go/kmip14"
	"github.com/gsealy/kmip-go/ttlv"
	"net"
	"time"
)

func Example_client() {

	conn, err := net.DialTimeout("tcp", "localhost:5696", 3*time.Second)
	if err != nil {
		panic(err)
	}

	biID := uuid.New()

	msg := RequestMessage{
		RequestHeader: RequestHeader{
			ProtocolVersion: ProtocolVersion{
				ProtocolVersionMajor: 1,
				ProtocolVersionMinor: 2,
			},
			BatchCount: 1,
		},
		BatchItem: []RequestBatchItem{
			{
				UniqueBatchItemID: biID[:],
				Operation:         kmip14.OperationDiscoverVersions,
				RequestPayload: DiscoverVersionsRequestPayload{
					ProtocolVersion: []ProtocolVersion{
						{ProtocolVersionMajor: 1, ProtocolVersionMinor: 2},
					},
				},
			},
		},
	}

	req, err := ttlv.Marshal(msg)
	if err != nil {
		panic(err)
	}

	fmt.Println(req)

	_, err = conn.Write(req)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 5000)
	_, err = bufio.NewReader(conn).Read(buf)
	if err != nil {
		panic(err)
	}

	resp := ttlv.TTLV(buf)
	fmt.Println(resp)

}

func ExampleServer() {
	listener, err := net.Listen("tcp", "0.0.0.0:5696")
	if err != nil {
		panic(err)
	}

	DefaultProtocolHandler.LogTraffic = true

	DefaultOperationMux.Handle(kmip14.OperationDiscoverVersions, &DiscoverVersionsHandler{
		SupportedVersions: []ProtocolVersion{
			{
				ProtocolVersionMajor: 1,
				ProtocolVersionMinor: 4,
			},
			{
				ProtocolVersionMajor: 1,
				ProtocolVersionMinor: 3,
			},
			{
				ProtocolVersionMajor: 1,
				ProtocolVersionMinor: 2,
			},
		},
	})
	srv := Server{}
	panic(srv.Serve(listener))

}
