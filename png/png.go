package png

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var pngSig = []byte{137, 80, 78, 71, 13, 10, 26, 10}

// File structure hold os.File handle, file path and separate PNG chunks in a slice.
type File struct {
	file     *os.File
	filePath string
	Chunks   []chunkHeader
}

type chunkHeader struct {
	file       *os.File
	Type       string
	Len        int64
	DataOffset int64
}

func (png *File) String() string {
	if png == nil {
		return fmt.Sprintf("%v", nil)
	}

	s := string(png.filePath) + "\n"
	for _, v := range png.Chunks {
		s += v.String() + "\n"
	}
	return s
}

// Close closes the PNGFile.file, rendering it unusable for I/O.
func (png *File) Close() error {
	if png == nil {
		return errors.New("PNG context is nil")
	}
	return png.file.Close()
}

func (c *chunkHeader) String() string {
	if c == nil {
		return fmt.Sprintf("%v", nil)
	}
	return fmt.Sprintf("%q: len: %v, offset: %v", c.Type, c.Len, c.DataOffset)
}

// WriteTo writes chunk into io.Writer.
func (c *chunkHeader) WriteTo(f io.Writer) error {
	_, err := c.file.Seek(c.DataOffset-8, 0)
	if err != nil {
		return err
	}
	return CopyFile(c.file, f, c.Len+12)
}

// Open opens the named file for reading and returns png.File structure if it's a valid PNG.
// If png.File is returned with an error that means that file was not fully read and data might be corrupted.
// But everything that was read should be correct.
func Open(filePath string) (*File, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	err = checkSignature(f)
	if err != nil {
		return nil, err
	}

	p := &File{
		file:     f,
		filePath: filePath,
		Chunks:   []chunkHeader{},
	}

	err = parseChunk(p)
	if err != nil {
		return nil, err
	}
	if p.Chunks[0].Type != "IHDR" || p.Chunks[0].Len != 13 {
		return nil, fmt.Errorf("incorrect IHDR chunk: %q %v", p.Chunks[0].Type, p.Chunks[0].Len)
	}

	for {
		err = parseChunk(p)
		if err != nil {
			return p, err
		}

		if p.Chunks[len(p.Chunks)-1].Type == "IEND" {
			break
		}
	}

	seekC, err := f.Seek(0, os.SEEK_CUR)
	if err != nil {
		return p, err
	}
	seekE, err := f.Seek(0, os.SEEK_END)
	if err != nil {
		return p, err
	}

	if seekC != seekE {
		return p, fmt.Errorf("data left after the last read PNG chunk. Seek offset %v/%v", seekC, seekE)
	}

	return p, nil
}

// WriteSignatureTo writes PNG header to io.Writer.
func WriteSignatureTo(f io.Writer) error {
	_, err := f.Write(pngSig)
	return err
}

// CopyFile copies up to len bytes from io.Reader to io.Writer.
func CopyFile(src io.Reader, dst io.Writer, len int64) error {
	blockSize := int64(1024)
	b := make([]byte, blockSize)
	for ; len > 0; len -= blockSize {
		if len < blockSize {
			blockSize = len
		}
		n, err := src.Read(b[0:blockSize])
		if err != nil {
			return err
		}
		_, err = dst.Write(b[0:n])
		if err != nil {
			return err
		}
	}
	return nil
}

// checkHeader reads 8 bytes from a file
// and checks if it is a valid PNG header.
func checkSignature(f *os.File) error {
	b := make([]byte, 8)
	n, err := f.Read(b)
	if len(b) != n || err != nil {
		return err
	}
	if string(b) != string(pngSig) {
		return errors.New("file does not containt a PNG signature")
	}
	return nil
}

// readChunkHeader returns *png.chunkHeader
// by reading 8 bytes from a file and treating them as Length and Chunk Type,
// *os.File offset is saved at the postion of Chunk Data start.
func readChunkHeader(f *os.File) (*chunkHeader, error) {
	c := &chunkHeader{file: f}
	// Read chunks length.
	b := make([]byte, 4)
	n, err := f.Read(b)
	if len(b) != n || err != nil {
		return nil, err
	}
	c.Len = int64(b[0]) << 32
	c.Len += int64(b[1]) << 16
	c.Len += int64(b[2]) << 8
	c.Len += int64(b[3])

	// Read chunks type.
	n, err = f.Read(b)
	if len(b) != n || err != nil {
		return nil, err
	}
	c.Type = string(b)

	// Save current file offset.
	c.DataOffset, err = f.Seek(0, os.SEEK_CUR)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// skipChunk offsets current seek to skip chunks data and CRC.
func skipChunk(f *os.File, c *chunkHeader) error {
	_, err := f.Seek(c.Len+4, os.SEEK_CUR) // +4 for CRC
	return err
}

func parseChunk(f *File) error {
	c, err := readChunkHeader(f.file)
	if err != nil {
		return err
	}

	err = skipChunk(f.file, c)
	if err != nil {
		return err
	}

	f.Chunks = append(f.Chunks, *c)

	return nil
}
