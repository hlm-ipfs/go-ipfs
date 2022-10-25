package corehttp

import (
	"context"
	"fmt"
	"github.com/ipfs/kubo/auth"
	"github.com/wumansgy/goEncrypt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	files "github.com/ipfs/go-ipfs-files"
	ipath "github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/ipfs/kubo/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func (i *gatewayHandler) serveUnixFS(ctx context.Context, w http.ResponseWriter, r *http.Request, resolvedPath ipath.Resolved, contentPath ipath.Path, begin time.Time, logger *zap.SugaredLogger) {
	ctx, span := tracing.Span(ctx, "Gateway", "ServeUnixFS", trace.WithAttributes(attribute.String("path", resolvedPath.String())))
	defer span.End()

	// Handling UnixFS
	dr, err := i.api.Unixfs().Get(ctx, resolvedPath)
	if err != nil {
		webError(w, "ipfs cat "+html.EscapeString(contentPath.String()), err, http.StatusNotFound)
		return
	}
	defer dr.Close()
	// Handling Unixfs file
	if f, ok := dr.(files.File); ok {
		r.Body.Close()
		//远程查看密码:
		args := r.URL.Query()
		password := args.Get("code")
		if password != "" {
			filename := filepath.Join(i.cache, resolvedPath.Cid().String())
			if _, err := os.Stat(filename); err == nil {
				if cacheFile, err := os.Open(filename); cacheFile != nil && err == nil {
					defer cacheFile.Close()
					i.serveCacheFile(ctx, w, r, resolvedPath, contentPath, cacheFile, begin)
					return
				}
			}
			decrypt := func() error {
				inFilePath := filepath.Join(i.cache, resolvedPath.Cid().String()+".enc")
				outFilePath := filepath.Join(i.cache, resolvedPath.Cid().String())

				inFile, err := os.OpenFile(inFilePath, os.O_RDWR|os.O_CREATE, 0644)
				if err != nil {
					return err
				}
				defer func() {
					inFile.Close()
					os.RemoveAll(inFilePath)
				}()
				_, err = io.Copy(inFile, f)
				if err != nil {
					return err
				}
				err = auth.DecryptBigFile(inFilePath, outFilePath, []byte(password))
				if err != nil {
					return err
				}
				return nil
			}
			if len(password) == 6 {
				if err := decrypt(); err != nil {
					internalWebError(w, err)
					return
				}
			} else {
				old, err := ioutil.ReadAll(f)
				if err != nil {
					internalWebError(w, err)
					return
				}
				defer f.Close()
				cryptText, err := goEncrypt.DesCbcDecrypt(old, []byte(password), []byte("wumansgy")) //解密得到密文,可以自己传入初始化向量,如果不传就使用默认的初始化向量,8字节
				if err != nil {
					internalWebError(w, err)
					return
				}
				dir, err := ioutil.ReadDir(i.cache)
				for _, d := range dir {
					os.RemoveAll(path.Join([]string{i.cache, d.Name()}...))
				}
				if err := ioutil.WriteFile(filename, cryptText, os.ModePerm); err != nil {
					internalWebError(w, err)
					return
				}
			}
			if cacheFile, err := os.Open(filename); cacheFile != nil && err == nil {
				defer cacheFile.Close()
				i.serveCacheFile(ctx, w, r, resolvedPath, contentPath, cacheFile, begin)
				return
			} else {
				internalWebError(w, err)
				return
			}
		}
		logger.Debugw("serving unixfs file", "path", contentPath)
		i.serveFile(ctx, w, r, resolvedPath, contentPath, f, begin)
		return
	}

	// Handling Unixfs directory
	dir, ok := dr.(files.Directory)
	if !ok {
		internalWebError(w, fmt.Errorf("unsupported UnixFS type"))
		return
	}

	logger.Debugw("serving unixfs directory", "path", contentPath)
	i.serveDirectory(ctx, w, r, resolvedPath, contentPath, dir, begin, logger)
}
