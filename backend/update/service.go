package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type ReleaseInfo struct {
	Tag         string    `json:"tag"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"publishedAt"`
	Notes       string    `json:"notes"`
}

type Status struct {
	Repo            string       `json:"repo"`
	CurrentVersion  string       `json:"currentVersion"`
	Latest          *ReleaseInfo `json:"latest,omitempty"`
	UpdateAvailable bool         `json:"updateAvailable"`
}

type ApplyResult struct {
	Applied bool        `json:"applied"`
	Info    ReleaseInfo `json:"info"`
}

type Service struct {
	repo          string
	current       string
	checkInterval time.Duration
	client        *http.Client
	token         string
	mu            sync.Mutex
	lastCheck     time.Time
	cached        *ReleaseInfo
}

func NewService(repo string, current string, interval time.Duration, token string) *Service {
	return &Service{
		repo:          repo,
		current:       current,
		checkInterval: interval,
		client:        &http.Client{Timeout: 30 * time.Second},
		token:         token,
	}
}

func (s *Service) Status(ctx context.Context) (Status, error) {
	info, err := s.latestRelease(ctx)
	if err != nil {
		return Status{Repo: s.repo, CurrentVersion: s.current, UpdateAvailable: false}, err
	}
	available := compareVersions(s.current, info.Tag) < 0
	return Status{
		Repo:            s.repo,
		CurrentVersion:  s.current,
		Latest:          info,
		UpdateAvailable: available,
	}, nil
}

func (s *Service) ApplyUpdate(ctx context.Context, tag string) (ApplyResult, error) {
	release, assets, err := s.fetchRelease(ctx, tag)
	if err != nil {
		return ApplyResult{}, err
	}
	assetName := releaseAssetName()
	assetURL, shaURL := findAssets(assets, assetName)
	if assetURL == "" {
		return ApplyResult{}, errors.New("khong tim thay release asset phu hop")
	}

	tmpDir, err := os.MkdirTemp("", "flowdb-update-*")
	if err != nil {
		return ApplyResult{}, err
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, assetName)
	if err := s.download(ctx, assetURL, archivePath); err != nil {
		return ApplyResult{}, err
	}

	if shaURL != "" {
		if err := s.verifySHA256(ctx, shaURL, archivePath); err != nil {
			return ApplyResult{}, err
		}
	}

	binPath, err := extractBinary(tmpDir, archivePath)
	if err != nil {
		return ApplyResult{}, err
	}
	if err := replaceBinary(binPath); err != nil {
		return ApplyResult{}, err
	}
	return ApplyResult{
		Applied: true,
		Info:    release,
	}, nil
}

func (s *Service) latestRelease(ctx context.Context) (*ReleaseInfo, error) {
	s.mu.Lock()
	if s.cached != nil && time.Since(s.lastCheck) < s.checkInterval {
		cached := *s.cached
		s.mu.Unlock()
		return &cached, nil
	}
	s.mu.Unlock()

	info, _, err := s.fetchRelease(ctx, "")
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.cached = &info
	s.lastCheck = time.Now()
	s.mu.Unlock()
	return &info, nil
}

type assetInfo struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

func (s *Service) fetchRelease(ctx context.Context, tag string) (ReleaseInfo, []assetInfo, error) {
	url := s.releaseURL(tag)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ReleaseInfo{}, nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return ReleaseInfo{}, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return ReleaseInfo{}, nil, errors.New("failed to fetch release")
	}
	var payload struct {
		TagName     string      `json:"tag_name"`
		HTMLURL     string      `json:"html_url"`
		PublishedAt string      `json:"published_at"`
		Body        string      `json:"body"`
		Assets      []assetInfo `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return ReleaseInfo{}, nil, err
	}
	parsedTime, _ := time.Parse(time.RFC3339, payload.PublishedAt)
	info := ReleaseInfo{
		Tag:         payload.TagName,
		URL:         payload.HTMLURL,
		PublishedAt: parsedTime,
		Notes:       strings.TrimSpace(payload.Body),
	}
	return info, payload.Assets, nil
}

func (s *Service) releaseURL(tag string) string {
	if tag == "" {
		return "https://api.github.com/repos/" + s.repo + "/releases/latest"
	}
	return "https://api.github.com/repos/" + s.repo + "/releases/tags/" + tag
}

func (s *Service) download(ctx context.Context, url string, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return errors.New("failed to download asset")
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func (s *Service) verifySHA256(ctx context.Context, url string, filePath string) error {
	tmp, err := os.CreateTemp("", "flowdb-sha-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := s.download(ctx, url, tmp.Name()); err != nil {
		return err
	}
	expectedRaw, err := os.ReadFile(tmp.Name())
	if err != nil {
		return err
	}
	expected := strings.Fields(string(expectedRaw))
	if len(expected) == 0 {
		return errors.New("invalid sha256 file")
	}
	actual, err := sha256File(filePath)
	if err != nil {
		return err
	}
	if !strings.EqualFold(expected[0], actual) {
		return errors.New("sha256 mismatch")
	}
	return nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func findAssets(assets []assetInfo, name string) (string, string) {
	var assetURL string
	var shaURL string
	for _, asset := range assets {
		if asset.Name == name {
			assetURL = asset.URL
		}
		if asset.Name == name+".sha256" {
			shaURL = asset.URL
		}
	}
	return assetURL, shaURL
}

func releaseAssetName() string {
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	return "flowdb_" + runtime.GOOS + "_" + runtime.GOARCH + "." + ext
}

func extractBinary(tmpDir string, archivePath string) (string, error) {
	if strings.HasSuffix(archivePath, ".zip") {
		return extractZip(tmpDir, archivePath)
	}
	return extractTarGz(tmpDir, archivePath)
}

func extractZip(tmpDir string, archivePath string) (string, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		if filepath.Base(file.Name) == "flowdb" || filepath.Base(file.Name) == "flowdb.exe" {
			dest := filepath.Join(tmpDir, filepath.Base(file.Name))
			rc, err := file.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()
			out, err := os.Create(dest)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(out, rc); err != nil {
				out.Close()
				return "", err
			}
			out.Close()
			return dest, nil
		}
	}
	return "", errors.New("khong tim thay binary trong zip")
}

func extractTarGz(tmpDir string, archivePath string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if hdr.FileInfo().IsDir() {
			continue
		}
		base := filepath.Base(hdr.Name)
		if base == "flowdb" {
			dest := filepath.Join(tmpDir, base)
			out, err := os.Create(dest)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return "", err
			}
			out.Close()
			return dest, nil
		}
	}
	return "", errors.New("khong tim thay binary trong tar.gz")
}

func replaceBinary(newBinary string) error {
	current, err := os.Executable()
	if err != nil {
		return err
	}
	current, err = filepath.Abs(current)
	if err != nil {
		return err
	}
	backup := current + ".old"
	_ = os.Remove(backup)
	if err := os.Rename(current, backup); err != nil {
		return err
	}
	if err := os.Rename(newBinary, current); err != nil {
		_ = os.Rename(backup, current)
		return err
	}
	return nil
}

func compareVersions(current string, latest string) int {
	clean := func(v string) string {
		return strings.TrimSpace(strings.TrimPrefix(v, "v"))
	}
	cParts := strings.Split(clean(current), ".")
	lParts := strings.Split(clean(latest), ".")
	for len(cParts) < 3 {
		cParts = append(cParts, "0")
	}
	for len(lParts) < 3 {
		lParts = append(lParts, "0")
	}
	for i := 0; i < 3; i++ {
		c := parseInt(cParts[i])
		l := parseInt(lParts[i])
		if c < l {
			return -1
		}
		if c > l {
			return 1
		}
	}
	return 0
}

func parseInt(value string) int {
	value = strings.TrimSpace(value)
	n := 0
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			break
		}
		n = n*10 + int(ch-'0')
	}
	return n
}
