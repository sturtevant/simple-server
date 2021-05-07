package simple_server

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
)

func SimpleServer(w http.ResponseWriter, r *http.Request) {
	bucketName := os.Getenv("GCS_BUCKET")
	defaultPrefix := os.Getenv("GCS_PREFIX")
	indexName := os.Getenv("GCS_INDEX")
	missingName := os.Getenv("GCS_MISSING")
	var suppress404 bool = false
	if os.Getenv("GCS_SUPPRESS404") == "TRUE" {
		suppress404 = true
	}

	if bucketName == "" {
		log.Fatal("Please specify Google Cloud Storage Bucket")
	}

	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create a storage client: %s", err)
	}

	bucketHandler := client.Bucket(bucketName)
	storageProxy := NewStorageProxy(bucketHandler, defaultPrefix, indexName, missingName, suppress404)
	storageProxy.Handle(w, r)
}

type StorageProxy struct {
	bucket        *storage.BucketHandle
	defaultPrefix string
	indexName     string
	missingName   string
	suppress404   bool
}

func NewStorageProxy(bucket *storage.BucketHandle, defaultPrefix string, indexName string, missingName string, suppress404 bool) *StorageProxy {
	return &StorageProxy{
		bucket:        bucket,
		defaultPrefix: defaultPrefix,
		indexName:     indexName,
		missingName:   missingName,
		suppress404:   suppress404,
	}
}

func (proxy StorageProxy) objectName(name string) string {
	if name == "" && proxy.indexName != "" {
		return proxy.defaultPrefix + proxy.indexName
	} else {
		return proxy.defaultPrefix + name
	}
}

func (proxy StorageProxy) Serve(address string, port int64) error {
	http.HandleFunc("/", proxy.Handle)

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, port))

	if err == nil {
		address := listener.Addr().String()
		listener.Close()
		log.Printf("Starting GCS proxy server %s\n", address)
		return http.ListenAndServe(address, nil)
	}
	return err
}

func (proxy StorageProxy) Handle(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path
	if key[0] == '/' {
		key = key[1:]
	}
	if r.Method == "GET" {
		proxy.downloadBlob(w, key)
	} else if r.Method == "HEAD" {
		proxy.checkBlobExists(w, key)
	}
}

func (proxy StorageProxy) downloadBlob(w http.ResponseWriter, name string) {
	// log.Printf("requested: %s\n", name)
	object := proxy.bucket.Object(proxy.objectName(name))
	if object == nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	reader, err := object.NewReader(context.Background())
	if err != nil {
		if err.Error() == "storage: object doesn't exist" && proxy.missingName != "" {
			object = proxy.bucket.Object(proxy.objectName(proxy.missingName))
			reader, err = object.NewReader(context.Background())
			if !proxy.suppress404 {
				w.WriteHeader(http.StatusNotFound)
			}
		}
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	defer reader.Close()
	bufferedReader := bufio.NewReader(reader)
	_, err = bufferedReader.WriteTo(w)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	// w.WriteHeader(http.StatusOK)
}

func (proxy StorageProxy) checkBlobExists(w http.ResponseWriter, name string) {
	object := proxy.bucket.Object(proxy.objectName(name))
	if object == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// lookup attributes to see if the object exists
	attrs, err := object.Attrs(context.Background())
	if err != nil || attrs == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}
