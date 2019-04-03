package jpg

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

type Marker byte

//go:generate stringer -type=Marker
// JPEG segment markers.
const (
	SOF0 Marker = 0xc0 + iota // Start Of Frame (Baseline Sequential)
	SOF1                      // Start Of Frame (Extended Sequential)
	SOF2                      // Start Of Frame (Progressive)
	SOF3                      // Start Of Frame (3)
	DHT                       // Define Huffman Table
	SOF5
	SOF6
	SOF7
	JPG // Reserved for JPEG extensions
	SOF9
	SOF10
	SOF11
	DAC // Define arithmetic coding conditioning(s)
	SOF13
	SOF14
	SOF15
)

const (
	RST0 Marker = 0xd0 + iota // ReSTart (0)
	RST1
	RST2
	RST3
	RST4
	RST5
	RST6
	RST7
	SOI // Start Of Image
	EOI // End Of Image
	SOS // Start Of Scan
	DQT // Define Quantization Table
	DNL // Define number of lines
	DRI // Define Restart Interval
	DHP // Define hierarchical progression
	EXP // Expand reference component(s)
)

const (
	APP0 Marker = 0xe0 + iota // JFIF application segment (0)
	APP1
	APP2
	APP3
	APP4
	APP5
	APP6
	APP7
	APP8
	APP9
	APP10
	APP11
	APP12
	APP13
	APP14
	APP15
)

const (
	JPG0 Marker = 0xf0 + iota // Reserved for JPEG extensions
	JPG1
	JPG2
	JPG3
	JPG4
	JPG5
	JPG6
	JPG7
	JPG8
	JPG9
	JPG10
	JPG11
	JPG12
	JPG13
	COM // COMment
)

type File struct {
	file     *os.File
	filePath string
	Chunks   []chunkHeader
}

func (f *File) String() string {
	if f == nil {
		return fmt.Sprintf("%v", nil)
	}

	s := string(f.filePath) + "\n"
	for _, v := range f.Chunks {
		s += v.String() + "\n"
	}
	return s
}

// Close closes the PNGFile.file, rendering it unusable for I/O.
func (f *File) Close() error {
	if f == nil {
		return errors.New("JPG context is nil")
	}
	return f.file.Close()
}

type chunkHeader struct {
	file       *os.File
	Type       Marker
	Len        int64
	DataOffset int64
	ImageData  [2]int64
}

func (c *chunkHeader) String() string {
	if c == nil {
		return fmt.Sprintf("%v", nil)
	}
	return fmt.Sprintf("%q: len: %v, offset: %v, image: %v", c.Type, c.Len, c.DataOffset, c.ImageData)
}

// WriteTo writes chunk into io.Writer.
func (c *chunkHeader) WriteTo(f io.Writer) error {
	_, err := c.file.Seek(c.DataOffset, 0)
	if err != nil {
		return err
	}
	err = CopyFile(c.file, f, c.Len+2)
	if c.Type == SOS {
		err = CopyFile(c.file, f, c.ImageData[1]-c.ImageData[0])
	}
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

func Open(filePath string) (*File, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	j := &File{
		file:     f,
		filePath: filePath,
		Chunks:   []chunkHeader{},
	}

	err = parseChunk(j)
	if err != nil {
		return nil, err
	}
	if j.Chunks[0].Type != SOI {
		return nil, fmt.Errorf("incorrect JPG header: %q", j.Chunks[0].Type)
	}

	for {
		err = parseChunk(j)
		if err != nil {
			return j, err
		}

		// Store image data after SOS (Start Of Scan) chunk.
		if j.Chunks[len(j.Chunks)-1].Type == SOS {
			j.Chunks[len(j.Chunks)-1].ImageData, err = parseImageData(j.file)
			if err != nil {
				return nil, err
			}
		}

		if j.Chunks[len(j.Chunks)-1].Type == EOI {
			break
		}
	}

	seekC, err := f.Seek(0, os.SEEK_CUR)
	if err != nil {
		return j, err
	}
	seekE, err := f.Seek(0, os.SEEK_END)
	if err != nil {
		return j, err
	}

	if seekC != seekE {
		return j, fmt.Errorf("data left after the last read JPG chunk. Seek offset %v/%v", seekC, seekE)
	}

	return j, nil
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

// readChunkHeader returns *chunkHeader.
func readChunkHeader(f *os.File) (*chunkHeader, error) {
	c := &chunkHeader{file: f}
	b := make([]byte, 1)

	// Read first byte.
	n, err := f.Read(b)
	if len(b) != n || err != nil {
		return nil, err
	}
	if b[0] != 0xff {
		fmt.Println(b[0])
		return nil, errors.New("chunk doesn't start with 0xff")
	}

	// Read chunks type.
	n, err = f.Read(b)
	if len(b) != n || err != nil {
		return nil, err
	}
	c.Type = Marker(b[0])

	switch c.Type {
	case SOI, EOI:
		// Save current file offset.
		c.DataOffset, err = f.Seek(0, os.SEEK_CUR)
		if err != nil {
			return nil, err
		}
		c.DataOffset += -2 // two chunk marker bytes
	default:
		// Read chunks length.
		b = make([]byte, 2)
		n, err = f.Read(b)
		if len(b) != n || err != nil {
			return nil, err
		}
		c.Len = int64(binary.BigEndian.Uint16(b))

		// Save current file offset.
		c.DataOffset, err = f.Seek(0, os.SEEK_CUR)
		if err != nil {
			return nil, err
		}
		c.DataOffset += -4 // two chunk marker bytes and two previous length bytes
	}
	return c, nil
}

// skipChunk offsets current seek to skip chunks data.
func skipChunk(f *os.File, c *chunkHeader) error {
	if c.Len == 0 {
		return nil
	}
	_, err := f.Seek(c.Len-2, os.SEEK_CUR)
	return err
}

// parseImageData return slice with start and end offsets of image data after SOS chunk.
func parseImageData(f *os.File) ([2]int64, error) {
	imageData := [2]int64{}
	var err error

	imageData[0], err = f.Seek(0, os.SEEK_CUR)
	if err != nil {
		return [2]int64{}, err
	}

	b := make([]byte, 1)
	for {
		n, err := f.Read(b)
		if len(b) != n || err != nil {
			return [2]int64{}, err
		}
		if b[0] == 0xff {
			n, err := f.Read(b)
			if len(b) != n || err != nil {
				return [2]int64{}, err
			}
			if !(b[0] == 0x00 || Marker(b[0]) == DNL || (RST0 <= Marker(b[0]) && Marker(b[0]) <= RST7)) {
				_, err = f.Seek(-2, os.SEEK_CUR)
				if err != nil {
					return [2]int64{}, err
				}
				break
			}
		}
	}

	imageData[1], err = f.Seek(0, os.SEEK_CUR)
	if err != nil {
		return [2]int64{}, err
	}

	return imageData, nil
}
