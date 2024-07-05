package main

import (
	"context"

	"flag"
	pg "grpc-transmitter/database"
	gRPC "grpc-transmitter/proto"
	"io"
	"log"
	"math"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Frequency struct {
	number []float64
}

var frequencyPool = sync.Pool{}

func main() {
	var k float64
	flag.Float64Var(&k, "k", 0, "anomaly")
	flag.Parse()
	conn, err := grpc.NewClient("localhost:3333", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("error connecting to server: %v", err)
	}
	defer conn.Close()

	db, err := pg.ConnectToDB()
	if err != nil {
		log.Fatalf("error connecting to db: %v", err)
	}
	var (
		mean float64
		STD  float64
	)
	client := gRPC.NewTransmitterClient(conn)
	resp, err := client.Transmit(context.Background(), &gRPC.Request{})
	if err != nil {
		log.Fatalf("error sending request: %v", err)
	}
	for i := 0; i < 100; i++ {
		next := getPool()
		r, err := resp.Recv()
		if err != nil {
			log.Fatalf("error recieving response: %v", err)
		}
		if i < 80 {
			mean, STD = putPool(next, r.Frequency)

			if (i+1)%10 == 0 {
				log.Printf("count: %d, mean: %f, CID: = %f", i+1, mean, STD)
			}
			if i == 79 {
				getPool().Reset()
			}

		} else if r.Frequency > mean+k*STD || r.Frequency < mean-k*STD {
			log.Println("Anomaly:", r.Frequency)
			db.Create(&pg.Users{
				SessionId: r.SessionId,
				Frequency: r.Frequency,
				Timestamp: r.Timestamp.AsTime(),
			})
		}
	}

	if err == io.EOF {
		return
	}
	if err != nil {
		log.Fatalf("error recieving data from stream: %v", err)
	}

}

func getPool() *Frequency {
	mem := frequencyPool.Get()
	if mem == nil {
		return &Frequency{}
	}
	return mem.(*Frequency)
}

func (q *Frequency) Reset() {
	*q = Frequency{}
}

func putPool(q *Frequency, value float64) (mean float64, STD float64) {
	q.number = append(q.number, value)
	size := len(q.number)
	mean = 0
	STD = 0
	for n := 0; n < size; n++ {
		mean += q.number[n]
	}
	mean /= float64(size)
	if size > 1 {
		for n := 0; n < size; n++ {
			STD += math.Pow(q.number[n]-mean, 2)
		}
		STD = math.Sqrt(STD / float64(size))
	}
	frequencyPool.Put(q)
	return mean, STD
}
