package corehttp

import (
	"context"
	"fmt"
	"github.com/wumansgy/goEncrypt"
	"html"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/go-ipfs/tracing"
	ipath "github.com/ipfs/interface-go-ipfs-core/path"
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
		//远程查看密码:
		args:=r.URL.Query()
		password:=args.Get("code")
		if password!="" {
			old, err := ioutil.ReadAll(f)
			if err != nil {
				internalWebError(w, err)
				return
			}
			cryptText, err := goEncrypt.DesCbcDecrypt(old, []byte(password),[]byte("wumansgy")) //解密得到密文,可以自己传入初始化向量,如果不传就使用默认的初始化向量,8字节
			if err != nil {
				internalWebError(w, err)
				return
			}
			f= files.NewBytesFile([]byte(cryptText))
			size, err := f.Size()
			w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
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
