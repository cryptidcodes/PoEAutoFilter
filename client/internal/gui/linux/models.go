//go:build linux

package linux

import (
	"fmt"
	"strings"

	"github.com/cryptidcodes/PoEAutoFilter/client/internal/core"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// --- GTK ListStore Models ---

type ColumnType int

const (
	StyleNameColumn ColumnType = iota
	StylePreviewColumn
)

const (
	TierNameColumn ColumnType = iota
	TierValueColumn
	TierCurrencyColumn
	TierStyleColumn
)

// StyleListStore manages the underlying GTK ListStore for the Style Library tab.
type StyleListStore struct {
	Store *gtk.ListStore
}

func NewStyleListStore() *StyleListStore {
	store := gtk.NewListStore([]glib.Type{glib.TypeString, glib.TypeString})
	return &StyleListStore{Store: store}
}

func (s *StyleListStore) Load(styles []core.Style) {
	s.Store.Clear()
	for _, style := range styles {
		preview := strings.ReplaceAll(style.ToFilterLines(), "\n", "; ")
		if len(preview) > 50 {
			preview = preview[:47] + "..."
		}
		s.Store.Set(s.Store.Append(),
			[]int{int(StyleNameColumn), int(StylePreviewColumn)},
			[]glib.Value{*glib.NewValue(style.Name), *glib.NewValue(preview)},
		)
	}
}

// TierListStore manages the underlying GTK ListStore for the Value Tiers tab.
type TierListStore struct {
	Store *gtk.ListStore
}

func NewTierListStore() *TierListStore {
	store := gtk.NewListStore([]glib.Type{glib.TypeString, glib.TypeString, glib.TypeString, glib.TypeString})
	return &TierListStore{Store: store}
}

func (t *TierListStore) Load(tiers []core.Tier) {
	t.Store.Clear()
	for _, tier := range tiers {
		t.Store.Set(t.Store.Append(),
			[]int{int(TierNameColumn), int(TierValueColumn), int(TierCurrencyColumn), int(TierStyleColumn)},
			[]glib.Value{
				*glib.NewValue(tier.Name),
				*glib.NewValue(fmt.Sprintf("%.2f", tier.Value)),
				*glib.NewValue(tier.Currency),
				*glib.NewValue(tier.StyleName),
			},
		)
	}
}

// createTextColumn is a helper to build generic TreeView text columns
func createTextColumn(title string, colId int) *gtk.TreeViewColumn {
	cellRenderer := gtk.NewCellRendererText()
	column := gtk.NewTreeViewColumn()
	column.SetTitle(title)
	column.PackEnd(cellRenderer, false)
	column.AddAttribute(cellRenderer, "text", colId)
	column.SetResizable(true)
	return column
}
