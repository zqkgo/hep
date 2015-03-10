// Copyright 2015 The go-hep Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rio

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"io"
	"reflect"
)

const (
	gAlign        = 0x00000003
	rioHdrVersion = Version(0)

	gMaskCodec = Options(0x00000fff)
	gMaskLevel = Options(0x0000f000)
	gMaskCompr = Options(0xffff0000)
)

// Version describes a rio on-disk version of a serialized block.
type Version uint32

// frameType frames blocks, records and footers
type frameType [4]byte

var (
	rioMagic = [4]byte{'r', 'i', 'o', '\x00'}

	recFrame = frameType{0xab, 0xad, 0xca, 0xfe} // 0xabadcafe
	blkFrame = frameType{0xde, 0xad, 0xbe, 0xef} // 0xdeadbeef
	ftrFrame = frameType{0xca, 0xfe, 0xba, 0xbe} // 0xcafebabe

	// Endian exposes the endianness of rio streams
	Endian = binary.LittleEndian

	hdrSize = int(reflect.TypeOf(rioHeader{}).Size())
	recSize = int(reflect.TypeOf(rioRecord{}).Size())
	blkSize = int(reflect.TypeOf(rioBlock{}).Size())
	ftrSize = int(reflect.TypeOf(rioFooter{}).Size())
)

// Marshaler is the interface implemented by an object that can
// marshal itself into a rio-binary form.
//
// RioMarshal marshals the receiver into a rio-binary form, writes that
// binary form to the io.Writer and returns an error if any.
type Marshaler interface {
	RioMarshal(w io.Writer) error
}

// Unmarshalr is the interface implemented by an object that can
// unmarshal a rio-binary representation of itself.
//
// RioUnmarshal must be able to unmarshal the form generated by RioMarshal.
type Unmarshaler interface {
	RioUnmarshal(r io.Reader) error
}

// Streamer is the interface implemented by an object that can
// marshal/unmarshal a rio-binary representation of itself
// to/from an io.Writer/io.Reader.
type Streamer interface {
	Marshaler
	Unmarshaler
	RioVersion() Version
}

// rioHeader
type rioHeader struct {
	// Length of payload in bytes (not counting Len nor Frame).
	// Always a multiple of four.
	Len uint32

	// Framing used to try identifying what kind of payload follows
	// (record or block)
	Frame frameType
}

func (hdr *rioHeader) RioMarshal(w io.Writer) error {
	var err error

	err = binary.Write(w, Endian, hdr.Len)
	if err != nil {
		return errorf("rio: write header length failed: %v", err)
	}

	err = binary.Write(w, Endian, hdr.Frame)
	if err != nil {
		return errorf("rio: write header frame failed: %v", err)
	}

	return err
}

func (hdr *rioHeader) RioUnmarshal(r io.Reader) error {
	var err error

	err = binary.Read(r, Endian, &hdr.Len)
	if err != nil {
		if err == io.EOF {
			return err
		}
		return errorf("rio: read header length failed: %v", err)
	}

	err = binary.Read(r, Endian, &hdr.Frame)
	if err != nil {
		return errorf("rio: read header frame failed: %v", err)
	}

	return err
}

func (hdr *rioHeader) RioVersion() Version {
	return rioHdrVersion
}

// Options describes the various options attached to a rio stream
// such as: compression method, compression level, codec, ...
type Options uint32

// CompressorKind extracts the CompressorKind from the Options value
func (o Options) CompressorKind() CompressorKind {
	return CompressorKind((o & gMaskCompr) >> 16)
}

// CompressorLevel extracts the compression level from the Options value
func (o Options) CompressorLevel() int {
	lvl := int((o & gMaskLevel) >> 12)
	if lvl == 0xf {
		return flate.DefaultCompression
	}
	return lvl
}

// CompressorCodec extracts the compression codec from the Options value
func (o Options) CompressorCodec() int {
	return int(o & gMaskCodec)
}

// NewOptions returns a new Options value carefully crafted from the CompressorKind and
// compression level
func NewOptions(compr CompressorKind, lvl int, codec int) Options {
	if lvl <= flate.DefaultCompression || lvl >= 0xf {
		lvl = 0xf
	}

	if compr == CompressDefault {
		compr = CompressZlib
	}

	// FIXME(sbinet): decide on how to handle different codecs (gob|cbor|xdr|riobin|...)
	opts := Options(Options(compr)<<16) |
		Options(Options(lvl)<<12) |
		Options(Options(codec)&gMaskCodec)
	return opts
}

// rioRecord
type rioRecord struct {
	Header rioHeader

	Options Options // options word (compression method, compression level, codec, ...)

	// length of compressed record content.
	// Total length in bytes for all the blocks in the record.
	// Always a multiple of four.
	// If the record is not compressed, same value than XLen.
	CLen uint32

	// length of un-compressed record content.
	// Total length in bytes for all the blocks in the record when decompressed.
	// Always a multiple of four.
	// When the record is not compressed, it is a count of the bytes that follow in the
	// record content.
	// When the record is compressed, this number is used to allocate a buffer into which
	// the record is decompressed.
	XLen uint32

	// name of the record. padded with zeros to a four byte boundary
	Name string
}

func (rec *rioRecord) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, recSize))
	err := rec.RioMarshal(buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (rec *rioRecord) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	return rec.RioUnmarshal(r)
}

func (rec *rioRecord) RioMarshal(w io.Writer) error {
	var err error

	err = rec.Header.RioMarshal(w)
	if err != nil {
		return errorf("rio: write record header failed: %v", err)
	}

	err = binary.Write(w, Endian, rec.Options)
	if err != nil {
		return errorf("rio: write record options failed: %v", err)
	}

	err = binary.Write(w, Endian, rec.CLen)
	if err != nil {
		return errorf("rio: write record compr-len failed: %v", err)
	}

	err = binary.Write(w, Endian, rec.XLen)
	if err != nil {
		return errorf("rio: write record len failed: %v", err)
	}

	err = binary.Write(w, Endian, uint32(len(rec.Name)))
	if err != nil {
		return errorf("rio: write record name-len failed: %v", err)
	}

	name := []byte(rec.Name)
	_, err = w.Write(name)
	if err != nil {
		return errorf("rio: write record name failed: %v", err)
	}

	size := rioAlign(len(name))
	if size > len(name) {
		_, err = w.Write(make([]byte, size-len(name)))
		if err != nil {
			return errorf("rio: write record name-padding failed: %v", err)
		}
	}

	return err
}

func (rec *rioRecord) RioUnmarshal(r io.Reader) error {
	var err error

	err = rec.unmarshalHeader(r)
	if err != nil {
		return err
	}

	err = rec.unmarshalData(r)
	if err != nil {
		return err
	}
	return err
}

func (rec *rioRecord) unmarshalHeader(r io.Reader) error {
	var err error

	err = rec.Header.RioUnmarshal(r)
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return err
		}
		return errorf("rio: read record header failed: %v", err)
	}

	if rec.Header.Frame != recFrame {
		return errorf("rio: read record header corrupted (frame=%#v)", rec.Header.Frame)
	}

	return err
}

func (rec *rioRecord) unmarshalData(r io.Reader) error {
	var err error

	err = binary.Read(r, Endian, &rec.Options)
	if err != nil {
		return errorf("rio: read record options failed: %v", err)
	}

	err = binary.Read(r, Endian, &rec.CLen)
	if err != nil {
		return errorf("rio: read record compr-len failed: %v", err)
	}

	err = binary.Read(r, Endian, &rec.XLen)
	if err != nil {
		return errorf("rio: read record len failed failed: %v", err)
	}

	nsize := uint32(0)
	err = binary.Read(r, Endian, &nsize)
	if err != nil {
		return errorf("rio: read record name-len failed: %v", err)
	}

	buf := make([]byte, rioAlign(int(nsize)))
	_, err = r.Read(buf)
	if err != nil {
		return errorf("rio: read record name failed: %v", err)
	}

	rec.Name = string(buf[:int(nsize)])

	return err
}

func (rec *rioRecord) RioVersion() Version {
	return rioHdrVersion
}

// rioBlock
type rioBlock struct {
	Header  rioHeader
	Version Version // block version
	Name    string  // block name
	Data    []byte  // block payload
}

func (blk *rioBlock) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, recSize))
	err := blk.RioMarshal(buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (blk *rioBlock) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	return blk.RioUnmarshal(r)
}

func (blk *rioBlock) RioMarshal(w io.Writer) error {
	var err error

	err = blk.Header.RioMarshal(w)
	if err != nil {
		return errorf("rio: write block header failed: %v", err)
	}

	err = binary.Write(w, Endian, blk.Version)
	if err != nil {
		return errorf("rio: write block version failed: %v", err)
	}

	name := []byte(blk.Name)
	err = binary.Write(w, Endian, uint32(len(name)))
	if err != nil {
		return errorf("rio: write block name-len failed: %v", err)
	}

	nb, err := w.Write(name)
	if err != nil {
		return errorf("rio: write block name failed: %v", err)
	}
	if nb != len(name) {
		return errorf("rio: wrote too few bytes (want=%d. got=%d)", len(name), nb)
	}

	nsize := rioAlign(len(name))
	if nsize > len(name) {
		nb, err = w.Write(make([]byte, nsize-len(name)))
		if err != nil {
			return errorf("rio: write block name-padding failed: %v", err)
		}
		if nb != nsize-len(name) {
			return errorf("rio: wrote too few bytes (want=%d. got=%d)", nsize-len(name), nb)
		}
	}

	nb, err = w.Write(blk.Data)
	if err != nil {
		return errorf("rio: write block data failed: %v", err)
	}
	if nb != len(blk.Data) {
		return errorf("rio: wrote too few bytes (want=%d. got=%d)", len(blk.Data), nb)
	}

	dsize := rioAlign(len(blk.Data))
	if dsize > len(blk.Data) {
		nb, err = w.Write(make([]byte, dsize-len(blk.Data)))
		if err != nil {
			return errorf("rio: write block data-padding failed: %v", err)
		}
		if nb != dsize-len(blk.Data) {
			return errorf("rio: wrote too few bytes (want=%d. got=%d)", dsize-len(blk.Data), nb)
		}
	}

	return err
}

func (blk *rioBlock) RioUnmarshal(r io.Reader) error {
	var err error

	err = blk.Header.RioUnmarshal(r)
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return err
		}
		return errorf("rio: read block header failed: %v", err)
	}

	if blk.Header.Frame != blkFrame {
		return errorf("rio: read block header corrupted (frame=%#v)", blk.Header.Frame)
	}

	err = binary.Read(r, Endian, &blk.Version)
	if err != nil {
		return errorf("rio: read block version failed: %v", err)
	}

	nsize := uint32(0)
	err = binary.Read(r, Endian, &nsize)
	if err != nil {
		return errorf("rio: read block name-len failed: %v", err)
	}
	name := make([]byte, rioAlign(int(nsize)))

	nb, err := io.ReadFull(r, name)
	if err != nil {
		return errorf("rio: read block name failed: %v", err)
	}
	if int(nb) != len(name) {
		return errorf("rio: read too few bytes for name (want=%d. got=%d)", len(name), nb)
	}

	blk.Name = string(name[:int(nsize)])

	data := make([]byte, rioAlign(int(blk.Header.Len)))
	nb, err = io.ReadFull(r, data)
	if err != nil {
		return errorf("rio: read block data failed: %v", err)
	}
	if int(nb) != len(data) {
		return errorf("rio: read too few bytes for data (want=%d. got=%d)", len(data), nb)
	}
	blk.Data = data[:int(blk.Header.Len)]

	return err
}

func (blk *rioBlock) RioVersion() Version {
	return blk.Version
}

// rioFooter marks the end of a rio stream
type rioFooter struct {
	Header rioHeader
	Meta   int64 // position of the record holding stream metadata, in bytes from rio-magic
}

func (ftr *rioFooter) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, recSize))
	err := ftr.RioMarshal(buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (ftr *rioFooter) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	return ftr.RioUnmarshal(r)
}

func (ftr *rioFooter) RioVersion() Version {
	return rioHdrVersion
}

func (ftr *rioFooter) RioMarshal(w io.Writer) error {
	var err error

	err = ftr.Header.RioMarshal(w)
	if err != nil {
		return errorf("rio: write footer header failed: %v", err)
	}

	err = binary.Write(w, Endian, ftr.Meta)
	if err != nil {
		return errorf("rio: write footer meta failed: %v", err)
	}

	return err
}

func (ftr *rioFooter) RioUnmarshal(r io.Reader) error {
	var err error

	err = ftr.unmarshalHeader(r)
	if err != nil {
		return err
	}

	err = ftr.unmarshalData(r)
	if err != nil {
		return err
	}

	return err
}

func (ftr *rioFooter) unmarshalHeader(r io.Reader) error {
	var err error
	err = ftr.Header.RioUnmarshal(r)
	if err != nil {
		if err == io.EOF {
			return err
		}
		return errorf("rio: read footer header failed: %v", err)
	}

	if ftr.Header.Frame != ftrFrame {
		return errorf("rio: read footer header corrupted (frame=%#v)", ftr.Header.Frame)
	}

	return err
}

func (ftr *rioFooter) unmarshalData(r io.Reader) error {
	var err error

	err = binary.Read(r, Endian, &ftr.Meta)
	if err != nil {
		return errorf("rio: read footer meta failed: %v", err)
	}

	return err
}

// Metadata stores metadata about a rio stream
type Metadata struct {
	Records []struct {
		Name   string
		Blocks []struct{ Name, Type string }
	}
	Offsets map[string][]int64
}
