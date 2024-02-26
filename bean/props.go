package bean

type Root struct {
	Packages []Package `json:"packages"`
	Desc     string    `json:"desc"`
	IsAdmin  string    `json:"isAdmin"`
}

type Package struct {
	Key         string        `json:"key"`
	Type        string        `json:"type"`
	Logotype    string        `json:"logotype"`
	Desc        string        `json:"desc"`
	CheckedList []interface{} `json:"checkedList"`
	InputValues []InputValue  `json:"inputValues"`
	CheckedAll  bool          `json:"checkedAll"`
	Backup      PackageStep   `json:"backup"`
	Update      PackageStep   `json:"update"`
	Rollback    PackageStep   `json:"rollback"`
	FileLists   []FileList    `json:"fileLists"`
}

type InputValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PackageStep struct {
	Active bool   `json:"active"`
	Shell  string `json:"shell"`
}

type FileList struct {
	Uid              string        `json:"uid"`
	LastModified     int64         `json:"lastModified"`
	LastModifiedDate string        `json:"lastModifiedDate"`
	Name             string        `json:"name"`
	Size             int           `json:"size"`
	Type             string        `json:"type"`
	Percent          int           `json:"percent"`
	OriginFileObj    OriginFileObj `json:"originFileObj"`
}

type OriginFileObj struct {
	Uid string `json:"uid"`
}
