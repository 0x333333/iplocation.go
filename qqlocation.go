package QQLocation

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io/ioutil"
	"net"
)

// Parser object
type Parser struct {
	data []byte

	p1 uint /* position of first record in Index zone */
	p2 uint /* position of last record in Index zone */
	p3 uint /* position of first record in Prefix zone */
	p4 uint /* position of last record in Prefix zone */

	prefixDict map[uint]PrefixIndex
}

// PrefixIndex object
type PrefixIndex struct {
	prefix     uint
	indexStart uint
	indexEnd   uint
}

// Information object
type Information struct {
	Continent   string
	Country     string
	Province    string
	City        string
	Isp         string
	CountryEn   string
	CountryCode string
	long        string
	lat         string
}

func (p *Parser) init(data []byte) {
	p.data = data
	p.p1 = uint(binary.LittleEndian.Uint32(data[:4]))
	p.p2 = uint(binary.LittleEndian.Uint32(data[4:8]))
	p.p3 = uint(binary.LittleEndian.Uint32(data[8:12]))
	p.p4 = uint(binary.LittleEndian.Uint32(data[12:16]))

	prefixCount := (p.p4-p.p3)/9 + 1

	indexBuffer := data[p.p3 : p.p3+prefixCount*9]
	p.prefixDict = make(map[uint]PrefixIndex)
	for k := uint(0); k < prefixCount; k++ {
		i := k * 9
		prefix := uint(indexBuffer[i])
		indexStart := uint(binary.LittleEndian.Uint32(indexBuffer[i+1 : i+5]))
		indexEnd := uint(binary.LittleEndian.Uint32(indexBuffer[i+5 : i+9]))
		p.prefixDict[prefix] = PrefixIndex{prefix, indexStart, indexEnd}
	}
}

func (p *Parser) binarySearch(low uint, high uint, ip uint) uint {
	mid := uint(0)
	for low <= high {
		m := (low + high) / 2
		endip := p.getEndIP(m)
		if endip > ip {
			mid = m
			if m == 0 {
				break
			}
			high = m - 1
		} else {
			low = m + 1
		}
	}
	return mid
}

func (p *Parser) getEndIP(left uint) uint {
	offset := p.p1 + left*12
	return uint(binary.LittleEndian.Uint32(p.data[4+offset : 8+offset]))
}

func (p *Parser) getIndex(left uint) (ipStart, ipEnd, locOffset, locLength uint) {
	offset := p.p1 + left*12
	ipStart = uint(binary.LittleEndian.Uint32(p.data[offset : offset+4]))
	ipEnd = uint(binary.LittleEndian.Uint32(p.data[offset+4 : offset+8]))
	locOffset = uint(p.data[offset+8]) + (uint(p.data[offset+9]) << 8) + (uint(p.data[offset+10]) << 16)
	locLength = uint(p.data[offset+11])
	return
}

func ip2int(ipStr string) (ipInt uint, ipPrefix uint, err error) {
	ip := net.ParseIP(ipStr)

	if ip == nil {
		err = errors.New("Fail to parse IP.")
		return
	}

	return uint(binary.BigEndian.Uint32([]byte(ip.To4()))), uint(ip[12]), nil
}

func newInformation(str []byte) (info *Information, err error) {
	fields := bytes.Split(str, []byte("|"))
	if len(fields) == 11 {
		info = &Information{
			Continent:   string(fields[0]),
			Country:     string(fields[1]),
			Province:    string(fields[2]),
			City:        string(fields[3]),
			Isp:         string(fields[5]),
			CountryEn:   string(fields[7]),
			CountryCode: string(fields[8]),
			long:        string(fields[9]),
			lat:         string(fields[10]),
		}
	} else {
		err = errors.New("unexpected ip info format: " + string(str))
	}

	return
}

// Init parser
func Init(file string) (*Parser, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	parser := new(Parser)
	parser.init(data)
	return parser, nil
}

// Search entry
func (p *Parser) Search(ipStr string) (info *Information, err error) {
	info = &Information{
		Continent:   "N/A",
		Country:     "N/A",
		Province:    "N/A",
		City:        "N/A",
		Isp:         "N/A",
		CountryEn:   "N/A",
		CountryCode: "N/A",
		long:        "N/A",
		lat:         "N/A",
	}

	ipInt, ipPrefix, err := ip2int(ipStr)

	if err != nil {
		return
	}

	low := uint(0)
	high := uint(0)

	if val, ok := p.prefixDict[ipPrefix]; ok {
		low = val.indexStart
		high = val.indexEnd
	} else {
		err = errors.New("IP not found")
		return
	}

	i := p.binarySearch(low, high, ipInt)
	ipStart, ipEnd, locOffset, locLength := p.getIndex(i)

	if ipStart <= ipInt && ipEnd >= ipInt {
		text := p.data[locOffset : locOffset+locLength]
		return newInformation(text)
	}

	return
}
