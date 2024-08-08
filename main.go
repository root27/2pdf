package main

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	tempFolder   = "temp"
	uploadFolder = "uploads"

	allowedExtensions = ".doc,.docx,.txt,.xls,.xlsx"
)

func main() {
	// Ensure temp and upload directories exist
	if _, err := os.Stat(tempFolder); os.IsNotExist(err) {
		os.Mkdir(tempFolder, 0755)
	}
	if _, err := os.Stat(uploadFolder); os.IsNotExist(err) {
		os.Mkdir(uploadFolder, 0755)
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", FileConverter)

	log.Printf("Starting server on port %s\n", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func FileConverter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case "POST":
		file, headers, _ := r.FormFile("file")

		fileData, err := io.ReadAll(file)

		defer file.Close()

		fileEx := filepath.Ext(headers.Filename)

		if !strings.Contains(allowedExtensions, fileEx) {
			http.Error(w, "Invalid file type", http.StatusBadRequest)
			return
		}

		tempFile, err := os.CreateTemp(tempFolder, headers.Filename)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer tempFile.Close()

		_, err = tempFile.Write(fileData)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		outputFilePath := filepath.Join(uploadFolder, "document.pdf")

		err = ConvertToPdf(tempFile.Name(), outputFilePath)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer os.Remove(outputFilePath)

		defer os.Remove(tempFile.Name())

		outputfileData, err := os.ReadFile(outputFilePath)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/pdf")
		_, err = io.Copy(w, bytes.NewReader(outputfileData))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case "GET":
		query := r.URL.Query().Get("url")
		if query == "" {
			t, _ := template.ParseFiles("./templates/index.html")

			t.Execute(w, nil)
			return
		}

		ext := strings.Split(query, "/")

		fileName := ext[len(ext)-1]

		fileExt := filepath.Ext(fileName)

		if !strings.Contains(allowedExtensions, fileExt) {

			http.Error(w, "Invalid file type", http.StatusBadRequest)
			return
		}

		client := &http.Client{}
		response, err := client.Get(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer response.Body.Close()

		tempFile, err := os.CreateTemp(tempFolder, fileName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tempFile.Close()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = tempFile.Write(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		outputFilePath := filepath.Join(uploadFolder, "document.pdf")
		err = ConvertToPdf(tempFile.Name(), outputFilePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer os.Remove(outputFilePath)

		defer os.Remove(tempFile.Name())

		fileData, err := os.ReadFile(outputFilePath)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/pdf")
		_, err = io.Copy(w, bytes.NewReader(fileData))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func ConvertToPdf(input string, output string) error {

	outputDir := filepath.Dir(output)
	cmd := exec.Command("libreoffice", "--headless", "--convert-to", "pdf", "--outdir", filepath.Dir(output), input)
	if err := cmd.Run(); err != nil {
		return err
	}

	inputBaseName := filepath.Base(input)
	pdfFileName := inputBaseName[:len(inputBaseName)-len(filepath.Ext(inputBaseName))] + ".pdf"
	tempPdfPath := filepath.Join(outputDir, pdfFileName)

	if err := os.Rename(tempPdfPath, output); err != nil {
		return err
	}
	return nil
}
