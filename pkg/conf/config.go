package conf

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type ValueType int

const (
	VtStr ValueType = iota
	VtInt
	VtFile
	VtDir
	VtList
)

type Config struct {
	Values  map[string]string
	Lists   map[string][]string
	Errors  []error
	Profile string
}

func NewConfig(profile string) *Config {
	conf := &Config{
		Values: make(map[string]string),
		Lists:  make(map[string][]string),
		Errors: []error{},
	}
	conf.Values["profile"] = profile
	conf.Profile = profile
	return conf
}

//////// Add functions

func (c *Config) Add(t ValueType, name, value string) {
	c.AddProfile(t, "", name, value)
}

func (c *Config) AddProfile(t ValueType, profile, name, value string) {
	value, ok := c.validate(t, profile, name, value)
	if !ok {
		return
	}
	c.Values[name] = value
}

func (c *Config) AddList(t ValueType, name, value string) {
	c.AddListProfile(t, "", name, value)
}

func (c *Config) AddListProfile(t ValueType, profile, name, value string) {
	value, ok := c.validate(t, profile, name, value)
	if !ok {
		return
	}
	list, isFound := c.Lists[name]
	if !isFound {
		list = []string{}
	}
	list = append(list, value)
	c.Lists[name] = list
	log.Printf("property list; name=%s; value=%s\n", name, value)
}

func (c *Config) validate(t ValueType, profile, name, value string) (string, bool) {
	if profile != "" && profile != c.Profile {
		return "", false // inactive profile. Not an error, but skip this value
	}
	if name == "" {
		c.addErr(fmt.Errorf("propery name was nil; value=%s", value))
		return "", false
	}
	value = c.envOverride(name, value)
	if !c.isValidType(t, value) {
		c.addErr(fmt.Errorf("propery value invalid; name=%s; value=%s", name, value))
		return "", false
	}
	return value, true
}

func (c *Config) isValidType(t ValueType, value string) bool {
	switch t {
	case VtStr:
		return true
	case VtInt:
		_, err := strconv.ParseInt(value, 10, 64)
		return err == nil
	case VtList:
		return true
	case VtDir:
		return isDir(value)
	case VtFile:
		return isFile(value)
	default:
		log.Printf("property type unknown; valueType=%d; value=%s\n", t, value)
		return false
	}
}

func (c *Config) envOverride(name, value string) string {
	envValue, ok := os.LookupEnv(name)
	if !ok {
		return value
	}
	return envValue
}

func (c *Config) AddStrProfile(profile, name, value string) {
	if name == "" {
		c.addErr(fmt.Errorf("propery name was nil; value=%s", value))
		return
	}
	if c.Profile != profile {
		return
	}
	c.Values[name] = value
}

func (c *Config) addErr(err error) {
	c.Errors = append(c.Errors, err)
}

////// Get functions

func (c *Config) Str(name string) string {
	s, isFound := c.Values[name]
	if !isFound {
		log.Printf("property missing; name=%s\n", name)
	}
	return s
}

func (c *Config) StrList(name string) []string {
	list, isFound := c.Lists[name]
	if !isFound {
		log.Printf("property missing; name=%s\n", name)
		list = []string{}
	}
	return list
}

func (c *Config) Int(name string) int64 {
	s := c.Str(name)
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Printf("property not numeric; name=%s; value=%s\n", name, s)
	}
	return n
}

func (c *Config) LogAndValidate() (*Config, error) {
	for k, v := range c.Values {
		logNameValue(k, v)
	}
	for i := range c.Errors {
		log.Printf("bad config; err=%+v\n", c.Errors[i])
	}
	if len(c.Errors) > 0 {
		return c, c.Errors[0]
	}
	return c, nil
}

func logNameValue(name, value string) {
	nameLower := strings.ToLower(name)
	if strings.Contains(nameLower, "pwd") || strings.Contains(nameLower, "password") {
		value = "***" // don't log passwords
	}
	log.Printf("config value; name=%s; value=%s\n", name, value)
}

func isDir(path string) bool {
	f, err := os.Stat(path)
	return err == nil && f.IsDir()
}

func isFile(path string) bool {
	f, err := os.Stat(path)
	return err == nil && !f.IsDir()
}
