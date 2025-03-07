package internal

type System interface {
	GetVolumePaths() ([]string, error)
	ListFilePaths(path string) ([]string, error)
	MonitorVolumes() (addedChan <-chan string, removedChan <-chan string, err error)
}
