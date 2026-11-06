# AGENTS.md

This project is **The World Clock on the Desktop**, a desktop application built with the [Viva](https://github.com/vivaui/viva) framework.

All AI coding assistants (including Gemini, Claude, Codex, GPT) working on this codebase MUST strictly follow the Viva framework rules outlined below.

---

## Building with Viva

Viva is a Go-first, GPU-native desktop framework. It draws its own UI: there is no WebView, HTML, CSS or JavaScript, and nothing is redrawn while the app is idle. Most mistakes come from writing it like a web or retained-mode toolkit. Read this before writing code, and read the guides it points to before guessing an API.

### 1. Decide whose code it is

Before building, sort each need into one of two kinds:
- **Application behavior**: your protocols, storage, workflows, product screens (e.g., world clock timezones, clock display, user preferences). It stays in your app.
- **A reusable desktop capability**: windows, input, text and IME, focus, layout, components, rendering, accessibility, lifecycle, headless control. If a second Viva app would copy the same helper or wiring, it belongs in Viva, not in your app. Stop and design the smallest self-contained API for it instead of working around the gap with platform calls, synthetic input or hand-written hit testing.

### 2. Rules that review rejects

- **No reflection, `map[string]any` or dynamic property bags** in layout, input, scene or render paths.
- **No goroutines, timers or polling loops that run while the app is idle.** Animation is a scheduled, temporary reason to draw, never a heartbeat. An idle window draws nothing. (For clock updates: schedule redraws precisely at second/minute boundaries or when active, never run a continuous busy-polling loop).
- **Declaring components and layout must not allocate per paint.** State that is expensive to compute (shaped text, line breaks) is cached by whoever owns it and invalidated explicitly.
- **Never number hit targets by hand.** Identities are names in a `ui.Tree`.
- **Do not add a dependency** to the code that ships in the app without measuring its stripped binary size, startup and idle memory.
- **Do not claim “zero allocation”, “fast” or “light”** without a reproducible measurement.

### 3. How a screen is written

The painting function is the UI. It runs again on every redraw, declares what is on screen with value types, and keeps no layout cache.
- **Declare identities once**: `tree = ui.Tree{}`, a `tree.Feature("name")` per area, `tree.ID(feature, "control")` per control. Keep them stable between frames; focus, hover and press are tracked by ID, and `Canvas.Focused`, `Canvas.FocusVisible` and `Canvas.Hovered` report them back.
- **Handle a click where the control is declared**: `f.Button(id, label).Clicked()` is true for the one paint that delivers the click, and only then may state change.
- **Lay out with `flex` inside `flex.View.Draw`**. Components are source-distributed, typed Go, one file per component; read the file rather than looking for a hidden style engine.
- **Prefer a few intention-revealing parameters**. When same-typed positional arguments pile up, use a small struct or chainable options.

### 4. Verify without opening a window

Do not open a real window unless the user explicitly agrees; it interrupts them and depends on which display it lands on. Use these in order of cost:
1. **`app.Driver` in a Go test**: runs your real painter, router and focus rules from explicit calls (`Click`, `Key`, `Type`, `Scroll`, `Advance`) and returns a `Snapshot` of labels, hit regions and focus. Test behavior a user depends on, not pixel offsets or exact primitive counts.
2. **Headless build and line protocol**: import `github.com/vivaui/viva/headless/auto` from a file guarded by `//go:build vivahead`, so release builds never contain it. Build with `-tags vivahead` and start the program with `VIVA_HEADLESS=1` and a throwaway `HOME`: `RunCanvas` reads commands per line from stdin (`click "Label"`, `type`, `key`, `waitfor`, `shot`, `texts`, `layout` …).
3. **`layout`**: lists text that is clipped, spills out of its control or lies over other text. **`tree`** lists every control with its role, identity, accessible name, value and state. **`expect`** asserts expectations cleanly without manual eye checks.
4. **Real windows**: only for platform-specific items that cannot be seen otherwise (IME, focus loss, window-manager close, physical keyboard, high-DPI). Report what was not tested.

Before finishing, run `gofmt`, `go vet ./...` and `go test ./... -count=1`. For painting or input paths, report allocations per paint and steady-state time with `go test -bench`.

### 5. Documentation and Reference

- Guides: <https://vivaui.dev/en/docs/> (programming model, declarative layout, handling input, text input, overlays, theming, performance, testing, platforms).
- Skill definition: `.agents/skills/building-with-viva/SKILL.md` and `.claude/skills/building-with-viva/SKILL.md`
