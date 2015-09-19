package govue

type Student struct {
	ID     int
	Name   string `xml:"ChildName"`
	School string `xml:"OrganizationName"`
	Grade  int
	Events []*Event `xml:"Events>Event"`
}

type Event struct {
	Date        string `xml:"EventDate"`
	Description string `xml:"EventDescription"`
	Module      string
}
