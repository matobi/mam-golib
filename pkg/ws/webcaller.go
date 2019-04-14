package ws

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/matobi/mam-golib/pkg/errid"
)

const (
	ContentJSON  = "application/json"
	ContentXML   = "application/xml"
	ContentPlain = "text/plain"
)

type Caller struct {
	Method      string
	URL         string
	contentType string
	accept      string
	user        string
	pwd         string
	headers     map[string]string
}

type input struct {
	contentType string
	buf         *bytes.Buffer
}

func NewCaller(method, url string) *Caller {
	return &Caller{
		Method:  method,
		URL:     url,
		headers: make(map[string]string),
	}
}

func (c *Caller) SetHeader(name, value string) *Caller {
	c.headers[name] = value
	return c
}

func (c *Caller) GetHeader(name string) (string, bool) {
	v, found := c.headers[name]
	return v, found
}

func (c *Caller) Accept(t string) *Caller {
	c.accept = t
	return c
}
func (c *Caller) Content(t string) *Caller {
	c.contentType = t
	return c
}

func (c *Caller) JSON() *Caller {
	c.contentType = ContentJSON
	c.accept = ContentJSON
	return c
}

func (c *Caller) XML() *Caller {
	c.contentType = ContentXML
	c.accept = ContentXML
	return c
}

func (c *Caller) Plain() *Caller {
	c.contentType = ContentPlain
	c.accept = ContentPlain
	return c
}

func (c *Caller) Auth(user, pwd string) *Caller {
	c.user = user
	c.pwd = pwd
	return c
}

func (c *Caller) getInBuffer(in interface{}) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	if in == nil {
		return buf, nil
	}
	if c.contentType == ContentJSON {
		if err := json.NewEncoder(buf).Encode(in); err != nil {
			return nil, errid.New("failed encode json").Cause(err).URL(c.URL)
		}
	} else if c.contentType == ContentXML {
		if err := xml.NewEncoder(buf).Encode(in); err != nil {
			return nil, errid.New("failed encode xml").Cause(err).URL(c.URL)
		}
	}
	return buf, nil
}

func (c *Caller) Call(client *http.Client, in interface{}, out interface{}) error {
	buf, webErr := c.getInBuffer(in)
	if webErr != nil {
		return webErr
	}

	req, err := http.NewRequest(c.Method, c.URL, buf)
	if err != nil {
		return errid.New("failed create request").Cause(err).URL(c.URL)
	}
	if in != nil && c.contentType != "" {
		req.Header.Set("Content-Type", c.contentType)
	}
	if out != nil && c.accept != "" {
		req.Header.Set("Accept", c.accept)
	}
	if c.user != "" || c.pwd != "" {
		req.SetBasicAuth(c.user, c.pwd)
	}
	for k, v := range c.headers {
		fmt.Printf("add header: %s=%s\n", k, v)
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("err=%v\n", err)
		fmt.Printf("c=%p\n", c)
		fmt.Printf("c.URL=%s\n", c.URL)
		return errid.New("failed call url").Cause(err).URL(c.URL)
	}
	defer resp.Body.Close()

	// save out headers
	for k, v := range resp.Header {
		if len(v) != 1 {
			continue // todo: currently only handles single value headers
		}
		c.headers[k] = v[0]
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		DiscardBody(resp)
		isTemp := resp.StatusCode != http.StatusBadRequest
		msg := fmt.Sprintf("http err reply: %s", resp.Status)
		return errid.New(msg).URL(c.URL).Code(resp.StatusCode).Temp(isTemp)
	}

	if out == nil {
		DiscardBody(resp)
		return nil
	}

	if c.accept == ContentXML {
		if err := xml.NewDecoder(resp.Body).Decode(out); err != nil {
			return errid.New("failed decode reply xml").Cause(err).URL(c.URL)
		}
	} else if c.accept == ContentJSON {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return errid.New("failed decode reply json").Cause(err).URL(c.URL)
		}
	} else if c.accept == ContentPlain {
		if v, ok := out.(*string); ok {
			buf, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return errid.New("failed read response").Cause(err).URL(c.URL)
			}
			*v = string(buf)
		}
	}
	return nil
}

func DiscardBody(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
}
