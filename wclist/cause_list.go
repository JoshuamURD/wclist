package wclist

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
)

// CauseListItem represents any item that can be included in a cause list
type CauseListItem interface {
	GetMatterNumber() uint64
	GetTenementNumber() string
	GetComments() string
	GetApplyingParty() string
	GetRespondingParty() string
}

// CauseList represents a cause list
type CauseList struct {
	Jurisdiction string
	Warden       string
	ReleaseDate  time.Time
	Items        []CauseListItem
}

// AssignedMatter represents a matter assigned to a lawyer (imported from lawyer package structure)
type AssignedMatter struct {
	ClientName      string
	TenementNumber  string
	OtherPartyNames []string
}

// MatchResult represents a match between an assigned matter and a cause list item
type MatchResult struct {
	AssignedMatter AssignedMatter
	CauseListItem  CauseListItem
	MatchReason    string
}

// NewCauseList creates a new cause list
func NewCauseList(jurisdiction string, warden string, releaseDate time.Time) *CauseList {
	return &CauseList{
		Jurisdiction: jurisdiction,
		Warden:       warden,
		ReleaseDate:  releaseDate,
		Items:        []CauseListItem{},
	}
}

// SearchAssignedMatters searches for assigned matters in the cause list
func (cl *CauseList) SearchAssignedMatters(assignedMatters []AssignedMatter) []MatchResult {
	var results []MatchResult

	for _, assignedMatter := range assignedMatters {
		for _, item := range cl.Items {
			if match, reason := cl.isMatch(assignedMatter, item); match {
				results = append(results, MatchResult{
					AssignedMatter: assignedMatter,
					CauseListItem:  item,
					MatchReason:    reason,
				})
			}
		}
	}

	return results
}

// isMatch checks if an assigned matter matches a cause list item
func (cl *CauseList) isMatch(assignedMatter AssignedMatter, item CauseListItem) (bool, string) {
	// Primary match: tenement number
	if assignedMatter.TenementNumber != "" &&
		strings.EqualFold(assignedMatter.TenementNumber, item.GetTenementNumber()) {
		return true, "Tenement number match"
	}

	// Secondary match: client name matches applying or responding party
	if assignedMatter.ClientName != "" {
		if cl.namesMatch(assignedMatter.ClientName, item.GetApplyingParty()) {
			return true, "Client name matches applying party"
		}
		if cl.namesMatch(assignedMatter.ClientName, item.GetRespondingParty()) {
			return true, "Client name matches responding party"
		}
	}

	// Tertiary match: other party names
	for _, otherParty := range assignedMatter.OtherPartyNames {
		if otherParty != "" {
			if cl.namesMatch(otherParty, item.GetApplyingParty()) {
				return true, "Other party matches applying party"
			}
			if cl.namesMatch(otherParty, item.GetRespondingParty()) {
				return true, "Other party matches responding party"
			}
		}
	}

	return false, ""
}

// namesMatch performs a case-insensitive comparison of names, handling common variations
func (cl *CauseList) namesMatch(name1, name2 string) bool {
	if name1 == "" || name2 == "" {
		return false
	}

	// Normalize both names
	norm1 := cl.normalizeName(name1)
	norm2 := cl.normalizeName(name2)

	// Exact match after normalization
	if norm1 == norm2 {
		return true
	}

	// Check if one name contains the other (for cases like "John Smith" vs "J. Smith")
	if strings.Contains(norm1, norm2) || strings.Contains(norm2, norm1) {
		return true
	}

	return false
}

// normalizeName normalizes a name for comparison
func (cl *CauseList) normalizeName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Remove common punctuation and extra spaces
	name = regexp.MustCompile(`[^\w\s]`).ReplaceAllString(name, " ")
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")
	name = strings.TrimSpace(name)

	return name
}

// ReadCauseList reads a cause list from a PDF file
func (cl *CauseList) ReadCauseList(file io.Reader, size int64) error {
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	pdfReader, err := pdf.NewReader(bytes.NewReader(data), size)
	if err != nil {
		return err
	}

	numPages := pdfReader.NumPage()
	fmt.Printf("PDF has %d pages\n", numPages)

	// Skip the first page (cover page) and process from page 2 onwards
	for i := 2; i <= numPages; i++ {
		fmt.Printf("Processing page %d...\n", i)

		page := pdfReader.Page(i)
		if page.V.IsNull() {
			fmt.Printf("Page %d is null, skipping\n", i)
			continue
		}

		// Extract text content from the page
		content, err := page.GetPlainText(nil)
		if err != nil {
			fmt.Printf("Error getting text from page %d: %v\n", i, err)
			continue // Skip pages that can't be read
		}

		// Parse the page content and extract items
		items := cl.parsePageText(content)
		cl.Items = append(cl.Items, items...)
		fmt.Printf("Extracted %d items from page %d\n", len(items), i)
	}

	fmt.Printf("Total items extracted: %d\n", len(cl.Items))
	return nil
}

// parsePageText parses plain text from a page and extracts cause list items
func (cl *CauseList) parsePageText(text string) []CauseListItem {
	var items []CauseListItem

	// Split text into lines and clean up
	lines := strings.Split(text, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "TNT-") {
			cleanLines = append(cleanLines, line)
		}
	}

	// Determine the current section type
	currentSection := cl.detectSectionType(text)
	fmt.Printf("Detected section type: %s\n", currentSection)

	// The text is formatted with data spread across multiple lines
	// We need to reconstruct table rows from the scattered text
	items = cl.reconstructTableRows(cleanLines, currentSection)

	return items
}

// reconstructTableRows attempts to reconstruct table rows from fragmented text
func (cl *CauseList) reconstructTableRows(lines []string, sectionType string) []CauseListItem {
	var items []CauseListItem

	// Look for patterns that indicate the start of a data row (matter numbers)
	matterNumberPattern := regexp.MustCompile(`^\d+$`)

	i := 0
	for i < len(lines) {
		line := lines[i]

		// Skip header lines
		if cl.isHeaderLine(line) {
			i++
			continue
		}

		// Check if this line starts with a matter number
		if matterNumberPattern.MatchString(line) {
			matterNumber, err := strconv.ParseUint(line, 10, 64)
			if err != nil {
				i++
				continue
			}

			// Try to extract the full row data starting from this matter number
			item := cl.extractRowFromPosition(lines, i, matterNumber, sectionType)
			if item != nil {
				items = append(items, item)
				fmt.Printf("Extracted item with matter number: %d\n", matterNumber)
			}
		}
		i++
	}

	return items
}

// extractRowFromPosition extracts a complete row starting from a matter number position
func (cl *CauseList) extractRowFromPosition(lines []string, startIdx int, matterNumber uint64, sectionType string) CauseListItem {
	if sectionType == "objection" {
		return cl.extractObjectionItem(lines, startIdx, matterNumber)
	} else if sectionType == "forfeiture" {
		return cl.extractForfeitureItem(lines, startIdx, matterNumber)
	} else if sectionType == "exemption" {
		return cl.extractExemptionItem(lines, startIdx, matterNumber)
	}

	return nil
}

// extractObjectionItem extracts an objection item from the text structure
func (cl *CauseList) extractObjectionItem(lines []string, startIdx int, matterNumber uint64) CauseListItem {
	// The text structure from the PDF is:
	// MATTER_NUMBER OBJECTION_NUMBER OBJECTOR_NAME TENEMENT_NUMBER APPLICANT_NAME
	// We need to reconstruct this from the fragmented lines

	// Collect all content starting from the matter number line until the next matter number
	var contentLines []string

	// Start from the current matter number line and collect subsequent lines
	for i := startIdx; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// Stop if we hit the next matter number (but not the current one)
		if i > startIdx && regexp.MustCompile(`^\d+$`).MatchString(line) && len(line) <= 3 {
			// Check if this might be the start of the next record
			if nextMatter, err := strconv.ParseUint(line, 10, 64); err == nil && nextMatter != matterNumber {
				break
			}
		}

		// Stop if we hit headers
		if cl.isHeaderLine(line) {
			break
		}

		contentLines = append(contentLines, line)

		// Limit how far we look ahead
		if i > startIdx+15 {
			break
		}
	}

	// Join all content and parse as a single string
	fullContent := strings.Join(contentLines, " ")
	fmt.Printf("DEBUG: Full content for matter %d: '%s'\n", matterNumber, fullContent)

	// Parse the content using regex to extract structured data
	return cl.parseObjectionFromContent(fullContent, matterNumber)
}

// parseObjectionFromContent parses objection data from the combined text content
func (cl *CauseList) parseObjectionFromContent(content string, matterNumber uint64) CauseListItem {
	// Expected pattern: MATTER_NUM OBJECTION_NUM OBJECTOR TENEMENT APPLICANT
	// Use regex to match the pattern

	// Remove the matter number from the beginning
	content = regexp.MustCompile(`^\d+\s+`).ReplaceAllString(content, "")

	// Extract objection number (first large number)
	objectionRegex := regexp.MustCompile(`^(\d{6,})\s+`)
	objectionMatches := objectionRegex.FindStringSubmatch(content)
	if len(objectionMatches) < 2 {
		return nil
	}

	objectionNum, err := strconv.ParseUint(objectionMatches[1], 10, 64)
	if err != nil {
		return nil
	}

	// Remove objection number from content
	content = objectionRegex.ReplaceAllString(content, "")

	// Find tenement pattern (like "E 15/2082", "L 28/100", etc.)
	tenementRegex := regexp.MustCompile(`\b([A-Z]+\s+\d+/\d+)\b`)
	tenementMatches := tenementRegex.FindStringSubmatch(content)
	if len(tenementMatches) < 2 {
		return nil
	}

	tenement := tenementMatches[1]
	tenementPos := strings.Index(content, tenement)

	// Everything before the tenement is the objector
	objector := strings.TrimSpace(content[:tenementPos])

	// Everything after the tenement (and tenement itself) contains the applicant
	afterTenement := content[tenementPos+len(tenement):]
	applicant := strings.TrimSpace(afterTenement)

	fmt.Printf("DEBUG: Matter %d, Objection %d, Objector: '%s', Tenement: '%s', Applicant: '%s'\n",
		matterNumber, objectionNum, objector, tenement, applicant)

	return ObjectionItems{
		CLIItems: CLIItems{
			MatterNumber:   matterNumber,
			TenementNumber: tenement,
			Comments:       "",
		},
		ObjectionNumber: objectionNum,
		ObjectorName:    objector,
		ApplicantName:   applicant,
	}
}

// extractForfeitureItem extracts a forfeiture item (placeholder for now)
func (cl *CauseList) extractForfeitureItem(lines []string, startIdx int, matterNumber uint64) CauseListItem {
	// TODO: Implement forfeiture parsing based on actual format
	return nil
}

// extractExemptionItem extracts an exemption item (placeholder for now)
func (cl *CauseList) extractExemptionItem(lines []string, startIdx int, matterNumber uint64) CauseListItem {
	// TODO: Implement exemption parsing based on actual format
	return nil
}

// detectSectionType determines what type of matters are being listed
func (cl *CauseList) detectSectionType(text string) string {
	text = strings.ToLower(text)

	if strings.Contains(text, "objection") || strings.Contains(text, "objector") {
		return "objection"
	} else if strings.Contains(text, "forfeiture") || strings.Contains(text, "forfeit") {
		return "forfeiture"
	} else if strings.Contains(text, "exemption") || strings.Contains(text, "exempt") {
		return "exemption"
	}

	return "unknown"
}

// isHeaderLine checks if a line is a header or section title
func (cl *CauseList) isHeaderLine(line string) bool {
	line = strings.ToLower(line)

	headers := []string{
		"matter number", "objection number", "objector", "tenement affected",
		"applicant", "comments", "respondent", "exemption",
	}

	for _, header := range headers {
		if strings.Contains(line, header) {
			return true
		}
	}

	return false
}

// parseTableRow attempts to parse a line as a table row and return the appropriate item type
func (cl *CauseList) parseTableRow(line string, sectionType string) CauseListItem {
	// Split the line by common delimiters (tabs, multiple spaces)
	fields := regexp.MustCompile(`\s{2,}|\t`).Split(line, -1)

	// Clean up fields
	for i, field := range fields {
		fields[i] = strings.TrimSpace(field)
	}

	// Need at least 3 fields to be a valid row
	if len(fields) < 3 {
		return nil
	}

	// Try to parse the first field as matter number
	matterNumber, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return nil // First field must be a matter number
	}

	switch sectionType {
	case "objection":
		return cl.parseObjectionRow(fields, matterNumber)
	case "forfeiture":
		return cl.parseForfeitureRow(fields, matterNumber)
	case "exemption":
		return cl.parseExemptionRow(fields, matterNumber)
	default:
		// Default to objection if section type is unknown
		return cl.parseObjectionRow(fields, matterNumber)
	}
}

// parseObjectionRow parses a row as an objection item
// Expected format: matter_number, objection_number, objector, tenement_affected, applicant, comments
func (cl *CauseList) parseObjectionRow(fields []string, matterNumber uint64) CauseListItem {
	if len(fields) < 6 {
		return nil
	}

	objectionNumber, _ := strconv.ParseUint(fields[1], 10, 64)

	return ObjectionItems{
		CLIItems: CLIItems{
			MatterNumber:   matterNumber,
			TenementNumber: fields[3], // tenement affected
			Comments:       fields[5], // comments
		},
		ObjectionNumber: objectionNumber,
		ObjectorName:    fields[2], // objector
		ApplicantName:   fields[4], // applicant
	}
}

// parseForfeitureRow parses a row as a forfeiture item
// Expected format: matter_number, tenement_affected, applicant, respondent, comments
func (cl *CauseList) parseForfeitureRow(fields []string, matterNumber uint64) CauseListItem {
	if len(fields) < 5 {
		return nil
	}

	return ForfeitureItems{
		CLIItems: CLIItems{
			MatterNumber:   matterNumber,
			TenementNumber: fields[1],                  // tenement affected
			Comments:       getFieldOrEmpty(fields, 4), // comments
		},
		ApplicantName:  fields[2], // applicant
		RespondentName: fields[3], // respondent
	}
}

// parseExemptionRow parses a row as an exemption item
// Expected format: matter_number, tenement_affected, applicant, respondent, comments
func (cl *CauseList) parseExemptionRow(fields []string, matterNumber uint64) CauseListItem {
	if len(fields) < 4 {
		return nil
	}

	return ExemptionItems{
		CLIItems: CLIItems{
			MatterNumber:   matterNumber,
			TenementNumber: fields[1],                  // tenement affected
			Comments:       getFieldOrEmpty(fields, 4), // comments
		},
		ApplicantName:  fields[2],                  // applicant
		RespondentName: getFieldOrEmpty(fields, 3), // respondent
	}
}

// getFieldOrEmpty safely gets a field from a slice or returns empty string
func getFieldOrEmpty(fields []string, index int) string {
	if index < len(fields) {
		return fields[index]
	}
	return ""
}
