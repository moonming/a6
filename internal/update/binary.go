package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func Download(asset Asset, progress io.Writer) (string, error) {
	if strings.TrimSpace(asset.BrowserDownloadURL) == "" {
		return "", fmt.Errorf("asset download URL is empty")
	}

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest(http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("build download request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download release asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp("", "a6-update-download-*")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	var dst io.Writer = tmp
	if progress != nil {
		dst = io.MultiWriter(tmp, progress)
	}

	if _, err := io.Copy(dst, resp.Body); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("write downloaded asset: %w", err)
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("close downloaded asset: %w", err)
	}

	return tmpPath, nil
}

func Install(downloadedPath string) error {
	if strings.TrimSpace(downloadedPath) == "" {
		return fmt.Errorf("download path is empty")
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve current executable: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("resolve executable symlink: %w", err)
	}

	oldInfo, err := os.Stat(execPath)
	if err != nil {
		return fmt.Errorf("stat current executable: %w", err)
	}

	dir := filepath.Dir(execPath)
	tmp, err := os.CreateTemp(dir, ".a6-update-*")
	if err != nil {
		return fmt.Errorf("create replacement file: %w", err)
	}
	tmpPath := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	if err := writeInstalledBinary(tmp, downloadedPath); err != nil {
		cleanup()
		return err
	}

	if err := tmp.Chmod(oldInfo.Mode().Perm()); err != nil {
		cleanup()
		return fmt.Errorf("set executable permissions: %w", err)
	}

	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close replacement file: %w", err)
	}

	if err := os.Rename(tmpPath, execPath); err != nil {
		_ = os.Remove(tmpPath)
		if errors.Is(err, os.ErrPermission) {
			return fmt.Errorf("permission denied replacing %s; try running with sudo or install manually: %w", execPath, err)
		}
		return fmt.Errorf("replace executable: %w", err)
	}

	return nil
}

func writeInstalledBinary(dst *os.File, downloadedPath string) error {
	format, err := detectArchive(downloadedPath)
	if err != nil {
		return err
	}

	switch format {
	case "tar.gz":
		if err := extractFromTarGz(downloadedPath, dst); err != nil {
			return fmt.Errorf("extract tar.gz asset: %w", err)
		}
	case "zip":
		if err := extractFromZip(downloadedPath, dst); err != nil {
			return fmt.Errorf("extract zip asset: %w", err)
		}
	default:
		f, err := os.Open(downloadedPath)
		if err != nil {
			return fmt.Errorf("open downloaded file: %w", err)
		}
		defer f.Close()

		if _, err := io.Copy(dst, f); err != nil {
			return fmt.Errorf("copy downloaded binary: %w", err)
		}
	}

	if runtime.GOOS != "windows" {
		if err := dst.Chmod(0o755); err != nil {
			return fmt.Errorf("set executable bit: %w", err)
		}
	}

	return nil
}

func detectArchive(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open downloaded file: %w", err)
	}
	defer f.Close()

	head := make([]byte, 4)
	n, err := io.ReadFull(f, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", fmt.Errorf("read downloaded file header: %w", err)
	}
	head = head[:n]

	if len(head) >= 2 && head[0] == 0x1f && head[1] == 0x8b {
		return "tar.gz", nil
	}
	if len(head) >= 4 && head[0] == 'P' && head[1] == 'K' && head[2] == 0x03 && head[3] == 0x04 {
		return "zip", nil
	}

	return "binary", nil
}

func extractFromTarGz(path string, dst io.Writer) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		name := filepath.Base(hdr.Name)
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if name == "a6" || name == "a6.exe" {
			if _, err := io.Copy(dst, tr); err != nil {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("a6 binary not found in archive")
}

func extractFromZip(path string, dst io.Writer) error {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer zr.Close()

	for _, f := range zr.File {
		name := filepath.Base(f.Name)
		if name != "a6" && name != "a6.exe" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(dst, rc)
		closeErr := rc.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		return nil
	}

	return fmt.Errorf("a6 binary not found in archive")
}
