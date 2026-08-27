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
	Execution      string `json:"execution"`
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
	Keywords       string `json:"keywords"`
	Mailto         string `json:"mailto"`
	Desc           string `json:"desc"`
	OpenedBy       string `json:"openedBy"`
	OpenedDate     string `json:"openedDate"`
	AssignedTo     string `json:"assignedTo"`
	AssignedDate   string `json:"assignedDate"`
	AssignedBy     string `json:"assignedBy"`
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
	Version        string `json:"version"`
	PriOrder       string `json:"priOrder"`
	Deleted        string `json:"deleted"`
}

// Story represents a ZenTao story/requirement.
type Story struct {
	ID             string `json:"id"`
	Product        string `json:"product"`
	Branch         string `json:"branch"`
	Module         string `json:"module"`
	Plan           string `json:"plan"`
	Source         string `json:"source"`
	SourceNote     string `json:"sourceNote"`
	FromBug        string `json:"fromBug"`
	Title          string `json:"title"`
	Keywords       string `json:"keywords"`
	Type           string `json:"type"`
	Category       string `json:"category"`
	Pri            string `json:"pri"`
	Estimate       string `json:"estimate"`
	Status         string `json:"status"`
	SubStatus      string `json:"subStatus"`
	Color          string `json:"color"`
	Stage          string `json:"stage"`
	StagedBy       string `json:"stagedBy"`
	Mailto         string `json:"mailto"`
	OpenedBy       string `json:"openedBy"`
	OpenedDate     string `json:"openedDate"`
	AssignedTo     string `json:"assignedTo"`
	AssignedDate   string `json:"assignedDate"`
	LastEditedBy   string `json:"lastEditedBy"`
	LastEditedDate string `json:"lastEditedDate"`
	ReviewedBy     string `json:"reviewedBy"`
	ReviewedDate   string `json:"reviewedDate"`
	ClosedBy       string `json:"closedBy"`
	ClosedDate     string `json:"closedDate"`
	ClosedReason   string `json:"closedReason"`
	ToBug          string `json:"toBug"`
	Spec           string `json:"spec"`
	Verify         string `json:"verify"`
	Version        string `json:"version"`
	Deleted        string `json:"deleted"`
}

// Todo represents a ZenTao todo item.
type Todo struct {
	ID           string `json:"id"`
	Account      string `json:"account"`
	Date         string `json:"date"`
	Begin        string `json:"begin"`
	End          string `json:"end"`
	Type         string `json:"type"`
	IDValue      string `json:"idvalue"`
	Pri          string `json:"pri"`
	Name         string `json:"name"`
	Desc         string `json:"desc"`
	Status       string `json:"status"`
	Private      string `json:"private"`
	AssignedBy   string `json:"assignedBy"`
	AssignedDate string `json:"assignedDate"`
	FinishedBy   string `json:"finishedBy"`
	FinishedDate string `json:"finishedDate"`
	ClosedBy     string `json:"closedBy"`
	ClosedDate   string `json:"closedDate"`
	Deleted      string `json:"deleted"`
}
