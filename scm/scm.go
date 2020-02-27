package scm

type SCM interface {
	Pull() error
	LocalPath() string
}

type scm struct {
	remoteUrl string
	localPath string
}

func makeLocalPath(srcUrl string) (string, error) {
	return "", nil
}

func New(srcUrl string) (SCM, error) {
	localPath, err := makeLocalPath(srcUrl)
	if err != nil {
		return nil, err
	}

	return &scm{
		remoteUrl: srcUrl,
		localPath: localPath,
	}, nil
}

func (scm *scm) Pull() error {
	return nil
}

func (scm *scm) LocalPath() string {
	return scm.localPath
}
