package certworker

type Uploader interface {
	Add(result *Result)
}
