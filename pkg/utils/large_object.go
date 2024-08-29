package utils

import (
	"fmt"
	"io"
)

func CopyBytes(receiver io.Writer, root io.Reader) (int64, error) {
	var written int64
	buffer := make([]byte, 4096)

	for {
		n, err := root.Read(buffer)
		if err != nil && err != io.EOF {
			return 0, fmt.Errorf("failed to read data: %v", err)
		}
		if n == 0 {
			break
		}
		_, err = receiver.Write(buffer[:n])
		if err != nil {
			return 0, fmt.Errorf("failed to write data: %v", err)
		}

		written += int64(n)
	}

	return written, nil
}
