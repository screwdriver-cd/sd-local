package scm

type SCM interface {
}

type scm struct {
	remoteUrl string
	localPath string
}

func makeLocalPath(srcUrl string) (string, error) {
	return "", nil
}

func New(srcUrl) *SCM {
	localPath, err := makeLocalPath(srcUrl)
	&scm{
		remoteUrl: srcUrl,
		localPath: localPath,
	}
}
