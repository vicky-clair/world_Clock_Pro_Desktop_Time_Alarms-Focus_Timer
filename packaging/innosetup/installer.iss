; 小瑞综合时钟桌面 (Xiaoru Integrated Clock Desktop) - Inno Setup 完整向导式安装配置
; 提供完整的：选择目标路径、选择开始菜单文件夹、勾选创建桌面快捷方式与开机自启、安装完成提示立即启动等交互向导

[Setup]
AppId={{E4F18B32-5819-4C5A-9D45-B7D8C5708812}
AppName=小瑞综合时钟桌面
AppVersion=1.0.0
AppPublisher=Xiaoru Studio
DefaultDirName={autopf}\XiaoruIntegratedClockDesktop
DefaultGroupName=小瑞综合时钟桌面
OutputDir=..\..\dist
OutputBaseFilename=小瑞综合时钟桌面_Setup
SetupIconFile=..\..\assets\icon.ico
UninstallDisplayIcon={app}\小瑞综合时钟桌面.exe
Compression=lzma2/ultra64
SolidCompression=yes
ArchitecturesInstallIn64BitMode=x64compatible
PrivilegesRequired=lowest

; 启用现代化向导界面与完整交互步骤
WizardStyle=modern
DisableDirPage=no
DisableProgramGroupPage=no
AlwaysShowDirOnReadyPage=yes
AlwaysShowGroupOnReadyPage=yes
ShowComponentSizes=yes

[Languages]
Name: "chinesesimplified"; MessagesFile: "compiler:Languages\ChineseSimplified.isl"
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "japanese"; MessagesFile: "compiler:Languages\Japanese.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"
Name: "autostart"; Description: "开机自动启动 小瑞综合时钟桌面"; GroupDescription: "系统启动选项:"; Flags: unchecked

[Files]
Source: "..\..\dist\小瑞综合时钟桌面.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\小瑞综合时钟桌面"; Filename: "{app}\小瑞综合时钟桌面.exe"
Name: "{group}\{cm:UninstallProgram,小瑞综合时钟桌面}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\小瑞综合时钟桌面"; Filename: "{app}\小瑞综合时钟桌面.exe"; Tasks: desktopicon

[Registry]
Root: HKCU; Subkey: "Software\Microsoft\Windows\CurrentVersion\Run"; ValueType: string; ValueName: "com.xiaoru.clockdesktop"; ValueData: """{app}\小瑞综合时钟桌面.exe"""; Flags: uninsdeletevalue; Tasks: autostart

[Run]
; 安装完成后在完成页面显示“立即启动 小瑞综合时钟桌面”复选框（默认勾选），点击“完成”按钮即可启动应用
Filename: "{app}\小瑞综合时钟桌面.exe"; Description: "{cm:LaunchProgram,小瑞综合时钟桌面}"; Flags: nowait postinstall skipifsilent
