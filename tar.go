// tar implements the tape archive format.
package carchivum

type Tar struct {
	car
}

func NewTar() *Tar {
	tar := &Tar{}
	tar.CompressionFormat = defaultCompressionFormat

}

func (t *Tar) Create(sources ...string) error {
	if len(sources) == 0 {
		return errors.New("Nothing to archive; no sources were received")
	}
	return nil

	//
}

func (t *Tar) Delete() error {
	return nil
}

func (t *Tar) Extract() error {
	return nil
}

