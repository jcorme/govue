package govue

import (
	"fmt"
	"strings"
)

type Changeset struct {
	a, b            *Gradebook
	aMap, bMap      map[int]*Course
	CourseSwitches  []*CourseSwitch
	CourseAdditions []*Course
	CourseDrops     []*Course
	CourseChanges   []*CourseChange
}

type CourseSwitch struct {
	Before, After             *Course
	BeforePeriod, AfterPeriod int
}

type CourseChange struct {
	Course              *Course
	GradeChange         *CourseGradeChange
	AssignmentChanges   []*CourseAssignmentChange
	AssignmentAdditions []*Assignment
	AssignmentRemovals  []*Assignment
}

type CourseGradeChange struct {
	GradeIncrease                       bool
	PreviousLetterGrade, NewLetterGrade string
	PreviousGradePct, NewGradePct       float64
	DeltaPct                            float64
}

type CourseAssignmentChange struct {
	Before, After                          *Assignment
	NameChange                             bool
	ScoreChange, PointsChange              bool
	ScoreIncrease, PossibleScoreChange     bool
	PointsIncrease, PossiblePointsIncrease bool
	PreviousScore, NewScore                *AssignmentScore
	PreviousPoints, NewPoints              *AssignmentPoints
}

type SemesterMismatchError struct {
	aSemester, bSemester int
}

func (s SemesterMismatchError) Error() string {
	return fmt.Sprintf("The current grading periods of the two Gradebooks do not match: one is in semester %d and the other is in semester %d", s.aSemester, s.bSemester)
}

func CalcChangeset(a *Gradebook, b *Gradebook) (*Changeset, error) {
	if as, bs, ok := gradebookSemestersMatch(a, b); !ok {
		return nil, SemesterMismatchError{
			aSemester: as,
			bSemester: bs,
		}
	}

	aMap, bMap := coursesAsMap(a.Courses, b.Courses)
	cs := &Changeset{
		a:    a,
		b:    b,
		aMap: aMap,
		bMap: bMap,
	}

	cs.diffCourseSets()
	cs.diffCourseAssignments()

	return cs, nil
}

func coursesAsMap(acs, bcs []*Course) (acsMap, bcsMap map[int]*Course) {
	acsMap, bcsMap = make(map[int]*Course), make(map[int]*Course)

	for _, ac := range acs {
		acsMap[ac.Period] = ac
	}

	for _, bc := range bcs {
		bcsMap[bc.Period] = bc
	}

	return acsMap, bcsMap
}

func (cs *Changeset) diffCourseSets() {
	aMap, bMap := cs.aMap, cs.bMap
	normalizedBMap := make(map[int]*Course)

	findCourseSwitch := func(p int, ac *Course) bool {
		c, k, found := findCourse(bMap, ac.ID.ID)

		if found {
			normalizedBMap[p] = c

			cswitch := &CourseSwitch{
				Before:       ac,
				After:        c,
				BeforePeriod: ac.Period,
				AfterPeriod:  c.Period,
			}

			cs.CourseSwitches = append(cs.CourseSwitches, cswitch)

			delete(cs.bMap, k)

			return true
		}

		return false
	}

	for p, ac := range aMap {
		bc, ok := bMap[p]

		if ok {
			if ac.ID.ID == bc.ID.ID {
				normalizedBMap[p] = bc

				delete(cs.bMap, p)

				continue
			} else {
				_ = findCourseSwitch(p, ac)
			}
		} else {
			found := findCourseSwitch(p, ac)

			if !found {
				cs.CourseDrops = append(cs.CourseDrops, ac)

				delete(cs.aMap, p)
			}
		}
	}

	for p, bc := range bMap {
		c, k, found := findCourse(aMap, bc.ID.ID)

		if found {
			normalizedBMap[k] = c
		} else {
			cs.CourseAdditions = append(cs.CourseAdditions, bc)
		}

		delete(cs.bMap, p)
	}

	cs.bMap = normalizedBMap
}

func (cs *Changeset) diffCourseAssignments() {
	aMap, bMap := cs.aMap, cs.bMap

	for p, ac := range aMap {
		bc := bMap[p]

		am := ac.CurrentMark
		bm := bc.CurrentMark
		cc := &CourseChange{Course: ac}

		bAssignments := make([]*Assignment, len(bm.Assignments))
		copy(bAssignments, bm.Assignments)

		notFoundAAssignments := make(map[string]*Assignment)
		notFoundBAssignments := make(map[string]*Assignment)

		aCount := len(am.Assignments)
		bCount := len(bAssignments)

		for k, a := range am.Assignments {
			if bCount < aCount && k >= bCount {
				notFoundAAssignments[a.GradebookID] = a

				continue
			}

			b := bAssignments[k]
			bAssignments[k] = nil

			if a.GradebookID == b.GradebookID {
				cc.diffAssignments(a, b)

				continue
			}

			notFoundAAssignments[a.GradebookID] = a
			notFoundBAssignments[b.GradebookID] = b
		}

		for k, b := range bAssignments {
			if b == nil {
				continue
			}

			gid := b.GradebookID

			if a, ok := notFoundAAssignments[gid]; ok {
				cc.diffAssignments(a, b)

				delete(notFoundAAssignments, gid)

				bAssignments[k] = nil

				continue
			}

			notFoundBAssignments[b.GradebookID] = b
		}

		for gid, a := range notFoundAAssignments {
			if b, ok := notFoundBAssignments[gid]; ok {
				cc.diffAssignments(a, b)

				delete(notFoundAAssignments, gid)
				delete(notFoundBAssignments, gid)

				continue
			}

			cc.AssignmentRemovals = append(cc.AssignmentRemovals, a)
		}

		for _, b := range notFoundBAssignments {
			cc.AssignmentAdditions = append(cc.AssignmentAdditions, b)
		}

		if ps, ns := am.RawGradeScore, bm.RawGradeScore; (ns - ps) != 0 {
			change := ns - ps

			cc.GradeChange = &CourseGradeChange{
				DeltaPct:            change,
				GradeIncrease:       change > 0,
				NewGradePct:         ns,
				NewLetterGrade:      bm.LetterGrade,
				PreviousGradePct:    ps,
				PreviousLetterGrade: am.LetterGrade,
			}
		}

		changed := len(cc.AssignmentAdditions) | len(cc.AssignmentChanges) | len(cc.AssignmentRemovals)

		if cc.GradeChange != nil || changed > 0 {
			cs.CourseChanges = append(cs.CourseChanges, cc)
		}
	}
}

func (cc *CourseChange) diffAssignments(a, b *Assignment) {
	nameChange := a.Name != b.Name

	scoreChange := (b.Score.Score - a.Score.Score) != 0
	possibleScoreChange := (b.Score.PossibleScore - a.Score.PossibleScore) != 0

	pointsChange := (b.Points.Points - a.Points.Points) != 0
	possiblePointsChange := (b.Points.PossiblePoints - a.Points.PossiblePoints) != 0

	if !nameChange && !scoreChange && !possibleScoreChange && !pointsChange && !possiblePointsChange {
		return
	}

	scoreIncrease := (b.Score.Score - a.Score.Score) > 0
	pointsIncrease := (b.Points.Points - a.Points.Points) > 0

	ca := &CourseAssignmentChange{
		Before:                 a,
		After:                  b,
		NameChange:             nameChange,
		ScoreChange:            scoreChange,
		PointsChange:           pointsChange,
		ScoreIncrease:          scoreIncrease,
		PossibleScoreChange:    possibleScoreChange,
		PointsIncrease:         pointsIncrease,
		PossiblePointsIncrease: possiblePointsChange,
		PreviousScore:          &a.Score,
		NewScore:               &b.Score,
		PreviousPoints:         &a.Points,
		NewPoints:              &b.Points,
	}

	cc.AssignmentChanges = append(cc.AssignmentChanges, ca)
}

func findCourse(courses map[int]*Course, id string) (*Course, int, bool) {
	for k, c := range courses {
		if c.ID.ID == id {
			return c, k, true
		}
	}

	return nil, 0, false
}

func gradebookSemestersMatch(a *Gradebook, b *Gradebook) (int, int, bool) {
	aGradePeriod := a.CurrentGradingPeriod.Name
	bGradePeriod := b.CurrentGradingPeriod.Name

	if strings.Contains(aGradePeriod, "Q1") || strings.Contains(aGradePeriod, "Q2") {
		if strings.Contains(bGradePeriod, "Q3") || strings.Contains(bGradePeriod, "Q4") {
			return 1, 2, false
		}
	} else if strings.Contains(aGradePeriod, "Q3") || strings.Contains(aGradePeriod, "Q4") {
		if strings.Contains(bGradePeriod, "Q1") || strings.Contains(bGradePeriod, "Q2") {
			return 2, 1, false
		}
	}

	return 0, 0, true
}
