package corehttp

import (
	"context"
	"fmt"
	"github.com/wumansgy/goEncrypt"
	"html"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	gopath "path"
	"strconv"
	"strings"
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
		args:=r.URL.Query()
		password:=args.Get("code")
		if password!="" {
			old, err := ioutil.ReadAll(f)
			if err != nil {
				internalWebError(w, err)
				return
			}
			defer f.Close()
			cryptText, err := goEncrypt.DesCbcDecrypt(old, []byte(password),[]byte("wumansgy")) //解密得到密文,可以自己传入初始化向量,如果不传就使用默认的初始化向量,8字节
			if err != nil {
				internalWebError(w, err)
				return
			}
			respFiles:= files.NewBytesFile([]byte(cryptText))
			size, err := respFiles.Size()
			defer respFiles.Close()

			// Set Content-Disposition
			name := addContentDispositionHeader(w, r, contentPath)
			var ctype string
			if _, isSymlink := respFiles.(*files.Symlink); isSymlink {
				// We should be smarter about resolving symlinks but this is the
				// "most correct" we can be without doing that.
				ctype = "inode/symlink"
			} else {
				ctype = mime.TypeByExtension(gopath.Ext(name))
				if ctype == "" {

				}
				// Strip the encoding from the HTML Content-Type header and let the
				// browser figure it out.
				//
				// Fixes https://github.com/ipfs/go-ipfs/issues/2203
				if strings.HasPrefix(ctype, "text/html;") {
					ctype = "text/html"
				}
			}
			// Setting explicit Content-Type to avoid mime-type sniffing on the client
			// (unifies behavior across gateways and web browsers)
			w.Header().Set("Accept-Ranges","none")
			w.Header().Set("Content-Type", ctype)
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
			xcontent := &lazySeeker{
				size:   size,
				reader: respFiles,
			}
			if _,err:=xcontent.Seek(0, io.SeekStart);err!=nil{
				internalWebError(w, err)
				return
			}
			if _, err = io.Copy(w, xcontent);err!=nil{
				internalWebError(w, err)
				return
			}
			return

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
