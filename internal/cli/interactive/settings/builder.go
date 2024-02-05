package settings

import (
	"strconv"
	"time"

	"github.com/cardinalby/vlc-sync-play/internal/app/static_features"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"

	"github.com/cardinalby/vlc-sync-play/internal/app"
	"github.com/cardinalby/vlc-sync-play/pkg/util/arr"
	"github.com/rivo/tview"
)

func BuildRoot(settings *app.Settings) tview.Primitive {
	form := tview.NewForm()

	addVlcInstances(form, settings)
	addNoVideo(form, settings)
	addReSeekSrc(form, settings)
	addPollingInterval(form, settings)
	if static_features.ClickPause {
		addClickPause(form, settings)
	}

	form.SetBorder(true).
		SetTitle("Settings").
		SetTitleAlign(tview.AlignLeft).
		SetRect(0, 0, 50, 13)

	return form
}

func addVlcInstances(form *tview.Form, settings *app.Settings) {
	options, initIndex := prepareOptions([]int{2, 3, 4}, settings.InstancesNumber.GetValue())
	strOptions := arr.Map(options, func(option int) string { return strconv.Itoa(option) })

	label := "VLC instances"

	form.AddDropDown(
		label,
		strOptions,
		initIndex,
		func(_ string, optionIndex int) {
			settings.InstancesNumber.SetValue(options[optionIndex])
		})
}

func addPollingInterval(form *tview.Form, settings *app.Settings) {
	options, initIndex := prepareOptions([]time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		500 * time.Millisecond,
		time.Second,
	}, settings.PollingInterval.GetValue())
	strOptions := arr.Map(options, func(option time.Duration) string { return option.String() })

	label := "Polling interval"

	form.AddDropDown(
		label,
		strOptions,
		initIndex,
		func(_ string, optionIndex int) {
			settings.PollingInterval.SetValue(options[optionIndex])
		})
}

func addClickPause(form *tview.Form, settings *app.Settings) {
	label := "Click to pause/resume playback"

	form.AddCheckbox(label, settings.ClickPause.GetValue(), func(checked bool) {
		settings.ClickPause.SetValue(checked)
	})
}

func addReSeekSrc(form *tview.Form, settings *app.Settings) {
	label := "Re-seek source player for precise sync"

	form.AddCheckbox(label, settings.ReSeekSrc.GetValue(), func(checked bool) {
		settings.ReSeekSrc.SetValue(checked)
	})
}

func addNoVideo(form *tview.Form, settings *app.Settings) {
	label := "Start new players without video"

	form.AddCheckbox(label, settings.NoVideo.GetValue(), func(checked bool) {
		settings.NoVideo.SetValue(checked)
	})
}

func prepareOptions[T constraints.Ordered](defaultOptions []T, initValue T) (options []T, initIndex int) {
	if index := slices.Index(defaultOptions, initValue); index != -1 {
		initIndex = index
		options = defaultOptions
	} else {
		options = append(defaultOptions, initValue)
		slices.Sort(options)
		initIndex = slices.Index(options, initValue)
	}
	return options, initIndex
}
