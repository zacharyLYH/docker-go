package main

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func getAuthToken(image string) (string, error) {
	client := &http.Client{}

	url := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/%s:pull", image)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	return data.Token, nil
}

func getImageManifest(token, imageName string, tag string) ([]byte, error) {
	url := fmt.Sprintf("https://registry.hub.docker.com/v2/library/%s/manifests/%s", imageName, tag)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decodeResp, decodeErr := io.ReadAll(resp.Body)

	return decodeResp, decodeErr
}

func pullLayer(token, layerDigest string, dir string, imageName string, mediaType string) error {
	url := fmt.Sprintf("https://registry.hub.docker.com/v2/library/%s/blobs/%s", imageName, layerDigest)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", mediaType)
	fmt.Println(token)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("\n\nIn pullLayer(): Failed to pull layer: %s, Response: %s\n", resp.Status, string(body))
		return fmt.Errorf("failed to pull layer, status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	fmt.Println("Got layer data from api call: ", resp.Body)
	// Assuming layer is a tarball
	// buf := bytes.NewBuffer(body)
	return extractTarball(resp.Body, dir)
}

func extractTarball(tarReader io.Reader, dir string) error {
	fmt.Println("Starting extraction process...")
	tr := tar.NewReader(tarReader)

	// Iterate through the files in the tar archive
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		// Target location where the dir entry will be created
		target := filepath.Join(dir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory if it doesn't exist
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			// Create file and write data from tar archive to it
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()

			// Set file permissions from tar header
			if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		default:
			fmt.Printf("Unsupported type: %c in %s\n", header.Typeflag, header.Name)
		}
	}

	return nil
}
