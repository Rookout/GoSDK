package prologue

import (
	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/assembler"
	"github.com/Rookout/GoSDK/pkg/services/disassembler"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/hooker/regbackup"
)


var ForceLongPrologue = false

const startLabel = "prologueStart"
const endLabel = "prologueEnd"
const fallbackLabel = "fallback"

type Generator struct {
	stackUsage           int
	fallbackAddr         uintptr
	goPrologueExists     bool
	epilogueInstructions []*disassembler.Instruction
	getRegsUsed          func() ([]assembler.Reg, rookoutErrors.RookoutError)
	morestackAddr        uintptr
}


func NewGenerator(funcEntry uintptr, funcEnd uintptr, stackUsage int, fallbackAddr uintptr, getRegsUsed func() ([]assembler.Reg, rookoutErrors.RookoutError)) (*Generator, rookoutErrors.RookoutError) {
	instructions, err := disassembler.Decode(funcEntry, funcEnd, true)
	if err != nil {
		return nil, err
	}

	g := &Generator{stackUsage: stackUsage, fallbackAddr: fallbackAddr, getRegsUsed: getRegsUsed}
	g.epilogueInstructions, g.morestackAddr, g.goPrologueExists = getOriginalEpilogue(instructions)
	if g.morestackAddr == 0 {
		g.morestackAddr = morestackAddr
	}
	return g, nil
}

func (g *Generator) Generate() ([]byte, rookoutErrors.RookoutError) {
	if !g.goPrologueExists || ForceLongPrologue {
		logger.Logger().Debug("Generating long prologue")
		return g.generateLongPrologue()
	}
	logger.Logger().Debug("Generating short prologue")
	return g.generateShortPrologue()
}

func (g *Generator) generateShortPrologue() ([]byte, rookoutErrors.RookoutError) {
	regBackup, regRestore := g.getOriginalRegBackup()
	b := assembler.NewBuilder()

	err := g.generateCheckStackUsage(b)
	if err != nil {
		return nil, err
	}
	err = b.AddInstructions(b.Bytes(regBackup))
	if err != nil {
		return nil, err
	}
	err = g.generateCallMorestack(b)
	if err != nil {
		return nil, err
	}
	err = b.AddInstructions(b.Bytes(regRestore))
	if err != nil {
		return nil, err
	}
	err = g.generateJumpToStart(b)
	if err != nil {
		return nil, err
	}

	return b.Assemble()
}

func (g *Generator) generateLongPrologue() ([]byte, rookoutErrors.RookoutError) {
	regsToUpdate, err := g.getRegsUsed()
	if err != nil {
		return nil, err
	}
	logger.Logger().Debugf("Updating regs: %v\n", regsToUpdate)

	regBackupGenerator := regbackup.NewGenerator(regsBackupBuffer, fallbackLabel, regsToUpdate)
	b := assembler.NewBuilder()

	err = g.generateCheckStackUsage(b)
	if err != nil {
		return nil, err
	}

	err = regBackupGenerator.GenerateRegBackup(b)
	if err != nil {
		return nil, err
	}
	err = g.generateCallMorestack(b)
	if err != nil {
		return nil, err
	}
	err = regBackupGenerator.GenerateRegRestore(b)
	if err != nil {
		return nil, err
	}

	err = g.generateCallFallback(b)
	if err != nil {
		return nil, err
	}
	err = g.generateJumpToStart(b)
	if err != nil {
		return nil, err
	}

	return b.Assemble()
}


func findFirstJmp(instructions []*disassembler.Instruction, startIndex int) (jmpIndex int, jmpDestIndex int, ok bool) {
	for i := startIndex; i < len(instructions); i++ {
		jmp, jmpIndex, ok := disassembler.GetFirstInstruction(instructions[i:], disassembler.IsDirectJump)
		if !ok {
			return 0, 0, false
		}
		jmpIndex += i

		jmpDest, err := jmp.GetDestPC()
		if err != nil {
			logger.Logger().WithError(err).Warningf("Failed to get first jmp dest PC")
			return 0, 0, false
		}

		destFound := false
		for j, inst := range instructions {
			if inst.PC == jmpDest {
				destFound = true
				jmpDestIndex = j
				break
			}
		}
		if !destFound {
			logger.Logger().Warningf("Unable to find destination instruction of jump. Jump PC: 0x%x, dest PC: 0x%x", jmp.PC, jmpDest)
			return 0, 0, false
		}

		
		if jmpDestIndex == jmpIndex+1 {
			continue
		}

		return jmpIndex, jmpDestIndex, true
	}

	return 0, 0, false
}

func findMorestackCall(instructions []*disassembler.Instruction, startIndex int) (callIndex int, callDest uintptr, ok bool) {
	call, callIndex, ok := disassembler.GetFirstInstruction(instructions[startIndex:], disassembler.IsDirectCall)
	if !ok {
		logger.Logger().Debug("Epilogue not found: no call in epilogue")
		return 0, 0, false
	}
	callIndex += startIndex

	callDest, err := call.GetDestPC()
	if err != nil {
		logger.Logger().WithError(err).Warningf("Failed to get call dest PC")
		return 0, 0, false
	}
	if _, ok := morestackAddrs[callDest]; !ok {
		logger.Logger().Debugf("Epilogue not found: call is not to morestack. callDest: %x, morestackAddrs: %x", callDest, morestackAddrs)
		return 0, 0, false
	}

	return callIndex, callDest, true
}

func findJmpToStart(instructions []*disassembler.Instruction, epilogueStart int) (jmpIndex int, ok bool) {
	jmpIndex, jmpDestIndex, ok := findFirstJmp(instructions, epilogueStart)
	if !ok || jmpDestIndex != 0 {
		return 0, false
	}
	return jmpIndex, true
}

func getOriginalEpilogue(instructions []*disassembler.Instruction) ([]*disassembler.Instruction, uintptr, bool) {
	_, epilogueStart, ok := findFirstJmp(instructions, 0)
	if !ok {
		return nil, 0, false
	}

	morestackCallIndex, morestackAddr, ok := findMorestackCall(instructions, epilogueStart)
	if !ok {
		return nil, 0, false
	}

	epilogueEnd, ok := findJmpToStart(instructions, morestackCallIndex)
	if !ok {
		return nil, 0, false
	}

	
	return instructions[epilogueStart:epilogueEnd], morestackAddr, true
}
