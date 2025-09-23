package horizon

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rotisserie/eris"
)

// FileSecurityService provides malware detection and file validation
type FileSecurityService interface {
	ValidateFileUpload(ctx context.Context, file *multipart.FileHeader) (*FileValidationResult, error)
	ScanFileContent(ctx context.Context, content []byte, filename string) (*FileScanResult, error)
	IsAllowedFileType(filename string, contentType string) bool
	CalculateFileHashes(content []byte) *FileHashes
}

type FileValidationResult struct {
	IsValid     bool     `json:"is_valid"`
	Filename    string   `json:"filename"`
	ContentType string   `json:"content_type"`
	Size        int64    `json:"size"`
	Errors      []string `json:"errors,omitempty"`
	Warnings    []string `json:"warnings,omitempty"`
	Hashes      *FileHashes `json:"hashes,omitempty"`
}

type FileScanResult struct {
	IsSafe           bool     `json:"is_safe"`
	ThreatLevel      string   `json:"threat_level"` // low, medium, high, critical
	DetectedThreats  []string `json:"detected_threats,omitempty"`
	SuspiciousItems  []string `json:"suspicious_items,omitempty"`
	Recommendations  []string `json:"recommendations,omitempty"`
}

type FileHashes struct {
	MD5    string `json:"md5"`
	SHA256 string `json:"sha256"`
}

type HorizonFileSecurity struct {
	maxFileSize       int64
	allowedExtensions map[string]bool
	allowedMimeTypes  map[string]bool
	malwareSignatures []MalwareSignature
}

type MalwareSignature struct {
	Name        string
	Pattern     []byte
	Description string
	ThreatLevel string
}

// NewFileSecurityService creates a new file security service
func NewFileSecurityService(maxFileSize int64) FileSecurityService {
	return &HorizonFileSecurity{
		maxFileSize:       maxFileSize,
		allowedExtensions: getAllowedExtensions(),
		allowedMimeTypes:  getAllowedMimeTypes(),
		malwareSignatures: getMalwareSignatures(),
	}
}

// ValidateFileUpload performs comprehensive validation on uploaded files
func (h *HorizonFileSecurity) ValidateFileUpload(ctx context.Context, file *multipart.FileHeader) (*FileValidationResult, error) {
	result := &FileValidationResult{
		IsValid:     true,
		Filename:    file.Filename,
		ContentType: file.Header.Get("Content-Type"),
		Size:        file.Size,
		Errors:      []string{},
		Warnings:    []string{},
	}

	// 1. Check file size
	if file.Size > h.maxFileSize {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("File size %d exceeds maximum allowed size %d", file.Size, h.maxFileSize))
	}

	// 2. Check filename for suspicious patterns
	if err := h.validateFilename(file.Filename); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, err.Error())
	}

	// 3. Check file extension
	if !h.IsAllowedFileType(file.Filename, result.ContentType) {
		result.IsValid = false
		result.Errors = append(result.Errors, "File type not allowed")
	}

	// 4. Read and validate file content
	src, err := file.Open()
	if err != nil {
		return nil, eris.Wrap(err, "failed to open uploaded file")
	}
	defer src.Close()

	content, err := io.ReadAll(src)
	if err != nil {
		return nil, eris.Wrap(err, "failed to read file content")
	}

	// 5. Verify content type matches file extension
	detectedType := http.DetectContentType(content)
	if !h.isContentTypeConsistent(file.Filename, detectedType, result.ContentType) {
		result.Warnings = append(result.Warnings, "Content type mismatch detected")
	}

	// 6. Calculate file hashes
	result.Hashes = h.CalculateFileHashes(content)

	// 7. Scan for malicious content
	scanResult, err := h.ScanFileContent(ctx, content, file.Filename)
	if err != nil {
		result.Warnings = append(result.Warnings, "Could not complete malware scan: "+err.Error())
	} else if !scanResult.IsSafe {
		result.IsValid = false
		result.Errors = append(result.Errors, "Malicious content detected")
		result.Errors = append(result.Errors, scanResult.DetectedThreats...)
	}

	return result, nil
}

// ScanFileContent scans file content for malicious patterns
func (h *HorizonFileSecurity) ScanFileContent(ctx context.Context, content []byte, filename string) (*FileScanResult, error) {
	result := &FileScanResult{
		IsSafe:          true,
		ThreatLevel:     "low",
		DetectedThreats: []string{},
		SuspiciousItems: []string{},
		Recommendations: []string{},
	}

	// 1. Check for known malware signatures
	for _, signature := range h.malwareSignatures {
		if bytes.Contains(content, signature.Pattern) {
			result.IsSafe = false
			result.DetectedThreats = append(result.DetectedThreats, signature.Name+": "+signature.Description)
			if signature.ThreatLevel == "critical" || signature.ThreatLevel == "high" {
				result.ThreatLevel = signature.ThreatLevel
			}
		}
	}

	// 2. Check for suspicious patterns in text files
	if h.isTextFile(filename) {
		contentStr := string(content)
		suspiciousPatterns := h.getSuspiciousTextPatterns()
		
		for pattern, description := range suspiciousPatterns {
			if matched, _ := regexp.MatchString(pattern, contentStr); matched {
				result.SuspiciousItems = append(result.SuspiciousItems, description)
				if result.ThreatLevel == "low" {
					result.ThreatLevel = "medium"
				}
			}
		}
	}

	// 3. Check for embedded executables in images/documents
	if h.isImageOrDocument(filename) {
		if h.containsEmbeddedExecutable(content) {
			result.IsSafe = false
			result.DetectedThreats = append(result.DetectedThreats, "Embedded executable detected in media file")
			result.ThreatLevel = "high"
		}
	}

	// 4. Check file structure integrity
	if err := h.validateFileStructure(content, filename); err != nil {
		result.SuspiciousItems = append(result.SuspiciousItems, "File structure anomaly: "+err.Error())
		if result.ThreatLevel == "low" {
			result.ThreatLevel = "medium"
		}
	}

	// 5. Generate recommendations
	if len(result.SuspiciousItems) > 0 {
		result.Recommendations = append(result.Recommendations, "Consider additional manual review")
	}
	if result.ThreatLevel != "low" {
		result.Recommendations = append(result.Recommendations, "Recommend scanning with updated antivirus")
	}

	return result, nil
}

// IsAllowedFileType checks if file type is allowed
func (h *HorizonFileSecurity) IsAllowedFileType(filename string, contentType string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	
	// Check extension
	if !h.allowedExtensions[ext] {
		return false
	}
	
	// Check MIME type
	if contentType != "" && !h.allowedMimeTypes[contentType] {
		return false
	}
	
	return true
}

// CalculateFileHashes generates MD5 and SHA256 hashes
func (h *HorizonFileSecurity) CalculateFileHashes(content []byte) *FileHashes {
	md5Hash := md5.Sum(content)
	sha256Hash := sha256.Sum256(content)
	
	return &FileHashes{
		MD5:    hex.EncodeToString(md5Hash[:]),
		SHA256: hex.EncodeToString(sha256Hash[:]),
	}
}

// Helper methods

func (h *HorizonFileSecurity) validateFilename(filename string) error {
	// Check for path traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return eris.New("filename contains path traversal characters")
	}
	
	// Check for suspicious patterns
	suspiciousPatterns := []string{
		`\.exe$`, `\.bat$`, `\.cmd$`, `\.com$`, `\.scr$`, `\.vbs$`, `\.js$`, `\.jar$`,
		`\.php$`, `\.asp$`, `\.aspx$`, `\.jsp$`, `\.pl$`, `\.py$`, `\.rb$`, `\.sh$`,
	}
	
	lowerFilename := strings.ToLower(filename)
	for _, pattern := range suspiciousPatterns {
		if matched, _ := regexp.MatchString(pattern, lowerFilename); matched {
			return eris.Errorf("suspicious file extension detected: %s", filepath.Ext(filename))
		}
	}
	
	return nil
}

func (h *HorizonFileSecurity) isContentTypeConsistent(filename, detectedType, declaredType string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	
	// Map of extensions to expected content types
	expectedTypes := map[string][]string{
		".jpg":  {"image/jpeg"},
		".jpeg": {"image/jpeg"},
		".png":  {"image/png"},
		".gif":  {"image/gif"},
		".pdf":  {"application/pdf"},
		".doc":  {"application/msword"},
		".docx": {"application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		".txt":  {"text/plain"},
		".csv":  {"text/csv", "application/csv"},
	}
	
	if expected, exists := expectedTypes[ext]; exists {
		for _, expectedType := range expected {
			if strings.HasPrefix(detectedType, expectedType) {
				return true
			}
		}
		return false
	}
	
	return true // Unknown extension, assume consistent
}

func (h *HorizonFileSecurity) isTextFile(filename string) bool {
	textExtensions := []string{".txt", ".csv", ".json", ".xml", ".html", ".css", ".js", ".sql"}
	ext := strings.ToLower(filepath.Ext(filename))
	
	for _, textExt := range textExtensions {
		if ext == textExt {
			return true
		}
	}
	return false
}

func (h *HorizonFileSecurity) isImageOrDocument(filename string) bool {
	mediaExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx", ".ppt", ".pptx"}
	ext := strings.ToLower(filepath.Ext(filename))
	
	for _, mediaExt := range mediaExtensions {
		if ext == mediaExt {
			return true
		}
	}
	return false
}

func (h *HorizonFileSecurity) containsEmbeddedExecutable(content []byte) bool {
	// Check for PE header (Windows executable)
	if bytes.Contains(content, []byte("MZ")) && bytes.Contains(content, []byte("PE\x00\x00")) {
		return true
	}
	
	// Check for ELF header (Linux executable)
	if bytes.HasPrefix(content, []byte("\x7fELF")) {
		return true
	}
	
	// Check for Mach-O header (macOS executable)
	if bytes.HasPrefix(content, []byte("\xfe\xed\xfa\xce")) || bytes.HasPrefix(content, []byte("\xfe\xed\xfa\xcf")) {
		return true
	}
	
	return false
}

func (h *HorizonFileSecurity) validateFileStructure(content []byte, filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	
	switch ext {
	case ".jpg", ".jpeg":
		if !bytes.HasPrefix(content, []byte("\xff\xd8\xff")) {
			return eris.New("invalid JPEG header")
		}
	case ".png":
		if !bytes.HasPrefix(content, []byte("\x89PNG\r\n\x1a\n")) {
			return eris.New("invalid PNG header")
		}
	case ".gif":
		if !bytes.HasPrefix(content, []byte("GIF87a")) && !bytes.HasPrefix(content, []byte("GIF89a")) {
			return eris.New("invalid GIF header")
		}
	case ".pdf":
		if !bytes.HasPrefix(content, []byte("%PDF-")) {
			return eris.New("invalid PDF header")
		}
	}
	
	return nil
}

func (h *HorizonFileSecurity) getSuspiciousTextPatterns() map[string]string {
	return map[string]string{
		`<script[^>]*>.*</script>`:                    "JavaScript code detected",
		`eval\s*\(`:                                  "Eval function detected",
		`exec\s*\(`:                                  "Exec function detected",
		`system\s*\(`:                                "System call detected",
		`shell_exec\s*\(`:                            "Shell execution detected",
		`\$_GET\[|_POST\[|\$_REQUEST\[`:              "PHP superglobal detected",
		`<\?php|<\?=`:                                "PHP code detected",
		`<%.*%>`:                                     "Server-side code detected",
		`SELECT.*FROM|INSERT.*INTO|UPDATE.*SET|DELETE.*FROM`: "SQL commands detected",
		`javascript:|vbscript:|data:`:                "Dangerous URI scheme detected",
	}
}

func getAllowedExtensions() map[string]bool {
	allowed := map[string]bool{
		// Images
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true,
		".webp": true, ".svg": true, ".ico": true,
		
		// Documents
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
		".ppt": true, ".pptx": true, ".txt": true, ".rtf": true,
		
		// Archives (be careful with these)
		".zip": true, ".rar": true, ".7z": true, ".tar": true, ".gz": true,
		
		// Audio/Video
		".mp3": true, ".wav": true, ".mp4": true, ".avi": true, ".mov": true,
		".wmv": true, ".flv": true, ".webm": true,
		
		// Data
		".csv": true, ".json": true, ".xml": true,
	}
	return allowed
}

func getAllowedMimeTypes() map[string]bool {
	allowed := map[string]bool{
		// Images
		"image/jpeg": true, "image/png": true, "image/gif": true, "image/bmp": true,
		"image/webp": true, "image/svg+xml": true, "image/x-icon": true,
		
		// Documents
		"application/pdf": true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/vnd.ms-excel": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
		"application/vnd.ms-powerpoint": true,
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
		"text/plain": true,
		"application/rtf": true,
		
		// Archives
		"application/zip": true,
		"application/x-rar-compressed": true,
		"application/x-7z-compressed": true,
		"application/x-tar": true,
		"application/gzip": true,
		
		// Audio/Video
		"audio/mpeg": true, "audio/wav": true,
		"video/mp4": true, "video/x-msvideo": true, "video/quicktime": true,
		"video/x-ms-wmv": true, "video/x-flv": true, "video/webm": true,
		
		// Data
		"text/csv": true, "application/json": true, "application/xml": true, "text/xml": true,
	}
	return allowed
}

func getMalwareSignatures() []MalwareSignature {
	return []MalwareSignature{
		{
			Name:        "EICAR Test String",
			Pattern:     []byte("X5O!P%@AP[4\\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*"),
			Description: "EICAR antivirus test signature",
			ThreatLevel: "high",
		},
		{
			Name:        "HTML Script Injection",
			Pattern:     []byte("<script>alert("),
			Description: "Basic XSS script injection",
			ThreatLevel: "medium",
		},
		{
			Name:        "PHP Web Shell",
			Pattern:     []byte("<?php system($_GET["),
			Description: "PHP web shell pattern",
			ThreatLevel: "critical",
		},
		{
			Name:        "SQL Injection",
			Pattern:     []byte("UNION SELECT"),
			Description: "SQL injection pattern",
			ThreatLevel: "high",
		},
		{
			Name:        "Command Injection",
			Pattern:     []byte("rm -rf /"),
			Description: "Dangerous command injection",
			ThreatLevel: "critical",
		},
	}
}
