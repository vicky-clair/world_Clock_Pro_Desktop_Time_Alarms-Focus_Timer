package ui

import (
	"fmt"

	"worldclock/pkg/clock"
	"worldclock/pkg/weather"
	"worldclock/pkg/world"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// DrawWorldClock 渲染多城市世界时钟列表、全球 193 国搜索添加面板以及重复添加提醒对话框。
// DrawWorldClock renders the multi-city world clock screen and city picker modal.
func DrawWorldClock(f *flex.Frame, state *AppState, theme ui.Theme) {
	cfg := state.Store.Get()
	localNow := state.Now.Local()
	sysLoc := clock.DetectSystemLocation(localNow)
	snapshots := state.World.Snapshots(state.UTC, localNow, cfg.TimeFormat24H, cfg.ShowSeconds, string(state.Lang()))

	f.Column().Fill().Pad(24).Gap(16).Children(func() {
		// Top Action Bar
		f.Row().FillWidth().Items(flex.Center).Gap(12).Children(func() {
			f.Text(state.T("world.title")).FontSize(22).Bold().Color(theme.Foreground)
			f.Text(sysLoc.FormatBadge(string(state.Lang()))).FontSize(12).Color(theme.Primary)
			f.Spacer()
			if f.Button(IDBtnAddCityModal, state.T("world.add_city")).Clicked() {
				state.ShowAddCityModal = true
			}
		})

		// City Cards
		if state.ShowAddCityModal {
			DrawAddCityModal(f, state, theme)
		} else {
			scrollBox := f.ScrollArea(IDScrollWorldList).Fill()
			scrollBox.Children(func() {
				f.Column().FillWidth().Gap(12).Children(func() {
					for _, snap := range snapshots {
						cityID := snap.City.ID
						f.Row().FillWidth().Pad(16).Background(theme.Card).Radius(12).Border(1, theme.Border).Items(flex.Center).Gap(16).Children(func() {
							// Left: City & Country
							f.Column().Width(180).Gap(4).Children(func() {
								f.Text(snap.DisplayName).FontSize(18).Bold().Color(theme.Foreground)
								f.Text(snap.DisplayCountry).FontSize(13).Color(theme.MutedForeground)
							})

							// Middle: Time & Period
							f.Row().Width(160).Items(flex.Baseline).Gap(6).Children(func() {
								f.Text(snap.TimeDisplay).FontSize(24).Bold().Mono().Color(theme.Foreground)
								if snap.Period != "" {
									f.Text(snap.Period).FontSize(14).Bold().Color(theme.Primary)
								}
							})

							// Right: Date & Badges
							f.Column().Width(200).Gap(4).Children(func() {
								f.Text(snap.DateDisplay).FontSize(13).Color(theme.MutedForeground)
								wSnap := state.Weather.GetWeather(cityID)
								tempStr := weather.FormatTemperature(wSnap.TempCelsius, cfg.TempFahrenheit)
								dayDiffStr := snap.DayDiff
								switch snap.DayDiff {
								case "Today":
									dayDiffStr = state.T("world.today")
								case "Tomorrow":
									dayDiffStr = state.T("world.tomorrow")
								case "Yesterday":
									dayDiffStr = state.T("world.yesterday")
								}
								f.Row().Gap(8).Children(func() {
									f.Text(snap.OffsetStr).FontSize(12).Bold().Color(theme.Accent)
									f.Text(fmt.Sprintf("[%s]", dayDiffStr)).FontSize(12).Color(theme.Primary)
									f.Text(fmt.Sprintf("%s %s", wSnap.Icon, tempStr)).FontSize(12).Bold()
								})
							})

							f.Spacer()

							// Move Up / Move Down Buttons
							if f.Button(CityMoveUpID(cityID), "↑").Small().Ghost().Clicked() {
								state.World.MoveUp(cityID)
								state.Persist()
							}
							if f.Button(CityMoveDownID(cityID), "↓").Small().Ghost().Clicked() {
								state.World.MoveDown(cityID)
								state.Persist()
							}

							// Favorite Star Button
							favLabel := "☆"
							btnFav := f.Button(CityToggleFavID(cityID), favLabel).Small()
							if snap.IsFavorite {
								btnFav.Label(state.T("world.fav")).Outline()
							} else {
								btnFav.Ghost()
							}
							if btnFav.Clicked() {
								state.World.ToggleFavorite(cityID)
								state.Persist()
							}

							// Remove Button
							if f.Button(CityRemoveID(cityID), state.T("world.remove")).DestructiveOutline().Small().Clicked() {
								state.World.RemoveCity(cityID)
								state.Persist()
							}
						})
					}
				})
			})
			DrawScrollbar(f.Canvas(), scrollBox, state.View.ScrollState(IDScrollWorldList), theme)
		}
	})
}

// DrawAddCityModal renders the city catalog picker dialog.
func DrawAddCityModal(f *flex.Frame, state *AppState, theme ui.Theme) {
	// Two-way sync between interactive Viva text field and CitySearchQuery
	fieldVal := state.CitySearchField.Text()
	if fieldVal != state.lastFieldText {
		state.lastFieldText = fieldVal
		state.CitySearchQuery = fieldVal
		state.ResetModalScroll()
	} else if state.CitySearchQuery != fieldVal {
		state.CitySearchField.SetText(state.CitySearchQuery)
		state.lastFieldText = state.CitySearchQuery
	}

	catalog := world.SearchCatalog(state.CitySearchQuery, state.SelectedContinent)

	f.Column().FillWidth().Pad(20).Background(theme.Card).Border(2, theme.Primary).Radius(10).Gap(10).Children(func() {
		// Modal Header
		f.Row().Items(flex.Center).Gap(10).Children(func() {
			f.Text(state.T("world.select_city")).FontSize(18).Bold().Color(theme.Primary)
			f.Spacer()
			if f.Button(IDBtnCloseAddCity, state.T("world.close")).Small().Outline().Clicked() {
				state.ShowAddCityModal = false
			}
		})

		// Interactive Search Input Box + Clear Button
		f.Row().FillWidth().Items(flex.Center).Gap(8).Children(func() {
			searchBox := f.Field(&state.CitySearchField, IDInputCitySearch, state.T("world.search_label"), state.T("world.search_placeholder")).Grow(1)
			if searchBox.Submitted() && len(catalog) > 0 {
				tryAddCityFromModal(state, catalog[0])
			}
			if state.CitySearchQuery != "" {
				if f.Button(IDBtnClearCitySearch, state.T("world.clear_search")).Small().Secondary().Clicked() {
					state.SetCitySearchQuery("")
				}
			}
		})

		// Continent Filter Tabs
		f.Row().Items(flex.Center).Gap(6).Children(func() {
			btnAll := f.Button(IDContAll, state.T("world.all_cont")).Small()
			if state.SelectedContinent == "" || state.SelectedContinent == "All" {
				btnAll.Outline()
			} else {
				btnAll.Ghost()
			}
			if btnAll.Clicked() {
				state.SelectedContinent = ""
				state.SetCitySearchQuery("")
			}

			btnAsia := f.Button(IDContAsia, state.T("world.asia")).Small()
			if state.SelectedContinent == "亚洲" || state.SelectedContinent == "Asia" {
				btnAsia.Outline()
			} else {
				btnAsia.Ghost()
			}
			if btnAsia.Clicked() {
				state.SelectedContinent = "亚洲"
				state.SetCitySearchQuery("")
			}

			btnEur := f.Button(IDContEurope, state.T("world.europe")).Small()
			if state.SelectedContinent == "欧洲" || state.SelectedContinent == "Europe" {
				btnEur.Outline()
			} else {
				btnEur.Ghost()
			}
			if btnEur.Clicked() {
				state.SelectedContinent = "欧洲"
				state.SetCitySearchQuery("")
			}

			btnAfr := f.Button(IDContAfrica, state.T("world.africa")).Small()
			if state.SelectedContinent == "非洲" || state.SelectedContinent == "Africa" {
				btnAfr.Outline()
			} else {
				btnAfr.Ghost()
			}
			if btnAfr.Clicked() {
				state.SelectedContinent = "非洲"
				state.SetCitySearchQuery("")
			}

			btnNA := f.Button(IDContNorthAmerica, state.T("world.north_america")).Small()
			if state.SelectedContinent == "北美洲" || state.SelectedContinent == "North America" {
				btnNA.Outline()
			} else {
				btnNA.Ghost()
			}
			if btnNA.Clicked() {
				state.SelectedContinent = "北美洲"
				state.SetCitySearchQuery("")
			}

			btnSA := f.Button(IDContSouthAmerica, state.T("world.south_america")).Small()
			if state.SelectedContinent == "南美洲" || state.SelectedContinent == "South America" {
				btnSA.Outline()
			} else {
				btnSA.Ghost()
			}
			if btnSA.Clicked() {
				state.SelectedContinent = "南美洲"
				state.SetCitySearchQuery("")
			}

			btnOC := f.Button(IDContOceania, state.T("world.oceania")).Small()
			if state.SelectedContinent == "大洋洲" || state.SelectedContinent == "Oceania" {
				btnOC.Outline()
			} else {
				btnOC.Ghost()
			}
			if btnOC.Clicked() {
				state.SelectedContinent = "大洋洲"
				state.SetCitySearchQuery("")
			}
		})
		// Quick country search shortcuts
		f.Row().Items(flex.Center).Gap(6).Children(func() {
			f.Text(state.T("world.quick_search")).FontSize(12).Muted()
			filters := []struct {
				id ui.ID
				q  string
				l  string
			}{
				{IDQuickFilterCN, "中国", state.T("world.qf_cn")},
				{IDQuickFilterUS, "美国", state.T("world.qf_us")},
				{IDQuickFilterJP, "日本", state.T("world.qf_jp")},
				{IDQuickFilterUK, "英国", state.T("world.qf_uk")},
				{IDQuickFilterFR, "法国", state.T("world.qf_fr")},
				{IDQuickFilterDE, "德国", state.T("world.qf_de")},
			}
			for _, item := range filters {
				it := item
				btn := f.Button(it.id, it.l).Small()
				if state.CitySearchQuery == it.q {
					btn.Outline()
				} else {
					btn.Ghost()
				}
				if btn.Clicked() {
					if state.CitySearchQuery == it.q {
						state.SetCitySearchQuery("")
					} else {
						state.SetCitySearchQuery(it.q)
					}
				}
			}
		})

		f.Text(fmt.Sprintf("%s (%d / %d)", state.T("world.available"), len(catalog), 193)).FontSize(12).Muted()

		// Scrollable List of 193 Countries with automatic clamping and sleek scrollbar
		scrollModal := f.ScrollArea(IDScrollCityModal).Height(210).FillWidth()
		scrollModal.Children(func() {
			f.Column().FillWidth().Gap(8).Children(func() {
				for _, c := range catalog {
					city := c
					f.Row().FillWidth().Pad(10).Background(theme.Muted).Radius(6).Items(flex.Center).Gap(12).Children(func() {
						f.Column().Width(240).Children(func() {
							countryLabel := city.LocalizedCountry(string(state.Lang()))
							f.Text(countryLabel).Bold().Color(theme.Foreground)

							capitalLabel := city.LocalizedName(string(state.Lang()))
							f.Text(capitalLabel).FontSize(12).Color(theme.MutedForeground)
						})

						f.Column().Width(200).Children(func() {
							f.Text(city.IanaTimeZoneID).FontSize(12).Color(theme.Accent)
							f.Text(city.LocalizedContinent(string(state.Lang()))).FontSize(11).Muted()
						})

						f.Spacer()

						alreadyAdded := state.World.HasCity(city.ID)
						btnLabel := state.T("world.btn_add")
						if alreadyAdded {
							btnLabel = state.T("world.btn_added")
						}
						btnAdd := f.Button(CityAddCatalogID(city.ID), btnLabel).Small()
						if alreadyAdded {
							btnAdd.Secondary()
						} else {
							btnAdd.Outline()
						}
						if btnAdd.Clicked() {
							tryAddCityFromModal(state, city)
						}
					})
				}
			})
		})
		DrawScrollbar(f.Canvas(), scrollModal, state.View.ScrollState(IDScrollCityModal), theme)
	})

	if state.DuplicateCityAlert != "" {
		d := f.Dialog(flex.DialogSpec{
			Title:          state.T("world.dup_title"),
			Description:    fmt.Sprintf(state.T("world.dup_desc"), state.DuplicateCityAlert),
			Width:          420,
			Scope:          IDDialogDupCityScope,
			DismissID:      IDDialogDupCityBackdrop,
			CloseID:        IDDialogDupCityClose,
			InitialFocusID: IDBtnDismissDuplicateCity,
		})
		d.Footer(func() {
			if f.Button(IDBtnDismissDuplicateCity, state.T("world.dup_ok")).Clicked() {
				state.DuplicateCityAlert = ""
			}
		})
		if d.Dismissed() {
			state.DuplicateCityAlert = ""
		}
	}
}

func tryAddCityFromModal(state *AppState, city world.City) {
	if state.World.AddCity(city.ID) {
		state.DuplicateCityAlert = ""
		state.Persist()
		state.ShowAddCityModal = false
		return
	}
	lang := string(state.Lang())
	state.DuplicateCityAlert = fmt.Sprintf("%s (%s)", city.LocalizedName(lang), city.LocalizedCountry(lang))
}
