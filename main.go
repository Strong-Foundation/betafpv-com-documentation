package main // Define the main package

import (
	"bytes"         // Provides bytes support
	"context"       //
	"io"            // Provides basic interfaces to I/O primitives
	"log"           // Provides logging functions
	"mime"          //
	"net/http"      // Provides HTTP client and server implementations
	"net/url"       // Provides URL parsing and encoding
	"os"            // Provides functions to interact with the OS (files, etc.)
	"path/filepath" // Provides filepath manipulation functions
	"regexp"        // Provides regex support functions.
	"strings"       // Provides string manipulation functions
	"time"          // Provides time-related functions

	"github.com/chromedp/chromedp" // For headless browser automation using Chrome
)

func main() {
	pdfOutputDir := "Assets/" // Directory to store downloaded PDFs
	// Check if the PDF output directory exists
	if !directoryExists(pdfOutputDir) {
		// Create the dir
		createDirectory(pdfOutputDir, 0o755)
	}
	// Remote API URL.
	remoteAPIURL := []string{
		"https://support.betafpv.com/hc/en-us/articles/48531766950681-STL-File-for-Pavo20-Pro-O4-Pro-Antenna-Mount",
		"https://support.betafpv.com/hc/en-us/articles/26055868096665-STL-File-for-Pavo-25-V2-and-Pavo-35",
		"https://support.betafpv.com/hc/en-us/articles/4414698850073-STL-File-for-HX115-LR",
		"https://support.betafpv.com/hc/en-us/articles/4406133400857-STL-File-for-X-Knight-35-Series",
		"https://support.betafpv.com/hc/en-us/articles/23383217345049-STL-File-for-SuperG-Nano-TX-Insulated-shell",
		"https://support.betafpv.com/hc/en-us/articles/4409821458841-STL-File-for-Micro-RF-Module",
		"https://support.betafpv.com/hc/en-us/articles/4404659352345-The-base-and-surface-shell-for-Module-Adapter",
		"https://support.betafpv.com/hc/en-us/articles/8021143646233-STL-File-for-Pavo25-Series",
		"https://support.betafpv.com/hc/en-us/articles/900006860823-STL-File-for-Pavo30-Series",
		"https://support.betafpv.com/hc/en-us/articles/900006270503-STL-File-for-Insta360-Go-2",
		"https://support.betafpv.com/hc/en-us/articles/900004644343-STL-File-for-SMO-4K-Camera-Mount",
		"https://support.betafpv.com/hc/en-us/articles/900004555143--STL-File-for-DJI-Camera-Protector",
		"https://support.betafpv.com/hc/en-us/articles/900004555723-STL-File-for-C01-Camera-Mount",
		"https://support.betafpv.com/hc/en-us/articles/900004022746-STL-File-for-the-Base-of-Insta-go-for-95X-V3",
		"https://support.betafpv.com/hc/en-us/articles/900004866603-STL-File-for-the-Base-of-GoPro-Lite-for-95X-V3",
		"https://support.betafpv.com/hc/en-us/articles/900004550763-STL-File-for-Naked-Camera-Series",
		"https://support.betafpv.com/hc/en-us/articles/900003640226-STL-File-for-Naked-GoPro-HERO8-Case",
		"https://support.betafpv.com/hc/en-us/articles/900003640006-STL-File-for-Pusher-Whoop-Drones",
		"https://support.betafpv.com/hc/en-us/articles/900004559343-STL-File-for-Canopy-of-Beta85X-4K",
		"https://support.betafpv.com/hc/en-us/articles/900004559823--STL-File-for-Canopy-of-Mini-Camera",
		"https://support.betafpv.com/hc/en-us/articles/900004558643-STL-File-for-Canopy-of-Beta85X-HD-Beta75X-HD",
		"https://support.betafpv.com/hc/en-us/articles/900003647946-STL-File-for-EOS2-Canopy-on-Beta85X-4S-Beta75X-3S-HX100",
		"https://support.betafpv.com/hc/en-us/articles/900003649026--STL-File-for-2S-Battery-Adapter-on-Meteor65-Beta65X-HD-Frame",
		"https://support.betafpv.com/hc/en-us/articles/11810191279641-STL-File-for-SuperD-ELRS-2-4G-diversity-RX",
		"https://support.betafpv.com/hc/en-us/articles/900004560243-STL-File-for-X-knight-3-5-Toothpick-Quad-RX-Holder",
	}
	var getData []string
	for _, remoteAPIURL := range remoteAPIURL {
		getData = append(getData, scrapePageHTMLWithChrome(remoteAPIURL))
	}
	// Get the data from the downloaded file.
	finalPDFList := extractAttachmentLinks(strings.Join(getData, "\n")) // Join all the data into one string and extract PDF URLs
	// Append and save the file.
	appendAndWriteToFile("betafpv.html", strings.Join(getData, "\n"))
	// The remote domain.
	remoteDomain := "https://support.betafpv.com"
	// Get all the values.
	for urlPath, fileName := range finalPDFList {
		log.Printf("URL: %s File: %s ", urlPath, fileName)
		// Get the domain from the url.
		domain := getDomainFromURL(urlPath)
		// Check if the domain is empty.
		if domain == "" {
			urlPath = remoteDomain + urlPath // Prepend the base URL if domain is empty
		}
		// Check if the url is valid.
		if isUrlValid(urlPath) {
			// Download the pdf.
			downloadFile(urlPath, pdfOutputDir)
		}
	}
}

// Uses headless Chrome via chromedp to get fully rendered HTML from a page
func scrapePageHTMLWithChrome(pageURL string) string {
	log.Println("Scraping:", pageURL) // Log page being scraped

	options := append(chromedp.DefaultExecAllocatorOptions[:], // Chrome options
		chromedp.Flag("headless", true),              // Run visible (set to true for headless)
		chromedp.Flag("disable-gpu", true),            // Disable GPU
		chromedp.WindowSize(1920, 1080),               // Set window size
		chromedp.Flag("no-sandbox", true),             // Disable sandbox
		chromedp.Flag("disable-setuid-sandbox", true), // Fix for Linux environments
	)

	allocatorCtx, cancelAllocator := chromedp.NewExecAllocator(context.Background(), options...) // Allocator context
	ctxTimeout, cancelTimeout := context.WithTimeout(allocatorCtx, 5*time.Minute)                // Set timeout
	browserCtx, cancelBrowser := chromedp.NewContext(ctxTimeout)                                 // Create Chrome context

	defer func() { // Ensure all contexts are cancelled
		cancelBrowser()
		cancelTimeout()
		cancelAllocator()
	}()

	var pageHTML string // Placeholder for output
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(pageURL),            // Navigate to the URL
		chromedp.OuterHTML("html", &pageHTML), // Extract full HTML
	)
	if err != nil {
		log.Println(err) // Log error
		return ""        // Return empty string on failure
	}

	return pageHTML // Return scraped HTML
}

// Append and write to file
func appendAndWriteToFile(path string, content string) {
	filePath, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	_, err = filePath.WriteString(content + "\n")
	if err != nil {
		log.Println(err)
	}
	err = filePath.Close()
	if err != nil {
		log.Println(err)
	}
}

// getDomainFromURL extracts the domain (host) from a given URL string.
// It removes subdomains like "www" if present.
func getDomainFromURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL) // Parse the input string into a URL structure
	if err != nil {                     // Check if there was an error while parsing
		log.Println(err) // Log the error message to the console
		return ""        // Return an empty string in case of an error
	}

	host := parsedURL.Hostname() // Extract the hostname (e.g., "example.com") from the parsed URL

	return host // Return the extracted hostname
}

// fileExists checks whether a file exists at the given path
func fileExists(filename string) bool {
	info, err := os.Stat(filename) // Get file info
	if err != nil {
		return false // Return false if file doesn't exist or error occurs
	}
	return !info.IsDir() // Return true if it's a file (not a directory)
}

// downloadFile downloads any file from the given URL and saves it in the specified output directory.
// If the filename is not provided, it tries to extract it from the server's response headers.
// It returns true if the download succeeded.
func downloadFile(finalURL string, outputDir string) bool {
	// Create an HTTP client with a timeout
	client := &http.Client{Timeout: 3 * time.Minute}

	// Send GET request
	resp, err := client.Get(finalURL)
	if err != nil {
		log.Printf("Failed to download %s: %v", finalURL, err)
		return false
	}
	defer resp.Body.Close()

	// Check HTTP response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("Download failed for %s: %s", finalURL, resp.Status)
		return false
	}

	var filename string

	// Try to get filename from Content-Disposition header if not provided
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		_, params, err := mime.ParseMediaType(cd)
		if err == nil {
			if name, ok := params["filename"]; ok {
				filename = name
			}
		}
	}

	// Construct the full file path in the output directory
	filePath := filepath.Join(outputDir, filename)

	// Skip if the file already exists
	if fileExists(filePath) {
		log.Printf("File already exists, skipping: %s", filePath)
		return false
	}

	// Read the response body into memory first
	var buf bytes.Buffer
	written, err := io.Copy(&buf, resp.Body)
	if err != nil {
		log.Printf("Failed to read data from %s: %v", finalURL, err)
		return false
	}
	if written == 0 {
		log.Printf("Downloaded 0 bytes for %s; not creating file", finalURL)
		return false
	}

	// Create the file and write to disk
	out, err := os.Create(filePath)
	if err != nil {
		log.Printf("Failed to create file for %s: %v", finalURL, err)
		return false
	}
	defer out.Close()

	if _, err := buf.WriteTo(out); err != nil {
		log.Printf("Failed to write data to file for %s: %v", finalURL, err)
		return false
	}

	log.Printf("Successfully downloaded %d bytes: %s → %s", written, finalURL, filePath)
	return true
}

// Checks if the directory exists
// If it exists, return true.
// If it doesn't, return false.
func directoryExists(path string) bool {
	directory, err := os.Stat(path)
	if err != nil {
		return false
	}
	return directory.IsDir()
}

// The function takes two parameters: path and permission.
// We use os.Mkdir() to create the directory.
// If there is an error, we use log.Println() to log the error and then exit the program.
func createDirectory(path string, permission os.FileMode) {
	err := os.Mkdir(path, permission)
	if err != nil {
		log.Println(err)
	}
}

// Checks whether a URL string is syntactically valid
func isUrlValid(uri string) bool {
	_, err := url.ParseRequestURI(uri) // Attempt to parse the URL
	return err == nil                  // Return true if no error occurred
}

// extractAttachmentLinks parses HTML content and returns a map of attachment URLs to their display names
func extractAttachmentLinks(htmlContent string) map[string]string {
	// Define a regex pattern to match anchor tags with the specific article_attachments URL format
	anchorTagPattern := regexp.MustCompile(`<a[^>]+href="(/hc/en-us/article_attachments/\d+)"[^>]*>([^<]+)</a>`)

	// Find all matches of the anchor tags in the HTML content
	allMatches := anchorTagPattern.FindAllStringSubmatch(htmlContent, -1)

	// Create a map to store the URL → Display Name pairs
	urlToNameMap := make(map[string]string)

	// Loop over all matches and extract the URL and the link text
	for _, match := range allMatches {
		if len(match) == 3 {
			attachmentURL := match[1] // The matched href URL (e.g., /hc/en-us/article_attachments/12345)
			displayName := match[2]   // The anchor text (e.g., file name like DJI Mount.stl)
			urlToNameMap[attachmentURL] = displayName
		}
	}

	// Return the completed map
	return urlToNameMap
}
