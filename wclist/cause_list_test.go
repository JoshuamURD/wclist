package wclist

import (
	"os"
	"testing"
	"time"
)

func TestCauseList(t *testing.T) {
	cl := NewCauseList("Warden's Court", "Warden's Court", time.Now())
	file, err := os.Open("testdata/cause_list.pdf")
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	t.Run("Read cause list", func(t *testing.T) {
		err := cl.ReadCauseList(file, stat.Size())
		if err != nil {
			t.Fatalf("Failed to read cause list: %v", err)
		}
	})
}
