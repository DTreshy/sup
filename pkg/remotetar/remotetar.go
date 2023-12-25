package remotetar

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Copying dirs/files over SSH using TAR.
// tar -C . -cvzf - $SRC | ssh $HOST "tar -C $DST -xvzf -"

// RemoteTarCommand returns command to be run on remote SSH host
// to properly receive the created TAR stream.
// TODO: Check for relative directory.
func RemoteTarCommand(dir string) string {
	return fmt.Sprintf("tar -C \"%q\" -xzf -", dir)
}

func LocalTarCmdArgs(path, exclude string) []string {
	args := []string{}

	// Added pattens to exclude from tar compress
	excludes := strings.Split(exclude, ",")

	for _, exclude := range excludes {
		trimmed := strings.TrimSpace(exclude)
		if trimmed != "" {
			args = append(args, `--exclude=`+trimmed)
		}
	}

	args = append(args, "-C", ".", "-czf", "-", path)

	return args
}

// NewTarStreamReader creates a tar stream reader from a local path.
// TODO: Refactor. Use "archive/tar" instead.
func NewTarStreamReader(cwd, path, exclude string) (io.Reader, error) {
	localTarCmdArgs := LocalTarCmdArgs(path, exclude)
	cmd := exec.Command("tar", localTarCmdArgs...)
	cmd.Dir = cwd

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Join(err, errors.New("tar: stdout pipe failed"))
	}

	if err := cmd.Start(); err != nil {
		return nil, errors.Join(err, errors.New("tar: starting cmd failed"))
	}

	return stdout, nil
}
