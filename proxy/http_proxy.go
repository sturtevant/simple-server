package http_cache

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"cloud.google.com/go/storage"
)

type StorageProxy struct {
	bucketHandler *storage.BucketHandle
	defaultPrefix string
	indexName     string
	missingName   string
	suppress404   bool
}

func NewStorageProxy(bucketHandler *storage.BucketHandle, defaultPrefix string, indexName string, missingName string, suppress404 bool) *StorageProxy {
	return &StorageProxy{
		bucketHandler: bucketHandler,
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
	http.HandleFunc("/", proxy.handler)

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, port))

	if err == nil {
		address := listener.Addr().String()
		listener.Close()
		log.Printf("Starting http cache server %s\n", address)
		return http.ListenAndServe(address, nil)
	}
	return err
}

func (proxy StorageProxy) handler(w http.ResponseWriter, r *http.Request) {
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
	object := proxy.bucketHandler.Object(proxy.objectName(name))
	if object == nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	reader, err := object.NewReader(context.Background())
	if err != nil {
		if err.Error() == "storage: object doesn't exist" && proxy.missingName != "" {
			object = proxy.bucketHandler.Object(proxy.objectName(proxy.missingName))
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
	object := proxy.bucketHandler.Object(proxy.objectName(name))
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
