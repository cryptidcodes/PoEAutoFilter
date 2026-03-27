//go:build windows
// +build windows

// models.go — Windows-only TableModel implementations for lxn/walk GUI.
// These bridge the core.Config data into walk.TableView-compatible models.

package windows

import (
	"fmt"
	"strings"

	"github.com/cryptidcodes/PoEAutoFilter/client/internal/core"

	"github.com/lxn/walk"
)

// --- Style Library Table Model ---

// StyleItem represents a single row in the Style Library table.
type StyleItem struct {
	Name    string
	Preview string
	Style   *core.Style
}

// StyleModel implements walk.TableModel for the Style Library TableView.
type StyleModel struct {
	walk.TableModelBase
	Items []*StyleItem
}

// NewStyleModel creates a StyleModel from the app's style library config.
func NewStyleModel(styles []core.Style) *StyleModel {
	m := &StyleModel{Items: make([]*StyleItem, len(styles))}
	for i := range styles {
		m.Items[i] = &StyleItem{
			Name:    styles[i].Name,
			Preview: styles[i].ToFilterLines(),
			Style:   &styles[i],
		}
	}
	return m
}

func (m *StyleModel) RowCount() int { return len(m.Items) }

func (m *StyleModel) Value(row, col int) interface{} {
	if row < 0 || row >= len(m.Items) {
		return ""
	}
	item := m.Items[row]
	switch col {
	case 0:
		return item.Name
	case 1:
		preview := strings.ReplaceAll(item.Preview, "\n", "; ")
		if len(preview) > 50 {
			return preview[:47] + "..."
		}
		return preview
	}
	return ""
}

// --- Value Tiers Table Model ---

// TierItem represents a single row in the Value Tiers table.
type TierItem struct {
	Name     string
	Value    string
	Currency string
	Style    string
	Tier     *core.Tier
}

// TierModel implements walk.TableModel for the Value Tiers TableView.
type TierModel struct {
	walk.TableModelBase
	Items []*TierItem
}

// NewTierModel creates a TierModel from the app's tier config.
func NewTierModel(tiers []core.Tier) *TierModel {
	m := &TierModel{Items: make([]*TierItem, len(tiers))}
	for i := range tiers {
		m.Items[i] = &TierItem{
			Name:     tiers[i].Name,
			Value:    fmt.Sprintf("%.2f", tiers[i].Value),
			Currency: tiers[i].Currency,
			Style:    tiers[i].StyleName,
			Tier:     &tiers[i],
		}
	}
	return m
}

func (m *TierModel) RowCount() int { return len(m.Items) }

func (m *TierModel) Value(row, col int) interface{} {
	if row < 0 || row >= len(m.Items) {
		return ""
	}
	item := m.Items[row]
	switch col {
	case 0:
		return item.Name
	case 1:
		return item.Value
	case 2:
		return item.Currency
	case 3:
		return item.Style
	}
	return ""
}
