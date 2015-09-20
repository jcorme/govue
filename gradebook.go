package govue

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

type Gradebook struct {
	XMLName             xml.Name        `xml:"Gradebook"`
	ReportingPeriods    []*ReportPeriod `xml:"ReportingPeriods>ReportPeriod"`
	CurrentReportPeriod *ReportPeriod   `xml:"ReportingPeriod"`
	Courses             []*Course       `xml:"Courses>Course"`
}

type ReportPeriod struct {
	Index       int           `xml:",attr"`
	GradePeriod string        `xml:",attr"`
	StartDate   GradebookDate `xml:",attr"`
	EndDate     GradebookDate `xml:",attr"`
}

type Course struct {
	Period       int           `xml:",attr"`
	Name         string        `xml:"Title,attr"`
	Room         string        `xml:",attr"`
	Teacher      string        `xml:"Staff,attr"`
	TeacherEmail string        `xml:"StaffEMail,attr"`
	Marks        []*CourseMark `xml:"Marks>Mark"`
}

type CourseMark struct {
	Name           string                 `xml:"MarkName,attr"`
	LetterGrade    string                 `xml:"CalculatedScoreString,attr"`
	RawGradeScore  float64                `xml:"CalculatedScoreRaw,attr"`
	GradeSummaries []*AssignmentGradeCalc `xml:"GradeCalculationSummary>AssignmentGradeCalc"`
	Assignments    []*Assignment          `xml:"Assignments>Assignment"`
}

type AssignmentGradeCalc struct {
	Type               string     `xml:",attr"`
	Weight             Percentage `xml:",attr"`
	Points             float64    `xml:",attr"`
	PointsPossible     float64    `xml:",attr"`
	WeightedPercentage Percentage `xml:"WeightedPct,attr"`
	LetterGrade        string     `xml:"CalculatedMark,attr"`
}

type Assignment struct {
	GradebookID string           `xml:",attr"`
	Name        string           `xml:"Measure,attr"`
	Type        string           `xml:",attr"`
	Date        GradebookDate    `xml:",attr"`
	DueDate     GradebookDate    `xml:",attr"`
	Score       AssignmentScore  `xml:",attr"`
	ScoreType   string           `xml:",attr"`
	Points      AssignmentPoints `xml:",attr"`
	Notes       string           `xml:",attr"`
}

type Percentage struct {
	float64
}

func (p *Percentage) UnmarshalXMLAttr(attr xml.Attr) error {
	pct := attr.Value

	if rune(pct[len(pct)-1]) != '%' {
		return fmt.Errorf("Expected percentage attribute in format `x%`, received %s", pct)
	}

	f, err := strconv.ParseFloat(pct[:len(pct)-1], 64)

	if err != nil {
		return err
	}

	*p = Percentage{f}

	return nil
}

type GradebookDate struct {
	time.Time
}

func (gd *GradebookDate) UnmarshalXMLAttr(attr xml.Attr) error {
	const gradebookDateFormat = "1/2/2006"

	dt, err := time.Parse(gradebookDateFormat, attr.Value)

	if err != nil {
		return err
	}

	*gd = GradebookDate{dt}

	return nil
}

type AssignmentScore struct {
	Graded               bool
	Score, PossibleScore float64
}

func (as *AssignmentScore) UnmarshalXMLAttr(attr xml.Attr) error {
	if attr.Value == "Not Graded" {
		*as = AssignmentScore{false, 0, 0}

		return nil
	}

	const scoreRegex = "([\\d\\.]+)\\s*out\\s*of\\s*([\\d\\.]+)"

	r, err := regexp.Compile(scoreRegex)

	if err != nil {
		return err
	}

	scores := r.FindStringSubmatch(attr.Value)

	if len(scores) != 3 {
		return fmt.Errorf("Expected assignment score in format `x out of y`, where x and y are real numbers, received %s and parsed %d numbers", attr.Value, len(scores))
	}

	fs, err := stringsToFloats(scores[1:])

	if err != nil {
		return err
	}

	*as = AssignmentScore{true, fs[0], fs[1]}

	return nil
}

type AssignmentPoints struct {
	Points, PossiblePoints float64
}

func (ap *AssignmentPoints) UnmarshalXMLAttr(attr xml.Attr) error {
	const pointsRegex = "([\\d\\.]+)\\/([\\d\\.]+)"

	r, err := regexp.Compile(pointsRegex)

	if err != nil {
		return err
	}

	points := r.FindStringSubmatch(attr.Value)

	if len(points) != 3 {
		return fmt.Errorf("Expected points attribute in format `x/y`, received %s and parsed %d numbers", attr.Value, len(points))
	}

	fs, err := stringsToFloats(points[1:])

	if err != nil {
		return err
	}

	*ap = AssignmentPoints{fs[0], fs[1]}

	return nil
}

func stringsToFloats(strs []string) ([]float64, error) {
	fs := make([]float64, 0, len(strs))

	for _, s := range strs {
		f, err := strconv.ParseFloat(s, 64)

		if err != nil {
			return nil, err
		}

		fs = append(fs, f)
	}

	return fs, nil
}
