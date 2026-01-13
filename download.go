package wasmtime

import (
	"archive/tar"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ulikunitz/xz"
)

const (
	wasmtimeVersion = "v40.0.0"
	wasmtimeBaseURL = "https://github.com/bytecodealliance/wasmtime/releases/download"
)

// getLibraryPath returns the path to the wasmtime library, downloading it if necessary.
func getLibraryPath() (string, error) {
	// Check for custom path in environment variable
	if customPath := os.Getenv("WASMTIME_LIB_PATH"); customPath != "" {
		return customPath, nil
	}

	// Get cache directory
	cacheDir, err := getCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to get cache directory: %w", err)
	}

	// Check if library already exists in cache
	libPath := filepath.Join(cacheDir, getLibraryFilename())
	if _, err := os.Stat(libPath); err == nil {
		return libPath, nil
	}

	// Download and extract the library
	if err := downloadAndExtractLibrary(cacheDir); err != nil {
		return "", fmt.Errorf("failed to download wasmtime library: %w", err)
	}

	return libPath, nil
}

// getCacheDir returns the cache directory for wasmtime libraries.
func getCacheDir() (string, error) {
	var baseDir string

	switch runtime.GOOS {
	case "darwin", "linux":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		baseDir = filepath.Join(home, ".local", "share", "purego-wasmtime", wasmtimeVersion)
	case "windows":
		baseDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "purego-wasmtime", wasmtimeVersion)
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", err
	}

	return baseDir, nil
}

// getLibraryFilename returns the platform-specific library filename.
func getLibraryFilename() string {
	switch runtime.GOOS {
	case "darwin":
		return "libwasmtime.dylib"
	case "linux":
		return "libwasmtime.so"
	case "windows":
		return "wasmtime.dll"
	default:
		return ""
	}
}

// getDownloadURL returns the download URL for the current platform.
func getDownloadURL() (string, error) {
	var platform string

	goos := runtime.GOOS
	goarch := runtime.GOARCH

	switch goos {
	case "darwin":
		switch goarch {
		case "amd64":
			platform = "x86_64-macos"
		case "arm64":
			platform = "aarch64-macos"
		default:
			return "", fmt.Errorf("unsupported macOS architecture: %s", goarch)
		}
	case "linux":
		switch goarch {
		case "amd64":
			platform = "x86_64-linux"
		case "arm64":
			platform = "aarch64-linux"
		default:
			return "", fmt.Errorf("unsupported Linux architecture: %s", goarch)
		}
	case "windows":
		switch goarch {
		case "amd64":
			platform = "x86_64-windows"
		case "arm64":
			platform = "aarch64-windows"
		default:
			return "", fmt.Errorf("unsupported Windows architecture: %s", goarch)
		}
	default:
		return "", fmt.Errorf("unsupported operating system: %s", goos)
	}

	filename := fmt.Sprintf("wasmtime-%s-%s-c-api", wasmtimeVersion, platform)
	if goos == "windows" {
		filename += ".zip"
	} else {
		filename += ".tar.xz"
	}

	return fmt.Sprintf("%s/%s/%s", wasmtimeBaseURL, wasmtimeVersion, filename), nil
}

// downloadAndExtractLibrary downloads and extracts the wasmtime library.
func downloadAndExtractLibrary(cacheDir string) error {
	url, err := getDownloadURL()
	if err != nil {
		return err
	}

	fmt.Printf("Downloading wasmtime %s for %s/%s...\n", wasmtimeVersion, runtime.GOOS, runtime.GOARCH)

	// Download the archive
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Create a temporary file for the download
	tmpFile, err := os.CreateTemp("", "wasmtime-*.tar.xz")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Copy the download to the temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("failed to save download: %w", err)
	}

	// Seek back to the beginning
	if _, err := tmpFile.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	fmt.Println("Extracting...")

	// Extract based on file type
	if runtime.GOOS == "windows" {
		return fmt.Errorf("Windows extraction not yet implemented")
	} else {
		return extractTarXz(tmpFile, cacheDir)
	}
}

// extractTarXz extracts a .tar.xz archive.
func extractTarXz(r io.Reader, destDir string) error {
	// Decompress XZ
	xzReader, err := xz.NewReader(r)
	if err != nil {
		return fmt.Errorf("failed to create xz reader: %w", err)
	}

	// Extract tar
	tarReader := tar.NewReader(xzReader)

	libFilename := getLibraryFilename()

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		// We only need to extract the library file
		if filepath.Base(header.Name) == libFilename {
			destPath := filepath.Join(destDir, libFilename)

			outFile, err := os.Create(destPath)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("failed to extract file: %w", err)
			}

			// Set executable permissions
			if err := os.Chmod(destPath, 0755); err != nil {
				return fmt.Errorf("failed to set permissions: %w", err)
			}

			fmt.Printf("Extracted %s\n", libFilename)
			return nil
		}
	}

	return fmt.Errorf("library file %s not found in archive", libFilename)
}
