package main

import (
	"bytes"
	"fmt"
	"os"
)

// Returns the filter blocks as a single string
func writeFilterBlocks(cfg Config, valueMap map[string]map[string]float64, sub1cMult float64) string {
	var buf bytes.Buffer
	for i := 0; i < len(typeSlice); i++ {
		for name, value := range valueMap[typeSlice[i]] {
			buf.WriteString("\n## Half Div Tier ##\n")
			buf.WriteString("Show\n")
			buf.WriteString(fmt.Sprintf("BaseType == \"%s\"\n", name))
			stacksize := int(0.5 * DivinePrice / value)
			if 0.5*DivinePrice != value {
				stacksize += 1
			}
			buf.WriteString(fmt.Sprintf("StackSize >= %v\n", int(stacksize)))
			for _, style := range cfg.Styles["Divine"] {
				buf.WriteString(style)
			}
			buf.WriteString("\n## Exalted Tier ##\n")
			buf.WriteString("Show\n")
			buf.WriteString(fmt.Sprintf("BaseType == \"%s\"\n", name))
			stacksize = int(ExaltedPrice / value)
			if ExaltedPrice != value {
				stacksize += 1
			}
			buf.WriteString(fmt.Sprintf("StackSize >= %v\n", int(stacksize)))
			for _, style := range cfg.Styles["Exalted"] {
				buf.WriteString(style)
			}
			buf.WriteString("\n## 5 Chaos Tier ##\n")
			buf.WriteString("Show\n")
			buf.WriteString(fmt.Sprintf("BaseType == \"%s\"\n", name))
			stacksize = int(5 * ChaosPrice / value)
			if 5*ChaosPrice != value {
				stacksize += 1
			}
			buf.WriteString(fmt.Sprintf("StackSize >= %v\n", int(stacksize)))
			for _, style := range cfg.Styles["5 Chaos"] {
				buf.WriteString(style)
			}
			buf.WriteString("\n## 1 Chaos Tier ##\n")
			buf.WriteString("Show\n")
			buf.WriteString(fmt.Sprintf("BaseType == \"%s\"\n", name))
			stacksize = int(ChaosPrice / value)
			if ChaosPrice != value {
				stacksize += 1
			}
			buf.WriteString(fmt.Sprintf("StackSize >= %v\n", int(stacksize)))
			for _, style := range cfg.Styles["1 Chaos"] {
				buf.WriteString(style)
			}

			// Disable sub 1 Chaos items
			buf.WriteString("\n## Sub 1 Chaos Tier ##\n")
			buf.WriteString("Show\n")
			buf.WriteString(fmt.Sprintf("BaseType == \"%s\"\n", name))
			buf.WriteString(fmt.Sprintf("StackSize >= %v\n", int(sub1cMult*ChaosPrice/value)+1))
			for _, style := range cfg.Styles["Sub 1 Chaos"] {
				buf.WriteString(style)
			}
			buf.WriteString("\nHide\n")
			buf.WriteString(fmt.Sprintf("BaseType == \"%s\"\n", name))
		}
	}
	return buf.String()
}

// Helper to update the filter file
func updateFilterFile(filename string, blocks ...string) error {
	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	// Find first "#=="
	idx := bytes.Index(content, []byte("#=="))
	if idx != -1 {
		content = content[idx:]
	}
	// Compose new content
	var buf bytes.Buffer
	for _, block := range blocks {
		buf.WriteString(block)
	}
	buf.Write(content)
	// Write back to file
	return os.WriteFile(filename, buf.Bytes(), 0644)
}
