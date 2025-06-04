package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/transparency-dev/tessera"
	"github.com/transparency-dev/tessera/storage/gcp"
	"golang.org/x/mod/sumdb/note"
)

var (
	bucket  = flag.String("bucket", "", "Bucket for log")
	spanner = flag.String("spanner", "", "Spanner resource URI")
)

func main() {
	ctx := context.Background()
	flag.Parse()

	s := getSignerOrDie()

	// driver, _ := posix.New(ctx, "/tmp/demolog")
	gcpCfg := gcp.Config{
		Bucket:  *bucket,
		Spanner: *spanner,
	}
	driver, err := gcp.New(ctx, gcpCfg)
	if err != nil {
		panic(err)
	}
	appender, shutdown, r, err := tessera.NewAppender(
		ctx,
		driver,
		tessera.NewAppendOptions().
			WithCheckpointSigner(s).
			WithCheckpointInterval(2*time.Second).
			WithBatching(1, time.Second))
	if err != nil {
		panic(err)
	}

	await := tessera.NewPublicationAwaiter(ctx, r.ReadCheckpoint, time.Second)

	f := []tessera.IndexFuture{}
	for i := range 100 {
		e := tessera.NewEntry(fmt.Appendf(nil, "Thing %d", i))
		f = append(f, appender.Add(ctx, e))
	}

	for _, f := range f {
		await.Await(ctx, f)
	}

	shutdown(ctx)
	fmt.Println("Added 100 things")

}

func getSignerOrDie() note.Signer {
	d, _ := os.ReadFile("demo.sec")
	s, _ := note.NewSigner(string(d))
	return s
}
