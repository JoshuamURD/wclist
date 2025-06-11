package wclist

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dslipak/pdf"
)

func TestReadCauseList(t *testing.T) {
	t.Log("Opening PDF file...")
	file, err := os.Open("../test/test.pdf")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	t.Log("Getting file info...")
	fileInfo, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("PDF file size: %d bytes", fileInfo.Size())

	t.Log("Creating PDF reader...")
	pdf, err := pdf.NewReader(file, fileInfo.Size())
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Testing PDF package", func(t *testing.T) {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Create a channel to receive the result
		done := make(chan struct{})
		var text string
		var textErr error

		go func() {
			t.Log("Getting first page...")
			page := pdf.Page(1)

			t.Log("Getting page content...")
			content := page.Content()

			t.Log("Extracting text...")
			var builder strings.Builder
			for _, text := range content.Text {
				builder.WriteString(text.S)
			}
			text = builder.String()
			close(done)
		}()

		// Wait for either completion or timeout
		select {
		case <-ctx.Done():
			t.Fatal("Test timed out after 5 seconds")
		case <-done:
			if textErr != nil {
				t.Fatalf("Failed to get text: %v", textErr)
			}
			if text == "" {
				t.Error("Expected non-empty text content from first page")
			}

			// Log the text content for debugging
			t.Logf("First page text content:\n%s", text)

			// Basic validation that we got some text
			if len(strings.TrimSpace(text)) == 0 {
				t.Error("Expected non-empty text content after trimming whitespace")
			}
		}
	})

	t.Run("Testing Kalgoorlie CauseList", func(t *testing.T) {
		cl := NewCauseList("Kalgoorlie", "John Doe", time.Now())
		file, err := os.Open("../test/test.pdf")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		cl.ReadCauseList(file, fileInfo.Size())

		if len(cl.Items) == 0 {
			t.Error("Expected at least one item in the cause list")
		}

		for _, item := range cl.Items {
			t.Logf("Item: %+v", item)
		}
	})
}
