package ui

import (
	"fmt"
	"time"

	"worldclock/pkg/clock"
	"worldclock/pkg/i18n"
	"worldclock/pkg/sleep"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// DrawSleepMode 渲染低亮度护眼睡眠时钟、环境音选择与音量渐弱倒计时界面。
// DrawSleepMode renders the minimal low-light sleep and ambient relaxation screen.
func DrawSleepMode(f *flex.Frame, state *AppState, theme ui.Theme) {
	now := state.Now
	sl := state.Sleep
	status, sound, vol := sl.Info()

	cfg := state.Store.Get()
	timeStr, period := clock.FormatTime(now, cfg.TimeFormat24H, cfg.ShowSeconds)

	// Low-light background for sleep comfort
	sleepBg := ui.RGB24(0x0E1210)

	f.Column().Fill().Pad(28).Background(sleepBg).Items(flex.Center).Gap(24).Children(func() {
		// Header
		f.Row().FillWidth().Items(flex.Center).Children(func() {
			f.Text(state.T("sleep.title")).FontSize(16).Bold().Color(theme.Accent)
			f.Spacer()
			if status != sleep.StateIdle {
				if f.Button(IDBtnSleepStop, state.T("sleep.exit")).Small().Outline().Clicked() {
					sl.Stop()
					if state.Audio != nil {
						state.Audio.Stop()
					}
				}
			}
		})

		f.Spacer()

		// Central Ambient Night Clock
		f.Column().Items(flex.Center).Gap(8).Children(func() {
			f.Row().Items(flex.Baseline).Gap(10).Children(func() {
				// Gentle warm/dim olive text for dark environment
				ambientColor := ui.RGB24(0x8F9779)
				f.Text(timeStr).FontSize(72).Bold().Mono().Color(ambientColor)
				if period != "" {
					f.Text(period).FontSize(24).Bold().Color(ambientColor)
				}
			})
			f.Text(i18n.FormatDateLocalized(state.Lang(), int(now.Month()), now.Day(), now.Year(), int(now.Weekday()), true)).FontSize(16).Color(ui.RGB24(0x556058))
		})

		f.Spacer()

		// Session Status or Configuration Panel
		if status == sleep.StateIdle {
			// Configuration
			f.Column().FillWidth().Pad(22).Background(theme.Card).Radius(16).Border(1, theme.Border).Items(flex.Center).Gap(14).Children(func() {
				f.Text(state.T("sleep.prepare")).FontSize(16).Bold().Color(theme.Primary)

				// Duration selector
				f.Row().Items(flex.Center).Gap(8).Children(func() {
					f.Text(state.T("sleep.duration")).FontSize(13).Muted()
					btn15 := f.Button(IDBtnSleep15m, state.T("sleep.dur_15m"))
					if state.SleepDuration == 15*time.Minute {
						btn15.Outline()
					} else {
						btn15.Secondary()
					}
					if btn15.Clicked() {
						state.SleepDuration = 15 * time.Minute
					}

					btn30 := f.Button(IDBtnSleep30m, state.T("sleep.dur_30m"))
					if state.SleepDuration == 30*time.Minute {
						btn30.Outline()
					} else {
						btn30.Secondary()
					}
					if btn30.Clicked() {
						state.SleepDuration = 30 * time.Minute
					}

					btn60 := f.Button(IDBtnSleep60m, state.T("sleep.dur_60m"))
					if state.SleepDuration == 60*time.Minute {
						btn60.Outline()
					} else {
						btn60.Secondary()
					}
					if btn60.Clicked() {
						state.SleepDuration = 60 * time.Minute
					}
				})

				// Sound selector
				f.Row().Items(flex.Center).Gap(8).Children(func() {
					f.Text(state.T("sleep.sound")).FontSize(13).Muted()
					btnRain := f.Button(IDBtnSleepRain, state.T("sleep.rain"))
					if state.SleepSound == sleep.SoundRain {
						btnRain.Outline()
					} else {
						btnRain.Secondary()
					}
					if btnRain.Clicked() {
						state.SleepSound = sleep.SoundRain
						if state.Audio != nil {
							state.Audio.PreviewAmbient(string(state.SleepSound))
						}
					}

					btnStream := f.Button(IDBtnSleepStream, state.T("sleep.stream"))
					if state.SleepSound == sleep.SoundStream {
						btnStream.Outline()
					} else {
						btnStream.Secondary()
					}
					if btnStream.Clicked() {
						state.SleepSound = sleep.SoundStream
						if state.Audio != nil {
							state.Audio.PreviewAmbient(string(state.SleepSound))
						}
					}

					btnWaves := f.Button(IDBtnSleepWaves, state.T("sleep.waves"))
					if state.SleepSound == sleep.SoundWaves {
						btnWaves.Outline()
					} else {
						btnWaves.Secondary()
					}
					if btnWaves.Clicked() {
						state.SleepSound = sleep.SoundWaves
						if state.Audio != nil {
							state.Audio.PreviewAmbient(string(state.SleepSound))
						}
					}

					btnWhite := f.Button(IDBtnSleepWhite, state.T("sleep.white"))
					if state.SleepSound == sleep.SoundWhite {
						btnWhite.Outline()
					} else {
						btnWhite.Secondary()
					}
					if btnWhite.Clicked() {
						state.SleepSound = sleep.SoundWhite
						if state.Audio != nil {
							state.Audio.PreviewAmbient(string(state.SleepSound))
						}
					}
				})

				// Start & Preview buttons
				f.Row().Items(flex.Center).Gap(12).Children(func() {
					if f.Button(IDBtnSleepPreview, state.T("sleep.preview")).Outline().Clicked() {
						if state.Audio != nil {
							state.Audio.PreviewAmbient(string(state.SleepSound))
						}
					}
					if f.Button(IDBtnSleepStart, state.T("sleep.start")).Clicked() {
						sl.Start(now, state.SleepDuration, state.SleepSound)
						if state.Audio != nil {
							state.Audio.SyncState(false, "", true, string(state.SleepSound), 0.8)
						}
					}
				})
			})
		} else {
			// Active Sleep Session Display
			soundLabel := string(sound)
			switch sound {
			case sleep.SoundRain:
				soundLabel = state.T("sleep.rain")
			case sleep.SoundStream:
				soundLabel = state.T("sleep.stream")
			case sleep.SoundWaves:
				soundLabel = state.T("sleep.waves")
			case sleep.SoundWhite:
				soundLabel = state.T("sleep.white")
			}
			f.Column().FillWidth().Pad(22).Background(theme.Card).Radius(16).Border(1, theme.Border).Items(flex.Center).Gap(10).Children(func() {
				f.Text(fmt.Sprintf(state.T("sleep.playing"), soundLabel, sl.FormatRemaining(now))).FontSize(16).Bold().Color(theme.Primary)

				volPercent := int(vol * 100)
				fadeText := state.T("sleep.normal")
				if status == sleep.StateFading {
					fadeText = state.T("sleep.fading")
				} else if status == sleep.StateCompleted {
					fadeText = state.T("sleep.ended")
				}
				f.Text(fmt.Sprintf(state.T("sleep.vol_info"), volPercent, fadeText)).FontSize(13).Color(theme.Accent)
			})
		}
	})
}
