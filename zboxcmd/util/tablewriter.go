package util

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

//WriteTable - Writes string data as a table
func WriteTable(writer io.Writer, header []string, footer []string, data [][]string) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader(header)
	table.SetFooter(footer)
	table.SetBorder(false)
	table.AppendBulk(data)
	table.Render()
}
