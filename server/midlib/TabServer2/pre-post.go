package TabServer2

type PrePostFlagType int

const (
	PrePostRVUpdatedSuccess PrePostFlagType = 1
	PrePostRVUpdatedFail    PrePostFlagType = 2
	PrePostNoAction         PrePostFlagType = 3
	PrePostFatalStatus500   PrePostFlagType = 4
)
