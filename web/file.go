package web

import (
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type FileUploader struct {
	// FileField 对应于文件在表单中的字段名字
	FileField string
	// DstPathFunc 用于计算目标路径
	DstPathFunc func(fh *multipart.FileHeader) string
}

func (f *FileUploader) Handle() HandleFunc {
	// 这里可以额外做一些检测
	// if f.FileField == "" {
	// 	// 这种方案默认值我其实不是很喜欢
	// 	// 因为我们需要教会用户说，这个 file 是指什么意思
	// 	f.FileField = "file"
	// }
	return func(ctx *Context) {
		src, srcHeader, err := ctx.Req.FormFile(f.FileField)
		if err != nil {
			ctx.RespStatusCode = 400
			ctx.RespData = []byte("上传失败，未找到数据")
			log.Fatalln(err)
			return
		}
		defer src.Close()
		dst, err := os.OpenFile(f.DstPathFunc(srcHeader),
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
		if err != nil {
			ctx.RespStatusCode = 500
			ctx.RespData = []byte("上传失败")
			log.Fatalln(err)
			return
		}
		defer dst.Close()

		_, err = io.CopyBuffer(dst, src, nil)
		if err != nil {
			ctx.RespStatusCode = 500
			ctx.RespData = []byte("上传失败")
			log.Fatalln(err)
			return
		}
		ctx.RespData = []byte("上传成功")
	}
}

// HandleFunc 这种设计方案也是可以的，但是不如上一种灵活。
// 它可以直接用来注册路由
// 上一种可以在返回 HandleFunc 之前可以继续检测一下传入的字段
// 这种形态和 Option 模式配合就很好
func (f *FileUploader) HandleFunc(ctx *Context) {
	src, srcHeader, err := ctx.Req.FormFile(f.FileField)
	if err != nil {
		ctx.RespStatusCode = 400
		ctx.RespData = []byte("上传失败，未找到数据")
		log.Fatalln(err)
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(f.DstPathFunc(srcHeader),
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		ctx.RespStatusCode = 500
		ctx.RespData = []byte("上传失败")
		log.Fatalln(err)
		return
	}
	defer dst.Close()

	_, err = io.CopyBuffer(dst, src, nil)
	if err != nil {
		ctx.RespStatusCode = 500
		ctx.RespData = []byte("上传失败")
		log.Fatalln(err)
		return
	}
	ctx.RespData = []byte("上传成功")
}

// FileDownloader 直接操作了 http.ResponseWriter
// 所以在 Middleware 里面将不能使用 RespData
// 因为没有赋值
type FileDownloader struct {
	Dir string
}

func (f *FileDownloader) Handle() HandleFunc {
	return func(ctx *Context) {
		req, _ := ctx.QueryValue("file")
		path := filepath.Join(f.Dir, filepath.Clean(req))
		fn := filepath.Base(path)
		header := ctx.Resp.Header()
		header.Set("Content-Disposition", "attachment;filename="+fn)
		header.Set("Content-Description", "File Transfer")
		header.Set("Content-Type", "application/octet-stream")
		header.Set("Content-Transfer-Encoding", "binary")
		header.Set("Expires", "0")
		header.Set("Cache-Control", "must-revalidate")
		header.Set("Pragma", "public")
		http.ServeFile(ctx.Resp, ctx.Req, path)
	}
}

type fileCacheItem struct {
	fileName    string
	fileSize    int
	contentType string
	data        []byte
}

type StaticResourceHandler struct {
	dir                     string
	extensionContentTypeMap map[string]string
	cache                   *lru.Cache
	maxFileSize             int
}

func NewStaticResourceHandler(dir string) (*StaticResourceHandler, error) {
	cache, err := lru.New(1000)
	if err != nil {
		return nil, err
	}

	return &StaticResourceHandler{
		dir:         dir,
		maxFileSize: 10 * 1024 * 1024,
		cache:       cache,
		extensionContentTypeMap: map[string]string{
			// 这里根据自己的需要不断添加
			"jpeg": "image/jpeg",
			"jpe":  "image/jpeg",
			"jpg":  "image/jpeg",
			"png":  "image/png",
			"pdf":  "image/pdf",
		},
	}, nil
}

func (h *StaticResourceHandler) Handle(ctx *Context) {
	req, _ := ctx.PathValue("filename")
	if item, ok := h.readFileFromData(req); ok {
		log.Printf("从缓存中读取数据...")
		h.writeItemAsResponse(item, ctx.Resp)
		return
	}
	path := filepath.Join(h.dir, req)
	f, err := os.Open(path)
	if err != nil {
		ctx.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	ext := filepath.Ext(path)[1:]
	t, ok := h.extensionContentTypeMap[ext]
	if !ok {
		ctx.Resp.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		ctx.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	item := &fileCacheItem{
		fileSize:    len(data),
		data:        data,
		contentType: t,
		fileName:    req,
	}

	h.cacheFile(item)
	h.writeItemAsResponse(item, ctx.Resp)
}

func (h *StaticResourceHandler) cacheFile(item *fileCacheItem) {
	if h.cache != nil && item.fileSize < h.maxFileSize {
		h.cache.Add(item.fileName, item)
	}
}

func (h *StaticResourceHandler) writeItemAsResponse(item *fileCacheItem, writer http.ResponseWriter) {
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", item.contentType)
	writer.Header().Set("Content-Length", fmt.Sprintf("%d", item.fileSize))
	_, _ = writer.Write(item.data)

}

func (h *StaticResourceHandler) readFileFromData(fileName string) (*fileCacheItem, bool) {
	if h.cache != nil {
		if item, ok := h.cache.Get(fileName); ok {
			return item.(*fileCacheItem), true
		}
	}
	return nil, false
}
