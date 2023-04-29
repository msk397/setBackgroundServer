package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// 设置路由
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/download", downloadHandler)
	// 启动服务器
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		fmt.Println("服务器启动失败：", err)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// 解析上传的文件
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("file")
	fish := r.FormValue("fish")
	if fish != "307090" {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err != nil {
		fmt.Println("文件解析失败：", err)
		return
	}
	defer file.Close()

	// 保存文件
	filename := handler.Filename
	ext := filepath.Ext(filename)
	filename = fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("文件保存失败：", err)
		return
	}
	defer f.Close()
	io.Copy(f, file)

	fmt.Println("文件保存成功：", filename)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	// 查找最新的一张图片
	var newestFile string
	var newestTime int64

	files, err := filepath.Glob("*.png")
	if err != nil || len(files) == 0 {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	for _, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			fmt.Println("获取文件信息失败：", err)
			continue
		}
		if stat.ModTime().UnixNano() > newestTime {
			newestTime = stat.ModTime().UnixNano()
			newestFile = file
		}
	}
	if newestFile == "" {
		fmt.Println("没有找到图片")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// 发送图片
	f, err := os.Open(newestFile)
	if err != nil {
		fmt.Println("打开文件失败：", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer func() {
		f.Close()
		// 删除文件
		err = os.Remove(newestFile)
		if err != nil {
			fmt.Println("删除文件失败：", err)
		}
	}()
	contentType := mime.TypeByExtension(filepath.Ext(newestFile))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	io.Copy(w, f)

}
