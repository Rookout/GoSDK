package collection

import (
	"errors"
	"fmt"

	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/collection/registers"
	"github.com/Rookout/GoSDK/pkg/services/collection/variable"
)

type Stackframe struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function,omitempty"`
}

type CollectionService struct {
	variables              []*variable.Variable
	StackTraceElements     []Stackframe
	variableLocators       []*variable.VariableLocator
	variablesCache         map[variable.VariablesCacheKey]variable.VariablesCacheValue
	dictVariableLocator    *variable.VariableLocator
	shouldLoadDictVariable bool
	dictAddr               uint64
	regs                   registers.Registers
	pointerSize            int
	goid                   int
}

const goDictionaryName = ".dict"

func NewCollectionService(regs registers.Registers, pointerSize int, stackTraceElements []Stackframe, variableLocators []*variable.VariableLocator, goid int) (*CollectionService, rookoutErrors.RookoutError) {
	c := &CollectionService{
		StackTraceElements:     stackTraceElements,
		regs:                   regs,
		shouldLoadDictVariable: false,
		pointerSize:            pointerSize,
		goid:                   goid,
		variablesCache:         make(map[variable.VariablesCacheKey]variable.VariablesCacheValue),
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
		dictVar := c.dictVariableLocator.Locate(c.regs, 0, c.variablesCache, config)
		dictAddr := dictVar.Addr

		c.dictAddr = uint64(dictAddr)
		c.shouldLoadDictVariable = false
	}
}

func (c *CollectionService) GetVariables(config config.ObjectDumpConfig) []*variable.Variable {
	fmt.Printf("Getting %d variables\n", len(c.variableLocators))
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

func (c *CollectionService) locateAndLoadVariable(varLocator *variable.VariableLocator, config config.ObjectDumpConfig) (v *variable.Variable) {
	fmt.Printf("Locating %s\n", varLocator.VariableName)
	v = varLocator.Locate(c.regs, c.dictAddr, c.variablesCache, config)
	if name := v.Name; len(name) > 1 && name[0] == '&' {
		v = v.MaybeDereference()
		if v.Addr == 0 && v.Unreadable == nil {
			v.Unreadable = fmt.Errorf("no address for escaped variable")
		}
		v.Name = name[1:]
	}

	if v.ObjectDumpConfig.ShouldTailor {
		v.ObjectDumpConfig.Tailor(v.Kind, int(v.Len))
	}

	fmt.Printf("Loading %s\n", varLocator.VariableName)
	v.LoadValue()
	c.variables = append(c.variables, v)
	fmt.Printf("Added %s\n", varLocator.VariableName)
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
