package govue

import (
	"encoding/xml"
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
	Type               string  `xml:",attr"`
	Weight             string  `xml:",attr"`
	Points             float64 `xml:",attr"`
	PointsPossible     float64 `xml:",attr"`
	WeightedPercentage string  `xml:"WeightedPct,attr"`
	LetterGrade        string  `xml:"CalculatedMark,attr"`
}

type Assignment struct {
	GradebookID string        `xml:",attr"`
	Name        string        `xml:"Measure,attr"`
	Type        string        `xml:",attr"`
	Date        GradebookDate `xml:",attr"`
	DueDate     GradebookDate `xml:",attr"`
	Score       string        `xml:",attr"`
	ScoreType   string        `xml:",attr"`
	Points      string        `xml:",attr"`
	Notes       string        `xml:",attr"`
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
