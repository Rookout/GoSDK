// The MIT License (MIT)

// Copyright (c) 2014 Derek Parker

// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package op

const (
	DW_OP_addr                Opcode = 0x03
	DW_OP_deref               Opcode = 0x06
	DW_OP_const1u             Opcode = 0x08
	DW_OP_const1s             Opcode = 0x09
	DW_OP_const2u             Opcode = 0x0a
	DW_OP_const2s             Opcode = 0x0b
	DW_OP_const4u             Opcode = 0x0c
	DW_OP_const4s             Opcode = 0x0d
	DW_OP_const8u             Opcode = 0x0e
	DW_OP_const8s             Opcode = 0x0f
	DW_OP_constu              Opcode = 0x10
	DW_OP_consts              Opcode = 0x11
	DW_OP_dup                 Opcode = 0x12
	DW_OP_drop                Opcode = 0x13
	DW_OP_over                Opcode = 0x14
	DW_OP_pick                Opcode = 0x15
	DW_OP_swap                Opcode = 0x16
	DW_OP_rot                 Opcode = 0x17
	DW_OP_xderef              Opcode = 0x18
	DW_OP_abs                 Opcode = 0x19
	DW_OP_and                 Opcode = 0x1a
	DW_OP_div                 Opcode = 0x1b
	DW_OP_minus               Opcode = 0x1c
	DW_OP_mod                 Opcode = 0x1d
	DW_OP_mul                 Opcode = 0x1e
	DW_OP_neg                 Opcode = 0x1f
	DW_OP_not                 Opcode = 0x20
	DW_OP_or                  Opcode = 0x21
	DW_OP_plus                Opcode = 0x22
	DW_OP_plus_uconst         Opcode = 0x23
	DW_OP_shl                 Opcode = 0x24
	DW_OP_shr                 Opcode = 0x25
	DW_OP_shra                Opcode = 0x26
	DW_OP_xor                 Opcode = 0x27
	DW_OP_bra                 Opcode = 0x28
	DW_OP_eq                  Opcode = 0x29
	DW_OP_ge                  Opcode = 0x2a
	DW_OP_gt                  Opcode = 0x2b
	DW_OP_le                  Opcode = 0x2c
	DW_OP_lt                  Opcode = 0x2d
	DW_OP_ne                  Opcode = 0x2e
	DW_OP_skip                Opcode = 0x2f
	DW_OP_lit0                Opcode = 0x30
	DW_OP_lit1                Opcode = 0x31
	DW_OP_lit2                Opcode = 0x32
	DW_OP_lit3                Opcode = 0x33
	DW_OP_lit4                Opcode = 0x34
	DW_OP_lit5                Opcode = 0x35
	DW_OP_lit6                Opcode = 0x36
	DW_OP_lit7                Opcode = 0x37
	DW_OP_lit8                Opcode = 0x38
	DW_OP_lit9                Opcode = 0x39
	DW_OP_lit10               Opcode = 0x3a
	DW_OP_lit11               Opcode = 0x3b
	DW_OP_lit12               Opcode = 0x3c
	DW_OP_lit13               Opcode = 0x3d
	DW_OP_lit14               Opcode = 0x3e
	DW_OP_lit15               Opcode = 0x3f
	DW_OP_lit16               Opcode = 0x40
	DW_OP_lit17               Opcode = 0x41
	DW_OP_lit18               Opcode = 0x42
	DW_OP_lit19               Opcode = 0x43
	DW_OP_lit20               Opcode = 0x44
	DW_OP_lit21               Opcode = 0x45
	DW_OP_lit22               Opcode = 0x46
	DW_OP_lit23               Opcode = 0x47
	DW_OP_lit24               Opcode = 0x48
	DW_OP_lit25               Opcode = 0x49
	DW_OP_lit26               Opcode = 0x4a
	DW_OP_lit27               Opcode = 0x4b
	DW_OP_lit28               Opcode = 0x4c
	DW_OP_lit29               Opcode = 0x4d
	DW_OP_lit30               Opcode = 0x4e
	DW_OP_lit31               Opcode = 0x4f
	DW_OP_reg0                Opcode = 0x50
	DW_OP_reg1                Opcode = 0x51
	DW_OP_reg2                Opcode = 0x52
	DW_OP_reg3                Opcode = 0x53
	DW_OP_reg4                Opcode = 0x54
	DW_OP_reg5                Opcode = 0x55
	DW_OP_reg6                Opcode = 0x56
	DW_OP_reg7                Opcode = 0x57
	DW_OP_reg8                Opcode = 0x58
	DW_OP_reg9                Opcode = 0x59
	DW_OP_reg10               Opcode = 0x5a
	DW_OP_reg11               Opcode = 0x5b
	DW_OP_reg12               Opcode = 0x5c
	DW_OP_reg13               Opcode = 0x5d
	DW_OP_reg14               Opcode = 0x5e
	DW_OP_reg15               Opcode = 0x5f
	DW_OP_reg16               Opcode = 0x60
	DW_OP_reg17               Opcode = 0x61
	DW_OP_reg18               Opcode = 0x62
	DW_OP_reg19               Opcode = 0x63
	DW_OP_reg20               Opcode = 0x64
	DW_OP_reg21               Opcode = 0x65
	DW_OP_reg22               Opcode = 0x66
	DW_OP_reg23               Opcode = 0x67
	DW_OP_reg24               Opcode = 0x68
	DW_OP_reg25               Opcode = 0x69
	DW_OP_reg26               Opcode = 0x6a
	DW_OP_reg27               Opcode = 0x6b
	DW_OP_reg28               Opcode = 0x6c
	DW_OP_reg29               Opcode = 0x6d
	DW_OP_reg30               Opcode = 0x6e
	DW_OP_reg31               Opcode = 0x6f
	DW_OP_breg0               Opcode = 0x70
	DW_OP_breg1               Opcode = 0x71
	DW_OP_breg2               Opcode = 0x72
	DW_OP_breg3               Opcode = 0x73
	DW_OP_breg4               Opcode = 0x74
	DW_OP_breg5               Opcode = 0x75
	DW_OP_breg6               Opcode = 0x76
	DW_OP_breg7               Opcode = 0x77
	DW_OP_breg8               Opcode = 0x78
	DW_OP_breg9               Opcode = 0x79
	DW_OP_breg10              Opcode = 0x7a
	DW_OP_breg11              Opcode = 0x7b
	DW_OP_breg12              Opcode = 0x7c
	DW_OP_breg13              Opcode = 0x7d
	DW_OP_breg14              Opcode = 0x7e
	DW_OP_breg15              Opcode = 0x7f
	DW_OP_breg16              Opcode = 0x80
	DW_OP_breg17              Opcode = 0x81
	DW_OP_breg18              Opcode = 0x82
	DW_OP_breg19              Opcode = 0x83
	DW_OP_breg20              Opcode = 0x84
	DW_OP_breg21              Opcode = 0x85
	DW_OP_breg22              Opcode = 0x86
	DW_OP_breg23              Opcode = 0x87
	DW_OP_breg24              Opcode = 0x88
	DW_OP_breg25              Opcode = 0x89
	DW_OP_breg26              Opcode = 0x8a
	DW_OP_breg27              Opcode = 0x8b
	DW_OP_breg28              Opcode = 0x8c
	DW_OP_breg29              Opcode = 0x8d
	DW_OP_breg30              Opcode = 0x8e
	DW_OP_breg31              Opcode = 0x8f
	DW_OP_regx                Opcode = 0x90
	DW_OP_fbreg               Opcode = 0x91
	DW_OP_bregx               Opcode = 0x92
	DW_OP_piece               Opcode = 0x93
	DW_OP_deref_size          Opcode = 0x94
	DW_OP_xderef_size         Opcode = 0x95
	DW_OP_nop                 Opcode = 0x96
	DW_OP_push_object_address Opcode = 0x97
	DW_OP_call2               Opcode = 0x98
	DW_OP_call4               Opcode = 0x99
	DW_OP_call_ref            Opcode = 0x9a
	DW_OP_form_tls_address    Opcode = 0x9b
	DW_OP_call_frame_cfa      Opcode = 0x9c
	DW_OP_bit_piece           Opcode = 0x9d
	DW_OP_implicit_value      Opcode = 0x9e
	DW_OP_stack_value         Opcode = 0x9f
)
