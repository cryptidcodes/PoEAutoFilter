package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/inconshreveable/go-update"
)

// AppVersion is the current version of the application.
// It is intended to be overridden at build time using -ldflags.
// Example: -ldflags="-X 'github.com/cryptidcodes/PoEAutoFilter/client/internal/core.AppVersion=v1.1.0'"
var AppVersion = "v0.0.0"

// VersionInfo matches the JSON format returned by the proxy server's /version endpoint.
type VersionInfo struct {
	Version      string `json:"version"`
	LinuxURL     string `json:"linux_url"`
	WindowsURL   string `json:"windows_url"`
	ReleaseNotes string `json:"release_notes"`
}

// CheckUpdate requests the version metadata from the server and compares it to AppVersion.
func CheckUpdate(serverURL string) (*VersionInfo, bool, error) {
	url := fmt.Sprintf("%s/version", serverURL)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var info VersionInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, false, err
	}

	// Basic string comparison assuming semantic versioning like "v1.1.0"
	hasUpdate := info.Version != AppVersion

	return &info, hasUpdate, nil
}

// ApplyUpdate downloads the binary from the given URL and performs an in-place update.
func ApplyUpdate(downloadURL string) error {
	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad HTTP status during download: %d", resp.StatusCode)
	}

	// update.Apply safely overwrites the current executable.
	// On Windows, it handles the file lock by moving the old executable to a temporary file.
	return update.Apply(resp.Body, update.Options{})
}
