package main

import (
	"context"
	"flag"
	"log"

	"cloud.google.com/go/storage"
	http_cache "github.com/sturtevant/simple-server/proxy"
	"google.golang.org/api/iterator"
)

func main() {
	var address string
	flag.StringVar(&address, "address", "127.0.0.1", "Address to listen on")
	var port int64
	flag.Int64Var(&port, "port", 8080, "Port to serve")
	var bucketName string
	flag.StringVar(&bucketName, "bucket", "", "Google Storage Bucket Name")
	var defaultPrefix string
	flag.StringVar(&defaultPrefix, "prefix", "", "Optional prefix for all objects. For example, use --prefix=foo/ to work under foo directory in a bucket.")
	var indexName string
	flag.StringVar(&indexName, "index", "", "Optional name of the index file to serve when there the root is requested. For example, use --index=index.html to serve index.html when / is requested.")
	var missingName string
	flag.StringVar(&missingName, "missing", "", "Optional name of the file to serve when there the requested file cannot be found. For example, use --missing=404.html to serve 404.html anytime a file cannot be found.")
	var suppress404 bool
	flag.BoolVar(&suppress404, "suppress404", false, "Option to suppress 404 response code when serving the missing / error file.")
	flag.Parse()

	if bucketName == "" {
		log.Fatal("Please specify Google Cloud Storage Bucket")
	}

	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create a storage client: %s", err)
	}

	bucketHandler := client.Bucket(bucketName)
	storageProxy := http_cache.NewStorageProxy(bucketHandler, defaultPrefix, indexName, missingName, suppress404)

	log.Printf("bucket: %s\n", bucketName)
	bucket := client.Bucket(bucketName)
	query := &storage.Query{}
	it := bucket.Objects(context.Background(), query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		log.Println(attrs.Name)
	}

	err = storageProxy.Serve(address, port)
	if err != nil {
		log.Fatalf("Failed to start proxy: %s", err)
	}
}
