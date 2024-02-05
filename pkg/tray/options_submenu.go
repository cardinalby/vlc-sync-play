package tray

import (
	"context"

	"fyne.io/systray"
	"github.com/cardinalby/vlc-sync-play/pkg/util/rx"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

type subMenuOption[T constraints.Ordered] struct {
	menuItem *systray.MenuItem
	value    T
}

func AddOptionsSubMenu[T constraints.Ordered](
	ctx context.Context,
	parent *systray.MenuItem,
	setting rx.Value[T],
	defaultValues []T,
	valFormatter func(T) string,
) {
	addOptionCheckboxFn := GetAddMenuItemCheckboxFn(parent)
	currValue := setting.GetValue()
	values := slices.Clone(defaultValues)
	if !slices.Contains(values, currValue) {
		values = append(values, currValue)
		slices.Sort(values)
	}

	var menuItems []*systray.MenuItem
	for _, v := range values {
		v := v
		menuItem := addOptionCheckboxFn(valFormatter(v), "", currValue == v)
		menuItems = append(menuItems, menuItem)
		OnClicked(ctx, menuItem, func() {
			setting.SetValue(v)
			menuItem.Check()
			for _, otherMenuItem := range menuItems {
				if otherMenuItem != menuItem {
					otherMenuItem.Uncheck()
				}
			}
		})
	}
}
