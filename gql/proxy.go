package gql

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Proxy intercepts GQL requests and logs them.
type Proxy struct {
	listenAddr   string
	targetURL    *url.URL
	outputDir    string
	logFile      *os.File
	reverseProxy *httputil.ReverseProxy
	operations   map[string]bool
	opMu         sync.Mutex
	verbose      bool
	logger       *log.Logger
}

// ProxyConfig configures the proxy server.
type ProxyConfig struct {
	ListenAddr string // Address to listen on (default: ":19808")
	OutputDir  string // Directory to save captured requests (default: "./gql_captures")
	Verbose    bool   // Log verbose output
}

// NewProxy creates a new GQL proxy server.
func NewProxy(cfg ProxyConfig) (*Proxy, error) {
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":19808"
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "./gql_captures"
	}

	targetURL, err := url.Parse(TwitchGQLEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target URL: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create log file
	logPath := filepath.Join(cfg.OutputDir, "proxy.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	p := &Proxy{
		listenAddr: cfg.ListenAddr,
		targetURL:  targetURL,
		outputDir:  cfg.OutputDir,
		logFile:    logFile,
		operations: make(map[string]bool),
		verbose:    cfg.Verbose,
		logger:     log.New(io.MultiWriter(os.Stdout, logFile), "[GQL Proxy] ", log.LstdFlags),
	}

	// Create reverse proxy
	p.reverseProxy = &httputil.ReverseProxy{
		Director:       p.director,
		ModifyResponse: p.modifyResponse,
		ErrorHandler:   p.errorHandler,
	}

	return p, nil
}

// Start starts the proxy server.
func (p *Proxy) Start() error {
	p.logger.Printf("Starting proxy on %s -> %s", p.listenAddr, p.targetURL.String())
	p.logger.Printf("Saving captures to: %s", p.outputDir)
	p.logger.Printf("Configure your browser to proxy through http://localhost%s", p.listenAddr)

	server := &http.Server{
		Addr:    p.listenAddr,
		Handler: p,
	}

	return server.ListenAndServe()
}

// Close closes the proxy and its resources.
func (p *Proxy) Close() error {
	if p.logFile != nil {
		return p.logFile.Close()
	}
	return nil
}

// ServeHTTP handles incoming requests.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		p.handleCORS(w)
		return
	}

	// Capture request body
	var bodyBytes []byte
	if r.Body != nil {
		bodyBytes, _ = io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Log and save the request
	p.captureRequest(r, bodyBytes)

	// Add CORS headers to response
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	// Forward to Twitch
	p.reverseProxy.ServeHTTP(w, r)
}

func (p *Proxy) director(r *http.Request) {
	r.URL.Scheme = p.targetURL.Scheme
	r.URL.Host = p.targetURL.Host
	r.URL.Path = p.targetURL.Path
	r.Host = p.targetURL.Host
}

func (p *Proxy) modifyResponse(resp *http.Response) error {
	// Add CORS headers
	resp.Header.Set("Access-Control-Allow-Origin", "*")
	resp.Header.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	resp.Header.Set("Access-Control-Allow-Headers", "*")
	return nil
}

func (p *Proxy) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	p.logger.Printf("Proxy error: %v", err)
	http.Error(w, err.Error(), http.StatusBadGateway)
}

func (p *Proxy) handleCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.WriteHeader(http.StatusNoContent)
}

func (p *Proxy) captureRequest(r *http.Request, body []byte) {
	if len(body) == 0 {
		return
	}

	// Try to parse as JSON
	var requests []json.RawMessage
	if err := json.Unmarshal(body, &requests); err != nil {
		// Try single request
		var single json.RawMessage
		if err := json.Unmarshal(body, &single); err != nil {
			if p.verbose {
				p.logger.Printf("Failed to parse request body as JSON")
			}
			return
		}
		requests = []json.RawMessage{single}
	}

	for _, rawReq := range requests {
		p.processRequest(rawReq)
	}
}

func (p *Proxy) processRequest(raw json.RawMessage) {
	var req struct {
		OperationName string                 `json:"operationName"`
		Query         string                 `json:"query"`
		Variables     map[string]interface{} `json:"variables"`
		Extensions    *struct {
			PersistedQuery *struct {
				Version    int    `json:"version"`
				SHA256Hash string `json:"sha256Hash"`
			} `json:"persistedQuery"`
		} `json:"extensions"`
	}

	if err := json.Unmarshal(raw, &req); err != nil {
		return
	}

	opName := req.OperationName
	if opName == "" {
		opName = "anonymous"
	}

	// Check if we've seen this operation
	p.opMu.Lock()
	isNew := !p.operations[opName]
	p.operations[opName] = true
	p.opMu.Unlock()

	// Log the operation
	timestamp := time.Now().Format("15:04:05")
	marker := " "
	if isNew {
		marker = "*"
	}

	var hashInfo string
	if req.Extensions != nil && req.Extensions.PersistedQuery != nil {
		hashInfo = fmt.Sprintf(" [hash: %s]", req.Extensions.PersistedQuery.SHA256Hash[:16]+"...")
	}

	p.logger.Printf("%s [%s] %s%s", marker, timestamp, opName, hashInfo)

	if p.verbose && len(req.Variables) > 0 {
		varsJSON, _ := json.Marshal(req.Variables)
		p.logger.Printf("  Variables: %s", string(varsJSON))
	}

	// Save to file
	p.saveOperation(opName, raw, req.Query, req.Extensions)
}

func (p *Proxy) saveOperation(opName string, raw json.RawMessage, query string, extensions interface{}) {
	// Create filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.json", sanitizeFilename(opName), timestamp)
	filepath := filepath.Join(p.outputDir, filename)

	// Format the JSON nicely
	var formatted bytes.Buffer
	if err := json.Indent(&formatted, raw, "", "  "); err != nil {
		formatted.Write(raw)
	}

	if err := os.WriteFile(filepath, formatted.Bytes(), 0644); err != nil {
		p.logger.Printf("Failed to save operation: %v", err)
		return
	}

	// Also save to a combined operations file
	p.appendToOperationsLog(opName, raw, query, extensions)
}

func (p *Proxy) appendToOperationsLog(opName string, raw json.RawMessage, query string, extensions interface{}) {
	opsFile := filepath.Join(p.outputDir, "operations.jsonl")
	f, err := os.OpenFile(opsFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	entry := map[string]interface{}{
		"timestamp":     time.Now().Format(time.RFC3339),
		"operationName": opName,
		"request":       raw,
	}

	line, _ := json.Marshal(entry)
	f.Write(line)
	f.WriteString("\n")
}

func sanitizeFilename(s string) string {
	// Replace invalid filename characters
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(s)
}

// GetCapturedOperations returns the list of captured operation names.
func (p *Proxy) GetCapturedOperations() []string {
	p.opMu.Lock()
	defer p.opMu.Unlock()

	ops := make([]string, 0, len(p.operations))
	for op := range p.operations {
		ops = append(ops, op)
	}
	return ops
}

// CapturedCount returns the number of unique operations captured.
func (p *Proxy) CapturedCount() int {
	p.opMu.Lock()
	defer p.opMu.Unlock()
	return len(p.operations)
}

// DecompressGzip decompresses gzipped data (useful for Spade events).
func DecompressGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}
