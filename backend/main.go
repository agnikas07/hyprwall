package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

const cacheDir = "/tmp/wallpaper_cache"

func main() {
	os.MkdirAll(cacheDir, os.ModePerm)

	http.HandleFunc("/api/search", searchHandler)
	http.HandleFunc("/api/set", setWallpaperHandler)

	fmt.Println("Backend running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	color := r.URL.Query().Get("colors")

	apiUrl := fmt.Sprintf("https://wallhaven.cc/api/v1/search?q=%s&colors=%s&sorting=random", query, color)

	resp, err := http.Get(apiUrl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}

func setWallpaperHandler(w http.ResponseWriter, r *http.Request) {
	imgUrl := r.URL.Query().Get("url")
	id := r.URL.Query().Get("id")

	if imgUrl == "" || id == "" {
		http.Error(w, "Missing url or id parameter", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(cacheDir, id+".jpg")
	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	resp, err := http.Get(imgUrl)
	if err != nil {
		http.Error(w, "Failed to download image", http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()
	io.Copy(out, resp.Body)

	cmd := exec.Command("awww", "img", filePath, "--transition-type", "wipe", "--transition-angle", "30", "--transition-step", "90")
	err = cmd.Run()

	if err != nil {
		fmt.Println("Error setting wallpaper:", err)
		http.Error(w, "Failed to set wallpaper", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Wallpaper updated successfully"))
}
