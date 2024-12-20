package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/shirou/gopsutil/process"
	"github.com/xackery/shindow/config"
	"github.com/xackery/wlk/cpl"
	"github.com/xackery/wlk/walk"
	"github.com/xackery/wlk/win"
	"golang.org/x/sys/windows"
)

var (
	Version string

	user32                = windows.NewLazySystemDLL("user32.dll")
	enumWindowsProc       = user32.NewProc("EnumWindows")
	getWindowThreadProcId = user32.NewProc("GetWindowThreadProcessId")

	settingsWnd         *walk.MainWindow
	cfg                 *config.CastConfiguration
	lstDevicesModel     *ProcessModel
	lstDevices          *walk.ListBox
	btnToggleBorderless *walk.PushButton
	txtResolutionX      *walk.TextEdit
	txtResolutionY      *walk.TextEdit
	txtResolutionW      *walk.TextEdit
	txtResolutionH      *walk.TextEdit
)

func main() {
	err := run()
	if err != nil {
		walk.MsgBox(nil, "Error", fmt.Sprintf("Failed to run: "+err.Error()), walk.MsgBoxOK)
	}
}

func run() error {
	var err error

	exePath := os.Args[0]

	cfg, err = config.LoadCastConfig(filepath.Dir(exePath) + "/shindow.ini")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	lstDevicesModel = &ProcessModel{}

	processes, err := listProcesses()
	if err != nil {
		return fmt.Errorf("list processes: %w", err)
	}
	lstDevicesModel.Set(processes)

	cmw := cpl.MainWindow{
		Title:    "Shindow Borderless v" + Version,
		Name:     "shindow",
		AssignTo: &settingsWnd,
		Layout:   cpl.VBox{},
		Children: []cpl.Widget{
			cpl.GroupBox{
				Title:  "Processes",
				Layout: cpl.VBox{},
				Children: []cpl.Widget{
					cpl.ListBox{
						ToolTipText: "Select which copy of EverQuest to fullscreen",
						AssignTo:    &lstDevices,
						Model:       lstDevicesModel,
						OnCurrentIndexChanged: func() {
						},
					},
					cpl.Composite{
						Layout: cpl.VBox{},
						Children: []cpl.Widget{
							cpl.PushButton{
								Text: "Refresh",
								//MaxSize: cpl.Size{Width: 45},
								OnClicked: func() {
									processes, err := listProcesses()
									if err != nil {
										fmt.Printf("listProcesses: %v\n", err)
									}
									lstDevicesModel.Set(processes)
								},
							},
							cpl.GroupBox{
								Title:  "Resolution",
								Layout: cpl.VBox{},
								Children: []cpl.Widget{
									cpl.PushButton{
										Text: "Set dimensions to fullscreen",
										//MaxSize: cpl.Size{Width: 45},
										OnClicked: func() {
											hwnd, err := hwndByPID(lstDevicesModel.SelectedProcess())
											if err != nil {
												walk.MsgBox(nil, "Error", fmt.Sprintf("Failed to find hwnd: "+err.Error()), walk.MsgBoxOK)
												return
											}

											monitor := win.MonitorFromWindow(hwnd, win.MONITOR_DEFAULTTOPRIMARY)
											var monitorInfo win.MONITORINFO
											monitorInfo.CbSize = uint32(unsafe.Sizeof(monitorInfo))
											if !win.GetMonitorInfo(monitor, &monitorInfo) {
												walk.MsgBox(nil, "Error", fmt.Sprintf("GetMonitorInfo failed: %s", lastErrorMessage()), walk.MsgBoxOK)
												return
											}

											rect := monitorInfo.RcMonitor
											txtResolutionX.SetText(fmt.Sprintf("%d", rect.Left))
											txtResolutionY.SetText(fmt.Sprintf("%d", rect.Top))
											txtResolutionW.SetText(fmt.Sprintf("%d", rect.Right-rect.Left))
											txtResolutionH.SetText(fmt.Sprintf("%d", rect.Bottom-rect.Top))
										},
									},
									cpl.PushButton{
										Text: "Set dimension based on current window",
										//MaxSize: cpl.Size{Width: 45},
										OnClicked: func() {
											hwnd, err := hwndByPID(lstDevicesModel.SelectedProcess())
											if err != nil {
												walk.MsgBox(nil, "Error", fmt.Sprintf("Failed to find hwnd: "+err.Error()), walk.MsgBoxOK)
												return
											}

											var rect win.RECT
											if !win.GetWindowRect(hwnd, &rect) {
												walk.MsgBox(nil, "Error", fmt.Sprintf("GetWindowRect failed: %s", lastErrorMessage()), walk.MsgBoxOK)
												return
											}

											txtResolutionX.SetText(fmt.Sprintf("%d", rect.Left+16))
											txtResolutionY.SetText(fmt.Sprintf("%d", rect.Top+39))
											txtResolutionW.SetText(fmt.Sprintf("%d", rect.Right-rect.Left-16))
											txtResolutionH.SetText(fmt.Sprintf("%d", rect.Bottom-rect.Top-39))
										},
									},
									cpl.Composite{
										Layout:  cpl.HBox{},
										MaxSize: cpl.Size{Height: 45},
										Children: []cpl.Widget{
											cpl.Label{Text: "X:"},
											cpl.TextEdit{
												AssignTo: &txtResolutionX,
												Text:     fmt.Sprintf("%d", cfg.EQWindowX),
												RowSpan:  1,
												MaxSize:  cpl.Size{Width: 45},
											}, cpl.Label{Text: "Y:"},
											cpl.TextEdit{
												AssignTo: &txtResolutionY,
												Text:     fmt.Sprintf("%d", cfg.EQWindowY),
												RowSpan:  1,
												MaxSize:  cpl.Size{Width: 45},
											},
										},
									},
									cpl.Composite{
										Layout:  cpl.HBox{},
										MaxSize: cpl.Size{Height: 45},
										Children: []cpl.Widget{
											cpl.Label{Text: "W:"},
											cpl.TextEdit{
												AssignTo: &txtResolutionW,
												Text:     fmt.Sprintf("%d", cfg.EQWindowW),
												RowSpan:  1,
												MaxSize:  cpl.Size{Width: 45},
											},
											cpl.Label{Text: "H:"},
											cpl.TextEdit{
												AssignTo: &txtResolutionH,
												Text:     fmt.Sprintf("%d", cfg.EQWindowH),
												RowSpan:  1,
												MaxSize:  cpl.Size{Width: 45},
											},
										},
									},
								},
							},

							cpl.PushButton{
								AssignTo: &btnToggleBorderless,
								Text:     "Set Fullscreen Borderless",
								//MaxSize:  cpl.Size{Width: 45},
								ToolTipText: "Go into options, display, turn off Allow window resizing, Overlap windows taskbar, and go to video modes and set it to your resolution",
								OnClicked: func() {

									hwnd, err := hwndByPID(lstDevicesModel.SelectedProcess())
									if err != nil {
										walk.MsgBox(nil, "Error", fmt.Sprintf("Failed to find hwnd: "+err.Error()), walk.MsgBoxOK)
										return
									}

									err = ToggleBorderlessWindow(hwnd, true)
									if err != nil {
										walk.MsgBox(nil, "Error", fmt.Sprintf("Failed to set fullscreen borderless: "+err.Error()), walk.MsgBoxOK)
									}
								},
							},
							cpl.PushButton{
								AssignTo: &btnToggleBorderless,
								Text:     "Remove Fullscreen Borderless",
								//MaxSize:  cpl.Size{Width: 45},
								ToolTipText: "Go into options, display, turn off Allow window resizing, Overlap windows taskbar, and go to video modes and set it to your resolution",
								OnClicked: func() {

									hwnd, err := hwndByPID(lstDevicesModel.SelectedProcess())
									if err != nil {
										walk.MsgBox(nil, "Error", fmt.Sprintf("Failed to find hwnd: "+err.Error()), walk.MsgBoxOK)
										return
									}

									err = ToggleBorderlessWindow(hwnd, false)
									if err != nil {
										walk.MsgBox(nil, "Error", fmt.Sprintf("Failed to reset fullscreen borderless: "+err.Error()), walk.MsgBoxOK)
									}
								},
							},
						},
					},
				},
			},
			cpl.PushButton{
				Text:    "Save",
				MaxSize: cpl.Size{Width: 45},
				OnClicked: func() {
					err := updateSave()
					if err != nil {
						walk.MsgBox(nil, "Error", fmt.Sprintf("Failed to save: "+err.Error()), walk.MsgBoxOK)
					}
				},
			},
		},
	}

	err = cmw.Create()
	if err != nil {
		return fmt.Errorf("create main window: %w", err)
	}

	if !cfg.IsNew {
		settingsWnd.SetX(cfg.SettingsX)
		settingsWnd.SetY(cfg.SettingsY)
	}

	settingsWnd.SetWidth(cfg.SettingsW)
	settingsWnd.SetHeight(cfg.SettingsH)

	if lstDevicesModel.ItemCount() == 1 {
		lstDevices.SetCurrentIndex(0)
	}

	settingsWnd.Closing().Attach(func(isCancel *bool, reason byte) {
		err := updateSave()
		if err != nil {
			fmt.Printf("updateSave post attach: %v\n", err)
		}
	})

	settingsWnd.Run()

	return nil
}

// ProcessModel represents a model for processes
type ProcessModel struct {
	walk.ListModelBase
	entries []*ProcessEntry
}

// ProcessEntry is an entry for a process, PID and name
type ProcessEntry struct {
	Name string
	PID  int
}

// Set sets the items
func (m *ProcessModel) Set(items []*ProcessEntry) {
	m.entries = items
	m.PublishItemsReset()
}

// ItemCount returns the number of items
func (m *ProcessModel) ItemCount() int {
	return len(m.entries)
}

// Value returns the value of an item
func (m *ProcessModel) Value(index int) interface{} {
	if index < 0 || index >= len(m.entries) {
		return nil
	}
	return fmt.Sprintf("%s (%d)", m.entries[index].Name, m.entries[index].PID)
}

func (m *ProcessModel) PID(index int) int {
	if index < 0 || index >= len(m.entries) {
		return 0
	}
	return m.entries[index].PID
}

func (m *ProcessModel) SelectedProcess() int {
	if lstDevices.CurrentIndex() == -1 {
		return 0
	}
	return m.PID(lstDevices.CurrentIndex())
}

func listProcesses() ([]*ProcessEntry, error) {
	var processes []*ProcessEntry
	procs, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("processes: %w", err)
	}
	for _, proc := range procs {

		name, err := proc.Name()
		if err != nil {
			continue
		}
		if !strings.Contains(name, "eqgame.exe") {
			continue
		}
		pid := proc.Pid
		processes = append(processes, &ProcessEntry{
			Name: name,
			PID:  int(pid),
		})
	}
	return processes, nil
}

func ToggleBorderlessWindow(hwnd windows.HWND, isBorderless bool) error {
	if win.IsZoomed(hwnd) {
		return fmt.Errorf("window is maximized, please restore it first")
	}
	//if isBorderless {
	//if win.IsZoomed(hwnd) && !win.ShowWindow(hwnd, win.SW_RESTORE) {
	//	return fmt.Errorf("failed to restore maximized window: %w", syscall.GetLastError())
	//}
	//}
	// Remove the title bar and borders by updating the window style
	style := win.GetWindowLong(hwnd, win.GWL_STYLE)
	if isBorderless {
		style &^= win.WS_CAPTION     // Remove WS_CAPTION
		style &^= win.WS_THICKFRAME  // Remove WS_THICKFRAME (border)
		style &^= win.WS_MINIMIZEBOX // Remove WS_MINIMIZEBOX
		style &^= win.WS_MAXIMIZEBOX // Remove WS_MAXIMIZEBOX
		style &^= win.WS_SYSMENU     // Remove WS_SYSMENU
	} else {
		style |= win.WS_CAPTION          // Add WS_CAPTION
		style |= win.WS_THICKFRAME       // Add WS_THICKFRAME (border)
		style |= win.WS_MINIMIZEBOX      // Add WS_MINIMIZEBOX
		style |= win.WS_MAXIMIZEBOX      // Add WS_MAXIMIZEBOX
		style |= win.WS_OVERLAPPEDWINDOW // Add WS_OVERLAPPEDWINDOW
		style |= win.WS_SYSMENU          // Add WS_SYSMENU
	}

	ret := win.SetWindowLong(hwnd, win.GWL_STYLE, style)
	if ret == 0 {
		return fmt.Errorf("SetWindowLong failed: %w", syscall.GetLastError())
	}

	// Remove extended styles that might conflict
	exStyle := win.GetWindowLong(hwnd, win.GWL_EXSTYLE)
	if isBorderless {
		exStyle &^= win.WS_EX_DLGMODALFRAME
		exStyle &^= win.WS_EX_CLIENTEDGE
		exStyle &^= win.WS_EX_STATICEDGE
	} else {
		exStyle |= win.WS_EX_DLGMODALFRAME
		exStyle |= win.WS_EX_CLIENTEDGE
		exStyle |= win.WS_EX_STATICEDGE
	}
	ret = win.SetWindowLong(hwnd, win.GWL_EXSTYLE, exStyle)
	if ret == 0 {
		return fmt.Errorf("SetWindowLong failed: %w", syscall.GetLastError())
	}

	if !isBorderless {
		//maximize the window
		// if !win.ShowWindow(hwnd, win.SW_MAXIMIZE) {
		// 	return fmt.Errorf("failed to maximize window: %w", syscall.GetLastError())
		// }
		return nil
	}

	// Get the monitor dimensions
	monitor := win.MonitorFromWindow(hwnd, win.MONITOR_DEFAULTTOPRIMARY)
	var monitorInfo win.MONITORINFO
	monitorInfo.CbSize = uint32(unsafe.Sizeof(monitorInfo))
	if !win.GetMonitorInfo(monitor, &monitorInfo) {
		return fmt.Errorf("GetMonitorInfo failed: %w", syscall.GetLastError())
	}

	// if the montior's full screen resolution matches the target resolution, then top/left is 0,0
	// otherwise, top/left is based onthe current window's position

	rect := win.RECT{Left: 0, Top: 0, Right: 0, Bottom: 0}
	if !win.GetWindowRect(hwnd, &rect) {
		return fmt.Errorf("GetWindowRect failed: %w", syscall.GetLastError())
	}

	val, err := strconv.Atoi(txtResolutionX.Text())
	if err != nil {
		return fmt.Errorf("x: %w", err)
	}
	rect.Left = int32(val)

	val, err = strconv.Atoi(txtResolutionY.Text())
	if err != nil {
		return fmt.Errorf("y: %w", err)
	}
	rect.Top = int32(val)

	val, err = strconv.Atoi(txtResolutionW.Text())
	if err != nil {
		return fmt.Errorf("w: %w", err)
	}
	rect.Right = rect.Left + int32(val)

	val, err = strconv.Atoi(txtResolutionH.Text())
	if err != nil {
		return fmt.Errorf("h: %w", err)
	}
	rect.Bottom = rect.Top + int32(val)

	// Set the window to cover the entire monitor area

	if !win.SetWindowPos(hwnd, win.HWND_NOTOPMOST, rect.Left, rect.Top,
		rect.Right-rect.Left, rect.Bottom-rect.Top,
		win.SWP_FRAMECHANGED|win.SWP_NOOWNERZORDER|win.SWP_NOZORDER) {
		return fmt.Errorf("SetWindowPos failed: %w", syscall.GetLastError())
	}

	fmt.Printf("Setting resolution to %dx%d top left and %dx%d bottom right\n", rect.Left, rect.Top, rect.Right, rect.Bottom)

	// Redraw the window
	if !win.RedrawWindow(hwnd, nil, 0, win.RDW_INVALIDATE|win.RDW_UPDATENOW|win.RDW_FRAME) {
		return fmt.Errorf("RedrawWindow failed: %w", syscall.GetLastError())
	}
	/*
		if !win.SetWindowPos(hwnd, win.HWND_NOTOPMOST, rect.Left, rect.Top,
			rect.Right-rect.Left, rect.Bottom-rect.Top,
			win.SWP_FRAMECHANGED|win.SWP_NOOWNERZORDER|win.SWP_NOZORDER) {
			return fmt.Errorf("SetWindowPos failed: %w", syscall.GetLastError())
		} */

	return nil
}

func StringToUTF16Ptr(s string) *uint16 {
	ptr, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		panic(err)
	}
	return ptr
}

func updateSave() error {
	if cfg == nil {
		return fmt.Errorf("config not loaded")
	}
	cfg.SettingsH = settingsWnd.Height()
	cfg.SettingsW = settingsWnd.Width()
	cfg.SettingsX = settingsWnd.X()
	cfg.SettingsY = settingsWnd.Y()
	val, err := strconv.Atoi(txtResolutionX.Text())
	if err != nil {
		return fmt.Errorf("x: %w", err)
	}
	cfg.EQWindowX = val

	val, err = strconv.Atoi(txtResolutionY.Text())
	if err != nil {
		return fmt.Errorf("y: %w", err)
	}
	cfg.EQWindowY = val

	val, err = strconv.Atoi(txtResolutionW.Text())
	if err != nil {
		return fmt.Errorf("w: %w", err)
	}
	cfg.EQWindowW = val

	val, err = strconv.Atoi(txtResolutionH.Text())
	if err != nil {
		return fmt.Errorf("h: %w", err)
	}
	cfg.EQWindowH = val

	err = cfg.Save()
	if err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	return nil
}

func hwndByPID(pid int) (windows.HWND, error) {
	hwnd := getHWNDFromPID(uint32(pid))
	if hwnd == 0 {
		return 0, fmt.Errorf("failed to find hwnd for pid %d", pid)
	}
	return windows.HWND(hwnd), nil
}

func getHWNDFromPID(pid uint32) uintptr {
	var hwnd uintptr
	enumWindows(enumWindowsCallback(func(h uintptr) bool {
		if pidByHWND(h) == pid {
			hwnd = h
			return false // Stop enumeration
		}
		return true // Continue enumeration
	}))
	return hwnd
}

func pidByHWND(hwnd uintptr) uint32 {
	var pid uint32
	getWindowThreadProcId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	return pid
}

func enumWindows(enumFunc func(h uintptr) bool) {
	// Wrap Go function in a syscall callback
	callback := syscall.NewCallback(func(hwnd uintptr, lparam uintptr) uintptr {
		if enumFunc(hwnd) {
			return 1 // Continue enumeration
		}
		return 0 // Stop enumeration
	})

	// Call EnumWindows
	enumWindowsProc.Call(callback, 0)
}

type enumWindowsCallback func(h uintptr) bool

func lastErrorMessage() string {
	return errorMessage(win.GetLastError())
}

func errorMessage(errCode uint32) string {
	var buf [512]uint16
	n, err := syscall.FormatMessage(
		syscall.FORMAT_MESSAGE_FROM_SYSTEM|syscall.FORMAT_MESSAGE_IGNORE_INSERTS,
		0,
		errCode,
		0, // Default language
		buf[:],
		nil,
	)
	if err != nil {
		return fmt.Sprintf("Unknown error code %d", errCode)
	}
	return syscall.UTF16ToString(buf[:n])
}
