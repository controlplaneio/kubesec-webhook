package webhook

// NOTE(ludo): not making this a CLI parameter
// since we will move away from the API in favor
// of kubesec v2 library.
const (
	scanURL     = "https://v2.kubesec.io"
	scanTimeOut = 120
)
