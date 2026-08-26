package zentao

// Department represents a ZenTao department.
type Department struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Parent   string `json:"parent"`
	Path     string `json:"path"`
	Grade    string `json:"grade"`
	Order    string `json:"order"`
	Position string `json:"position"`
	Function string `json:"function"`
	Manager  string `json:"manager"`
}

// User represents a ZenTao user.
type User struct {
	ID       string `json:"id"`
	Dept     string `json:"dept"`
	Account  string `json:"account"`
	Realname string `json:"realname"`
	Role     string `json:"role"`
	Email    string `json:"email"`
	Gender   string `json:"gender"`
	Mobile   string `json:"mobile"`
	Phone    string `json:"phone"`
	QQ       string `json:"qq"`
	Address  string `json:"address"`
	Zipcode  string `json:"zipcode"`
	Join     string `json:"join"`
	Visits   string `json:"visits"`
	IP       string `json:"ip"`
	Last     string `json:"last"`
	Fails    string `json:"fails"`
	Locked   string `json:"locked"`
	Ranzhi   string `json:"ranzhi"`
	Deleted  string `json:"deleted"`
}

// Product represents a ZenTao product.
type Product struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Code           string `json:"code"`
	Line           string `json:"line"`
	Type           string `json:"type"`
	Status         string `json:"status"`
	Desc           string `json:"desc"`
	PO             string `json:"PO"`
	QD             string `json:"QD"`
	RD             string `json:"RD"`
	ACL            string `json:"acl"`
	Whitelist      string `json:"whitelist"`
	CreatedBy      string `json:"createdBy"`
	CreatedDate    string `json:"createdDate"`
	CreatedVersion string `json:"createdVersion"`
	Order          string `json:"order"`
	Deleted        string `json:"deleted"`
}

// Project represents a ZenTao project / execution.
type Project struct {
	ID            string `json:"id"`
	IsCat         string `json:"isCat"`
	CatID         string `json:"catID"`
	Type          string `json:"type"`
	Parent        string `json:"parent"`
	Name          string `json:"name"`
	Code          string `json:"code"`
	Begin         string `json:"begin"`
	End           string `json:"end"`
	Days          string `json:"days"`
	Status        string `json:"status"`
	Statge        string `json:"statge"`
	Pri           string `json:"pri"`
	Desc          string `json:"desc"`
	OpenedBy      string `json:"openedBy"`
	OpenedDate    string `json:"openedDate"`
	OpenedVersion string `json:"openedVersion"`
	ClosedBy      string `json:"closedBy"`
	ClosedDate    string `json:"closedDate"`
	CanceledBy    string `json:"canceledBy"`
	CanceledDate  string `json:"canceledDate"`
	PO            string `json:"PO"`
	PM            string `json:"PM"`
	QD            string `json:"QD"`
	RD            string `json:"RD"`
	Team          string `json:"team"`
	ACL           string `json:"acl"`
	Whitelist     string `json:"whitelist"`
	Order         string `json:"order"`
	Deleted       string `json:"deleted"`
}

// Task represents a ZenTao task.
type Task struct {
	ID             string `json:"id"`
	Parent         string `json:"parent"`
	Project        string `json:"project"`
	Module         string `json:"module"`
	Story          string `json:"story"`
	StoryVersion   string `json:"storyVersion"`
	FromBug        string `json:"fromBug"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	Pri            string `json:"pri"`
	Estimate       string `json:"estimate"`
	Consumed       string `json:"consumed"`
	Left           string `json:"left"`
	Deadline       string `json:"deadline"`
	Status         string `json:"status"`
	Color          string `json:"color"`
	Mailto         string `json:"mailto"`
	Desc           string `json:"desc"`
	OpenedBy       string `json:"openedBy"`
	OpenedDate     string `json:"openedDate"`
	AssignedTo     string `json:"assignedTo"`
	AssignedDate   string `json:"assignedDate"`
	EstStarted     string `json:"estStarted"`
	RealStarted    string `json:"realStarted"`
	FinishedBy     string `json:"finishedBy"`
	FinishedDate   string `json:"finishedDate"`
	FinishedList   string `json:"finishedList"`
	CanceledBy     string `json:"canceledBy"`
	CanceledDate   string `json:"canceledDate"`
	ClosedBy       string `json:"closedBy"`
	ClosedDate     string `json:"closedDate"`
	ClosedReason   string `json:"closedReason"`
	LastEditedBy   string `json:"lastEditedBy"`
	LastEditedDate string `json:"lastEditedDate"`
	Deleted        string `json:"deleted"`
}

// Bug represents a ZenTao bug.
type Bug struct {
	ID             string `json:"id"`
	Product        string `json:"product"`
	Branch         string `json:"branch"`
	Module         string `json:"module"`
	Project        string `json:"project"`
	Story          string `json:"story"`
	StoryVersion   string `json:"storyVersion"`
	Task           string `json:"task"`
	ToTask         string `json:"toTask"`
	ToStory        string `json:"toStory"`
	Title          string `json:"title"`
	Keywords       string `json:"keywords"`
	Severity       string `json:"severity"`
	Pri            string `json:"pri"`
	Type           string `json:"type"`
	OS             string `json:"os"`
	Browser        string `json:"browser"`
	Hardware       string `json:"hardware"`
	Found          string `json:"found"`
	Steps          string `json:"steps"`
	Status         string `json:"status"`
	SubStatus      string `json:"subStatus"`
	Color          string `json:"color"`
	Confirmed      string `json:"confirmed"`
	ActivatedCount string `json:"activatedCount"`
	ActivatedDate  string `json:"activatedDate"`
	Mailto         string `json:"mailto"`
	OpenedBy       string `json:"openedBy"`
	OpenedDate     string `json:"openedDate"`
	OpenedBuild    string `json:"openedBuild"`
	AssignedTo     string `json:"assignedTo"`
	AssignedDate   string `json:"assignedDate"`
	Deadline       string `json:"deadline"`
	ResolvedBy     string `json:"resolvedBy"`
	Resolution     string `json:"resolution"`
	ResolvedBuild  string `json:"resolvedBuild"`
	ResolvedDate   string `json:"resolvedDate"`
	ClosedBy       string `json:"closedBy"`
	ClosedDate     string `json:"closedDate"`
	DuplicateBug   string `json:"duplicateBug"`
	LinkBug        string `json:"linkBug"`
	Case           string `json:"case"`
	CaseVersion    string `json:"caseVersion"`
	Result         string `json:"result"`
	Repo           string `json:"repo"`
	Entry          string `json:"entry"`
	Lines          string `json:"lines"`
	V1             string `json:"v1"`
	V2             string `json:"v2"`
	RepoType       string `json:"repoType"`
	Testtask       string `json:"testtask"`
	LastEditedBy   string `json:"lastEditedBy"`
	LastEditedDate string `json:"lastEditedDate"`
	Deleted        string `json:"deleted"`
}
