package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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

	apiUrl, _ := url.Parse("https://wallhaven.cc/api/v1/search")
	params := url.Values{}

	if query != "" {
		params.Add("q", query)
	}
	if color != "" {
		params.Add("colors", color)
	}
	params.Add("sorting", "random")

	apiUrl.RawQuery = params.Encode()

	fmt.Printf("🔍 Searching Wallhaven: %s\n", apiUrl.String())

	resp, err := http.Get(apiUrl.String())
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

	fmt.Printf("\n--- NEW WALLPAPER REQUEST ---\n")
	fmt.Printf("1. Target ID: %s\n", id)
	fmt.Printf("2. Downloading from: %s\n", imgUrl)

	req, _ := http.NewRequest("GET", imgUrl, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Network Error: %v\n", err)
		http.Error(w, "Failed to connect to Wallhaven", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Wallhaven rejected the download! HTTP Status: %d\n", resp.StatusCode)
		http.Error(w, "Wallhaven rejected the request", http.StatusInternalServerError)
		return
	}

	ext := filepath.Ext(imgUrl)
	if ext == "" {
		ext = ".jpg"
	}

	filePath := filepath.Join(cacheDir, id+ext)
	out, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("❌ Failed to create local file: %v\n", err)
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}

	io.Copy(out, resp.Body)
	out.Close()

	fmt.Printf("3. Saved successfully to: %s\n", filePath)

	scriptPath := "/home/agnikas/.config/quickshell/ii/scripts/colors/switchwall.sh"
	cmd := exec.Command(scriptPath, filePath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Printf("❌ awww failed! Exit Status: %v\n", err)
		fmt.Printf("❌ awww Output: %s\n", string(output))
		http.Error(w, "Failed to set wallpaper", http.StatusInternalServerError)
		return
	}

	fmt.Printf("✅ Wallpaper successfully changed!\n-----------------------------\n")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Wallpaper updated successfully"))
}
