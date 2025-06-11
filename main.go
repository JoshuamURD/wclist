package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joshuamURD/wclist/wclist"
)

func main() {
	// Create a new cause list
	causeList := wclist.NewCauseList("Queensland", "Brisbane", time.Now())

	// Open and read the PDF file
	file, err := os.Open("test/test.pdf")
	if err != nil {
		log.Fatalf("Error opening PDF file: %v", err)
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		log.Fatalf("Error getting file stats: %v", err)
	}

	// Read the cause list from PDF
	fmt.Println("Reading cause list from PDF...")
	err = causeList.ReadCauseList(file, stat.Size())
	if err != nil {
		log.Fatalf("Error reading cause list: %v", err)
	}

	fmt.Printf("Successfully parsed %d items from the cause list\n", len(causeList.Items))

	// Display parsed items
	for i, item := range causeList.Items {
		if i >= 5 { // Show only first 5 items for brevity
			fmt.Printf("... and %d more items\n", len(causeList.Items)-5)
			break
		}
		fmt.Printf("Item %d:\n", i+1)
		fmt.Printf("  Matter Number: %d\n", item.GetMatterNumber())

		// Check if this is an objection item to display objection number
		if objItem, ok := item.(wclist.ObjectionItems); ok {
			fmt.Printf("  Objection Number: %d\n", objItem.GetObjectionNumber())
		}

		fmt.Printf("  Tenement: %s\n", item.GetTenementNumber())
		fmt.Printf("  Applying Party (Applicant): %s\n", item.GetApplyingParty())
		fmt.Printf("  Responding Party (Objector): %s\n", item.GetRespondingParty())
		fmt.Printf("  Comments: %s\n", item.GetComments())
		fmt.Println()
	}

	// Example: Search for lawyer's assigned matters
	fmt.Println("Searching for assigned matters...")

	// Example assigned matters (would come from lawyer.go)
	assignedMatters := []wclist.AssignedMatter{
		{
			ClientName:      "Karorra (Higginsville) Pty Ltd",
			TenementNumber:  "E 15/2082",
			OtherPartyNames: []string{"XYZ Corp", "DEF Industries"},
		},
		{
			ClientName:      "FOCUS MINERALS LTD",
			TenementNumber:  "L 15/474",
			OtherPartyNames: []string{"Jones Mining", "Brown Resources"},
		},
	}

	// Search for matches
	matches := causeList.SearchAssignedMatters(assignedMatters)

	if len(matches) > 0 {
		fmt.Printf("Found %d matches:\n", len(matches))
		for i, match := range matches {
			fmt.Printf("Match %d:\n", i+1)
			fmt.Printf("  Client: %s\n", match.AssignedMatter.ClientName)
			fmt.Printf("  Tenement: %s\n", match.AssignedMatter.TenementNumber)
			fmt.Printf("  Matter Number: %d\n", match.CauseListItem.GetMatterNumber())
			fmt.Printf("  Match Reason: %s\n", match.MatchReason)
			fmt.Println()
		}
	} else {
		fmt.Println("No matches found for assigned matters.")
	}
}
