package govue

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// A Gradebook holds a student's courses, including their grades and assignments in
// those courses, and their school's reporting periods ((mid-)terms, semesters, etc...).
type Gradebook struct {
	XMLName xml.Name `xml:"Gradebook"`

	// GradingPeriod holds all the grading periods of the student's school.
	// If the school uses a semester schedule with (mid-)terms, there will be
	// eight reporting periods.
	GradingPeriods []*GradingPeriod `xml:"ReportingPeriods>ReportPeriod"`

	// CurrentGradingPeriod is the school's current grading period.
	CurrentGradingPeriod *GradingPeriod `xml:"ReportingPeriod"`

	// Courses holds all of the student's classes, which should be ordered by
	// the class's period in the student's schedule.
	Courses []*Course `xml:"Courses>Course"`
}

// A GradingPeriod represents one grading period for a school.
// This usually means each GradingPeriod is a half of a quarter (term).
type GradingPeriod struct {
	// Index is a zero-based index representing the GradingPeriod 's place in
	// the GradingPeriods set.
	Index int `xml:",attr"`

	// Name is the name of the grading period.
	Name string `xml:"GradePeriod,attr"`

	// StartDate is when the grading period begins.
	StartDate GradebookDate `xml:",attr"`

	// EndDate is when the grading period ends.
	EndDate GradebookDate `xml:",attr"`
}

// A Course represents one of a student's classes.
type Course struct {
	// Period is the period of the day in which the student has this class.
	Period int `xml:",attr"`

	// ID holds identification information for this class, which includes
	// its Name and ID within the school's/StudentVUE's systems.
	ID CourseID `xml:"Title,attr"`

	// Room is the room number of this class inside the school.
	Room string `xml:",attr"`

	// Teacher is the name of the instructor of this class.
	Teacher string `xml:"Staff,attr"`

	// TeacherEmail is the email of this class's instructor.
	TeacherEmail string `xml:"StaffEMail,attr"`

	// Marks holds the student's grading, including assignments, information
	//for each grading period.
	Marks []*CourseMark `xml:"Marks>Mark"`

	// CurrentMark points to the mark for the current grading period.
	CurrentMark *CourseMark `xml:"-"`
}

// A CourseMark holds a student's grades and assignments for a single grading period.
type CourseMark struct {
	// Name is the name of the grading period.
	Name string `xml:"MarkName,attr"`

	// LetterGrade is the student's raw (number) grade mapped to a letter.
	// Usually mapped as such:
	//		90+ -> A
	//		80+ -> B
	//		70+ -> C
	//		60+ -> D
	//		Else -> F
	LetterGrade string `xml:"CalculatedScoreString,attr"`

	// RawGradeScore is the student's raw percentage grade for the grading period.
	RawGradeScore float64 `xml:"CalculatedScoreRaw,attr"`

	// GradeSummaries holds the grade summaries for each of the course's weighted categories.
	// For example, if a course weighs Tests and Homework as separate categories, those will
	// be contained here with information including the category's weighted percentage and
	// letter grade.
	GradeSummaries []*AssignmentGradeCalc `xml:"GradeCalculationSummary>AssignmentGradeCalc"`

	// Assignments holds all of the course's assignments for the grading period.
	Assignments []*Assignment `xml:"Assignments>Assignment"`
}

// AssignmentGradeCalc represents one of a course's weighted categories.
// This may include Tests, Homework, Class Work, etc... These are created and decided
// by the course's instructor.
type AssignmentGradeCalc struct {
	// Type is the name of the weighted category.
	Type string `xml:",attr"`

	// Weight is the weight of the category of the student's grade in percent.
	Weight Percentage `xml:",attr"`

	// Points is the number of points earned by the student in this category.
	Points float64 `xml:",attr"`

	// PointsPossible is the number of points that can be earned by the student in this category.
	PointsPossible float64 `xml:",attr"`

	// WeightedPercentage is the impact of this category on the student's overall
	// grade in percent.
	WeightedPercentage Percentage `xml:"WeightedPct,attr"`

	// LetterGrade is the student's raw (number) grade mapped to a letter for this category.
	LetterGrade string `xml:"CalculatedMark,attr"`
}

// An Assignment is a single entry into a course's gradebook by an instructor.
type Assignment struct {
	// GradebookID is the internal ID given to the assignment by StudentVUE.
	GradebookID string `xml:",attr"`

	// Name is the name of the assignment entry.
	Name string `xml:"Measure,attr"`

	// Type is the weighted category to which the assignment belongs.
	Type string `xml:",attr"`

	// Date is the date on which the assignment was entered into the gradebook
	// by the instructor.
	Date GradebookDate `xml:",attr"`

	// DueDate is the date on which the assignment was due for the student.
	DueDate GradebookDate `xml:",attr"`

	// Score holds the student's earned and possible raw score of the assignment.
	Score AssignmentScore `xml:",attr"`

	// ScoreType is the kind of score represented by the Score field; e.g. `Raw Score.`
	ScoreType string `xml:",attr"`

	// Points is the number of points for which the assignment actually counted.
	// For example, an assignment score may be out of 20, but the instructor may
	// choose to scale it down to only be worth 5 points (towards calculating the
	// student's grade) or scale it up to be worth 80 points.
	Points AssignmentPoints `xml:",attr"`

	// Notes is any comment added by the instructor on the assignment entry.
	Notes string `xml:",attr"`
}

// A CourseID holds the identification information for a class.
type CourseID struct {
	// ID is the school's/StudentVUE's internal ID for the class.
	ID string

	// Name is the official name of the class.
	Name string
}

func (cid *CourseID) UnmarshalXMLAttr(attr xml.Attr) error {
	const nameRegex = "(.+?)\\s*(\\(.+?\\))"

	r, err := regexp.Compile(nameRegex)

	if err != nil {
		return err
	}

	name := r.FindStringSubmatch(attr.Value)

	if len(name) != 3 {
		return fmt.Errorf("Expected course name attribute in format `Course (ID)`, received %s and found %d regex matches", attr.Value, len(name)-1)
	}

	var (
		id    string
		cname string
	)

	for i, g := range name[1:] {
		if rune(g[0]) == '(' && rune(g[len(g)-1]) == ')' {
			id = g[1 : len(g)-1]
			cname = name[i]

			break
		}
	}

	if id == "" {
		return fmt.Errorf("Unable to parse out course name and ID from `%s`, got `%v`", attr.Value, name)
	}

	*cid = CourseID{id, cname}

	return nil
}

// A Percentage is a floating-point number representing a percentage.
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

// A GradebookDate holds a timestamp parsed from the format of StudentVUE's systems.
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

// An AssignmentScore holds the score information for a single assignment for a student.
type AssignmentScore struct {
	// Graded denotes whether the assignment has been graded or not.
	Graded bool

	// Score is the number of points earned on the assignment by the student.
	Score float64

	// PossibleScore is the number of points that could be earned by the student.
	PossibleScore float64
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

// An AssignmentPoints holds an assignment's actual score for a student.
// The different between AssignmentScore and AssignmentPoints is that an assignment's
// score is a raw score, while the points may be either the score scaled up or down
// to affect the student's actual grade differently.
type AssignmentPoints struct {
	// Points is the number of points that the student received on the assignment.
	Points float64

	// PossiblePoints is the number of points the student could receive on the assignment.
	PossiblePoints float64
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

// CurrentGradingPeriodIndex returns the position of the current grading
// period in the ReportingPeriods field of a Gradebook.
func (g *Gradebook) CurrentGradingPeriodIndex() int {
	gradePeriod := g.CurrentGradingPeriod.Name

	for _, p := range g.GradingPeriods {
		if p.Name == gradePeriod {
			return p.Index / 2
		}
	}

	return 0
}
