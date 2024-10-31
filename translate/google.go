package translate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	// "net/http/httputil"
	"net/url"
)

// GoogleTranslateRequest represents the request body for Google Cloud Translation API
type GoogleTranslateRequest struct {
	Q      []string `json:"q"`      // Array of texts to translate
	Source string   `json:"source"`  // Source language
	Target string   `json:"target"`  // Target language
	Format string   `json:"format"`  // Text format
}

// GoogleTranslateItem represents a single translation result from Google
type GoogleTranslateItem struct {
	TranslatedText string `json:"translatedText"`
}

// GoogleTranslateResponse represents the response from Google Cloud Translation API
type GoogleTranslateResponse struct {
	Data struct {
		Translations []GoogleTranslateItem `json:"translations"`
	} `json:"data"`
}

// TranslateByGoogle translates texts using Google Cloud Translation API V2
func TranslateByGoogle(sourceLang string, targetLang string, texts []string, apiKey string, proxyURL string) (DeepLXTranslationResult, error) {
	// Parameter validation
	if len(texts) == 0 {
		return DeepLXTranslationResult{
			Code:    http.StatusBadRequest,
			Message: "No text to translate",
		}, nil
	}

	if apiKey == "" {
		return DeepLXTranslationResult{
			Code:    http.StatusBadRequest,
			Message: "API key is required",
		}, nil
	}

	// Set default languages if not specified
	if sourceLang == "" {
		sourceLang = "en"
	}
	if targetLang == "" {
		targetLang = "en"
	}

	// Prepare request body
	requestBody := GoogleTranslateRequest{
		Q:      texts,
		Source: sourceLang,
		Target: targetLang,
		Format: "text",
	}

	// Marshal request body to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return DeepLXTranslationResult{
			Code:    http.StatusInternalServerError,
			Message: "Failed to marshal request",
		}, err
	}

	// log.Printf("Request Body: %s\n", string(jsonData))

	// Build URL with API key
	baseURL := "https://translation.googleapis.com/language/translate/v2"
	fullURL := fmt.Sprintf("%s?key=%s", baseURL, apiKey)
	// log.Printf("Request URL: %s\n", fullURL)

	// Create HTTP request
	request, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return DeepLXTranslationResult{
			Code:    http.StatusServiceUnavailable,
			Message: "Failed to create request",
		}, err
	}

	// Set request headers
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	// Print full request for debugging
	// requestDump, err := httputil.DumpRequestOut(request, true)
	// if err != nil {
	// 	log.Printf("Failed to dump request: %v\n", err)
	// } else {
	// 	log.Printf("Full Request:\n%s\n", string(requestDump))
	// }

	// Configure HTTP client with proxy if specified
	var client *http.Client
	if proxyURL != "" {
		// Parse proxy URL
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			log.Printf("Failed to parse proxy URL: %v\n", err)
			return DeepLXTranslationResult{
				Code:    http.StatusServiceUnavailable,
				Message: fmt.Sprintf("Invalid proxy URL: %v", err),
			}, err
		}

		// Create transport with proxy
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}

		// Create client with custom transport
		client = &http.Client{Transport: transport}
		log.Printf("Using proxy: %s\n", proxyURL)
	} else {
		client = &http.Client{}
		log.Println("No proxy specified, using direct connection")
	}

	// Send request
	resp, err := client.Do(request)
	if err != nil {
		log.Printf("Request error: %v\n", err)
		return DeepLXTranslationResult{
			Code:    http.StatusServiceUnavailable,
			Message: fmt.Sprintf("Translation request failed: %v", err),
		}, err
	}
	defer resp.Body.Close()

	// Print response status and headers
	// log.Printf("Response Status: %s\n", resp.Status)
	// log.Printf("Response Headers: %v\n", resp.Header)

	// Print full response
	// responseDump, err := httputil.DumpResponse(resp, true)
	// if err != nil {
	// 	log.Printf("Failed to dump response: %v\n", err)
	// } else {
	// 	log.Printf("Full Response:\n%s\n", string(responseDump))
	// }

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v\n", err)
		return DeepLXTranslationResult{
			Code:    http.StatusServiceUnavailable,
			Message: fmt.Sprintf("Failed to read response: %v", err),
		}, err
	}

	// log.Printf("Response Body: %s\n", string(body))

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		log.Printf("API returned non-200 status code: %d\n", resp.StatusCode)
		return DeepLXTranslationResult{
			Code:    resp.StatusCode,
			Message: fmt.Sprintf("API error: %s", string(body)),
		}, nil
	}

	// Parse response
	var response GoogleTranslateResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Failed to parse response: %v\n", err)
		return DeepLXTranslationResult{
			Code:    http.StatusServiceUnavailable,
			Message: fmt.Sprintf("Failed to parse response: %v", err),
		}, err
	}

	// Extract translation result
	if len(response.Data.Translations) == 0 {
		log.Println("API returned empty translations array")
		return DeepLXTranslationResult{
			Code:    http.StatusServiceUnavailable,
			Message: "Translation failed, API returns an empty result",
		}, nil
	}

	// Get the first translation result
	translatedText := response.Data.Translations[0].TranslatedText
	// log.Printf("Translated text: %s\n", translatedText)

	// Return successful result
	return DeepLXTranslationResult{
		Code:       http.StatusOK,
		Message:    "Success",
		Data:       translatedText,
		SourceLang: sourceLang,
		TargetLang: targetLang,
		Method:     "GoogleCloud",
	}, nil
}