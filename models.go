package main

type Company struct {
	Name string
	INN  string
}

type Uni struct {
	Name string
}

type Student struct {
	UniName string
	Course  string
	FIO     string
}

type DiplomaTopic struct {
	ID          int
	Name        string
	Description string
	CompanyINN  string
	TargetUni   string
	Status      string
}

type Internship struct {
	ID           int
	Name         string
	Requirements string
	Places       int
	CompanyINN   string
	TargetUni    string
	Status       string
}

type Application struct {
	ID         int
	StudentID  int64
	Type       string
	ItemID     int
	Status     string
	CompanyINN string
	UniName    string
}

var (
	nextTopicID           int = 1
	nextInternID          int = 1
	nextAppID             int = 1
	companies                 = make(map[string]Company)
	unis                      = make(map[string]Uni)
	userRoles                 = make(map[int64]string)
	students                  = make(map[int64]Student)
	diplomaTopics             = []DiplomaTopic{}
	specificDiplomaTopics     = make(map[string][]DiplomaTopic)
	uniDiplomaPools           = make(map[string][]int)
	internshipOffers          = []Internship{}
	specificInternships       = make(map[string][]Internship)
	uniInternshipPools        = make(map[string][]int)
	applications              = []Application{}
	userStates                = make(map[int64]string)
	userTempData              = make(map[int64]map[string]string)
)
