package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"sort"
	"time"
)

// writeFilterBlocks generates the filter syntax for each item tier based on dynamic configuration.
func writeFilterBlocks(cfg Config, valueMap map[string]map[string]float64) string {
	var buf bytes.Buffer

	// 1. Prepare and sort tiers by absolute Chaos value (descending)
	type sortedTier struct {
		Tier        Tier
		AbsChaosVal float64
	}
	var sorted []sortedTier
	for _, t := range cfg.Tiers {
		absVal := t.Value
		switch t.Currency {
		case "Exalted":
			absVal *= ExaltedPrice
		case "Divine":
			absVal *= DivinePrice
		}
		sorted = append(sorted, sortedTier{Tier: t, AbsChaosVal: absVal})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].AbsChaosVal > sorted[j].AbsChaosVal
	})

	// 2. Map styles for quick lookup
	styleMap := make(map[string]Style)
	for _, s := range cfg.StyleLibrary {
		styleMap[s.Name] = s
	}

	// 3. Generate blocks for each item
	for i := 0; i < len(typeSlice); i++ {
		category := typeSlice[i]
		for name, itemPrice := range valueMap[category] {
			if itemPrice <= 0 {
				continue
			}

			for _, st := range sorted {
				buf.WriteString(fmt.Sprintf("\n## %s Tier ##\n", st.Tier.Name))
				buf.WriteString("Show\n")
				buf.WriteString(fmt.Sprintf("BaseType == \"%s\"\n", name))

				// Calculate stack size to match this tier's absolute chaos value
				// StackSize * itemPrice >= st.AbsChaosVal
				// StackSize >= st.AbsChaosVal / itemPrice
				stackSize := math.Ceil(st.AbsChaosVal / itemPrice)
				if stackSize < 1 {
					stackSize = 1
				}
				buf.WriteString(fmt.Sprintf("StackSize >= %d\n", int(stackSize)))

				// Apply actions from the referenced style
				if style, ok := styleMap[st.Tier.StyleName]; ok {
					for _, action := range style.Actions {
						buf.WriteString(action.ToFilterLine() + "\n")
					}
				}
			}

			// Final Hide block for this item (to ensure it's hidden if it doesn't meet any tier)
			buf.WriteString("\nHide\n")
			buf.WriteString(fmt.Sprintf("BaseType == \"%s\"\n", name))
		}
	}

	return buf.String()
}

// updateFilterFile generates the new filter content and writes it to the output file.
// It places the auto-generated economy rules at the top, followed by the entirety of the base filter.
func updateFilterFile(basePath string, outputPath string, blocks ...string) error {
	var buf bytes.Buffer

	// 1. Add header for the auto-generated section
	buf.WriteString("# PoEAutoFilter - Dynamic Economy Rules\n")
	buf.WriteString("# Generated: " + time.Now().Format(time.RFC1123) + "\n")
	buf.WriteString("# Everything below this header and above the Base Filter section is auto-generated.\n\n")

	// 2. Append all generated blocks
	for _, block := range blocks {
		buf.WriteString(block)
	}

	// 3. Add separator and then the base filter
	buf.WriteString("\n\n#===============================================================================================================\n")
	buf.WriteString("# SECTION: Base Filter Content (Template)\n")
	buf.WriteString("# Everything below this line is copied directly from your selected base filter.\n")
	buf.WriteString("#===============================================================================================================\n\n")

	if basePath != "" {
		content, err := os.ReadFile(basePath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to read base filter: %w", err)
		}
		buf.Write(content)
	}

	// 4. Write final content to output file
	return os.WriteFile(outputPath, buf.Bytes(), 0644)
}
