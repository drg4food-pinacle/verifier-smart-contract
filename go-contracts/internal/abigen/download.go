package abigen

import (
	"archive/tar"
	"compress/gzip"
	"crypto/md5"
	"errors"
	"fmt"
	"go-contracts/internal/logger"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

var (
	emptyStr = ""
)

// Download the abigen binary for the given version.
func checkAbigen(version Version) (string, error) {
	v, ok := abigenVersions[version]
	if !ok {
		return emptyStr, fmt.Errorf("abigen: checksum missing for version %q", version)
	}

	if err := createBinDir(); err != nil {
		return emptyStr, err
	}

	absAbigenPath := filepath.Join(Root, abigenDir)
	abigenArchive := filepath.Join(absAbigenPath, v.Path)
	abigenExtractedBinary := filepath.Join(absAbigenPath, "abigen")

	_, err, _ := dg.Do("abigen_"+version.String(), func() (any, error) {
		if _, err := os.Stat(abigenArchive); errors.Is(err, os.ErrNotExist) {
			var (
				try int
				err error
			)
			for ; try < maxTries && (try == 0 || err != nil); try++ {
				err = downloadAbigen(abigenArchive, v)
			}
			if try >= maxTries {
				return emptyStr, fmt.Errorf("abigen: failed to download %q: %w", version, err)
			}
		}

		// Open the archive for checksum verification
		f, err := os.Open(abigenArchive)
		if err != nil {
			return emptyStr, err
		}
		defer f.Close()

		// Verify checksum
		if err := verifyChecksum(f, v); err != nil {
			return emptyStr, err
		}

		// Extract abigen binary
		if _, err := os.Stat(abigenExtractedBinary); errors.Is(err, os.ErrNotExist) {
			outpath, err := extractAbigen(abigenArchive, absAbigenPath)
			if err != nil {
				return emptyStr, fmt.Errorf("failed to extract abigen: %w", err)
			}
			logger.Logger.Info().Msgf("Abigen extracted to: %s", outpath)
		}

		return nil, nil
	})

	if err != nil {
		return emptyStr, err
	}

	return abigenExtractedBinary, nil
}

// Fetch the tar.gz from the source
func downloadAbigen(path string, v abigenVersion) error {
	resp, err := http.Get(abigenBaseURL + v.Path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0o0764)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}
	return nil
}

// extracts the downloaded tar.gz from geth-tools and takes the abigen binary
func extractAbigen(tarGzPath, destDir string) (string, error) {
	// Open the tar.gz file
	f, err := os.Open(tarGzPath)
	if err != nil {
		return emptyStr, fmt.Errorf("failed to open archive: %w", err)
	}
	defer f.Close()

	// Create a gzip reader
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return emptyStr, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Create a tar reader
	tr := tar.NewReader(gzr)

	// Iterate through the files in the tar archive
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return emptyStr, fmt.Errorf("error reading tar: %w", err)
		}

		// Look for the 'abigen' binary
		if filepath.Base(header.Name) == "abigen" {
			outPath := filepath.Join(destDir, "abigen")
			outFile, err := os.Create(outPath)
			if err != nil {
				return emptyStr, fmt.Errorf("failed to create output file: %w", err)
			}
			defer outFile.Close()

			// Copy the file contents
			if _, err := io.Copy(outFile, tr); err != nil {
				return emptyStr, fmt.Errorf("failed to copy abigen: %w", err)
			}

			// Make it executable
			if err := os.Chmod(outPath, 0o755); err != nil {
				return emptyStr, fmt.Errorf("failed to chmod abigen: %w", err)
			}

			return outPath, nil // Done
		}
	}

	return emptyStr, fmt.Errorf("abigen binary not found in archive")
}

// Verify the checksum of the downloaded abigen binary.
func verifyChecksum(r io.Reader, v abigenVersion) error {
	// Calculate SHA256 checksum of the file
	hash := md5.New()
	if _, err := io.Copy(hash, r); err != nil {
		return err
	}

	var gotMD5 [16]byte
	hash.Sum(gotMD5[:0])

	// Compare the calculated checksum with the expected checksum
	if v.MD5 != gotMD5 {
		return errors.New("abigen: checksum mismatch for downloaded abigen binary")
	}
	return nil
}

// Create the directory if it does not exist
func createBinDir() error {
	if _, err := os.Stat(abigenDir); os.IsNotExist(err) {
		if err := os.MkdirAll(abigenDir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory .abigen/bin: %w", err)
		}
	}
	return nil
}
