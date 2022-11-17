package collection

import (
	"errors"
	"fmt"
	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/services/collection/registers"
	"github.com/Rookout/GoSDK/pkg/services/collection/variable"
	"go/constant"
)

type Stackframe struct {
	File     string    `json:"file"`
	Line     int       `json:"line"`
	Function *Function `json:"function,omitempty"`
	PCs      []uint64  `json:"pcs,omitempty"`
}


type Function struct {
	
	Name   string `json:"name"`
	Value  uint64 `json:"value"`
	Type   byte   `json:"type"`
	GoType uint64 `json:"goType"`
	
	Optimized bool `json:"optimized"`
}

type CollectionService struct {
	
	variables              []*variable.Variable
	StackTraceElements     []Stackframe
	variableLocators       []*variable.VariableLocator
	dictVariableLocator    *variable.VariableLocator
	shouldLoadDictVariable bool
	dictAddr               uint64
	regs                   registers.Registers
	pointerSize            int
	goid                   int 
}

const goDictionaryName = ".dict"

func NewCollectionService(regs registers.Registers, pointerSize int, stackTraceElements []Stackframe, variableLocators []*variable.VariableLocator, goid int) (*CollectionService, error) {
	c := &CollectionService{
		StackTraceElements:     stackTraceElements,
		regs:                   regs,
		shouldLoadDictVariable: false,
		pointerSize:            pointerSize,
		goid:                   goid,
	}

	for _, variableLocator := range variableLocators {
		if variableLocator.VariableName == goDictionaryName {
			c.dictVariableLocator = variableLocator
			c.shouldLoadDictVariable = true
		} else {
			c.variableLocators = append(c.variableLocators, variableLocator)
		}
	}

	return c, nil
}

func (c *CollectionService) GetFrame() *Stackframe {
	return &c.StackTraceElements[0]
}

func (c *CollectionService) loadDictVariable(config config.ObjectDumpConfig) {
	if c.shouldLoadDictVariable {
		dictVar := c.dictVariableLocator.Locate(c.regs, 0, config)
		dictVar.LoadValue()
		dictAddr, _ := constant.Int64Val(dictVar.Value)

		c.dictAddr = uint64(dictAddr)
		c.shouldLoadDictVariable = false
	}
}

func (c *CollectionService) GetVariables(config config.ObjectDumpConfig) []*variable.Variable {
	c.loadDictVariable(config)

	var variables []*variable.Variable
	for _, varLocator := range c.variableLocators {
		variables = append(variables, c.locateAndLoadVariable(varLocator, config))
	}

	return variables
}

func (c *CollectionService) GetVariable(name string, config config.ObjectDumpConfig) (*variable.Variable, error) {
	c.loadDictVariable(config)

	for _, varLocator := range c.variableLocators {
		if varLocator.VariableName == name || varLocator.VariableName == "&"+name {
			return c.locateAndLoadVariable(varLocator, config), nil
		}
	}
	return nil, errors.New("variable not found")
}

func (c *CollectionService) locateAndLoadVariable(varLocator *variable.VariableLocator, config config.ObjectDumpConfig) *variable.Variable {
	v := varLocator.Locate(c.regs, c.dictAddr, config)
	if name := v.Name; len(name) > 1 && name[0] == '&' {
		v = v.MaybeDereference()
		if v.Addr == 0 && v.Unreadable == nil {
			v.Unreadable = fmt.Errorf("no address for escaped variable")
		}
		v.Name = name[1:]
	}

	if v.ObjectDumpConfig.ShouldTailor {
		v.ObjectDumpConfig.Tailor(v.Kind)
	}

	v.LoadValue()
	c.variables = append(c.variables, v)
	return v
}

func (c *CollectionService) Close() error {
	for _, v := range c.variables {
		_ = v.Close()
	}

	return nil
}

func (c *CollectionService) GoroutineID() int {
	return c.goid
}
