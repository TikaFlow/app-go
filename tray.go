package app

import (
	"fmt"

	"fyne.io/systray"
)

const (
	item_TYPE_SYSTRAY = iota
	item_TYPE_ITEM
	item_TYPE_CHECKBOX
	item_TYPE_FOLDER
)

type TrayItem struct {
	itemType int
	menuItem *systray.MenuItem
	callback func(*TrayItem)
}

func InitTray(title string) {
	systray.SetTitle(title)
	systray.SetTooltip(title)
}

func Tooltip(tip string) {
	systray.SetTooltip(tip)
}

func (this *TrayItem) AddSeparator() {
	if this.itemType == item_TYPE_SYSTRAY {
		systray.AddSeparator()
	} else {
		this.menuItem.AddSeparator()
		// if it has a subitem, set type to folder
		this.itemType = item_TYPE_FOLDER
	}
}

func (this *TrayItem) AddItem(icon []byte, title string) *TrayItem {
	// get item
	var item *systray.MenuItem
	if this.itemType == item_TYPE_SYSTRAY {
		item = systray.AddMenuItem(title, "")
	} else {
		item = this.menuItem.AddSubMenuItem(title, "")
		// if it has a subitem, set type to folder
		this.itemType = item_TYPE_FOLDER
	}

	// set icon
	if icon != nil {
		item.SetIcon(icon)
	}

	if this.callback != nil {
		panic("cannot set callback for a folder item")
	}

	return &TrayItem{
		itemType: item_TYPE_ITEM,
		menuItem: item,
	}
}

func (this *TrayItem) AddCheckbox(icon []byte, title string, checked bool) *TrayItem {
	// get item
	var item *systray.MenuItem
	if this.itemType == item_TYPE_SYSTRAY {
		item = systray.AddMenuItemCheckbox(title, "", checked)
	} else {
		item = this.menuItem.AddSubMenuItemCheckbox(title, "", checked)
		// if it has a subitem, set type to folder
		this.itemType = item_TYPE_FOLDER
	}

	// set icon
	if icon != nil {
		item.SetIcon(icon)
	}

	if this.callback != nil {
		panic("cannot set callback for a folder item")
	}

	return &TrayItem{
		itemType: item_TYPE_CHECKBOX,
		menuItem: item,
	}
}

func (this *TrayItem) OnClick(action func(item *TrayItem)) {
	if this.itemType != item_TYPE_ITEM {
		panic("can only set callback for a item")
	}

	this.callback = action
	go func() {
		for {
			_, ok := <-this.menuItem.ClickedCh
			if ok {
				if action != nil {
					action(this)
				}
			} else {
				// if channel is closed, exit
				this.callback = nil
				return
			}
		}
	}()
}

func (this *TrayItem) OffClick() {
	close(this.menuItem.ClickedCh)
}

func (this *TrayItem) ToggleChecked() bool {
	if this.itemType != item_TYPE_CHECKBOX {
		fmt.Println("can only toggle checked for a checkbox")
		return false
	}

	if this.menuItem.Checked() {
		this.menuItem.Uncheck()
	} else {
		this.menuItem.Check()
	}

	return this.menuItem.Checked()
}

func (this *TrayItem) ToggleEnabled() bool {
	if this.itemType == item_TYPE_SYSTRAY {
		fmt.Println("can not toggle enabled for systray")
		return false
	}

	if this.menuItem.Disabled() {
		this.menuItem.Enable()
	} else {
		this.menuItem.Disable()
	}

	return !this.menuItem.Disabled()
}
