package op

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/util"
)

type ReadMemoryFunc func([]byte, uint64) (int, error)

type OpcodeExecutorCreator func(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error)
type OpcodeExecutorCreatorContext struct {
	buf         *bytes.Buffer
	prog        []byte
	pointerSize int
}
type OpcodeExecutorContext struct {
	Stack      []int64
	Pieces     []Piece
	prog       []byte
	PtrSize    int
	readMemory ReadMemoryFunc

	DwarfRegisters
}
type OpcodeExecutor interface {
	Execute(ctx *OpcodeExecutorContext) error
}





type CloseLoc struct {
	isLastInstruction bool
	piece             Piece
	size              int
}

func newCloseLocExecutor(buf *bytes.Buffer, piece Piece) *CloseLoc {
	c := &CloseLoc{piece: piece}
	if buf.Len() == 0 {
		return c
	}

	b, err := buf.ReadByte()
	if err != nil {
		return nil
	}

	opcode := Opcode(b)
	switch opcode {
	case DW_OP_piece:
		sz, _ := util.DecodeULEB128(buf)
		c.piece.Size = int(sz)
		return c

	case DW_OP_bit_piece:
		
		return nil
	default:
		return nil
	}
}

func (c *CloseLoc) Execute(ctx *OpcodeExecutorContext) error {
	ctx.Pieces = append(ctx.Pieces, c.piece)
	return nil
}

type callframeCFAExecutor struct {
}

func newCallframeCFAExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	return &callframeCFAExecutor{}, nil
}

func (c *callframeCFAExecutor) Execute(ctx *OpcodeExecutorContext) error {
	if ctx.CFA == 0 {
		return fmt.Errorf("could not retrieve CFA for current PC")
	}
	ctx.Stack = append(ctx.Stack, int64(ctx.CFA))
	return nil
}

type addrExecutor struct {
	stack uint64
}

func newAddrExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	b := ctx.buf.Next(ctx.pointerSize)
	stack, err := util.ReadUintRaw(bytes.NewReader(b), binary.LittleEndian, ctx.pointerSize)
	if err != nil {
		return nil, err
	}
	return &addrExecutor{stack: stack}, nil
}

func (a *addrExecutor) Execute(ctx *OpcodeExecutorContext) error {
	ctx.Stack = append(ctx.Stack, int64(a.stack+ctx.StaticBase))
	return nil
}

type plusExecutor struct{}

func newPlusExector(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	return &plusExecutor{}, nil
}

func (p *plusExecutor) Execute(ctx *OpcodeExecutorContext) error {
	var (
		slen   = len(ctx.Stack)
		digits = ctx.Stack[slen-2 : slen]
		st     = ctx.Stack[:slen-2]
	)

	ctx.Stack = append(st, digits[0]+digits[1])
	return nil
}

type plusUconstsExecutor struct {
	num uint64
}

func newPlusUconstsExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	num, _ := util.DecodeULEB128(ctx.buf)
	return &plusUconstsExecutor{num: num}, nil
}

func (p *plusUconstsExecutor) Execute(ctx *OpcodeExecutorContext) error {
	slen := len(ctx.Stack)
	ctx.Stack[slen-1] = ctx.Stack[slen-1] + int64(p.num)
	return nil
}

type constsExecutor struct {
	num int64
}

func newConstsExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	num, _ := util.DecodeSLEB128(ctx.buf)
	return &constsExecutor{num: num}, nil
}

func (c *constsExecutor) Execute(ctx *OpcodeExecutorContext) error {
	ctx.Stack = append(ctx.Stack, c.num)
	return nil
}

type framebaseExecutor struct {
	num int64
}

func newFramebaseExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	num, _ := util.DecodeSLEB128(ctx.buf)
	return &framebaseExecutor{num: num}, nil
}

func (f *framebaseExecutor) Execute(ctx *OpcodeExecutorContext) error {
	ctx.Stack = append(ctx.Stack, ctx.FrameBase+f.num)
	return nil
}

func newRegisterExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	var regNum uint64
	if opcode == DW_OP_regx {
		n, _ := util.DecodeSLEB128(ctx.buf)
		regNum = uint64(n)
	} else {
		regNum = uint64(opcode - DW_OP_reg0)
	}

	return newCloseLocExecutor(ctx.buf, Piece{Kind: RegPiece, Val: regNum}), nil
}

type bRegisterExecutor struct {
	regNum uint64
	offset int64
}

func newBRegisterExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	var regnum uint64
	if opcode == DW_OP_bregx {
		regnum, _ = util.DecodeULEB128(ctx.buf)
	} else {
		regnum = uint64(opcode - DW_OP_breg0)
	}
	offset, _ := util.DecodeSLEB128(ctx.buf)
	return &bRegisterExecutor{regNum: regnum, offset: offset}, nil
}

func (b *bRegisterExecutor) Execute(ctx *OpcodeExecutorContext) error {
	if ctx.Reg(b.regNum) == nil {
		return fmt.Errorf("register %d not available", b.regNum)
	}
	ctx.Stack = append(ctx.Stack, int64(ctx.Uint64Val(b.regNum))+b.offset)
	return nil
}

type pieceExecutor struct {
	size uint64
}

func newPieceExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	sz, _ := util.DecodeULEB128(ctx.buf)
	return &pieceExecutor{size: sz}, nil
}

func (p *pieceExecutor) Execute(ctx *OpcodeExecutorContext) error {
	if len(ctx.Stack) == 0 {
		
		
		ctx.Pieces = append(ctx.Pieces, Piece{Size: int(p.size), Kind: ImmPiece, Val: 0})
		return nil
	}

	addr := ctx.Stack[len(ctx.Stack)-1]
	ctx.Pieces = append(ctx.Pieces, Piece{Size: int(p.size), Kind: AddrPiece, Val: uint64(addr)})
	ctx.Stack = ctx.Stack[:0]
	return nil
}

type literalExecutor struct {
	opcode Opcode
}

func newLiteralExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	return &literalExecutor{opcode: opcode}, nil
}

func (l *literalExecutor) Execute(ctx *OpcodeExecutorContext) error {
	ctx.Stack = append(ctx.Stack, int64(l.opcode-DW_OP_lit0))
	return nil
}

type constnuExecutor struct {
	n uint64
}

func newConstnuExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	var (
		n   uint64
		err error
	)
	switch opcode {
	case DW_OP_const1u:
		var b uint8
		b, err = ctx.buf.ReadByte()
		n = uint64(b)
	case DW_OP_const2u:
		n, err = util.ReadUintRaw(ctx.buf, binary.LittleEndian, 2)
	case DW_OP_const4u:
		n, err = util.ReadUintRaw(ctx.buf, binary.LittleEndian, 4)
	case DW_OP_const8u:
		n, err = util.ReadUintRaw(ctx.buf, binary.LittleEndian, 8)
	default:
		err = fmt.Errorf("unknown opcode: %v", opcode)
	}
	if err != nil {
		return nil, err
	}

	return &constnuExecutor{n: n}, nil
}

func (c *constnuExecutor) Execute(ctx *OpcodeExecutorContext) error {
	ctx.Stack = append(ctx.Stack, int64(c.n))
	return nil
}

type constnsExecutor struct {
	n uint64
}

func newConstnsExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	var (
		n   uint64
		err error
	)
	switch opcode {
	case DW_OP_const1s:
		var b uint8
		b, err = ctx.buf.ReadByte()
		n = uint64(int64(int8(b)))
	case DW_OP_const2s:
		n, err = util.ReadUintRaw(ctx.buf, binary.LittleEndian, 2)
		n = uint64(int64(int16(n)))
	case DW_OP_const4s:
		n, err = util.ReadUintRaw(ctx.buf, binary.LittleEndian, 4)
		n = uint64(int64(int32(n)))
	case DW_OP_const8s:
		n, err = util.ReadUintRaw(ctx.buf, binary.LittleEndian, 8)
	default:
		err = fmt.Errorf("unknown opcode: %v", opcode)
	}
	if err != nil {
		return nil, err
	}

	return &constnsExecutor{n: n}, nil
}

func (c *constnsExecutor) Execute(ctx *OpcodeExecutorContext) error {
	ctx.Stack = append(ctx.Stack, int64(c.n))
	return nil
}

type constuExecutor struct {
	num uint64
}

func newConstuExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	num, _ := util.DecodeULEB128(ctx.buf)
	return &constuExecutor{num: num}, nil
}

func (c *constuExecutor) Execute(ctx *OpcodeExecutorContext) error {
	ctx.Stack = append(ctx.Stack, int64(c.num))
	return nil
}

type dupExecutor struct {
}

func newDupExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	return &dupExecutor{}, nil
}

func (d *dupExecutor) Execute(ctx *OpcodeExecutorContext) error {
	if len(ctx.Stack) <= 0 {
		return fmt.Errorf("stack underflow: expected value in context stack")
	}
	ctx.Stack = append(ctx.Stack, ctx.Stack[len(ctx.Stack)-1])
	return nil
}

type dropExecutor struct {
}

func newDropExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	return &dropExecutor{}, nil
}

func (d *dropExecutor) Execute(ctx *OpcodeExecutorContext) error {
	if len(ctx.Stack) <= 0 {
		return fmt.Errorf("stack underflow: expected value in context stack")
	}
	ctx.Stack = ctx.Stack[:len(ctx.Stack)-1]
	return nil
}

type pickExecutor struct {
	n byte
}

func newPickExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	var n byte
	switch opcode {
	case DW_OP_pick:
		n, _ = ctx.buf.ReadByte()
	case DW_OP_over:
		n = 1
	default:
		return nil, fmt.Errorf("unexpected opcode: %v", opcode)
	}
	return &pickExecutor{n: n}, nil
}

func (p *pickExecutor) Execute(ctx *OpcodeExecutorContext) error {
	idx := len(ctx.Stack) - 1 - int(uint8(p.n))
	if idx < 0 || idx >= len(ctx.Stack) {
		return fmt.Errorf("stack index out of bounds: %d/%d", idx, len(ctx.Stack))
	}
	ctx.Stack = append(ctx.Stack, ctx.Stack[idx])
	return nil
}

type swapExecutor struct{}

func newSwapExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	return &swapExecutor{}, nil
}

func (s *swapExecutor) Execute(ctx *OpcodeExecutorContext) error {
	if len(ctx.Stack) < 2 {
		return fmt.Errorf("stack underflow: expected value on stack")
	}
	ctx.Stack[len(ctx.Stack)-1], ctx.Stack[len(ctx.Stack)-2] = ctx.Stack[len(ctx.Stack)-2], ctx.Stack[len(ctx.Stack)-1]
	return nil
}

type rotExecutor struct {
}

func newRotExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	return &rotExecutor{}, nil
}

func (r *rotExecutor) Execute(ctx *OpcodeExecutorContext) error {
	if len(ctx.Stack) < 3 {
		return fmt.Errorf("stack underflow: expected value on stack")
	}
	ctx.Stack[len(ctx.Stack)-1], ctx.Stack[len(ctx.Stack)-2], ctx.Stack[len(ctx.Stack)-3] = ctx.Stack[len(ctx.Stack)-2], ctx.Stack[len(ctx.Stack)-3], ctx.Stack[len(ctx.Stack)-1]
	return nil
}

type unaryOpExecutor struct {
	opcode Opcode
}

func newUnaryOpExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	return &unaryOpExecutor{opcode: opcode}, nil
}

func (u *unaryOpExecutor) Execute(ctx *OpcodeExecutorContext) error {
	if len(ctx.Stack) < 1 {
		return fmt.Errorf("stack underflow: expected value on stack")
	}
	operand := ctx.Stack[len(ctx.Stack)-1]
	switch u.opcode {
	case DW_OP_abs:
		if operand < 0 {
			operand = -operand
		}
	case DW_OP_neg:
		operand = -operand
	case DW_OP_not:
		operand = ^operand
	default:
		return fmt.Errorf("unexpected opcode: %v", u.opcode)
	}
	ctx.Stack[len(ctx.Stack)-1] = operand
	return nil
}

type binaryOpExecutor struct {
	opcode Opcode
}

func newBinaryOpExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	return &binaryOpExecutor{opcode: opcode}, nil
}

func (b *binaryOpExecutor) Execute(ctx *OpcodeExecutorContext) error {
	if len(ctx.Stack) < 2 {
		return fmt.Errorf("stack underflow: expected value on stack")
	}
	second := ctx.Stack[len(ctx.Stack)-2]
	top := ctx.Stack[len(ctx.Stack)-1]
	var r int64
	ctx.Stack = ctx.Stack[:len(ctx.Stack)-2]
	switch b.opcode {
	case DW_OP_and:
		r = second & top
	case DW_OP_div:
		r = second / top
	case DW_OP_minus:
		r = second - top
	case DW_OP_mod:
		r = second % top
	case DW_OP_mul:
		r = second * top
	case DW_OP_or:
		r = second | top
	case DW_OP_plus:
		r = second + top
	case DW_OP_shl:
		r = second << uint64(top)
	case DW_OP_shr:
		r = second >> uint64(top)
	case DW_OP_shra:
		r = int64(uint64(second) >> uint64(top))
	case DW_OP_xor:
		r = second ^ top
	case DW_OP_le:
		r = boolToInt(second <= top)
	case DW_OP_ge:
		r = boolToInt(second >= top)
	case DW_OP_eq:
		r = boolToInt(second == top)
	case DW_OP_lt:
		r = boolToInt(second < top)
	case DW_OP_gt:
		r = boolToInt(second > top)
	case DW_OP_ne:
		r = boolToInt(second != top)
	default:
		return fmt.Errorf("unexpected opcode: %v", b.opcode)
	}
	ctx.Stack = append(ctx.Stack, r)
	return nil
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

type skipExecutor struct{}

func newSkipExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	var n int16
	err := binary.Read(ctx.buf, binary.LittleEndian, &n)
	if err != nil {
		return nil, err
	}
	if err = ctx.jump(n); err != nil {
		return nil, err
	}
	return &skipExecutor{}, nil
}

func (c *OpcodeExecutorCreatorContext) jump(n int16) error {
	i := len(c.prog) - c.buf.Len() + int(n)
	if i < 0 {
		return fmt.Errorf("stack underflow")
	}
	if i >= len(c.prog) {
		i = len(c.prog)
	}
	c.buf = bytes.NewBuffer(c.prog[i:])
	return nil
}

func (s *skipExecutor) Execute(ctx *OpcodeExecutorContext) error {
	return nil
}

type braExecutor struct {
	n           int16
	withJump    []OpcodeExecutor
	withoutJump []OpcodeExecutor
}

func newBraExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	var n int16
	err := binary.Read(ctx.buf, binary.LittleEndian, &n)
	if err != nil {
		return nil, err
	}

	b := &braExecutor{n: n}
	bufBytes := make([]byte, ctx.buf.Len())
	_, err = ctx.buf.Read(bufBytes)
	if err != nil {
		return nil, err
	}

	b.withJump, err = b.buildBraExecutors(ctx, bufBytes, true)
	if err != nil {
		return nil, err
	}

	b.withoutJump, err = b.buildBraExecutors(ctx, bufBytes, false)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (b braExecutor) buildBraExecutors(ctx *OpcodeExecutorCreatorContext, bufBytes []byte, withJump bool) ([]OpcodeExecutor, error) {
	var executors []OpcodeExecutor
	bytesCopy := make([]byte, len(bufBytes))
	copy(bytesCopy, bufBytes)
	ctx.buf = bytes.NewBuffer(bytesCopy)

	if withJump {
		err := ctx.jump(b.n)
		if err != nil {
			return nil, err
		}
	}

	for i := 0; i < len(bufBytes); i++ {
		opcodeByte, err := ctx.buf.ReadByte()
		if err != nil {
			break
		}
		opcode := Opcode(opcodeByte)
		if opcode == DW_OP_nop {
			continue
		}
		executorCreator, ok := OpcodeToExecutorCreator(opcode)
		if !ok {
			return nil, fmt.Errorf("invalid instruction %#v", opcode)
		}

		executor, err := executorCreator(opcode, ctx)
		if err != nil {
			return nil, err
		}
		executors = append(executors, executor)
	}

	return executors, nil
}

func (b braExecutor) Execute(ctx *OpcodeExecutorContext) error {
	if len(ctx.Stack) < 1 {
		return fmt.Errorf("stack underflow: expected value on context stack")
	}

	top := ctx.Stack[len(ctx.Stack)-1]
	ctx.Stack = ctx.Stack[:len(ctx.Stack)-1]
	executors := b.withoutJump
	if top != 0 {
		executors = b.withJump
	}

	for _, executor := range executors {
		err := executor.Execute(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

type stackValueExecutor struct {
	closeLoc *CloseLoc
}

func newStackValueExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	piece := Piece{Kind: ImmPiece}
	closeLoc := newCloseLocExecutor(ctx.buf, piece)
	return &stackValueExecutor{closeLoc: closeLoc}, nil
}

func (s *stackValueExecutor) Execute(ctx *OpcodeExecutorContext) error {
	if len(ctx.Stack) < 1 {
		return fmt.Errorf("stack underflow: expected value on context stack")
	}
	val := ctx.Stack[len(ctx.Stack)-1]
	ctx.Stack = ctx.Stack[:len(ctx.Stack)-1]
	s.closeLoc.piece.Val = uint64(val)
	return s.closeLoc.Execute(ctx)
}

func newImplicitValueExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	sz, _ := util.DecodeULEB128(ctx.buf)
	block := make([]byte, sz)
	n, _ := ctx.buf.Read(block)
	if uint64(n) != sz {
		return nil, fmt.Errorf("insufficient bytes read while reading DW_OP_implicit_value's block %d (expected: %d)", n, sz)
	}
	return newCloseLocExecutor(ctx.buf, Piece{Kind: ImmPiece, Bytes: block, Size: int(sz)}), nil
}

type derefExecutor struct {
	size   int
	opcode Opcode
}

func newDerefExecutor(opcode Opcode, ctx *OpcodeExecutorCreatorContext) (OpcodeExecutor, error) {
	size := ctx.pointerSize
	if opcode == DW_OP_deref_size || opcode == DW_OP_xderef_size {
		n, err := ctx.buf.ReadByte()
		if err != nil {
			return nil, err
		}
		size = int(n)
	}
	return &derefExecutor{size: size, opcode: opcode}, nil
}

func (d *derefExecutor) Execute(ctx *OpcodeExecutorContext) error {
	if ctx.readMemory == nil {
		return fmt.Errorf("memory read is unavailable")
	}
	if len(ctx.Stack) <= 0 {
		return fmt.Errorf("stack underflow: expected a value on context stack")
	}

	addr := ctx.Stack[len(ctx.Stack)-1]
	ctx.Stack = ctx.Stack[:len(ctx.Stack)-1]
	if d.opcode == DW_OP_xderef || d.opcode == DW_OP_xderef_size {
		if len(ctx.Stack) <= 0 {
			return fmt.Errorf("stack underflow: expected a value on context stack")
		}
		
		ctx.Stack = ctx.Stack[:len(ctx.Stack)-1]
	}

	buf := make([]byte, d.size)
	_, err := ctx.readMemory(buf, uint64(addr))
	if err != nil {
		return err
	}
	x, err := util.ReadUintRaw(bytes.NewReader(buf), binary.LittleEndian, d.size)
	if err != nil {
		return err
	}

	ctx.Stack = append(ctx.Stack, int64(x))
	return nil
}

func OpcodeToExecutorCreator(opcode Opcode) (OpcodeExecutorCreator, bool) {
	switch opcode {
	case DW_OP_addr:
		return newAddrExecutor, true
	case DW_OP_deref, DW_OP_xderef, DW_OP_deref_size, DW_OP_xderef_size:
		return newDerefExecutor, true
	case DW_OP_const1s, DW_OP_const2s, DW_OP_const4s, DW_OP_const8s:
		return newConstnsExecutor, true
	case DW_OP_const1u, DW_OP_const2u, DW_OP_const4u, DW_OP_const8u:
		return newConstnuExecutor, true
	case DW_OP_constu:
		return newConstuExecutor, true
	case DW_OP_consts:
		return newConstsExecutor, true
	case DW_OP_dup:
		return newDupExecutor, true
	case DW_OP_drop:
		return newDropExecutor, true
	case DW_OP_over, DW_OP_pick:
		return newPickExecutor, true
	case DW_OP_swap:
		return newSwapExecutor, true
	case DW_OP_rot:
		return newRotExecutor, true
	case DW_OP_abs:
		return newUnaryOpExecutor, true
	case DW_OP_and, DW_OP_div, DW_OP_minus, DW_OP_mod, DW_OP_mul, DW_OP_or, DW_OP_plus, DW_OP_shl, DW_OP_shr, DW_OP_shra,
		DW_OP_xor, DW_OP_eq, DW_OP_ge, DW_OP_gt, DW_OP_le, DW_OP_lt, DW_OP_ne:
		return newBinaryOpExecutor, true
	case DW_OP_neg, DW_OP_not:
		return newUnaryOpExecutor, true
	case DW_OP_plus_uconst:
		return newPlusUconstsExecutor, true
	case DW_OP_skip:
		return newSkipExecutor, true
	case DW_OP_lit0, DW_OP_lit1, DW_OP_lit2, DW_OP_lit3, DW_OP_lit4, DW_OP_lit5, DW_OP_lit6, DW_OP_lit7, DW_OP_lit8,
		DW_OP_lit9, DW_OP_lit10, DW_OP_lit11, DW_OP_lit12, DW_OP_lit13, DW_OP_lit14, DW_OP_lit15, DW_OP_lit16, DW_OP_lit17,
		DW_OP_lit18, DW_OP_lit19, DW_OP_lit20, DW_OP_lit21, DW_OP_lit22, DW_OP_lit23, DW_OP_lit24, DW_OP_lit25, DW_OP_lit26,
		DW_OP_lit27, DW_OP_lit28, DW_OP_lit29, DW_OP_lit30, DW_OP_lit31:
		return newLiteralExecutor, true
	case DW_OP_reg0, DW_OP_reg1, DW_OP_reg2, DW_OP_reg3, DW_OP_reg4, DW_OP_reg5, DW_OP_reg6, DW_OP_reg7, DW_OP_reg8,
		DW_OP_reg9, DW_OP_reg10, DW_OP_reg11, DW_OP_reg12, DW_OP_reg13, DW_OP_reg14, DW_OP_reg15, DW_OP_reg16, DW_OP_reg17,
		DW_OP_reg18, DW_OP_reg19, DW_OP_reg20, DW_OP_reg21, DW_OP_reg22, DW_OP_reg23, DW_OP_reg24, DW_OP_reg25, DW_OP_reg26,
		DW_OP_reg27, DW_OP_reg28, DW_OP_reg29, DW_OP_reg30, DW_OP_reg31, DW_OP_regx:
		return newRegisterExecutor, true
	case DW_OP_breg0, DW_OP_breg1, DW_OP_breg2, DW_OP_breg3, DW_OP_breg4, DW_OP_breg5, DW_OP_breg6, DW_OP_breg7,
		DW_OP_breg8, DW_OP_breg9, DW_OP_breg10, DW_OP_breg11, DW_OP_breg12, DW_OP_breg13, DW_OP_breg14, DW_OP_breg15,
		DW_OP_breg16, DW_OP_breg17, DW_OP_breg18, DW_OP_breg19, DW_OP_breg20, DW_OP_breg21, DW_OP_breg22, DW_OP_breg23,
		DW_OP_breg24, DW_OP_breg25, DW_OP_breg26, DW_OP_breg27, DW_OP_breg28, DW_OP_breg29, DW_OP_breg30, DW_OP_breg31,
		DW_OP_bregx:
		return newBRegisterExecutor, true
	case DW_OP_fbreg:
		return newFramebaseExecutor, true
	case DW_OP_piece:
		return newPieceExecutor, true
	case DW_OP_call_frame_cfa:
		return newCallframeCFAExecutor, true
	case DW_OP_implicit_value:
		return newImplicitValueExecutor, true
	case DW_OP_stack_value:
		return newStackValueExecutor, true
	case DW_OP_bra:
		return newBraExecutor, true
	default:
		return nil, false
	}
}
