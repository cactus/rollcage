package core

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"
)

type WriteFlusher interface {
	io.Writer
	Flush() error
}

type OutputWriter struct {
	w             WriteFlusher
	MachineOutput bool
}

func (ow *OutputWriter) Write(data []byte) (int, error) {
	return ow.w.Write(data)
}

func (ow *OutputWriter) Flush() error {
	return ow.w.Flush()
}

func NewOutputWriter(headers []string, machineOutput bool) *OutputWriter {
	ow := &OutputWriter{
		MachineOutput: machineOutput,
	}

	if machineOutput {
		ow.w = bufio.NewWriter(os.Stdout)
	} else {
		ow.w = tabwriter.NewWriter(os.Stdout, 4, 0, 2, ' ', 0)
	}

	if !ow.MachineOutput {
		capHeaders := []string{}
		for _, h := range headers {
			capHeaders = append(capHeaders, strings.ToUpper(h))
		}

		header := strings.Join(capHeaders, "\t")
		fmt.Fprintln(ow.w, header)
	}
	return ow
}

type TemplateOutputWriter struct {
	BodyTemplate *template.Template
	*OutputWriter
}

func (tw *TemplateOutputWriter) WriteTemplate(data interface{}) error {
	return tw.BodyTemplate.Execute(tw, data)
}

func NewTemplateOutputWriter(headers []string, machineOutput bool) *TemplateOutputWriter {
	titleHeaders := []string{}
	for _, h := range headers {
		titleHeaders = append(titleHeaders, strings.Title(strings.ToLower(h)))
	}

	tplText := "{{." + strings.Join(titleHeaders, "}}\t{{.") + "}}\n"
	tw := &TemplateOutputWriter{
		BodyTemplate: template.Must(template.New("out").Parse(tplText)),
		OutputWriter: NewOutputWriter(headers, machineOutput),
	}

	return tw
}

type ocols []string

type OutputCols struct {
	valid []string
	ocols
}

func (c *OutputCols) Type() string {
	return "string"
}

func (c *OutputCols) String() string {
	return fmt.Sprint(c.ocols)
}

func (c *OutputCols) GetCols() []string {
	return c.ocols
}

func (c *OutputCols) GetValidCols() []string {
	return c.valid
}

func (c *OutputCols) Set(value string) error {
	for _, p := range strings.Split(value, ",") {
		if !StringInSlice(c.valid, p) {
			return fmt.Errorf("Invalid column name '%s'", p)
		}
		c.ocols = append(c.ocols, p)
	}
	return nil
}

func NewOutputCols(valid []string) *OutputCols {
	return &OutputCols{valid, make([]string, 0)}
}
