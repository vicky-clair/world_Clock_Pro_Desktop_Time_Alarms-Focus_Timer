package ui

import (
	"fmt"
	"path/filepath"

	"worldclock/pkg/clock"
	"worldclock/pkg/i18n"
	"worldclock/pkg/storage"
	"worldclock/pkg/wallpaper"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// DrawSettings 渲染设置与偏好中心（包含字体定制、壁纸与轮播图、语言与温度单位、时钟与关闭按钮行为、关于与配置文件五大子标签页）。
// DrawSettings renders configuration preferences with sub-tabs, full scrolling, and wallpaper controls.
func DrawSettings(f *flex.Frame, state *AppState, theme ui.Theme) {
	cfg := state.Store.Get()
	currentLang := state.Lang()
	wpCfg := state.Wallpaper.Config()

	if state.SettingsTab == "" {
		state.SettingsTab = "wallpaper"
	}

	f.Column().Fill().Pad(20).Gap(14).Children(func() {
		// Top Bar: Title + Sub-Tabs + On-screen Scroll Buttons
		f.Row().FillWidth().Items(flex.Center).Gap(10).Children(func() {
			f.Text(state.T("settings.title")).FontSize(20).Bold().Color(theme.Foreground)

			f.Spacer()

			// Sub-Category Tabs
			btnFont := f.Button(IDSettingsTabFont, state.T("settings.subtab_font")).Small()
			if state.SettingsTab == "font" {
				btnFont.Outline()
			} else {
				btnFont.Ghost()
			}
			if btnFont.Clicked() {
				state.SettingsTab = "font"
				state.ResetScroll()
			}

			btnWp := f.Button(IDSettingsTabWallpaper, state.T("settings.subtab_wallpaper")).Small()
			if state.SettingsTab == "wallpaper" {
				btnWp.Outline()
			} else {
				btnWp.Ghost()
			}
			if btnWp.Clicked() {
				state.SettingsTab = "wallpaper"
				state.ResetScroll()
			}
			btnLang := f.Button(IDSettingsTabLanguage, state.T("settings.subtab_language")).Small()
			if state.SettingsTab == "lang" {
				btnLang.Outline()
			} else {
				btnLang.Ghost()
			}
			if btnLang.Clicked() {
				state.SettingsTab = "lang"
				state.ResetScroll()
			}

			btnClk := f.Button(IDSettingsTabClock, state.T("settings.subtab_clock")).Small()
			if state.SettingsTab == "clock" {
				btnClk.Outline()
			} else {
				btnClk.Ghost()
			}
			if btnClk.Clicked() {
				state.SettingsTab = "clock"
				state.ResetScroll()
			}

			btnAbt := f.Button(IDSettingsTabAbout, state.T("settings.subtab_about")).Small()
			if state.SettingsTab == "about" {
				btnAbt.Outline()
			} else {
				btnAbt.Ghost()
			}
			if btnAbt.Clicked() {
				state.SettingsTab = "about"
				state.ResetScroll()
			}

			// Buttons removed per user request; scrollbar and wheel active
		})

		// Modern ScrollArea with automatic clamping and sleek scrollbar
		scrollBox := f.ScrollArea(IDScrollSettings).Fill()
		scrollBox.Children(func() {
			f.Column().FillWidth().Gap(16).Children(func() {
				switch state.SettingsTab {
				case "font":
					drawSettingsFont(f, state, theme, cfg)
				case "wallpaper":
					drawSettingsWallpaper(f, state, theme, cfg, wpCfg)
				case "lang":
					drawSettingsLanguage(f, state, theme, cfg, currentLang)
				case "clock":
					drawSettingsClock(f, state, theme, cfg)
				case "about":
					drawSettingsAbout(f, state, theme)
				default:
					drawSettingsFont(f, state, theme, cfg)
				}
			})
		})
		ss := state.View.ScrollState(IDScrollSettings)
		DrawScrollbar(f.Canvas(), scrollBox, ss, theme)
	})
}

func drawSettingsWallpaper(f *flex.Frame, state *AppState, theme ui.Theme, cfg storage.Settings, wpCfg wallpaper.Config) {
	// Fresh config
	wpCfg = state.Wallpaper.Config()

	// Card 0: Live Wallpaper & Transparency Preview Box
	f.Column().FillWidth().Pad(16).Background(theme.Card).Radius(12).Border(1, theme.Border).Items(flex.Center).Gap(10).Children(func() {
		f.Text(state.T("settings.wp_preview_title")).FontSize(14).Bold().Color(theme.Primary)

		f.Column().Width(460).Height(110).Radius(10).Border(1, theme.Border).Background(theme.Background).Items(flex.Center).Justify(flex.Center).Children(func() {
			f.Row().Items(flex.Center).PadX(24).PadY(12).Background(theme.Card).Radius(8).Border(1, theme.Border).Children(func() {
				f.Text(state.T("settings.wp_glass_demo")).Bold().Color(theme.Foreground)
			})
		})
	})

	// Card 1: Wallpaper Settings
	f.Column().FillWidth().Pad(16).Background(theme.Card).Radius(12).Border(1, theme.Border).Gap(12).Children(func() {
		f.Text(state.T("settings.wallpaper")).FontSize(16).Bold().Color(theme.Primary)
		f.Text(state.T("settings.wallpaper_desc")).FontSize(13).Muted()

		// Toggle Wallpaper ON / OFF
		f.Row().Items(flex.Center).Children(func() {
			f.Text(state.T("settings.wp_status")).Bold()
			f.Spacer()
			toggleText := state.T("common.off")
			btn := f.Button(IDWallpaperToggle, toggleText)
			if wpCfg.Enabled {
				btn.Label(state.T("common.on")).Outline()
			} else {
				btn.Secondary()
			}
			if btn.Clicked() {
				newEnabled := !wpCfg.Enabled
				state.Wallpaper.SetEnabled(newEnabled)
				state.Persist()
			}
		})

		if wpCfg.Enabled {
			f.Divider()

			// Browse Local File(s) / Photo Folder & Sample Controls
			f.Row().Items(flex.Center).Gap(8).Children(func() {
				if f.Button(IDBtnBrowseWallpaper, state.T("settings.browse_file")).Small().Clicked() {
					dialogTitle := state.T("settings.browse_file")
					go func() {
						paths, err := app.OpenFiles(app.OpenFileOptions{
							Title:         dialogTitle,
							AllowMultiple: true,
						})
						if err == nil && len(paths) > 0 {
							state.QueueWallpaperFiles(paths)
						}
					}()
				}

				if f.Button(IDBtnBrowseFolder, state.T("settings.browse_folder")).Small().Outline().Clicked() {
					dialogTitle := state.T("settings.browse_folder")
					go func() {
						dir, err := app.OpenDirectory(app.OpenDirectoryOptions{
							Title: dialogTitle,
						})
						if err == nil && dir != "" {
							state.QueueWallpaperFolder(dir)
						}
					}()
				}

				if f.Button(IDBtnSampleSlideshow, state.T("settings.sample_slideshow")).Secondary().Small().Clicked() {
					state.Wallpaper.LoadImages(wallpaper.SampleSlideshowURIs())
					state.Persist()
				}

				if f.Button(IDBtnSampleWallpaper, state.T("settings.sample_image")).Secondary().Small().Clicked() {
					if samplePath, err := wallpaper.EnsureSampleLocalFile(); err == nil {
						_ = state.Wallpaper.LoadImage(samplePath)
						state.Persist()
					}
				}

				if wpCfg.ImagePath != "" || len(wpCfg.ImagePaths) > 0 {
					if f.Button(IDBtnClearWallpaper, state.T("settings.clear_custom")).Secondary().Small().Clicked() {
						state.Wallpaper.ClearCustomImage()
						state.Persist()
					}
				}
			})

			// File path or slideshow playlist status preview
			statusPreview := state.T("settings.wp_using_preset")
			if len(wpCfg.ImagePaths) > 1 {
				statusPreview = fmt.Sprintf(
					state.T("settings.slideshow_status"),
					len(wpCfg.ImagePaths),
					wpCfg.CurrentIndex+1,
					len(wpCfg.ImagePaths),
					filepath.Base(wpCfg.ImagePath),
				)
			} else if wpCfg.ImagePath != "" {
				statusPreview = fmt.Sprintf(state.T("settings.wp_loaded_fmt"), filepath.Base(wpCfg.ImagePath))
			}
			f.Text(statusPreview).FontSize(12).Bold().Color(theme.Accent)
			f.Text(state.T("settings.drag_hint")).FontSize(12).Muted()

			f.Divider()

			// Wallpaper Carousel / Slideshow Controls
			f.Text(state.T("settings.slideshow_title")).FontSize(14).Bold().Color(theme.Primary)
			f.Text(state.T("settings.slideshow_desc")).FontSize(12).Muted()

			// Row 1: Auto-Play Toggle + Prev / Next Slide Buttons
			f.Row().Items(flex.Center).Gap(8).Children(func() {
				f.Text(state.T("settings.slideshow_toggle")).FontSize(13).Bold()
				f.Spacer()

				if f.Button(IDWallpaperSlidePrev, state.T("settings.slideshow_prev")).Small().Ghost().Clicked() {
					state.Wallpaper.PrevSlide(state.Now)
					state.Persist()
				}

				slideToggleLbl := state.T("settings.slideshow_off")
				btnSlideToggle := f.Button(IDWallpaperSlideshowToggle, slideToggleLbl).Small()
				if wpCfg.SlideshowEnabled {
					btnSlideToggle.Label(state.T("settings.slideshow_on")).Outline()
				} else {
					btnSlideToggle.Secondary()
				}
				if btnSlideToggle.Clicked() {
					state.Wallpaper.SetSlideshowEnabled(!wpCfg.SlideshowEnabled)
					state.Persist()
				}

				if f.Button(IDWallpaperSlideNext, state.T("settings.slideshow_next")).Small().Ghost().Clicked() {
					state.Wallpaper.NextSlide(state.Now)
					state.Persist()
				}
			})

			// Row 2: Rotation Interval (1m / 5m / 10m / 15m)
			f.Row().Items(flex.Center).Gap(8).Children(func() {
				f.Text(state.T("settings.slideshow_int")).FontSize(13).Bold()
				f.Spacer()

				btn1m := f.Button(IDWallpaperInt1m, state.T("settings.interval_1m")).Small()
				if wpCfg.IntervalMinutes == 1 {
					btn1m.Outline()
				} else {
					btn1m.Secondary()
				}
				if btn1m.Clicked() {
					state.Wallpaper.SetIntervalMinutes(1)
					state.Persist()
				}

				btn5m := f.Button(IDWallpaperInt5m, state.T("settings.interval_5m")).Small()
				if wpCfg.IntervalMinutes == 5 || wpCfg.IntervalMinutes <= 0 {
					btn5m.Outline()
				} else {
					btn5m.Secondary()
				}
				if btn5m.Clicked() {
					state.Wallpaper.SetIntervalMinutes(5)
					state.Persist()
				}

				btn10m := f.Button(IDWallpaperInt10m, state.T("settings.interval_10m")).Small()
				if wpCfg.IntervalMinutes == 10 {
					btn10m.Outline()
				} else {
					btn10m.Secondary()
				}
				if btn10m.Clicked() {
					state.Wallpaper.SetIntervalMinutes(10)
					state.Persist()
				}

				btn15m := f.Button(IDWallpaperInt15m, state.T("settings.interval_15m")).Small()
				if wpCfg.IntervalMinutes == 15 {
					btn15m.Outline()
				} else {
					btn15m.Secondary()
				}
				if btn15m.Clicked() {
					state.Wallpaper.SetIntervalMinutes(15)
					state.Persist()
				}
			})

			// Row 3: Playback Order (Shuffle / Sequential)
			f.Row().Items(flex.Center).Gap(8).Children(func() {
				f.Text(state.T("settings.slideshow_order")).FontSize(13).Bold()
				f.Spacer()

				btnShuffle := f.Button(IDWallpaperOrderShuffle, state.T("settings.order_shuffle")).Small()
				if wpCfg.SlideshowOrder == "" || wpCfg.SlideshowOrder == wallpaper.OrderShuffle {
					btnShuffle.Outline()
				} else {
					btnShuffle.Secondary()
				}
				if btnShuffle.Clicked() {
					state.Wallpaper.SetSlideshowOrder(wallpaper.OrderShuffle)
					state.Persist()
				}

				btnSeq := f.Button(IDWallpaperOrderSeq, state.T("settings.order_seq")).Small()
				if wpCfg.SlideshowOrder == wallpaper.OrderSequential {
					btnSeq.Outline()
				} else {
					btnSeq.Secondary()
				}
				if btnSeq.Clicked() {
					state.Wallpaper.SetSlideshowOrder(wallpaper.OrderSequential)
					state.Persist()
				}
			})

			f.Divider()

			// Preset Wallpapers
			f.Text(state.T("settings.wp_presets_label")).FontSize(13).Bold()
			f.Row().Gap(8).Children(func() {
				btnGrid := f.Button(IDPresetGrid, state.T("settings.preset_grid")).Small()
				if wpCfg.ImagePath == "" && wpCfg.Preset == wallpaper.PresetGrid {
					btnGrid.Outline()
				} else {
					btnGrid.Secondary()
				}
				if btnGrid.Clicked() {
					state.Wallpaper.SetPreset(wallpaper.PresetGrid)
					state.Persist()
				}

				btnAurora := f.Button(IDPresetAurora, state.T("settings.preset_aurora")).Small()
				if wpCfg.ImagePath == "" && wpCfg.Preset == wallpaper.PresetAurora {
					btnAurora.Outline()
				} else {
					btnAurora.Secondary()
				}
				if btnAurora.Clicked() {
					state.Wallpaper.SetPreset(wallpaper.PresetAurora)
					state.Persist()
				}

				btnForest := f.Button(IDPresetForest, state.T("settings.preset_forest")).Small()
				if wpCfg.ImagePath == "" && wpCfg.Preset == wallpaper.PresetForest {
					btnForest.Outline()
				} else {
					btnForest.Secondary()
				}
				if btnForest.Clicked() {
					state.Wallpaper.SetPreset(wallpaper.PresetForest)
					state.Persist()
				}

				btnSunset := f.Button(IDPresetSunset, state.T("settings.preset_sunset")).Small()
				if wpCfg.ImagePath == "" && wpCfg.Preset == wallpaper.PresetSunset {
					btnSunset.Outline()
				} else {
					btnSunset.Secondary()
				}
				if btnSunset.Clicked() {
					state.Wallpaper.SetPreset(wallpaper.PresetSunset)
					state.Persist()
				}
			})

			f.Divider()

			// Fit Mode Selector
			f.Text(state.T("settings.fit_mode")).FontSize(13).Bold()
			f.Row().Gap(8).Children(func() {
				btnCover := f.Button(IDWallpaperFitFill, state.T("settings.fit_fill")).Small()
				if wpCfg.FitMode == wallpaper.FitCover {
					btnCover.Outline()
				} else {
					btnCover.Secondary()
				}
				if btnCover.Clicked() {
					state.Wallpaper.SetFitMode(wallpaper.FitCover)
					state.Persist()
				}

				btnContain := f.Button(IDWallpaperFitFit, state.T("settings.fit_fit")).Small()
				if wpCfg.FitMode == wallpaper.FitContain {
					btnContain.Outline()
				} else {
					btnContain.Secondary()
				}
				if btnContain.Clicked() {
					state.Wallpaper.SetFitMode(wallpaper.FitContain)
					state.Persist()
				}

				btnStretch := f.Button(IDWallpaperFitStretch, state.T("settings.fit_stretch")).Small()
				if wpCfg.FitMode == wallpaper.FitStretch {
					btnStretch.Outline()
				} else {
					btnStretch.Secondary()
				}
				if btnStretch.Clicked() {
					state.Wallpaper.SetFitMode(wallpaper.FitStretch)
					state.Persist()
				}

				btnTile := f.Button(IDWallpaperFitTile, state.T("settings.fit_tile")).Small()
				if wpCfg.FitMode == wallpaper.FitTile {
					btnTile.Outline()
				} else {
					btnTile.Secondary()
				}
				if btnTile.Clicked() {
					state.Wallpaper.SetFitMode(wallpaper.FitTile)
					state.Persist()
				}

				btnCenter := f.Button(IDWallpaperFitCenter, state.T("settings.fit_center")).Small()
				if wpCfg.FitMode == wallpaper.FitCenter {
					btnCenter.Outline()
				} else {
					btnCenter.Secondary()
				}
				if btnCenter.Clicked() {
					state.Wallpaper.SetFitMode(wallpaper.FitCenter)
					state.Persist()
				}
			})

			f.Divider()

			// Opacity Controls
			f.Row().Items(flex.Center).Children(func() {
				f.Text(state.T("settings.opacity")).Bold()
				f.Spacer()
				f.Row().Items(flex.Center).Gap(8).Children(func() {
					if f.Button(IDWallpaperOpacityDown, "- 10%").Small().Ghost().Clicked() {
						newOp := wpCfg.Opacity - 0.10
						if newOp < 0.10 {
							newOp = 0.10
						}
						state.Wallpaper.SetOpacity(newOp)
						state.Persist()
					}

					f.Text(fmt.Sprintf("%.0f%%", wpCfg.Opacity*100)).Bold().Mono().Width(50)

					if f.Button(IDWallpaperOpacityUp, "+ 10%").Small().Ghost().Clicked() {
						newOp := wpCfg.Opacity + 0.10
						if newOp > 1.0 {
							newOp = 1.0
						}
						state.Wallpaper.SetOpacity(newOp)
						state.Persist()
					}
				})
			})

			// Card Glass Translucency Controls
			f.Row().Items(flex.Center).Children(func() {
				f.Text(state.T("settings.card_glass")).Bold()
				f.Spacer()
				f.Row().Items(flex.Center).Gap(8).Children(func() {
					if f.Button(IDCardGlassDown, "- 10%").Small().Ghost().Clicked() {
						newOp := wpCfg.CardOpacity - 0.10
						if newOp < 0.20 {
							newOp = 0.20
						}
						state.Wallpaper.SetCardOpacity(newOp)
						state.Persist()
					}

					f.Text(fmt.Sprintf("%.0f%%", wpCfg.CardOpacity*100)).Bold().Mono().Width(50)

					if f.Button(IDCardGlassUp, "+ 10%").Small().Ghost().Clicked() {
						newOp := wpCfg.CardOpacity + 0.10
						if newOp > 1.0 {
							newOp = 1.0
						}
						state.Wallpaper.SetCardOpacity(newOp)
						state.Persist()
					}
				})
			})
		}
	})

	// Card 2: Theme Selector
	f.Column().FillWidth().Pad(16).Background(theme.Card).Radius(8).Border(1, theme.Border).Gap(12).Children(func() {
		f.Text(state.T("settings.theme")).FontSize(16).Bold().Color(theme.Primary)
		f.Text(state.T("settings.theme_desc")).FontSize(13).Muted()

		f.Row().Gap(8).Children(func() {
			btnOlive := f.Button(IDThemeOlive, state.T("settings.theme_olive"))
			if cfg.Theme == storage.ThemeOlive {
				btnOlive.Outline()
			} else {
				btnOlive.Secondary()
			}
			if btnOlive.Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					s.Theme = storage.ThemeOlive
				})
			}

			btnDark := f.Button(IDThemeDark, state.T("settings.theme_dark"))
			if cfg.Theme == storage.ThemeDark {
				btnDark.Outline()
			} else {
				btnDark.Secondary()
			}
			if btnDark.Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					s.Theme = storage.ThemeDark
				})
			}

			btnLight := f.Button(IDThemeLight, state.T("settings.theme_light"))
			if cfg.Theme == storage.ThemeLight {
				btnLight.Outline()
			} else {
				btnLight.Secondary()
			}
			if btnLight.Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					s.Theme = storage.ThemeLight
				})
			}

			btnNord := f.Button(IDThemeNord, state.T("settings.theme_nord"))
			if cfg.Theme == storage.ThemeNord {
				btnNord.Outline()
			} else {
				btnNord.Secondary()
			}
			if btnNord.Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					s.Theme = storage.ThemeNord
				})
			}
		})
	})
}

func drawSettingsLanguage(f *flex.Frame, state *AppState, theme ui.Theme, cfg storage.Settings, currentLang i18n.Language) {
	f.Column().FillWidth().Pad(16).Background(theme.Card).Radius(8).Border(1, theme.Border).Gap(12).Children(func() {
		f.Text(state.T("settings.lang")).FontSize(16).Bold().Color(theme.Primary)
		f.Text(state.T("settings.lang_desc")).FontSize(13).Muted()

		f.Row().Gap(8).Children(func() {
			btnZH := f.Button(IDLangZH, "🇨🇳 简体中文")
			if currentLang == i18n.LangZH {
				btnZH.Outline()
			} else {
				btnZH.Secondary()
			}
			if btnZH.Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					s.Language = string(i18n.LangZH)
				})
			}

			btnEN := f.Button(IDLangEN, "🇺🇸 English")
			if currentLang == i18n.LangEN {
				btnEN.Outline()
			} else {
				btnEN.Secondary()
			}
			if btnEN.Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					s.Language = string(i18n.LangEN)
				})
			}

			btnJA := f.Button(IDLangJA, "🇯🇵 日本語")
			if currentLang == i18n.LangJA {
				btnJA.Outline()
			} else {
				btnJA.Secondary()
			}
			if btnJA.Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					s.Language = string(i18n.LangJA)
				})
			}
		})

		f.Divider()

		// Temperature Unit
		f.Row().Items(flex.Center).Children(func() {
			f.Text(state.T("settings.temp_unit")).Bold()
			f.Spacer()
			unitLabel := state.T("settings.unit_celsius")
			if cfg.TempFahrenheit {
				unitLabel = state.T("settings.unit_fahrenheit")
			}
			btnUnit := f.Button(IDOptTempFahrenheit, unitLabel).Outline()
			if btnUnit.Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					s.TempFahrenheit = !s.TempFahrenheit
				})
			}
		})
	})
}
func drawSettingsFont(f *flex.Frame, state *AppState, theme ui.Theme, cfg storage.Settings) {
	// Card 1: 实时效果预览 (Live Preview)
	f.Column().FillWidth().Pad(20).Background(theme.Card).Radius(12).Border(1, theme.Border).Items(flex.Center).Gap(10).Children(func() {
		f.Text(state.T("settings.font_preview_title")).FontSize(14).Bold().Color(theme.Primary)
		f.Text("12:34:56").FontSize(state.EffectiveClockFontSize()).Bold().Mono().Color(state.ClockColor(theme))
		f.Text(i18n.FormatDateLocalized(state.Lang(), 11, 5, 2026, 4, true)).FontSize(14).Muted()
	})

	// Card 2: 时钟数字大小 (Clock Font Size)
	f.Column().FillWidth().Pad(16).Background(theme.Card).Radius(12).Border(1, theme.Border).Gap(12).Children(func() {
		f.Text(state.T("settings.font_size")).FontSize(16).Bold().Color(theme.Primary)

		// Preset buttons
		f.Row().Items(flex.Center).Gap(8).Children(func() {
			sizes := []struct {
				id  ui.ID
				val int
				lbl string
			}{
				{IDFontSize48, 48, state.T("settings.fs_48")},
				{IDFontSize64, 64, state.T("settings.fs_64")},
				{IDFontSize76, 76, state.T("settings.fs_76")},
				{IDFontSize88, 88, state.T("settings.fs_88")},
				{IDFontSize96, 96, state.T("settings.fs_96")},
			}
			curSz := cfg.ClockFontSize
			if curSz <= 0 {
				curSz = 64
			}
			for _, s := range sizes {
				item := s
				btn := f.Button(item.id, item.lbl).Small()
				if curSz == item.val {
					btn.Outline()
				} else {
					btn.Secondary()
				}
				if btn.Clicked() {
					state.UpdateSettings(func(c *storage.Settings) {
						c.ClockFontSize = item.val
					})
				}
			}
		})

		f.Divider()

		// Fine-tuning stepper
		f.Row().Items(flex.Center).Children(func() {
			f.Text(state.T("settings.fine_stepper")).FontSize(13).Bold()
			f.Spacer()
			f.Row().Items(flex.Center).Gap(8).Children(func() {
				if f.Button(IDFontSizeCustomDown, "- 2px").Small().Ghost().Clicked() {
					newSz := cfg.ClockFontSize - 2
					if newSz < 32 {
						newSz = 32
					}
					state.UpdateSettings(func(c *storage.Settings) {
						c.ClockFontSize = newSz
					})
				}

				f.Text(fmt.Sprintf("%d px", cfg.ClockFontSize)).Bold().Mono().Width(60)

				if f.Button(IDFontSizeCustomUp, "+ 2px").Small().Ghost().Clicked() {
					newSz := cfg.ClockFontSize + 2
					if newSz > 120 {
						newSz = 120
					}
					state.UpdateSettings(func(c *storage.Settings) {
						c.ClockFontSize = newSz
					})
				}
			})
		})
	})

	// Card 3: 时钟数字颜色 (Clock Font Color)
	f.Column().FillWidth().Pad(16).Background(theme.Card).Radius(12).Border(1, theme.Border).Gap(12).Children(func() {
		f.Text(state.T("settings.font_color")).FontSize(16).Bold().Color(theme.Primary)

		f.Row().Gap(8).Children(func() {
			colors := []struct {
				id  ui.ID
				val string
				lbl string
			}{
				{IDFontColorWhite, "white", state.T("settings.color_white")},
				{IDFontColorGold, "gold", state.T("settings.color_gold")},
				{IDFontColorOlive, "olive", state.T("settings.color_olive")},
				{IDFontColorCyan, "cyan", state.T("settings.color_cyan")},
				{IDFontColorPurple, "purple", state.T("settings.color_purple")},
				{IDFontColorCoral, "coral", state.T("settings.color_coral")},
			}
			curCol := cfg.ClockFontColor
			if curCol == "" {
				curCol = "white"
			}
			for _, c := range colors {
				item := c
				btn := f.Button(item.id, item.lbl).Small()
				if curCol == item.val {
					btn.Outline()
				} else {
					btn.Secondary()
				}
				if btn.Clicked() {
					state.UpdateSettings(func(c *storage.Settings) {
						c.ClockFontColor = item.val
					})
				}
			}
		})
	})

	// Card 4: 界面全域文字颜色 (UI Interface Font Color)
	f.Column().FillWidth().Pad(16).Background(theme.Card).Radius(12).Border(1, theme.Border).Gap(12).Children(func() {
		f.Text(state.T("settings.ui_font_color")).FontSize(16).Bold().Color(theme.Primary)
		f.Text(state.T("settings.ui_color_desc")).FontSize(13).Muted()

		f.Row().Gap(8).Children(func() {
			uiColors := []struct {
				id  ui.ID
				val string
				lbl string
			}{
				{IDUIColorDefault, "default", state.T("settings.color_theme")},
				{IDUIColorWhite, "white", state.T("settings.color_white")},
				{IDUIColorBlack, "black", state.T("settings.color_black")},
				{IDUIColorGold, "gold", state.T("settings.color_gold")},
				{IDUIColorCyan, "cyan", state.T("settings.color_cyan")},
				{IDUIColorPurple, "purple", state.T("settings.color_purple")},
				{IDUIColorOlive, "olive", state.T("settings.color_olive")},
			}
			curUICol := cfg.UIFontColor
			if curUICol == "" {
				curUICol = "default"
			}
			for _, uc := range uiColors {
				item := uc
				btn := f.Button(item.id, item.lbl).Small()
				if curUICol == item.val {
					btn.Outline()
				} else {
					btn.Secondary()
				}
				if btn.Clicked() {
					state.UpdateSettings(func(c *storage.Settings) {
						c.UIFontColor = item.val
					})
				}
			}
		})
	})

	// Card 4: 界面整体字号缩放 (Global UI Font Scale)
	f.Column().FillWidth().Pad(16).Background(theme.Card).Radius(12).Border(1, theme.Border).Gap(12).Children(func() {
		f.Text(state.T("settings.font_scale")).FontSize(16).Bold().Color(theme.Primary)

		f.Row().Gap(8).Children(func() {
			scales := []struct {
				id  ui.ID
				val float32
				lbl string
			}{
				{IDFontScale90, 0.90, state.T("settings.fscale_90")},
				{IDFontScale100, 1.00, state.T("settings.fscale_100")},
				{IDFontScale110, 1.10, state.T("settings.fscale_110")},
				{IDFontScale120, 1.20, state.T("settings.fscale_120")},
			}
			curScale := cfg.FontScale
			if curScale <= 0 {
				curScale = 1.00
			}
			for _, sc := range scales {
				item := sc
				btn := f.Button(item.id, item.lbl).Small()
				if curScale == item.val {
					btn.Outline()
				} else {
					btn.Secondary()
				}
				if btn.Clicked() {
					state.UpdateSettings(func(c *storage.Settings) {
						c.FontScale = item.val
					})
				}
			}
		})
	})
}

func drawSettingsClock(f *flex.Frame, state *AppState, theme ui.Theme, cfg storage.Settings) {
	// Display Options
	f.Column().FillWidth().Pad(16).Background(theme.Card).Radius(8).Border(1, theme.Border).Gap(12).Children(func() {
		f.Text(state.T("settings.display")).FontSize(16).Bold().Color(theme.Primary)
		f.Row().Items(flex.Center).Children(func() {
			f.Text(state.T("settings.format24h"))
			f.Spacer()
			btn := f.Button(IDOptFormat24H, state.T("clock.12h"))
			if cfg.TimeFormat24H {
				btn.Label(state.T("clock.24h")).Outline()
			} else {
				btn.Secondary()
			}
			if btn.Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					s.TimeFormat24H = !s.TimeFormat24H
				})
			}
		})

		f.Divider()

		f.Row().Items(flex.Center).Children(func() {
			f.Text(state.T("settings.seconds"))
			f.Spacer()
			btn := f.Button(IDOptShowSeconds, state.T("clock.seconds_off"))
			if cfg.ShowSeconds {
				btn.Label(state.T("clock.seconds_on")).Outline()
			} else {
				btn.Secondary()
			}
			if btn.Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					s.ShowSeconds = !s.ShowSeconds
				})
			}
		})

		f.Divider()

		f.Row().Items(flex.Center).Children(func() {
			f.Text(state.T("settings.fulldate"))
			f.Spacer()
			btn := f.Button(IDOptFullDate, state.T("settings.date_short"))
			if cfg.ShowFullDate {
				btn.Label(state.T("settings.date_full")).Outline()
			} else {
				btn.Secondary()
			}
			if btn.Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					s.ShowFullDate = !s.ShowFullDate
				})
			}
		})

		f.Divider()

		f.Row().Items(flex.Center).Children(func() {
			f.Text(state.T("settings.clockface"))
			f.Spacer()
			f.Row().Gap(8).Children(func() {
				btnDig := f.Button(IDOptStyleDigital, state.T("clock.style_digital"))
				if cfg.ClockStyle == string(clock.StyleDigital) {
					btnDig.Outline()
				} else {
					btnDig.Secondary()
				}
				if btnDig.Clicked() {
					state.UpdateSettings(func(s *storage.Settings) {
						s.ClockStyle = string(clock.StyleDigital)
					})
				}

				btnAna := f.Button(IDOptStyleAnalog, state.T("clock.style_analog"))
				if cfg.ClockStyle == string(clock.StyleAnalog) {
					btnAna.Outline()
				} else {
					btnAna.Secondary()
				}
				if btnAna.Clicked() {
					state.UpdateSettings(func(s *storage.Settings) {
						s.ClockStyle = string(clock.StyleAnalog)
					})
				}
			})
		})

		f.Divider()

		f.Row().Items(flex.Center).Children(func() {
			f.Text(state.T("settings.popout_pos"))
			f.Spacer()
			f.Row().Gap(8).Children(func() {
				btnTop := f.Button(IDOptPopoutPosTop, state.T("popout.pos_top")).Small()
				if cfg.PopoutPosition == storage.PopoutPosTop {
					btnTop.Outline()
				} else {
					btnTop.Secondary()
				}
				if btnTop.Clicked() {
					state.UpdateSettings(func(s *storage.Settings) {
						s.PopoutPosition = storage.PopoutPosTop
					})
				}

				btnCenter := f.Button(IDOptPopoutPosCenter, state.T("popout.pos_center")).Small()
				if cfg.PopoutPosition == storage.PopoutPosCenter || cfg.PopoutPosition == "" {
					btnCenter.Outline()
				} else {
					btnCenter.Secondary()
				}
				if btnCenter.Clicked() {
					state.UpdateSettings(func(s *storage.Settings) {
						s.PopoutPosition = storage.PopoutPosCenter
					})
				}

				btnBottom := f.Button(IDOptPopoutPosBottom, state.T("popout.pos_bottom")).Small()
				if cfg.PopoutPosition == storage.PopoutPosBottom {
					btnBottom.Outline()
				} else {
					btnBottom.Secondary()
				}
				if btnBottom.Clicked() {
					state.UpdateSettings(func(s *storage.Settings) {
						s.PopoutPosition = storage.PopoutPosBottom
					})
				}
			})
		})
	})

	// Profiles
	f.Column().FillWidth().Pad(16).Background(theme.Card).Radius(8).Border(1, theme.Border).Gap(12).Children(func() {
		f.Text(state.T("settings.profile")).FontSize(16).Bold().Color(theme.Primary)
		f.Text(state.T("settings.profile_desc")).FontSize(13).Muted()

		f.Row().Gap(8).Children(func() {
			btnDef := f.Button(IDProfileDefault, state.T("settings.profile_default"))
			if cfg.ActiveProfile == "Default" {
				btnDef.Outline()
			} else {
				btnDef.Secondary()
			}
			if btnDef.Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					s.ActiveProfile = "Default"
				})
			}

			btnWork := f.Button(IDProfileWork, state.T("settings.profile_work"))
			if cfg.ActiveProfile == "Work" {
				btnWork.Outline()
			} else {
				btnWork.Secondary()
			}
			if btnWork.Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					s.ActiveProfile = "Work"
				})
			}
		})
	})

	// Close Button Behavior (Minimize to Tray vs Exit Directly)
	f.Column().FillWidth().Pad(16).Background(theme.Card).Radius(8).Border(1, theme.Border).Gap(12).Children(func() {
		f.Text(state.T("settings.close_behavior")).FontSize(16).Bold().Color(theme.Primary)
		f.Text(state.T("settings.close_desc")).FontSize(13).Muted()

		closeBehavior := cfg.CloseBehavior
		if closeBehavior != storage.CloseBehaviorExit {
			closeBehavior = storage.CloseBehaviorMinimize
		}

		f.Row().Gap(8).Children(func() {
			btnMin := f.Button(IDCloseBehaviorMinimize, state.T("settings.close_minimize"))
			if closeBehavior == storage.CloseBehaviorMinimize {
				btnMin.Outline()
			} else {
				btnMin.Secondary()
			}
			if btnMin.Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					s.CloseBehavior = storage.CloseBehaviorMinimize
				})
			}

			btnExit := f.Button(IDCloseBehaviorExit, state.T("settings.close_exit"))
			if closeBehavior == storage.CloseBehaviorExit {
				btnExit.Outline()
			} else {
				btnExit.Secondary()
			}
			if btnExit.Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					s.CloseBehavior = storage.CloseBehaviorExit
				})
			}
		})
	})
}

func drawSettingsAbout(f *flex.Frame, state *AppState, theme ui.Theme) {
	f.Column().FillWidth().Pad(16).Background(theme.Card).Radius(8).Border(1, theme.Border).Gap(10).Children(func() {
		f.Text(state.T("settings.about")).FontSize(16).Bold().Color(theme.Foreground)
		f.Text(state.T("settings.about_arch")).FontSize(13).Muted()
		f.Text(state.T("settings.about_idle")).FontSize(13).Muted()
		f.Text(state.T("settings.about_wp")).FontSize(13).Color(theme.Primary)
		f.Text(state.T("settings.about_ver")).FontSize(12).Color(theme.Accent)
	})
}
