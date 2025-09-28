package montoya

import "io"

// IniFile contains a parsed Ini file
type IniFile struct {
	// The start of the file
	Head IniLine
	// The end of the file
	Tail IniLine

	readLine IniLine
	done     bool
}

// Reset all reader state, prepare to be Read again
func (f *IniFile) Reset() {
	for line := f.Head; line != nil; line = line.Next() {
		line.Reset()
	}
	f.readLine = nil
	f.done = false
}

// Read file contents to a slice
func (f *IniFile) Read(dst []byte) (int, error) {
	if f.readLine == nil && !f.done {
		f.readLine = f.Head
	}

	if f.readLine == nil {
		return 0, io.EOF // empty file or Read after finish
	}

	totalWritten := 0
	for totalWritten < len(dst) && f.readLine != nil {
		n, err := f.readLine.Read(dst[totalWritten:])
		totalWritten += n

		if err == io.EOF {
			if f.readLine == f.Tail {
				// reached end of last line
				f.readLine = nil
				f.done = true
				if totalWritten == 0 {
					return 0, io.EOF
				}
				return totalWritten, nil
			}
			// advance to next line
			f.readLine = f.readLine.Next()
			continue
		}
		if err != nil {
			return totalWritten, err
		}

		// line produced data, break to return immediately
		if n > 0 {
			break
		}
	}

	return totalWritten, nil
}
