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
Compression=lzma2/ultra64
SolidCompression=yes
ArchitecturesInstallIn64BitMode=x64compatible
PrivilegesRequired=lowest
DisableProgramGroupPage=yes

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "japanese"; MessagesFile: "compiler:Languages\Japanese.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"

[Files]
Source: "..\..\dist\小瑞综合时钟桌面.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\小瑞综合时钟桌面"; Filename: "{app}\小瑞综合时钟桌面.exe"
Name: "{autodesktop}\小瑞综合时钟桌面"; Filename: "{app}\小瑞综合时钟桌面.exe"; Tasks: desktopicon

[Run]
Filename: "{app}\小瑞综合时钟桌面.exe"; Description: "{cm:LaunchProgram,小瑞综合时钟桌面}"; Flags: nowait postinstall skipifsilent
