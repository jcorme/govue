package govue

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
	Assignment                                  *Assignment
	ScoreIncrease, PossibleScoreChange          bool
	PreviousAssignmentScore, NewAssignmentScore float64
	PreviousPossibleScore, NewPossibleScore     float64
}

func CalcChangeset(a *Gradebook, b *Gradebook) *Changeset {
	aMap, bMap := coursesAsMap(a.Courses, b.Courses)
	cs := &Changeset{
		a:    a,
		b:    b,
		aMap: aMap,
		bMap: bMap,
	}

	cs.diffCourseSets()
	cs.diffCourseAssignments()

	return cs
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

			cswitch := &CourseSwitch{ac, c, ac.Period, c.Period}

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
		cc := &CourseChange{Course: ac}

		for i, am := range ac.Marks {
			bm := bc.Marks[i]
			notFoundAAssignments := make(map[string]*Assignment)
			notFoundBAssignments := make(map[string]*Assignment)

			for k, a := range am.Assignments {
				b := bm.Assignments[k]

				if a.GradebookID == b.GradebookID {
					cc.diffAssignments(a, b)

					continue
				}

				notFoundAAssignments[a.GradebookID] = a
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
		}

		if cc.GradeChange != nil {
			cs.CourseChanges = append(cs.CourseChanges, cc)
		}
	}
}

func (cc *CourseChange) diffAssignments(a, b *Assignment) {
	if a.Score.Score == b.Score.Score {
		return
	}

	scoreIncrease := (b.Score.Score - a.Score.Score) > 0
	possibleScoreChange := (b.Score.PossibleScore - a.Score.PossibleScore) != 0

	ca := &CourseAssignmentChange{
		b,
		scoreIncrease,
		possibleScoreChange,
		a.Score.Score,
		b.Score.Score,
		a.Score.PossibleScore,
		b.Score.PossibleScore,
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