package function

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	mullinsZip "github.com/alexmullins/zip" // 外部ライブラリをインポート
)

// HTTPエントリポイント
func CreateZip(w http.ResponseWriter, r *http.Request) {
	password := r.URL.Query().Get("password")
	zipFilename := r.URL.Query().Get("zip_filename")
	if zipFilename == "" {
		zipFilename = "protected.zip"
	}
	if password == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	tempDir, err := os.MkdirTemp("", "upload")
	if err != nil {
		http.Error(w, "Failed to create temp directory", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	// ZIP ファイルの作成
	zipPath := filepath.Join(tempDir, zipFilename)
	zipFile, err := os.Create(zipPath)
	if err != nil {
		http.Error(w, "Failed to create zip file", http.StatusInternalServerError)
		return
	}
	defer zipFile.Close()

	zipWriter := mullinsZip.NewWriter(zipFile) // パスワード対応ライブラリを使用
	defer func() {
		if err := zipWriter.Close(); err != nil {
			http.Error(w, "Failed to close zip writer", http.StatusInternalServerError)
		}
	}()

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Failed to open file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// ZIP 内のファイルに書き込み
		zipEntry, err := zipWriter.Encrypt(fileHeader.Filename, password)
		if err != nil {
			http.Error(w, "Failed to create zip entry", http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(zipEntry, file); err != nil {
			http.Error(w, "Failed to write file to zip", http.StatusInternalServerError)
			return
		}
	}

	// zipWriter.Close() の完了を待ち、ZIP ファイルを読み込み、HTTP レスポンスに返す
	if err := zipWriter.Close(); err != nil {
		http.Error(w, "Failed to finalize zip file", http.StatusInternalServerError)
		return
	}

	zipData, err := os.ReadFile(zipPath)
	if err != nil {
		http.Error(w, "Failed to read zip file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipFilename))
	w.Write(zipData)
}
