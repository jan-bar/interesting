/*Package LzmaSpec

本文件照抄官方SDK: [https://www.7-zip.org/a/lzma-specification.7z]
本文件用法参考DecodeLzma.go

懒得将压缩代码也抄一份,因为平时代码多数用解压,需要压缩就用下面的jar包吧
lzma.jar: [https://www.7-zip.org/a/lzma2201.7z]的java代码打包
bash -x make_java.sh 打包java代码为jar文件
java -jar lzma.jar: 查看帮助
java -jar lzma.jar e a.txt a.lzma: 压缩a.txt到a.lzma
java -jar lzma.jar d a.lzma a.txt: 解压a.lzma到a.txt

另外也可以用7z官网下载的lzma.exe来压缩和解压缩,但只是windows下用
*/
package LzmaSpec

import (
	"bufio"
	"errors"
	"io"
	"sync/atomic"
)

type cOutStream struct {
	File      *bufio.Writer
	Processed uint64
}

func (c *cOutStream) WriteByte(b byte) error {
	atomic.AddUint64(&c.Processed, 1)
	return c.File.WriteByte(b)
}

type cOutWindow struct {
	buf    []byte
	pos    uint32
	size   uint32
	isFull bool

	TotalPos  uint32
	OutStream *cOutStream
}

func (c *cOutWindow) create(dictSize uint32) {
	c.buf = make([]byte, dictSize)
	c.pos = 0
	c.size = dictSize
	c.isFull = false
	c.TotalPos = 0
}

func (c *cOutWindow) PutByte(b byte) error {
	c.TotalPos++
	c.buf[c.pos] = b
	c.pos++
	if c.pos == c.size {
		c.pos = 0
		c.isFull = true
	}
	return c.OutStream.WriteByte(b)
}

func (c *cOutWindow) GetByte(dist uint32) byte {
	if dist <= c.pos {
		return c.buf[c.pos-dist]
	}
	return c.buf[c.size-dist+c.pos]
}

func (c *cOutWindow) CopyMatch(dist, l uint32) error {
	var err error
	for ; l > 0; l-- {
		if err = c.PutByte(c.GetByte(dist)); err != nil {
			return err
		}
	}
	return nil
}

func (c *cOutWindow) CheckDistance(dist uint32) bool {
	return dist <= c.pos || c.isFull
}

func (c *cOutWindow) IsEmpty() bool {
	return c.pos == 0 && !c.isFull
}

type cProb uint16

const (
	kNumBitModelTotalBits = 11
	kNumMoveBits          = 5

	uint32One  uint32 = 1
	uint32Zero uint32 = 0

	probInitVal = cProb((1 << kNumBitModelTotalBits) / 2)
	kTopValue   = uint32One << 24
)

func initProbs(p []cProb) {
	for i := 0; i < len(p); i++ {
		p[i] = probInitVal
	}
}

type cRangeDecoder struct {
	Range uint32
	Code  uint32

	InStream  *bufio.Reader
	Corrupted bool
}

func (c *cRangeDecoder) Normalize() error {
	if c.Range < kTopValue {
		c.Range <<= 8
		b, err := c.InStream.ReadByte()
		if err != nil {
			return err
		}
		c.Code = (c.Code << 8) | uint32(b)
	}
	return nil
}

func (c *cRangeDecoder) Init() (bool, error) {
	c.Corrupted = false
	c.Range = 0xFFFFFFFF
	c.Code = 0
	b0, err := c.InStream.ReadByte()
	if err != nil {
		return false, err
	}
	var b1 byte
	for i := 0; i < 4; i++ {
		b1, err = c.InStream.ReadByte()
		if err != nil {
			return false, err
		}
		c.Code = (c.Code << 8) | uint32(b1)
	}
	if b0 != 0 || c.Code == c.Range {
		c.Corrupted = true
	}
	return b0 == 0, nil
}

func (c *cRangeDecoder) IsFinishedOK() bool {
	return c.Code == 0
}

func (c *cRangeDecoder) DecodeDirectBits(numBits uint32) (uint32, error) {
	res := uint32Zero
	for {
		c.Range >>= 1
		c.Code -= c.Range
		t := 0 - (c.Code >> 31)
		c.Code += c.Range & t

		if c.Code == c.Range {
			c.Corrupted = true
		}
		err := c.Normalize()
		if err != nil {
			return 0, err
		}
		res <<= 1
		res += t + 1
		if numBits--; numBits == 0 {
			break
		}
	}
	return res, nil
}

func (c *cRangeDecoder) DecodeBit(prob *cProb) (uint32, error) {
	v, symbol := uint32(*prob), uint32Zero
	bound := (c.Range >> kNumBitModelTotalBits) * v
	if c.Code < bound {
		v += ((1 << kNumBitModelTotalBits) - v) >> kNumMoveBits
		c.Range = bound
	} else {
		v -= v >> kNumMoveBits
		c.Code -= bound
		c.Range -= bound
		symbol = 1
	}
	*prob = cProb(v)
	return symbol, c.Normalize()
}

func bitTreeReverseDecode(probs []cProb, numBits uint32, rc *cRangeDecoder) (uint32, error) {
	m, symbol, i := uint32One, uint32Zero, uint32Zero
	for ; i < numBits; i++ {
		bit, err := rc.DecodeBit(&probs[m])
		if err != nil {
			return 0, err
		}
		m <<= 1
		m += bit
		symbol |= bit << i
	}
	return symbol, nil
}

type cBitTreeDecoder struct {
	numBits uint32
	Probs   []cProb
}

func newCBitTreeDecoder(numBits uint32) *cBitTreeDecoder {
	return &cBitTreeDecoder{
		numBits: numBits,
		Probs:   make([]cProb, uint32(1<<numBits)),
	}
}

func (c *cBitTreeDecoder) init() {
	initProbs(c.Probs)
}

func (c *cBitTreeDecoder) Decode(rc *cRangeDecoder) (uint32, error) {
	m, i := uint32One, uint32Zero
	for ; i < c.numBits; i++ {
		tmp, err := rc.DecodeBit(&c.Probs[m])
		if err != nil {
			return 0, err
		}
		m = (m << 1) + tmp
	}
	return m - (uint32One << c.numBits), nil
}

func (c *cBitTreeDecoder) ReverseDecode(rc *cRangeDecoder) (uint32, error) {
	return bitTreeReverseDecode(c.Probs, c.numBits, rc)
}

const (
	kNumPosBitsMax = 4

	kNumStates         = 12
	kNumLenToPosStates = 4
	kNumAlignBits      = 4
	// kStartPosModelIndex = 4
	kEndPosModelIndex = 14
	kNumFullDistances = 1 << (kEndPosModelIndex >> 1)
	kMatchMinLen      = 2
)

type cLenDecoder struct {
	Choice  cProb
	Choice2 cProb

	lLM       int
	LowCoder  []*cBitTreeDecoder
	MidCoder  []*cBitTreeDecoder
	HighCoder *cBitTreeDecoder
}

func newCLenDecoder() *cLenDecoder {
	val := &cLenDecoder{
		Choice:    probInitVal,
		Choice2:   probInitVal,
		lLM:       1 << kNumPosBitsMax,
		HighCoder: newCBitTreeDecoder(8),
	}
	val.LowCoder = make([]*cBitTreeDecoder, val.lLM)
	val.MidCoder = make([]*cBitTreeDecoder, val.lLM)
	for i := 0; i < val.lLM; i++ {
		val.LowCoder[i] = newCBitTreeDecoder(3)
		val.MidCoder[i] = newCBitTreeDecoder(3)
	}
	return val
}

func (c *cLenDecoder) Init() {
	c.Choice = probInitVal
	c.Choice2 = probInitVal
	c.HighCoder.init()
	for i := 0; i < c.lLM; i++ {
		c.LowCoder[i].init()
		c.MidCoder[i].init()
	}
}

func (c *cLenDecoder) Decode(rc *cRangeDecoder, posState uint32) (uint32, error) {
	tmp, err := rc.DecodeBit(&c.Choice)
	if err != nil {
		return 0, err
	}
	if tmp == 0 {
		return c.LowCoder[posState].Decode(rc)
	}
	tmp, err = rc.DecodeBit(&c.Choice2)
	if err != nil {
		return 0, err
	}
	if tmp == 0 {
		tmp, err = c.MidCoder[posState].Decode(rc)
		if err != nil {
			return 0, err
		}
		return 8 + tmp, nil
	}
	tmp, err = c.HighCoder.Decode(rc)
	if err != nil {
		return 0, err
	}
	return 16 + tmp, nil
}

func updateStateLiteral(state uint32) uint32 {
	if state < 4 {
		return 0
	}
	if state < 10 {
		return state - 3
	}
	return state - 6
}

func updateStateMatch(state uint32) uint32 {
	if state < 7 {
		return 7
	}
	return 10
}

func updateStateRep(state uint32) uint32 {
	if state < 7 {
		return 8
	}
	return 11
}

func updateStateShortRep(state uint32) uint32 {
	if state < 7 {
		return 9
	}
	return 11
}

const lzmaDicMin = uint32(1 << 12)

type cLzmaDecoder struct {
	rangeDec  *cRangeDecoder
	outWindow *cOutWindow

	header               []byte
	markerIsMandatory    bool
	lc, pb, lp           uint32
	dictSize             uint32
	dictSizeInProperties uint32

	unpackSize uint64
	litProbs   []cProb

	posSlotDecoder []*cBitTreeDecoder
	alignDecoder   *cBitTreeDecoder
	posDecoders    []cProb

	isMatch    []cProb
	isRep      []cProb
	isRepG0    []cProb
	isRepG1    []cProb
	isRepG2    []cProb
	isRep0Long []cProb

	lenDecoder    *cLenDecoder
	repLenDecoder *cLenDecoder
}

func NewCLzmaDecoder(r io.Reader, w io.Writer) *cLzmaDecoder {
	val := &cLzmaDecoder{
		rangeDec: &cRangeDecoder{InStream: bufio.NewReader(r)},
		outWindow: &cOutWindow{
			OutStream: &cOutStream{
				File:      bufio.NewWriter(w),
				Processed: 0,
			},
		},
		posSlotDecoder: make([]*cBitTreeDecoder, kNumLenToPosStates),
		alignDecoder:   newCBitTreeDecoder(kNumAlignBits),
		posDecoders:    make([]cProb, 1+kNumFullDistances-kEndPosModelIndex),
		isMatch:        make([]cProb, kNumStates<<kNumPosBitsMax),
		isRep:          make([]cProb, kNumStates),
		isRepG0:        make([]cProb, kNumStates),
		isRepG1:        make([]cProb, kNumStates),
		isRepG2:        make([]cProb, kNumStates),
		isRep0Long:     make([]cProb, kNumStates<<kNumPosBitsMax),
		lenDecoder:     newCLenDecoder(),
		repLenDecoder:  newCLenDecoder(),
	}
	for i := 0; i < kNumLenToPosStates; i++ {
		val.posSlotDecoder[i] = newCBitTreeDecoder(6)
	}
	return val
}

func (c *cLzmaDecoder) init() {
	c.initLiterals()
	c.initDist()
	initProbs(c.isMatch)
	initProbs(c.isRep)
	initProbs(c.isRepG0)
	initProbs(c.isRepG1)
	initProbs(c.isRepG2)
	initProbs(c.isRep0Long)
	c.lenDecoder.Init()
	c.repLenDecoder.Init()
}

func (c *cLzmaDecoder) DecodeProperties() (val [5]uint32, unpackSize uint64, unpackSizeDefined bool, err error) {
	c.header = make([]byte, 13)
	if _, err = io.ReadFull(c.rangeDec.InStream, c.header); err != nil {
		return
	}

	d := uint32(c.header[0])
	if d >= (9 * 5 * 5) {
		err = errors.New("incorrect LZMA properties")
		return
	}
	c.lc = d % 9
	d /= 9
	c.pb = d / 5
	c.lp = d % 5
	c.dictSizeInProperties = 0
	for i := 0; i < 4; i++ {
		c.dictSizeInProperties |= uint32(c.header[i+1]) << (8 * i)
	}
	c.dictSize = c.dictSizeInProperties
	if c.dictSize < lzmaDicMin {
		c.dictSize = lzmaDicMin
	}

	val[0], val[1], val[2] = c.lc, c.lp, c.pb
	val[3] = c.dictSizeInProperties
	val[4] = c.dictSize

	for i := 0; i < 8; i++ {
		b := c.header[5+i]
		if b != 0xff {
			unpackSizeDefined = true
		}
		unpackSize |= uint64(b) << (8 * i)
	}
	c.markerIsMandatory = !unpackSizeDefined
	c.unpackSize = unpackSize
	return
}

func (c *cLzmaDecoder) create() {
	c.outWindow.create(c.dictSize)
	c.createLiterals()
}

func (c *cLzmaDecoder) createLiterals() {
	c.litProbs = make([]cProb, uint32(0x300)<<(c.lc+c.lp))
}

func (c *cLzmaDecoder) initLiterals() {
	num := uint32(0x300) << (c.lc + c.lp)
	for i := uint32Zero; i < num; i++ {
		c.litProbs[i] = probInitVal
	}
}

func (c *cLzmaDecoder) decodeLiteral(state, rep0 uint32) error {
	prevByte := uint32Zero
	if !c.outWindow.IsEmpty() {
		prevByte = uint32(c.outWindow.GetByte(1))
	}
	symbol := uint32One
	litState := ((c.outWindow.TotalPos & ((1 << c.lp) - 1)) << c.lc) + (prevByte >> (8 - c.lc))
	probs := c.litProbs[uint32(0x300)*litState:]
	if state >= 7 {
		matchByte := uint32(c.outWindow.GetByte(rep0 + 1))
		for {
			matchBit := (matchByte >> 7) & 1
			matchByte <<= 1
			bit, err := c.rangeDec.DecodeBit(&probs[((1+matchBit)<<8)+symbol])
			if err != nil {
				return err
			}
			symbol = (symbol << 1) | bit
			if matchBit != bit || symbol >= 0x100 {
				break
			}
		}
	}
	for symbol < 0x100 {
		bit, err := c.rangeDec.DecodeBit(&probs[symbol])
		if err != nil {
			return err
		}
		symbol = (symbol << 1) | bit
	}
	return c.outWindow.PutByte(byte(symbol - 0x100))
}

func (c *cLzmaDecoder) initDist() {
	for i := 0; i < kNumLenToPosStates; i++ {
		c.posSlotDecoder[i].init()
	}
	c.alignDecoder.init()
	initProbs(c.posDecoders)
}

func (c *cLzmaDecoder) decodeDistance(l uint32) (uint32, error) {
	lenState := l
	if lenState > kNumLenToPosStates-1 {
		lenState = kNumLenToPosStates - 1
	}
	posSlot, err := c.posSlotDecoder[lenState].Decode(c.rangeDec)
	if err != nil {
		return 0, err
	}
	if posSlot < 4 {
		return posSlot, nil
	}

	var tmp uint32
	numDirectBits := (posSlot >> 1) - 1
	dist := (2 | (posSlot & 1)) << numDirectBits
	if posSlot < kEndPosModelIndex {
		tmp, err = bitTreeReverseDecode(c.posDecoders[dist-posSlot:], numDirectBits, c.rangeDec)
		if err != nil {
			return 0, err
		}
		dist += tmp
	} else {
		tmp, err = c.rangeDec.DecodeDirectBits(numDirectBits - kNumAlignBits)
		if err != nil {
			return 0, err
		}
		dist += tmp << kNumAlignBits
		tmp, err = c.alignDecoder.ReverseDecode(c.rangeDec)
		if err != nil {
			return 0, err
		}
		dist += tmp
	}
	return dist, nil
}

const (
	LzmaResError                 = 0
	LzmaResFinishedWithMarker    = 1
	LzmaResFinishedWithoutMarker = 2
)

func (c *cLzmaDecoder) Decode(unpackSizeDefined bool, unpackSize uint64) (int, error) {
	defer c.outWindow.OutStream.File.Flush()

	c.create()
	isError, err := c.rangeDec.Init()
	if err != nil {
		return 0, err
	}
	if !isError {
		return LzmaResError, nil
	}
	c.init()
	var rep0, rep1, rep2, rep3, state, posState, bit, l, dist uint32
	for {
		if unpackSizeDefined && unpackSize == 0 && !c.markerIsMandatory {
			if c.rangeDec.IsFinishedOK() {
				return LzmaResFinishedWithoutMarker, nil
			}
		}
		posState = c.outWindow.TotalPos & ((1 << c.pb) - 1)
		bit, err = c.rangeDec.DecodeBit(&c.isMatch[(state<<kNumPosBitsMax)+posState])
		if err != nil {
			return 0, err
		}
		if bit == 0 {
			if unpackSizeDefined && unpackSize == 0 {
				return LzmaResError, nil
			}
			err = c.decodeLiteral(state, rep0)
			if err != nil {
				return 0, err
			}
			state = updateStateLiteral(state)
			unpackSize--
			continue
		}

		bit, err = c.rangeDec.DecodeBit(&c.isRep[state])
		if err != nil {
			return 0, err
		}
		l = uint32Zero
		if bit != 0 {
			if unpackSizeDefined && unpackSize == 0 {
				return LzmaResError, nil
			}
			if c.outWindow.IsEmpty() {
				return LzmaResError, nil
			}
			bit, err = c.rangeDec.DecodeBit(&c.isRepG0[state])
			if err != nil {
				return 0, err
			}
			if bit == 0 {
				bit, err = c.rangeDec.DecodeBit(&c.isRep0Long[(state<<kNumPosBitsMax)+posState])
				if err != nil {
					return 0, err
				}
				if bit == 0 {
					state = updateStateShortRep(state)
					err = c.outWindow.PutByte(c.outWindow.GetByte(rep0 + 1))
					if err != nil {
						return 0, err
					}
					unpackSize--
					continue
				}
			} else {
				bit, err = c.rangeDec.DecodeBit(&c.isRepG1[state])
				if err != nil {
					return 0, err
				}
				dist = uint32Zero
				if bit == 0 {
					dist = rep1
				} else {
					bit, err = c.rangeDec.DecodeBit(&c.isRepG2[state])
					if err != nil {
						return 0, err
					}
					if bit == 0 {
						dist = rep2
					} else {
						dist = rep3
						rep3 = rep2
					}
					rep2 = rep1
				}
				rep1 = rep0
				rep0 = dist
			}
			l, err = c.repLenDecoder.Decode(c.rangeDec, posState)
			if err != nil {
				return 0, err
			}
			state = updateStateRep(state)
		} else {
			rep3 = rep2
			rep2 = rep1
			rep1 = rep0
			l, err = c.lenDecoder.Decode(c.rangeDec, posState)
			if err != nil {
				return 0, err
			}
			state = updateStateMatch(state)
			bit, err = c.decodeDistance(l)
			if err != nil {
				return 0, err
			}
			rep0 = bit
			if rep0 == 0xFFFFFFFF {
				if c.rangeDec.IsFinishedOK() {
					return LzmaResFinishedWithMarker, nil
				}
				return LzmaResError, nil
			}
			if unpackSizeDefined && unpackSize == 0 {
				return LzmaResError, nil
			}
			if rep0 >= c.dictSize || !c.outWindow.CheckDistance(rep0) {
				return LzmaResError, nil
			}
		}
		l += kMatchMinLen
		isError = false
		if unpackSizeDefined && unpackSize < uint64(l) {
			l = uint32(unpackSize)
			isError = true
		}
		if err = c.outWindow.CopyMatch(rep0+1, l); err != nil {
			return 0, err
		}
		unpackSize -= uint64(l)
		if isError {
			return LzmaResError, nil
		}
	}
}

func (c *cLzmaDecoder) GetOutStreamProcessed() uint64 {
	return atomic.LoadUint64(&c.outWindow.OutStream.Processed)
}

func (c *cLzmaDecoder) IsCorrupted() bool {
	return c.rangeDec.Corrupted
}
