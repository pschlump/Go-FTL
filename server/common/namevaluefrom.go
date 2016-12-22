package common

// rv += fmt.Sprintf("%35s  %7d %s\n", vv.Name, vv.UsedAt, vv.UsedFile)

type NameValueFrom struct {
	Name      string
	Value     string
	From      string
	ParamType string
	UsedAt    int    // if 0, not used
	UsedFile  string // file where used, last used.
}
