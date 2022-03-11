package main

import (
	"bytes"
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type GeneralsHeader struct {
	GameType       *ByteString `byte:"size=6"`
	TimeStampBegin *ByteInt32  `byte:"size=4"`
	TimeStampEnd   *ByteInt32  `byte:"size=4"`
	NumTimeStamps  *ByteInt16  `byte:"size=2"`
	Unknown1       *ByteRaw    `byte:"size=12"`
	FileName       *ByteString `byte:"size=2,nullterm,smallchar"`
	Junk           *ByteRaw    `byte:"size=16"`
	Version        *ByteString `byte:"size=2,nullterm,smallchar"`
	DateTime       *ByteString `byte:"size=2,nullterm,smallchar"`
	Junk2          *ByteRaw    `byte:"size=12"`
	Map            *ByteString `byte:"size=2,nullterm"`
	Unknown2       *ByteInt16  `byte:"size=2"`
	Unknown3       *ByteInt32  `byte:"size=4"`
	Unknown4       *ByteInt32  `byte:"size=4"`
	Unknown5       *ByteInt32  `byte:"size=4"`
	Unknown6       *ByteInt32  `byte:"size=4"`
}

type GeneralsBody struct {
	TimeCode       *ByteInt32 `byte:"size=4"`
	Order          *ByteInt32 `byte:"size=4"`
	Number         *ByteInt32 `byte:"size=4"`
	UniqueArgCount *ByteRaw   `byte:"size=1"`
	Orders
}

type GeneralsOrder struct {
	OrderType    *ByteRaw `byte:"size=1"`
	NumberOfArgs *ByteRaw `byte:"size=1"`
}

type ByteInterface interface {
	Write([]byte)
}

type ByteRaw []byte

func NewByteRaw(b []byte) *ByteRaw {
	br := ByteRaw(b)
	return &br
}

func (br *ByteRaw) Write(b []byte) {
	*br = ByteRaw(b)
}

type ByteString string

func NewByteString(s string) *ByteString {
	bs := ByteString(s)
	return &bs
}

func (bs *ByteString) Write(b []byte) {
	*bs = ByteString(b)
}

type ByteInt32 int32

func NewByteInt32(i int32) *ByteInt32 {
	bs := ByteInt32(i)
	return &bs
}

func (bs *ByteInt32) Write(b []byte) {
	*bs = ByteInt32(binary.LittleEndian.Uint32(b))
}

type ByteInt16 int16

func NewByteInt16(i int16) *ByteInt16 {
	bs := ByteInt16(i)
	return &bs
}

func (bs *ByteInt16) Write(b []byte) {
	*bs = ByteInt16(binary.LittleEndian.Uint16(b))
}

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.WithError(err).Fatal("could not open file")
	}
	header := GeneralsHeader{
		GameType:       NewByteString(""),
		TimeStampBegin: NewByteInt32(0),
		TimeStampEnd:   NewByteInt32(0),
		NumTimeStamps:  NewByteInt16(0),
		Unknown1:       NewByteRaw([]byte{}),
		FileName:       NewByteString(""),
		Junk:           NewByteRaw([]byte{}),
		Version:        NewByteString(""),
		DateTime:       NewByteString(""),
		Junk2:          NewByteRaw([]byte{}),
		Map:            NewByteString(""),
		Unknown2:       NewByteInt16(0),
		Unknown3:       NewByteInt32(0),
		Unknown4:       NewByteInt32(0),
		Unknown5:       NewByteInt32(0),
		Unknown6:       NewByteInt32(0),
	}
	inst := reflect.ValueOf(&header).Elem()
	gtype := reflect.TypeOf(header)
	for i := 0; i < gtype.NumField(); i++ {
		tagKv := parseTag(gtype.Field(i).Tag.Get("byte"))
		size, err := strconv.Atoi(tagKv["size"])
		if err != nil {
			log.WithError(err).Error("could not parse fieldRaw size")
			continue
		}
		//fieldValue := []byte{}
		_, nullterm := tagKv["nullterm"]
		_, smallchar := tagKv["smallchar"]
		switch {
		case !nullterm:
			fieldRaw := make([]byte, size)
			sizeRead, err := file.Read(fieldRaw)
			if sizeRead != size {
				log.Errorf("unable to read fieldRaw. expected size: %d, got: %d", size, sizeRead)
			}

			if err != nil {
				log.WithError(err).Error("unable to read fieldRaw")
			}

			field := inst.Field(i).Interface().(ByteInterface)

			field.Write(fieldRaw)

		case nullterm:
			fieldRaw := bytes.Buffer{}
			for {
				b := make([]byte, size)
				sizeRead, err := file.Read(b)
				if sizeRead != size {
					log.Errorf("unable to read fieldRaw. expected size: %d, got: %d", size, sizeRead)
					break
				}

				if err != nil {
					log.WithError(err).Error("unable to read fieldRaw")
				}
				if b[0] == byte(0) && b[1] == byte(0) {
					break
				}
				switch {
				case smallchar:
					fieldRaw.Write(b[:1])
				case !smallchar:
					fieldRaw.Write(b)
				}
			}
			field := inst.Field(i).Interface().(ByteInterface)

			field.Write(fieldRaw.Bytes())
		}
	}
	log.Printf("%+v", *header.Map)
}

func parseTag(rawTag string) map[string]string {
	out := map[string]string{}
	pairs := strings.Split(rawTag, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		switch len(kv) {
		case 1:
			out[kv[0]] = "true"
		case 2:
			out[kv[0]] = kv[1]
		default:
			log.WithField("pair", pair).Warn("unexpected tag pair format")
		}
	}
	return out
}
