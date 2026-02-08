package provider

import (
	"bufio"
	"context"
	"io"
)

func ReadLines(ctx context.Context, reader io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(reader)

	//add context checking to exit earlier
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			lines = append(lines, scanner.Text())
		}
	}

	return lines, scanner.Err()
}

func WriteLines(ctx context.Context, w io.Writer, lines []string) error {
	bw := bufio.NewWriter(w)

	defer bw.Flush()

	for _, line := range lines {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if _, err := bw.WriteString(line + "\n"); err != nil {
				return err
			}
		}
	}

	return nil
}
