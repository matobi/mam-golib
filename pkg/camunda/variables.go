package camunda

import "log"

type Variables struct {
	values  map[string]string
	updated map[string]bool
}

// Get Return variable or empty string if not found.
func (v Variables) Get(name string) string {
	value, found := v.values[name]
	if !found {
		log.Printf("missing camunda variable: %s", name)
	}
	return value
}

// Update Changes value for a variable.
func (v Variables) Update(name, value string) {
	v.values[name] = value
	v.updated[name] = true
}

func getVariables(task jsonTask) Variables {
	vars := Variables{values: make(map[string]string), updated: make(map[string]bool)}
	for name := range task.Vars {
		vars.values[name] = getVarStr(task.Vars, name)
	}
	return vars
}

func updateTaskVariables(vars Variables) map[string]jsonVariable {
	jsonVars := make(map[string]jsonVariable)
	for name := range vars.updated {
		value := vars.Get(name)
		addVarStr(jsonVars, name, value)
	}
	return jsonVars
}

func addVarStr(vars map[string]jsonVariable, name string, value string) {
	vars[name] = jsonVariable{Type: "String", Value: value}
}

func getVarStr(vars map[string]jsonVariable, name string) string {
	arg, hasArg := vars[name]
	if !hasArg {
		return ""
	}
	v, isOk := arg.Value.(string)
	if !isOk {
		return ""
	}
	return v
}
