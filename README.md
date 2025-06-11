# WCList - Cause List Parser and Search System

A Go application that parses Wardens Court cause lists and helps lawyers identify their matters scheduled for hearings.

## Features

- **PDF Parsing**: Extracts cause list data from PDF files with table structures
- **Multiple Matter Types**: Supports objection, forfeiture, and exemption matters
- **Intelligent Search**: Matches lawyer's assigned matters against cause list items
- **Flexible Matching**: Matches by tenement number, client names, and other party names
- **Name Normalization**: Handles common name variations and formatting differences

## Structure

### Core Components

- `wclist/cause_list.go` - Main parsing and search logic
- `wclist/cause_list_items.go` - Data structures for different matter types
- `lawyer/lawyer.go` - Lawyer and assigned matter structures
- `main.go` - Example usage

### Data Types

#### Cause List Items

- **ObjectionItems**: Objection matters with objection number, objector, and applicant
- **ForfeitureItems**: Forfeiture matters with applicant and respondent
- **ExemptionItems**: Exemption matters with applicant and respondent

All items implement the `CauseListItem` interface with methods:
- `GetMatterNumber() uint64`
- `GetTenementNumber() string`
- `GetComments() string`
- `GetApplyingParty() string`
- `GetRespondingParty() string`

#### Search Results

- **MatchResult**: Contains the matched assigned matter, cause list item, and match reason

## Usage

### Basic PDF Parsing

```go
package main

import (
    "os"
    "time"
    "github.com/joshuamURD/wclist/wclist"
)

func main() {
    // Create a new cause list
    causeList := wclist.NewCauseList("Queensland", "Brisbane", time.Now())
    
    // Open PDF file
    file, err := os.Open("cause_list.pdf")
    if err != nil {
        panic(err)
    }
    defer file.Close()
    
    // Get file size
    stat, _ := file.Stat()
    
    // Parse the PDF
    err = causeList.ReadCauseList(file, stat.Size())
    if err != nil {
        panic(err)
    }
    
    // Access parsed items
    for _, item := range causeList.Items {
        fmt.Printf("Matter %d: %s\n", item.GetMatterNumber(), item.GetTenementNumber())
    }
}
```

### Searching for Assigned Matters

```go
// Define assigned matters
assignedMatters := []wclist.AssignedMatter{
    {
        ClientName:      "ABC Mining Company",
        TenementNumber:  "ML12345",
        OtherPartyNames: []string{"XYZ Corp", "DEF Industries"},
    },
}

// Search for matches
matches := causeList.SearchAssignedMatters(assignedMatters)

// Process results
for _, match := range matches {
    fmt.Printf("Match found: %s (Reason: %s)\n", 
        match.AssignedMatter.ClientName, 
        match.MatchReason)
}
```

## PDF Format Requirements

The system expects PDF files with:

1. **Cover page** (page 1) - skipped during parsing
2. **Table data** (page 2 onwards) with columns:
   - Matter Number
   - Objection Number (objection matters only)
   - Objector/Applicant
   - Tenement Affected
   - Applicant/Respondent
   - Comments

### Supported Table Formats

#### Objection Matters
| Matter Number | Objection Number | Objector | Tenement Affected | Applicant | Comments |

#### Forfeiture/Exemption Matters
| Matter Number | Tenement Affected | Applicant | Respondent | Comments |

## Matching Logic

The search system uses a hierarchical matching approach:

1. **Primary Match**: Exact tenement number match (case-insensitive)
2. **Secondary Match**: Client name matches applying or responding party
3. **Tertiary Match**: Other party names match applying or responding party

### Name Matching Features

- Case-insensitive comparison
- Punctuation normalization
- Partial name matching (e.g., "J. Smith" matches "John Smith")
- Extra whitespace handling

## Dependencies

- `github.com/dslipak/pdf` - PDF parsing library

## Installation

```bash
go get github.com/joshuamURD/wclist
```

## Running the Example

```bash
# Build the application
go build .

# Run with test PDF
./wclist
```

## Error Handling

The system gracefully handles:
- Corrupted or unreadable PDF pages
- Missing table data
- Malformed rows
- Different table structures within the same document

## Contributing

This system is designed to be extensible. To add new matter types:

1. Define a new struct in `cause_list_items.go`
2. Implement the `CauseListItem` interface
3. Add parsing logic in `parseTableRow()` method
4. Update section detection in `detectSectionType()`

## Notes

- The first page of PDFs is always skipped (assumed to be a cover page)
- Text extraction relies on the PDF library's ability to group text by rows
- The system attempts to auto-detect different matter types based on content
- Name matching is designed to handle common legal document formatting variations 