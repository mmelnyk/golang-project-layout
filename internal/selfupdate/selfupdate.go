//go:build selfupdate
// +build selfupdate

package selfupdate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/m-sign/msign"
)

// Release collects data about a single release on GitHub.
type Release struct {
	Name        string    `json:"name"`
	TagName     string    `json:"tag_name"`
	Draft       bool      `json:"draft"`
	PreRelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []Asset   `json:"assets"`
}

// Asset is a file uploaded and attached to a release.
type Asset struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

func (r Release) String() string {
	return fmt.Sprintf("%v %v, %d assets",
		r.TagName,
		r.PublishedAt.Local().Format("2006-01-02 15:04:05"),
		len(r.Assets))
}

const (
	githubAPITimeout        = 30 * time.Second
	gitRegex                = `^(https?|git)(:\/\/|@)([^\/:]+)[\/:]([^\/:]+)\/([^\/\.:]+)(|\.git)$`
	githubDomain            = "github.com"
	githubReleaseFormat     = "https://api.github.com/repos/%s/%s/releases/latest"
	githubAssetFormat       = "https://api.github.com/repos/%s/%s/releases/assets/%d"
	githubAPIAccept         = "application/vnd.github.v3+json"
	githubAPIContent        = "application/json"
	githubAPIAcceptBinaries = "application/octet-stream"
)

var (
	// msignPublic is the public key of the msign keypair used to sign the binaries.
	// It is used to verify the signature of the downloaded binary.
	// It is not a secret and can be shared publicly.
	// This value MUST be updated if the keypair is changed.
	msignPublic = "PUB:ARi1u_Ij_5AStTTLT3JfYmVFgWOS4lGPvrtqEuVLsKnsOzbh5oHZ\n"
)

// githubError is returned by the GitHub API, e.g. for rate-limiting.
type githubError struct {
	Message string
}

// gitHubLatestRelease uses the GitHub API to get information about the latest
// release of a repository.
func githubLatestRelease(ctx context.Context, git string) (Release, error) {
	re := regexp.MustCompile(gitRegex)
	matches := re.FindStringSubmatch(git)
	if len(matches) < 6 {
		return Release{}, fmt.Errorf("invalid GitHub URL %q", git)
	}

	if matches[3] != githubDomain {
		return Release{}, fmt.Errorf("invalid GitHub domain %q", matches[3])
	}

	owner := matches[4]
	repo := matches[5]

	ctx, cancel := context.WithTimeout(ctx, githubAPITimeout)
	defer cancel()

	url := fmt.Sprintf(githubReleaseFormat, owner, repo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Release{}, err
	}

	// pin API version 3
	req.Header.Set("Accept", githubAPIAccept)

	res, err := http.DefaultClient.Do(req.WithContext(ctx))
	// If we got an error, and the context has been canceled,
	// the context's error is probably more useful.
	if err != nil {
		select {
		case <-ctx.Done():
			err = ctx.Err()
		default:
		}
		return Release{}, err
	}

	if res.StatusCode != http.StatusOK {
		content := res.Header.Get("Content-Type")
		if strings.Contains(content, githubAPIContent) {
			// try to decode error message
			var msg githubError
			jerr := json.NewDecoder(res.Body).Decode(&msg)
			if jerr == nil {
				return Release{}, fmt.Errorf("unexpected status %v (%v) returned, message:\n  %v", res.StatusCode, res.Status, msg.Message)
			}
		}

		_ = res.Body.Close()
		return Release{}, fmt.Errorf("unexpected status %v (%v) returned", res.StatusCode, res.Status)
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		_ = res.Body.Close()
		return Release{}, err
	}

	err = res.Body.Close()
	if err != nil {
		return Release{}, err
	}

	var release Release
	err = json.Unmarshal(buf, &release)
	if err != nil {
		return Release{}, err
	}

	if release.TagName == "" {
		return Release{}, errors.New("tag name for latest release is empty")
	}

	return release, nil
}

// githubDownloadAsset uses the GitHub API to download an asset
func githubDownloadAsset(ctx context.Context, asset Asset) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, asset.URL, nil)
	if err != nil {
		return nil, err
	}

	// request binary data
	req.Header.Set("Accept", githubAPIAcceptBinaries)

	res, err := http.DefaultClient.Do(req.WithContext(ctx))
	// If we got an error, and the context has been canceled,
	// the context's error is probably more useful.
	if err != nil {
		select {
		case <-ctx.Done():
			err = ctx.Err()
		default:
		}
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %v (%v) returned", res.StatusCode, res.Status)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		_ = res.Body.Close()
		return nil, err
	}

	err = res.Body.Close()
	if err != nil {
		return nil, err
	}

	return data, nil
}

// GetLatestVersion returns the latest version of released binary on GitHub.
func GetLatestVersion(giturl string) (string, error) {
	release, err := githubLatestRelease(context.Background(), giturl)

	if err != nil {
		return "", err
	}

	return release.TagName, nil
}

// DownloadLatestVersion downloads the latest version of released binary on GitHub.
func DownloadLatestVersion(giturl string, binary string, currentRelease string) error {

	// 1. Get current binary name and path
	currentBinary, err := os.Executable()
	if err != nil {
		fmt.Println(err)
		return err
	}
	currentBinary = path.Clean(currentBinary)
	if unlink, err := filepath.EvalSymlinks(currentBinary); err == nil {
		currentBinary = unlink
	}

	// 2. Get latest version of released assets on GitHub
	release, err := githubLatestRelease(context.Background(), giturl)
	if err != nil {
		return err
	}

	if release.TagName == currentRelease {
		fmt.Printf("Already up to date: %v\n", release.TagName)
		return nil
	}

	fmt.Printf("Update to latest release: %v\n", release.TagName)

	// 3. Find binary and sign assets for current binary/OS/ARCH
	binarySign := fmt.Sprintf("%s.msign", binary)

	var binaryAsset Asset
	var binarySignAsset Asset

	for _, asset := range release.Assets {
		if asset.Name == binary {
			binaryAsset = asset
		}
		if asset.Name == binarySign {
			binarySignAsset = asset
		}
	}

	if binaryAsset.Name == "" {
		return fmt.Errorf("binary asset %q not found", binary)
	}

	if binarySignAsset.Name == "" {
		return fmt.Errorf("binary sign asset %q not found", binarySign)
	}

	// 4. Download binary and sign assets
	fmt.Printf("Downloading %s... ", binaryAsset.Name)
	binaryData, err := githubDownloadAsset(context.Background(), binaryAsset)
	if err != nil {
		fmt.Println("failed")
		return err
	}
	fmt.Println("done")

	fmt.Printf("Downloading %s... ", binarySignAsset.Name)
	binarySignData, err := githubDownloadAsset(context.Background(), binarySignAsset)
	if err != nil {
		fmt.Println("failed")
		return err
	}
	fmt.Println("done")

	// 5. Verify signature
	fmt.Printf("Verifying %s... ", binaryAsset.Name)

	pub, err := msign.ImportPublicKey(strings.NewReader(msignPublic))
	if err != nil {
		fmt.Println("failed")
		return err
	}

	sig, err := msign.ImportSignature(bytes.NewReader(binarySignData))
	if err != nil {
		fmt.Println("failed")
		return err
	}

	valid, err := pub.Verify(bytes.NewReader(binaryData), sig)

	if err != nil {
		fmt.Println("failed")
		return err
	}

	if !valid {
		fmt.Println("failed")
		return errors.New("signature verification failed")
	}

	fmt.Println("done")
	// 6. Replace current binary with downloaded binary
	// 6.1. Save current binary to new name

	//create new file
	fmt.Printf("Saving downloaded update... ")
	newBinary := currentBinary + ".new"
	newBinaryFile, err := os.Create(newBinary)
	if err != nil {
		fmt.Println("failed")
		return err
	}
	_, err = newBinaryFile.Write(binaryData)
	newBinaryFile.Close()
	if err != nil {
		fmt.Println("failed")
		os.Remove(newBinary) // clean up
		return err
	}

	//copy permissions
	info, err := os.Stat(currentBinary)
	if err != nil {
		fmt.Println("failed")
		os.Remove(newBinary) // clean up
		return err
	}

	err = os.Chmod(newBinary, info.Mode())
	if err != nil {
		fmt.Println("failed")
		os.Remove(newBinary) // clean up
		return err
	}
	fmt.Println("done")

	// 6.2. Rename old binary to backup name
	fmt.Printf("Updating... ")
	err = os.Rename(currentBinary, currentBinary+".bak")
	if err != nil {
		fmt.Println("failed")
		os.Remove(newBinary) // clean up
		return err
	}
	// 6.3. Rename new binary to old binary name
	err = os.Rename(newBinary, currentBinary)
	if err != nil {
		fmt.Println("failed")
		os.Rename(currentBinary+".bak", currentBinary) // revert backup
		return err
	}

	fmt.Println("done")

	return nil
}

func init() {
	msignPublic += "\n" // add newline in case public key is not terminated with one
}
