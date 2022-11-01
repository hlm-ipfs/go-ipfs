package corehttp

import (
	"encoding/json"
	"fmt"
	"github.com/ipfs/kubo/core"

	"github.com/tus/tusd/pkg/filestore"
	tusd "github.com/tus/tusd/pkg/handler"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

func HelloWorld(path string) ServeOption {
	return func(_ *core.IpfsNode, _ net.Listener, mux *http.ServeMux) (*http.ServeMux, error) {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			//if r.Method != http.MethodPost {
			//	http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
			//	return
			//}
			//if err := r.ParseForm(); err != nil {
			//	http.Error(w, err.Error(), http.StatusBadRequest)
			//	return
			//}
			//
			//rateStr := r.Form.Get("rate")
			//if len(rateStr) == 0 {
			//	http.Error(w, "parameter 'rate' must be set", http.StatusBadRequest)
			//	return
			//}
			//
			//rate, err := strconv.Atoi(rateStr)
			//if err != nil {
			//	http.Error(w, err.Error(), http.StatusBadRequest)
			//	return
			//}
			log.Infof("hello world")

			var data struct {
				Name string
			}
			data.Name = "hello world"
			ret_json, _ := json.Marshal(data)
			io.WriteString(w, string(ret_json))
			//runtime.SetBlockProfileRate(rate)
		})
		return mux, nil
	}
}

func TusFiles(path string) ServeOption {
	return func(node *core.IpfsNode, _ net.Listener, mux *http.ServeMux) (*http.ServeMux, error) {

		storePath := "./uploads"
		if ex, err := os.Executable(); err == nil {
			storePath = filepath.Dir(ex) + "/uploads"
		}
		fmt.Println("===storePath=====", storePath)
		b, err := PathExists(storePath)
		if !b {
			os.Mkdir(storePath, 0777)
			// 再修改权限
			os.Chmod(storePath, 0777)
		}

		store := filestore.FileStore{
			Path: storePath,
		}

		composer := tusd.NewStoreComposer()
		store.UseIn(composer)
		handler, err := tusd.NewHandler(tusd.Config{
			BasePath:              path,
			StoreComposer:         composer,
			NotifyCompleteUploads: true,
		})
		if err != nil {
			fmt.Println("========", err)
		}
		go func() {
			for {
				event := <-handler.CompleteUploads
				fmt.Printf("Upload %s finished\n", event.Upload.ID)
			}
		}()

		//http.Handle(path,handler)
		mux.Handle(path, http.StripPrefix(path, handler))
		return mux, nil
	}
}

// 判断所给路径文件/文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	//isnotexist来判断，是不是不存在的错误
	if os.IsNotExist(err) { //如果返回的错误类型使用os.isNotExist()判断为true，说明文件或者文件夹不存在
		return false, nil
	}
	return false, err //如果有错误了，但是不是不存在的错误，所以把这个错误原封不动的返回
}
